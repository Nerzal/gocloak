package benchmarks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v15"
)

// fakeKeycloak serves the payload set on gocloak's URL paths. It runs in-process, so its CPU cost is part of every
// number; it is identical for all clients, so compare absolute ns/op and allocs/op deltas, not percentages.
// Input: p [*payloadSet]. Output: [http.Handler].
func fakeKeycloak(p *payloadSet) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/protocol/openid-connect/token"):
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			body = p.token
		case strings.HasSuffix(path, "/protocol/openid-connect/certs"):
			body = p.certs
		case strings.HasSuffix(path, "/users"):
			body = p.users100
			if r.URL.Query().Get("max") == "1000" {
				body = p.users1000
			}
		case strings.Contains(path, "/users/"):
			body = p.user
		case strings.HasSuffix(path, "/groups"):
			body = p.groups
		case strings.HasSuffix(path, "/clients"):
			body = p.clients
		default:
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
}

// startFake starts the fake Keycloak. Input: b [*testing.B], mode [string] "http", "tls-h1" or "tls-h2".
// Output: [*target].
func startFake(b *testing.B, mode string) *target {
	b.Helper()
	srv := httptest.NewUnstartedServer(fakeKeycloak(payloads()))
	switch mode {
	case "http":
		srv.Start()
	case "tls-h1":
		srv.StartTLS()
	case "tls-h2":
		srv.EnableHTTP2 = true
		srv.StartTLS()
	}
	b.Cleanup(srv.Close)
	t := &target{
		base: srv.URL, realm: benchRealm, userID: "0f1e2d3c-4b5a-6978-8796-000000000000",
		clientID: "gocloak", clientSecret: "gocloak-secret",
		token:    func() string { return "offline-token" },
		cooldown: 2 * time.Second,
	}
	if srv.TLS != nil {
		tr, ok := srv.Client().Transport.(*http.Transport)
		if !ok {
			b.Fatal("test server client has no *http.Transport")
		}
		t.tls = tr.TLSClientConfig.Clone()
	}
	return t
}

// BenchmarkOfflineHTTP compares HTTP stacks per endpoint over plain HTTP/1.1 against an in-process fake Keycloak.
func BenchmarkOfflineHTTP(b *testing.B) {
	runEndpointBenchmarks(b, startFake(b, "http"), nil)
}

// BenchmarkOfflinePool measures many concurrent callers over TLS: connection reuse and handshake churn.
func BenchmarkOfflinePool(b *testing.B) {
	for _, mode := range []string{"tls-h1", "tls-h2"} {
		b.Run("proto="+mode, func(b *testing.B) {
			runPoolBenchmark(b, startFake(b, mode), "user", poolParallelism)
		})
	}
}

// BenchmarkOfflineDecodeAccessToken measures local token verification with warm certs (no network per call).
func BenchmarkOfflineDecodeAccessToken(b *testing.B) {
	t := startFake(b, "http")
	g := gocloak.NewClient(t.base)
	ctx := context.Background()
	token := payloads().accessToken
	if _, _, err := g.DecodeAccessToken(ctx, token, t.realm); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, _, err := g.DecodeAccessToken(ctx, token, t.realm); err != nil {
			b.Fatal(err)
		}
	}
}

// poolParallelism multiplies GOMAXPROCS to get the number of concurrent callers in the pool benchmarks.
const poolParallelism = 4
