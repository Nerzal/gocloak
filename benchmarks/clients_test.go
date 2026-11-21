package benchmarks

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v15"
	"github.com/go-resty/resty/v2"
	"github.com/valyala/fasthttp"
)

// target describes the Keycloak (real or fake) the HTTP benchmarks talk to.
type target struct {
	base         string
	realm        string
	tls          *tls.Config // nil for plain HTTP
	token        func() string
	userID       string
	clientID     string
	clientSecret string
	cooldown     time.Duration // pause before each pool benchmark client, so the previous client's sockets drain
}

// endpoint is one typical Keycloak call, expressed for every client.
type endpoint struct {
	name   string
	method string
	path   func(t *target) string // path plus query, relative to base
	form   func(t *target) url.Values
	bearer bool
	result func() any
	// viaGocloak performs the call through the gocloak API.
	viaGocloak func(ctx context.Context, g *gocloak.GoCloak, t *target) error
}

var endpoints = []endpoint{
	{
		name: "token", method: http.MethodPost,
		path: func(t *target) string { return "/realms/" + t.realm + "/protocol/openid-connect/token" },
		form: func(t *target) url.Values {
			return url.Values{"grant_type": {"client_credentials"}, "client_id": {t.clientID}, "client_secret": {t.clientSecret}}
		},
		result: func() any { return new(gocloak.JWT) },
		viaGocloak: func(ctx context.Context, g *gocloak.GoCloak, t *target) error {
			_, err := g.LoginClient(ctx, t.clientID, t.clientSecret, t.realm)
			return err
		},
	},
	{
		name: "user", method: http.MethodGet, bearer: true,
		path:   func(t *target) string { return "/admin/realms/" + t.realm + "/users/" + t.userID },
		result: func() any { return new(gocloak.User) },
		viaGocloak: func(ctx context.Context, g *gocloak.GoCloak, t *target) error {
			_, err := g.GetUserByID(ctx, t.token(), t.realm, t.userID)
			return err
		},
	},
	usersEndpoint(100),
	usersEndpoint(1000),
	{
		name: "groups", method: http.MethodGet, bearer: true,
		path:   func(t *target) string { return "/admin/realms/" + t.realm + "/groups" },
		result: func() any { return new([]*gocloak.Group) },
		viaGocloak: func(ctx context.Context, g *gocloak.GoCloak, t *target) error {
			_, err := g.GetGroups(ctx, t.token(), t.realm, gocloak.GetGroupsParams{})
			return err
		},
	},
	{
		name: "clients", method: http.MethodGet, bearer: true,
		path:   func(t *target) string { return "/admin/realms/" + t.realm + "/clients" },
		result: func() any { return new([]*gocloak.Client) },
		viaGocloak: func(ctx context.Context, g *gocloak.GoCloak, t *target) error {
			_, err := g.GetClients(ctx, t.token(), t.realm, gocloak.GetClientsParams{})
			return err
		},
	},
}

// usersEndpoint lists max users. Input: maxUsers [int]. Output: [endpoint].
func usersEndpoint(maxUsers int) endpoint {
	return endpoint{
		name: fmt.Sprintf("users%d", maxUsers), method: http.MethodGet, bearer: true,
		path: func(t *target) string {
			return fmt.Sprintf("/admin/realms/%s/users?first=0&max=%d", t.realm, maxUsers)
		},
		result: func() any { return new([]*gocloak.User) },
		viaGocloak: func(ctx context.Context, g *gocloak.GoCloak, t *target) error {
			_, err := g.GetUsers(ctx, t.token(), t.realm, gocloak.GetUsersParams{First: new(0), Max: new(maxUsers)})
			return err
		},
	}
}

// caller performs one endpoint call with a specific HTTP stack.
type caller func(ctx context.Context, ep *endpoint) error

// clientKind names an HTTP stack and how to build it.
type clientKind struct {
	name string
	// build returns a caller; dials counts new TCP connections.
	build func(t *target, dials *atomic.Int64) caller
}

// tunedPool is the idle pool size of the net/http and fasthttp variants: enough idle connections for one host.
const tunedPool = 256

// keepPool tells configureTransport to keep gocloak's idle pool setting.
const keepPool = 0

