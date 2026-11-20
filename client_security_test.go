package gocloak

// Unit tests for the security hardening of v15.0.1, running without a Keycloak instance.

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// TestUnit_Redirects checks that redirects are followed within the same origin but not to another host or port,
// where Go would re-send passwords in the form body or keep the Authorization header.
// Input: t [*testing.T]. Output: none.
func TestUnit_Redirects(t *testing.T) {
	t.Parallel()
	var leaked atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		leaked.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer other.Close()
	otherHost := strings.Replace(other.URL, "127.0.0.1", "localhost", 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/realms/r/protocol/openid-connect/token":
			http.Redirect(w, r, other.URL+r.URL.Path, http.StatusTemporaryRedirect) // same host, other port
		case "/admin/realms/r/users/u":
			http.Redirect(w, r, otherHost+r.URL.Path, http.StatusTemporaryRedirect) // other host
		case "/admin/realms/r/users/moved":
			http.Redirect(w, r, "/admin/realms/r/users/new", http.StatusMovedPermanently) // same origin
		case "/admin/realms/r/users/loop":
			http.Redirect(w, r, r.URL.Path, http.StatusFound)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"new"}`))
		}
	}))
	defer server.Close()
	client := NewClient(server.URL)
	ctx := context.Background()

	_, err := client.Login(ctx, "client", "secret", "r", "alice", "password")
	require.ErrorContains(t, err, "redirect to another origin")
	_, err = client.GetUserByID(ctx, "token", "r", "u")
	require.ErrorContains(t, err, "redirect to another origin")
	require.Zero(t, leaked.Load(), "credentials must not reach another origin")

	user, err := client.GetUserByID(ctx, "token", "r", "moved")
	require.NoError(t, err)
	require.Equal(t, "new", PString(user.ID))

	_, err = client.GetUserByID(ctx, "token", "r", "loop")
	require.ErrorContains(t, err, "stopped after 10 redirects")
}

// TestUnit_GetCerts_UnknownRealmsAreNotCached checks that realms whose certs were never fetched do not stay in the
// cache, while a realm with known certs keeps its entry when a later refresh fails.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetCerts_UnknownRealmsAreNotCached(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var down atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if down.Load() || r.URL.Path != "/realms/known/protocol/openid-connect/certs" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CertResponse{Keys: []CertResponseKey{jwkFor("k", &key.PublicKey)}})
	}))
	defer server.Close()
	client := NewClient(server.URL)
	ctx := context.Background()

	for i := range 50 {
		_, err := client.GetCerts(ctx, fmt.Sprintf("unknown-%d", i))
		require.Error(t, err)
	}
	_, err = client.GetCerts(ctx, "known")
	require.NoError(t, err)

	down.Store(true)
	ageCerts(t, client, "known", 3*client.Config.CertsInvalidateTime)
	_, err = client.GetCerts(ctx, "known")
	require.Error(t, err)

	var realms []string
	client.certsCache.Range(func(realm, _ any) bool {
		name, _ := realm.(string)
		realms = append(realms, name)
		return true
	})
	require.Equal(t, []string{"known"}, realms)
}

// TestUnit_TransportErrorsHideQuery checks that the query string (e.g. an email search) is removed from transport
// errors while errors.Is still reaches the cause.
// Input: t [*testing.T]. Output: none.
func TestUnit_TransportErrorsHideQuery(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewClient("http://127.0.0.1:1").GetUsers(ctx, "token", "r", GetUsersParams{Email: new("alice@example.com")})
	require.ErrorIs(t, err, context.Canceled)
	require.NotContains(t, err.Error(), "alice")
	require.Contains(t, err.Error(), "/admin/realms/r/users?REDACTED")
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.NotContains(t, apiErr.Err.Error(), "alice")

	plain := errors.New("no url")
	require.Same(t, plain, redactURLError(plain))
	noQuery := &url.Error{Op: "Get", URL: "http://kc/x", Err: plain}
	require.Same(t, error(noQuery), redactURLError(noQuery))
}

// TestUnit_ResponseBodyLimit checks the default response size limit and that it can be changed.
// Input: t [*testing.T]. Output: none.
func TestUnit_ResponseBodyLimit(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"` + strings.Repeat("x", 1024) + `"}`))
	}))
	defer server.Close()
	client := NewClient(server.URL)
	require.Equal(t, defaultResponseBodyLimit, client.RestyClient().ResponseBodyLimit)

	_, err := client.GetUserByID(context.Background(), "token", "r", "u")
	require.NoError(t, err)

	client.RestyClient().SetResponseBodyLimit(100)
	_, err = client.GetUserByID(context.Background(), "token", "r", "u")
	require.ErrorIs(t, err, resty.ErrResponseBodyTooLarge)
}

// TestUnit_CheckTokenIssuer checks the realm match of the iss claim, including escaped realm names and other realms
// whose name ends with the expected one.
// Input: t [*testing.T]. Output: none.
func TestUnit_CheckTokenIssuer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		iss  any
		ok   bool
		name string
	}{
		{"https://kc/realms/r", true, "public host"},
		{"http://internal:8080/auth/realms/r", true, "legacy prefix"},
		{"https://kc/realms/other", false, "other realm"},
		{"https://kc/realms/r/extra", false, "extra path"},
		{"https://kc/realms/xr", false, "realm name suffix"},
		{"", false, "empty"},
		{42, false, "not a string"},
	}
	for _, tc := range tests {
		err := checkTokenIssuer(jwt.MapClaims{"iss": tc.iss}, "r")
		if tc.ok {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
	require.NoError(t, checkTokenIssuer(jwt.MapClaims{"iss": "https://kc/realms/my%20realm"}, "my realm"))
	require.NoError(t, checkTokenIssuer(jwt.MapClaims{"iss": "https://kc/realms/my realm"}, "my realm"))
}

// TestUnit_EnforcedString_EmptyInput checks that empty input returns an error instead of panicking.
// Input: t [*testing.T]. Output: none.
func TestUnit_EnforcedString_EmptyInput(t *testing.T) {
	t.Parallel()
	var s EnforcedString
	require.Error(t, s.UnmarshalJSON(nil))
	require.Error(t, s.UnmarshalJSON([]byte{}))
}
