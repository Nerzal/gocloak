package gocloak

// Unit tests for the connection pool default and SetDirectBodyRead, running without a Keycloak instance.

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

// TestUnit_ConnectionPool checks that NewClient keeps enough idle connections per host for concurrent callers.
// Input: t [*testing.T]. Output: none.
func TestUnit_ConnectionPool(t *testing.T) {
	t.Parallel()
	transport, err := NewClient("http://kc").RestyClient().Transport()
	require.NoError(t, err)
	require.Equal(t, defaultMaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
	require.Equal(t, defaultMaxIdleConnsPerHost, transport.MaxIdleConns)
}

// bodyReadServer answers GET .../users/{case} with the response named by case.
// Input: t [*testing.T]. Output: [*httptest.Server].
func bodyReadServer(t *testing.T) *httptest.Server {
	t.Helper()
	user := `{"id":"u1","username":"alice","attributes":{"a":["b"]}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch path.Base(r.URL.Path) {
		case "ok":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(user))
		case "error":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not_found","error_description":"User not found"}`))
		case "error-garbage":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`<html>`))
		case "invalid":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":`))
		case "no-content":
			w.WriteHeader(http.StatusNoContent)
		case "text":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("hello"))
		case "gzip":
			// sent without the client asking for it, so the transport does not decompress it
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			_, _ = gz.Write([]byte(user))
			_ = gz.Close()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Encoding", "gzip")
			_, _ = w.Write(buf.Bytes())
		case "big":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"` + strings.Repeat("x", 1024) + `"}`))
		case "big-chunked":
			// flushing before the end drops Content-Length, so only the read limit can stop it
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"`))
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			_, _ = w.Write([]byte(strings.Repeat("x", 1024) + `"}`))
		}
	}))
	t.Cleanup(server.Close)
	return server
}

// TestUnit_DirectBodyRead checks that SetDirectBodyRead returns the same results and errors as resty's body
// handling, including error bodies, 204, non-JSON content, gzip and the response body limit.
// Input: t [*testing.T]. Output: none.
func TestUnit_DirectBodyRead(t *testing.T) {
	t.Parallel()
	server := bodyReadServer(t)
	cases := []struct {
		name    string
		limit   int
		wantErr bool
	}{
		{name: "ok"},
		{name: "error", wantErr: true},
		{name: "error-garbage", wantErr: true},
		{name: "invalid", wantErr: true},
		{name: "no-content"},
		{name: "text"},
		{name: "gzip"},
		{name: "big", limit: 100, wantErr: true},
		{name: "big-chunked", limit: 100, wantErr: true},
		{name: "big", limit: 4096},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			viaResty := NewClient(server.URL)
			direct := NewClient(server.URL, SetDirectBodyRead())
			if tc.limit > 0 {
				viaResty.RestyClient().SetResponseBodyLimit(tc.limit)
				direct.RestyClient().SetResponseBodyLimit(tc.limit)
			}
			want, wantErr := viaResty.GetUserByID(context.Background(), "token", "r", tc.name)
			got, gotErr := direct.GetUserByID(context.Background(), "token", "r", tc.name)

			require.Equal(t, want, got)
			if !tc.wantErr {
				require.NoError(t, wantErr)
				require.NoError(t, gotErr)
				return
			}
			require.Error(t, wantErr)
			require.Error(t, gotErr)
			require.Equal(t, wantErr.Error(), gotErr.Error())
			var wantAPI, gotAPI *APIError
			require.ErrorAs(t, wantErr, &wantAPI)
			require.ErrorAs(t, gotErr, &gotAPI)
			require.Equal(t, wantAPI.Code, gotAPI.Code)
			require.Equal(t, wantAPI.Type, gotAPI.Type)
			if tc.limit > 0 {
				require.ErrorIs(t, gotErr, resty.ErrResponseBodyTooLarge)
			}
		})
	}
}

// TestUnit_ReadBodyPrealloc checks that readBody reads a body with a known length in a single allocation.
// Input: t [*testing.T]. Output: none.
func TestUnit_ReadBodyPrealloc(t *testing.T) {
	body := strings.Repeat("x", 512<<10)
	allocs := testing.AllocsPerRun(10, func() {
		resp := &http.Response{
			Body:          http.NoBody,
			ContentLength: int64(len(body)),
			Header:        http.Header{},
		}
		resp.Body = readCloser{strings.NewReader(body)}
		got, err := readBody(resp, defaultResponseBodyLimit)
		if err != nil || len(got) != len(body) {
			t.Fatalf("readBody: %v, %d bytes", err, len(got))
		}
	})
	// the buffer, plus small fixed allocations (reader wrappers); resty's appends need dozens for 512 KiB
	require.LessOrEqual(t, allocs, 5.0)
}

// TestUnit_ReadBody checks readBody's limit handling and error paths.
// Input: t [*testing.T]. Output: none.
func TestUnit_ReadBody(t *testing.T) {
	t.Parallel()
	response := func(body io.Reader, contentLength int64, encoding string) *http.Response {
		header := http.Header{}
		if encoding != "" {
			header.Set("Content-Encoding", encoding)
		}
		return &http.Response{Body: io.NopCloser(body), ContentLength: contentLength, Header: header}
	}

	got, err := readBody(response(strings.NewReader("abc"), -1, ""), 0)
	require.NoError(t, err, "no limit")
	require.Equal(t, "abc", string(got))

	got, err = readBody(response(strings.NewReader("abc"), 3, ""), 3)
	require.NoError(t, err, "exactly at the limit")
	require.Equal(t, "abc", string(got))

	_, err = readBody(response(strings.NewReader("abcd"), -1, ""), 3)
	require.ErrorIs(t, err, resty.ErrResponseBodyTooLarge, "unknown length over the limit")

	_, err = readBody(response(strings.NewReader("abc"), 1<<40, ""), 1<<20)
	require.ErrorIs(t, err, resty.ErrResponseBodyTooLarge, "announced length over the limit")

	_, err = readBody(response(strings.NewReader("not gzip"), 8, "gzip"), 0)
	require.Error(t, err, "invalid gzip")

	readErr := errors.New("connection reset")
	_, err = readBody(response(iotest.ErrReader(readErr), -1, ""), 100)
	require.ErrorIs(t, err, readErr, "read error")
}

// TestUnit_DirectBodyReadTransportError checks that a transport error with SetDirectBodyRead becomes an *APIError
// like without it.
// Input: t [*testing.T]. Output: none.
func TestUnit_DirectBodyReadTransportError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	_, err := NewClient(server.URL, SetDirectBodyRead()).GetUserByID(context.Background(), "token", "r", "u")
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, 0, apiErr.Code)
}

// readCloser adds a no-op Close to a reader.
type readCloser struct{ *strings.Reader }

// Close implements io.Closer. Output: [error] always nil.
func (readCloser) Close() error { return nil }