var clientKinds = []clientKind{
	{"gocloak", func(t *target, d *atomic.Int64) caller { return gocloakCaller(t, d, keepPool) }},
	// resty's default pool size, which gocloak used before v15.1.0
	{"gocloak-restypool", func(t *target, d *atomic.Int64) caller { return gocloakCaller(t, d, runtime.GOMAXPROCS(0)+1) }},
	{"resty", restyCaller},
	{"nethttp", netHTTPCaller},
	{"fasthttp", fastHTTPCaller},
}

// countingDial wraps dial so every new connection increments dials.
// Input: dial [func], dials [*atomic.Int64]. Output: [func] wrapped dial.
func countingDial(dial func(ctx context.Context, network, addr string) (net.Conn, error), dials *atomic.Int64) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dials.Add(1)
		return dial(ctx, network, addr)
	}
}

// gocloakCaller builds a gocloak client, optionally with another idle pool size.
// Input: t [*target], dials [*atomic.Int64], pool [int] idle connections per host, keepPool for gocloak's default.
// Output: [caller].
func gocloakCaller(t *target, dials *atomic.Int64, pool int) caller {
	g := gocloak.NewClient(t.base)
	configureTransport(g, t, dials, pool)
	return func(ctx context.Context, ep *endpoint) error { return ep.viaGocloak(ctx, g, t) }
}

// configureTransport counts dials and sets the TLS config of the target on the transport of g; a pool other than
// keepPool replaces gocloak's idle pool size.
// Input: g [*gocloak.GoCloak], t [*target], dials [*atomic.Int64], pool [int] idle connections per host.
func configureTransport(g *gocloak.GoCloak, t *target, dials *atomic.Int64, pool int) {
	tr, err := g.RestyClient().Transport()
	must(err)
	tr.DialContext = countingDial(tr.DialContext, dials)
	tr.TLSClientConfig = t.tls
	if pool != keepPool {
		tr.MaxIdleConnsPerHost = pool
		tr.MaxIdleConns = pool
	}
}

// restyCaller calls Keycloak with a bare resty client (default transport): gocloak overhead minus gocloak.
// Input: t [*target], dials [*atomic.Int64]. Output: [caller].
func restyCaller(t *target, dials *atomic.Int64) caller {
	c := resty.New()
	tr, err := c.Transport()
	must(err)
	tr.DialContext = countingDial(tr.DialContext, dials)
	tr.TLSClientConfig = t.tls
	return func(ctx context.Context, ep *endpoint) error {
		r := c.R().SetContext(ctx).SetResult(ep.result())
		if ep.bearer {
			r.SetAuthToken(t.token())
		}
		if ep.form != nil {
			r.SetFormDataFromValues(ep.form(t))
		}
		resp, err := r.Execute(ep.method, t.base+ep.path(t))
		if err != nil {
			return err
		}
		if resp.IsError() {
			return fmt.Errorf("resty: %s", resp.Status())
		}
		return nil
	}
}

// netHTTPCaller calls Keycloak with plain net/http, a tuned pool and a streaming JSON decoder.
// Input: t [*target], dials [*atomic.Int64]. Output: [caller].
func netHTTPCaller(t *target, dials *atomic.Int64) caller {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("http.DefaultTransport is not *http.Transport")
	}
	tr := defaultTransport.Clone()
	tr.DialContext = countingDial((&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext, dials)
	tr.TLSClientConfig = t.tls
	tr.MaxIdleConns = tunedPool
	tr.MaxIdleConnsPerHost = tunedPool
	c := &http.Client{Transport: tr, Timeout: time.Minute}
	return func(ctx context.Context, ep *endpoint) error {
		var body io.Reader
		if ep.form != nil {
			body = strings.NewReader(ep.form(t).Encode())
		}
		req, err := http.NewRequestWithContext(ctx, ep.method, t.base+ep.path(t), body)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		if ep.form != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if ep.bearer {
			req.Header.Set("Authorization", "Bearer "+t.token())
		}
		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= http.StatusBadRequest {
			_, _ = io.Copy(io.Discard, resp.Body)
			return fmt.Errorf("nethttp: %s", resp.Status)
		}
		return json.NewDecoder(resp.Body).Decode(ep.result())
	}
}

// fastHTTPCaller calls Keycloak with fasthttp (HTTP/1.1 only, own connection pool, pooled buffers).
// Input: t [*target], dials [*atomic.Int64]. Output: [caller].
func fastHTTPCaller(t *target, dials *atomic.Int64) caller {
	var tlsCfg *tls.Config
	if t.tls != nil {
		tlsCfg = t.tls.Clone()
		tlsCfg.NextProtos = []string{"http/1.1"} // fasthttp speaks HTTP/1.1 only, so it must not offer h2 via ALPN
	}
	c := &fasthttp.Client{
		TLSConfig:       tlsCfg,
		MaxConnsPerHost: tunedPool,
		ReadTimeout:     time.Minute,
		WriteTimeout:    time.Minute,
		Dial: func(addr string) (net.Conn, error) {
			dials.Add(1)
			return fasthttp.DialTimeout(addr, 30*time.Second)
		},
	}
	return func(_ context.Context, ep *endpoint) error {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)
		req.SetRequestURI(t.base + ep.path(t))
		req.Header.SetMethod(ep.method)
		req.Header.Set("Accept", "application/json")
		if ep.bearer {
			req.Header.Set("Authorization", "Bearer "+t.token())
		}
		if ep.form != nil {
			req.Header.SetContentType("application/x-www-form-urlencoded")
			req.SetBodyString(ep.form(t).Encode())
		}
		if err := c.Do(req, resp); err != nil {
			return err
		}
		if resp.StatusCode() >= http.StatusBadRequest {
			return fmt.Errorf("fasthttp: status %d", resp.StatusCode())
		}
		// Unmarshal copies all strings out of the pooled body buffer before it is released.
		return json.Unmarshal(resp.Body(), ep.result())
	}
}

// runEndpointBenchmarks runs every endpoint with every client kind, sequentially.
// Input: b [*testing.B], t [*target], skip [func] reports endpoints the target cannot serve.
func runEndpointBenchmarks(b *testing.B, t *target, skip func(ep string) bool) {
	b.Helper()
	ctx := context.Background()
	for i := range endpoints {
		ep := &endpoints[i]
		if skip != nil && skip(ep.name) {
			continue
		}
		for _, kind := range clientKinds {
			b.Run("endpoint="+ep.name+"/client="+kind.name, func(b *testing.B) {
				var dials atomic.Int64
				call := kind.build(t, &dials)
				must(call(ctx, ep)) // warm up the connection
				b.ReportAllocs()
				for b.Loop() {
					if err := call(ctx, ep); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// runPoolBenchmark runs one endpoint from many goroutines (GOMAXPROCS*parallelism) and reports new connections per
// 1000 calls; resty's default pool keeps only GOMAXPROCS+1 idle connections per host.
// Input: b [*testing.B], t [*target], epName [string], parallelism [int].
func runPoolBenchmark(b *testing.B, t *target, epName string, parallelism int) {
	b.Helper()
	ep := findEndpoint(epName)
	for _, kind := range clientKinds {
		b.Run("client="+kind.name, func(b *testing.B) {
			if b.N == 1 {
				// first ramp-up round: let sockets closed by the previous client drain, so it does not inherit a
				// full listener backlog (or, on Docker Desktop, a congested port forwarder)
				time.Sleep(t.cooldown)
			}
			var dials, failures atomic.Int64
			call := kind.build(t, &dials)
			if err := call(context.Background(), ep); err != nil {
				b.Fatal(err)
			}
			dials.Store(0)
			b.ResetTimer()
			b.ReportAllocs()
			b.SetParallelism(parallelism)
			b.RunParallel(func(pb *testing.PB) {
				ctx := context.Background()
				for pb.Next() {
					// connection churn can exhaust the listener backlog; count failures instead of aborting
					if err := call(ctx, ep); err != nil {
						failures.Add(1)
					}
				}
			})
			b.ReportMetric(float64(dials.Load())*1000/float64(b.N), "dials/1k-op")
			b.ReportMetric(float64(failures.Load())*1000/float64(b.N), "fails/1k-op")
		})
	}
}

// findEndpoint returns the endpoint named name. Input: name [string]. Output: [*endpoint].
func findEndpoint(name string) *endpoint {
	for i := range endpoints {
		if endpoints[i].name == name {
			return &endpoints[i]
		}
	}
	panic("unknown endpoint " + name)
}
