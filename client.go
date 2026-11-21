// Package gocloak is a golang keycloak adaptor.
package gocloak

import (
	"context"
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/Nerzal/gocloak/v15/pkg/jwx"
)

// GoCloak provides functionalities to talk to Keycloak.
type GoCloak struct {
	basePath    string
	certsCache  sync.Map
	issuerCache sync.Map   // realm -> client assertion audience
	versionLock sync.Mutex // guards Config.version and versionCall
	versionCall *versionCall
	restyClient *resty.Client
	Config      struct {
		CertsInvalidateTime time.Duration
		authAdminRealms     string
		authRealms          string
		tokenEndpoint       string
		revokeEndpoint      string
		logoutEndpoint      string
		openIDConnect       string
		attackDetection     string
		version             string
	}
}

// Verify struct implements interface
var _ GoCloakIface = &GoCloak{}

const (
	adminClientID string = "admin-cli"
	urlSeparator  string = "/"
)

// makeURL joins the given path segments with "/".
// Input: path [...string] URL segments.
// Output: [string] joined URL.
func makeURL(path ...string) string {
	return strings.Join(path, urlSeparator)
}

// versionCall is a running serverinfo request; done is closed once version and err are set.
type versionCall struct {
	done    chan struct{}
	token   string // token the request is made with; only callers with the same token join it
	version string
	err     error
}

// GetServerVersion returns the server version from the serverinfo endpoint.
// If the version is already set, it will return the cached version.
// Otherwise, it will fetch the version from the serverinfo endpoint and cache it.
// Safe for concurrent use: concurrent callers share one request and its result; a failure is not cached,
// so callers do not queue up behind each other while serverinfo fails. Only callers with the same token share a
// request, so one caller's expired token never fails another caller.
// Input: ctx [context.Context], token [string] admin token.
// Output: [string] server version, [error] if serverinfo fails or has no version
// (e.g. admins of non-master realms; use SetServerVersion then).
func (g *GoCloak) GetServerVersion(ctx context.Context, token string) (string, error) {
	g.versionLock.Lock()
	if g.Config.version != "" {
		defer g.versionLock.Unlock()
		return g.Config.version, nil
	}
	call := g.versionCall
	switch {
	case call == nil:
		call = &versionCall{done: make(chan struct{}), token: token}
		g.versionCall = call
		go g.fetchServerVersion(context.WithoutCancel(ctx), token, call)
	case call.token != token:
		// a request with another token is running: fetch with our own token, without replacing the shared call
		call = &versionCall{done: make(chan struct{}), token: token}
		go g.fetchServerVersion(context.WithoutCancel(ctx), token, call)
	}
	g.versionLock.Unlock()

	select {
	case <-call.done:
		return call.version, call.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// fetchServerVersion reads the version from serverinfo, caches it on success and completes call [*versionCall].
// Input: ctx [context.Context] (values only), token [string] admin token, call [*versionCall]. Output: none.
func (g *GoCloak) fetchServerVersion(ctx context.Context, token string, call *versionCall) {
	ctx, cancel := context.WithTimeout(ctx, lockedFetchTimeout)
	defer cancel()

	serverInfo, err := g.GetServerInfo(ctx, token)
	switch {
	case err != nil:
		call.err = err
	case serverInfo == nil || serverInfo.SystemInfo.Version == "":
		call.err = errors.New("could not get server version: serverinfo has no systemInfo.version, use SetServerVersion")
	default:
		call.version = serverInfo.SystemInfo.Version
	}

	g.versionLock.Lock()
	if call.err == nil {
		g.Config.version = call.version
	}
	if g.versionCall == call {
		g.versionCall = nil
	}
	g.versionLock.Unlock()
	close(call.done)
}

// GetRequest returns a base request bound to ctx with tracing headers and HTTPErrorResponse as error result.
// Input: ctx [context.Context].
// Output: [*resty.Request] base request.
func (g *GoCloak) GetRequest(ctx context.Context) *resty.Request {
	var err HTTPErrorResponse
	return injectTracingHeaders(ctx, g.restyClient.R().
		SetContext(ctx).
		SetError(&err))
}

// GetRequestWithBearerAuthNoCache returns a JSON base request configured with an auth token and no-cache header.
// Input: ctx [context.Context], token [string] bearer token.
// Output: [*resty.Request] prepared request.
func (g *GoCloak) GetRequestWithBearerAuthNoCache(ctx context.Context, token string) *resty.Request {
	return g.GetRequest(ctx).
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json").
		SetHeader("Cache-Control", "no-cache")
}

// GetRequestWithBearerAuth returns a JSON base request configured with an auth token.
// Input: ctx [context.Context], token [string] bearer token.
// Output: [*resty.Request] prepared request.
func (g *GoCloak) GetRequestWithBearerAuth(ctx context.Context, token string) *resty.Request {
	return g.GetRequest(ctx).
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json")
}

// GetRequestWithBearerAuthXMLHeader returns an XML base request configured with an auth token.
// Input: ctx [context.Context], token [string] bearer token.
// Output: [*resty.Request] prepared request.
func (g *GoCloak) GetRequestWithBearerAuthXMLHeader(ctx context.Context, token string) *resty.Request {
	return g.GetRequest(ctx).
		SetAuthToken(token).
		SetHeader("Content-Type", "application/xml;charset=UTF-8")
}

// GetRequestWithBasicAuth returns a form data base request configured with basic auth.
// Basic auth is only set when both clientID and clientSecret are given (public clients need none).
// Input: ctx [context.Context], clientID [string], clientSecret [string].
// Output: [*resty.Request] prepared request.
func (g *GoCloak) GetRequestWithBasicAuth(ctx context.Context, clientID, clientSecret string) *resty.Request {
	req := g.GetRequest(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")
	// Public client doesn't require Basic Auth
	if len(clientID) > 0 && len(clientSecret) > 0 {
		httpBasicAuth := base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(clientID) + ":" + url.QueryEscape(clientSecret)))
		req.SetHeader("Authorization", "Basic "+httpBasicAuth)
	}

	return req
}

// getRequestingParty posts an UMA ticket request to the token endpoint and decodes the response into a new T.
// Input: ctx [context.Context], g [*GoCloak], token [string], realm [string], options [RequestingPartyTokenOptions],
// errMessage [string] error prefix. Output: [T] decoded result, [error] *APIError if the call fails.
func getRequestingParty[T any](ctx context.Context, g *GoCloak, token, realm string, options RequestingPartyTokenOptions, errMessage string) (T, error) {
	return doRequest[T](g.GetRequestWithBearerAuth(ctx, token).
		SetFormData(options.FormData()).
		SetFormDataFromValues(url.Values{"permission": options.Permissions}),
		http.MethodPost, g.getOIDCURL(realm, g.Config.tokenEndpoint), errMessage)
}

// checkForError converts a transport error, a nil response or an HTTP error status into an *APIError.
// Input: resp [*resty.Response], err [error] transport error, errMessage [string] error prefix.
// Output: [error] *APIError, or nil if the request succeeded.
func checkForError(resp *resty.Response, err error, errMessage string) error {
	if err != nil {
		err = redactURLError(err)
		return &APIError{
			Code:    0,
			Message: fmt.Sprintf("%s: %v", errMessage, err),
			Type:    ParseAPIErrType(err),
			Err:     err,
		}
	}

	if resp == nil {
		return &APIError{
			Message: "empty response",
			Type:    APIErrTypeUnknown,
		}
	}

	if resp.IsError() {
		var msg string

		if e, ok := resp.Error().(*HTTPErrorResponse); ok && e.NotEmpty() {
			msg = fmt.Sprintf("%s: %s", resp.Status(), e)
		} else {
			msg = resp.Status()
		}

		return &APIError{
			Code:    resp.StatusCode(),
			Message: msg,
			Type:    ParseAPIErrType(errors.New(msg)),
		}
	}

	return nil
}

// redactURLError removes the query string from the URL of a transport error, so search parameters like email or
// username do not end up in logs. Other errors are returned unchanged; errors.Is/As still reach the cause.
// Input: err [error]. Output: [error] err with redacted URL.
func redactURLError(err error) error {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return err
	}
	redacted, _, found := strings.Cut(urlErr.URL, "?")
	if !found {
		return err
	}
	return &url.Error{Op: urlErr.Op, URL: redacted + "?REDACTED", Err: urlErr.Err}
}

// doRequest sends req with the given HTTP method to url and decodes the JSON response into a new T.
// Input: req [*resty.Request] prepared request, method [string] HTTP method, url [string], errMessage [string] error prefix.
// Output: [T] decoded result, [error] *APIError if the call fails.
func doRequest[T any](req *resty.Request, method, url, errMessage string) (T, error) {
	var result T
	resp, err := req.SetResult(&result).Execute(method, url)
	if err = checkForError(resp, err, errMessage); err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

// doRequestPtr works like doRequest but returns a pointer to the decoded result.
// Input: req [*resty.Request] prepared request, method [string] HTTP method, url [string], errMessage [string] error prefix.
// Output: [*T] decoded result, [error] *APIError if the call fails.
func doRequestPtr[T any](req *resty.Request, method, url, errMessage string) (*T, error) {
	result, err := doRequest[T](req, method, url, errMessage)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// doNoResult sends req with the given HTTP method to url and ignores the response body.
// Input: req [*resty.Request] prepared request, method [string] HTTP method, url [string], errMessage [string] error prefix.
// Output: [error] *APIError if the call fails.
func doNoResult(req *resty.Request, method, url, errMessage string) error {
	resp, err := req.Execute(method, url)
	return checkForError(resp, err, errMessage)
}

// doCreate posts req to url and returns the ID of the created resource from the Location header.
// Input: req [*resty.Request] prepared request, url [string], errMessage [string] error prefix.
// Output: [string] ID of the created resource, [error] *APIError if the call fails.
func doCreate(req *resty.Request, url, errMessage string) (string, error) {
	resp, err := req.Post(url)
	if err = checkForError(resp, err, errMessage); err != nil {
		return "", err
	}
	return getID(resp), nil
}

// getID returns the unescaped last path segment of the Location header, i.e. the ID of a created resource.
// Keycloak percent-encodes names (e.g. role names with spaces) there; the URL helpers escape again on reuse.
// Input: resp [*resty.Response].
// Output: [string] resource ID.
func getID(resp *resty.Response) string {
	header := resp.Header().Get("Location")
	last := header[strings.LastIndex(header, urlSeparator)+1:]
	if id, err := url.PathUnescape(last); err == nil {
		return id
	}
	return last
}

// injectTracingHeaders writes the OpenTelemetry trace context of ctx into the request headers.
// It uses the propagator attached via WithPropagator, otherwise the global one; without a span in ctx it adds nothing.
// Input: ctx [context.Context], req [*resty.Request]. Output: [*resty.Request] the same request.
func injectTracingHeaders(ctx context.Context, req *resty.Request) *resty.Request {
	// look for propagator in context, use the global propagator if not found
	propagator, ok := ctx.Value(propagatorContextKey).(propagation.TextMapPropagator)
	if !ok || propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}

	// inject tracing headers into request; this is a no-op when ctx carries no span
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	return req
}

// ===============
// Keycloak client
// ===============

// defaultRequestTimeout bounds every request, so a hanging Keycloak cannot block callers without a context deadline.
const defaultRequestTimeout = 30 * time.Second

// defaultResponseBodyLimit bounds the size of a response body, so a misbehaving server cannot exhaust memory.
// It is far above what Keycloak returns for normal admin and token calls.
const defaultResponseBodyLimit = 100 << 20 // 100 MiB

// maxRedirects is the number of redirects followed, matching the net/http default.
const maxRedirects = 10

// sameOriginRedirectPolicy follows at most maxRedirects redirects and only to the scheme, host and port of the
// original request. Go re-sends form bodies (passwords, refresh tokens, client assertions) on 307/308 and keeps the
// Authorization header when only the port changes, so a redirect to another origin could leak credentials.
var sameOriginRedirectPolicy = resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return fmt.Errorf("stopped after %d redirects", maxRedirects)
	}
	if origin := via[0].URL; req.URL.Scheme != origin.Scheme || req.URL.Host != origin.Host {
		return fmt.Errorf("redirect to another origin (%s://%s) is not allowed", req.URL.Scheme, req.URL.Host)
	}
	return nil
})

// NewClient creates a client for the Keycloak at basePath [string], configured by options [...func(*GoCloak)].
// Requests time out after 30 seconds, response bodies are limited to 100 MiB and redirects are only followed to the
// same origin; change these via RestyClient().SetTimeout, SetResponseBodyLimit and SetRedirectPolicy.
// Output: [*GoCloak].
func NewClient(basePath string, options ...func(*GoCloak)) *GoCloak {
	c := GoCloak{
		basePath: strings.TrimRight(basePath, urlSeparator),
		restyClient: resty.New().
			SetTimeout(defaultRequestTimeout).
			SetResponseBodyLimit(defaultResponseBodyLimit).
			SetRedirectPolicy(sameOriginRedirectPolicy),
	}

	c.Config.CertsInvalidateTime = 10 * time.Minute
	c.Config.authAdminRealms = makeURL("admin", "realms")
	c.Config.authRealms = makeURL("realms")
	c.Config.tokenEndpoint = makeURL("protocol", "openid-connect", "token")
	c.Config.logoutEndpoint = makeURL("protocol", "openid-connect", "logout")
	c.Config.revokeEndpoint = makeURL("protocol", "openid-connect", "revoke")
	c.Config.openIDConnect = makeURL("protocol", "openid-connect")
	c.Config.attackDetection = makeURL("attack-detection", "brute-force")

	for _, option := range options {
		option(&c)
	}

	return &c
}

// RestyClient returns the internal resty client, e.g. to configure TLS, proxies, timeouts or debug mode.
// Output: [*resty.Client].
func (g *GoCloak) RestyClient() *resty.Client {
	return g.restyClient
}

// SetRestyClient replaces the internal resty client with restyClient [*resty.Client]; its timeout settings replace the default.
func (g *GoCloak) SetRestyClient(restyClient *resty.Client) {
	g.restyClient = restyClient
}

// escapeSegment path-escapes one URL segment, so IDs containing "/", "?", "#" or ".." stay a single segment.
// Input: segment [string]. Output: [string] escaped segment.
func escapeSegment(segment string) string {
	if segment == "." || segment == ".." {
		return strings.ReplaceAll(segment, ".", "%2E")
	}
	return url.PathEscape(segment)
}

// escapeSegments applies escapeSegment to every element.
// Input: segments [...string]. Output: [[]string] escaped segments.
func escapeSegments(segments ...string) []string {
	escaped := make([]string, len(segments))
	for i, segment := range segments {
		escaped[i] = escapeSegment(segment)
	}
	return escaped
}

// getRealmURL builds {basePath}/{authRealms}/{realm}/{path...} with escaped realm and path segments.
// Input: realm [string], path [...string]. Output: [string] URL.
func (g *GoCloak) getRealmURL(realm string, path ...string) string {
	return makeURL(slices.Concat([]string{g.basePath, g.Config.authRealms, escapeSegment(realm)}, escapeSegments(path...))...)
}

// getOIDCURL builds {basePath}/{authRealms}/{realm}/{endpoint}/{path...}; endpoint is a configured path and stays unescaped.
// Input: realm [string], endpoint [string], path [...string]. Output: [string] URL.
func (g *GoCloak) getOIDCURL(realm, endpoint string, path ...string) string {
	return makeURL(slices.Concat([]string{g.basePath, g.Config.authRealms, escapeSegment(realm), endpoint}, escapeSegments(path...))...)
}

// getAdminRealmURL builds {basePath}/{authAdminRealms}/{realm}/{path...} with escaped realm and path segments.
// Input: realm [string], path [...string]. Output: [string] URL.
func (g *GoCloak) getAdminRealmURL(realm string, path ...string) string {
	return makeURL(slices.Concat([]string{g.basePath, g.Config.authAdminRealms, escapeSegment(realm)}, escapeSegments(path...))...)
}

// getAttackDetectionURL builds the brute-force detection URL with escaped realm, user and path segments.
// Input: realm [string], user [string], path [...string]. Output: [string] URL.
func (g *GoCloak) getAttackDetectionURL(realm string, user string, path ...string) string {
	return makeURL(slices.Concat([]string{g.basePath, g.Config.authAdminRealms, escapeSegment(realm), g.Config.attackDetection, escapeSegment(user)}, escapeSegments(path...))...)
}

// ==== Functional Options ===

// SetServerVersion returns an option that sets the server version to bypass the serverinfo call.
// Input: version [string] Keycloak server version.
// Output: [func(g *GoCloak)] option for NewClient.
func SetServerVersion(version string) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.version = version
	}
}

// SetLegacyWildFlySupport returns an option that switches to the legacy WildFly "/auth" path prefixes.
// Output: [func(g *GoCloak)] option for NewClient.
func SetLegacyWildFlySupport() func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.authAdminRealms = makeURL("auth", "admin", "realms")
		g.Config.authRealms = makeURL("auth", "realms")
	}
}

// SetAuthRealms returns an option that sets the path prefix of the realms endpoints.
// Input: url [string] realms path.
// Output: [func(g *GoCloak)] option for NewClient.
func SetAuthRealms(url string) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.authRealms = url
	}
}

// SetAuthAdminRealms returns an option that sets the path prefix of the admin realms endpoints.
// Input: url [string] admin realms path.
// Output: [func(g *GoCloak)] option for NewClient.
func SetAuthAdminRealms(url string) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.authAdminRealms = url
	}
}

// SetTokenEndpoint returns an option that sets the token endpoint path.
// Input: url [string] token endpoint path.
// Output: [func(g *GoCloak)] option for NewClient.
func SetTokenEndpoint(url string) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.tokenEndpoint = url
	}
}

// SetRevokeEndpoint returns an option that sets the revoke endpoint path.
// Input: url [string] revoke endpoint path.
// Output: [func(g *GoCloak)] option for NewClient.
func SetRevokeEndpoint(url string) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.revokeEndpoint = url
	}
}

// SetLogoutEndpoint sets the logout endpoint path url [string] relative to the realm.
func SetLogoutEndpoint(url string) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.logoutEndpoint = url
	}
}

// SetOpenIDConnectEndpoint sets the OpenID Connect base path url [string] relative to the realm.
func SetOpenIDConnectEndpoint(url string) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.openIDConnect = url
	}
}

// SetCertCacheInvalidationTime sets how long fetched certs are cached, duration [time.Duration].
// Zero disables caching: every token decode fetches the certs, one fetch per realm at a time.
func SetCertCacheInvalidationTime(duration time.Duration) func(g *GoCloak) {
	return func(g *GoCloak) {
		g.Config.CertsInvalidateTime = duration
	}
}

// GetServerInfo fetches the server info from GET /admin/serverinfo.
// Input: ctx [context.Context], accessToken [string] admin token.
// Output: [*ServerInfoRepresentation] server info, [error] if the call fails.
func (g *GoCloak) GetServerInfo(ctx context.Context, accessToken string) (*ServerInfoRepresentation, error) {
	errMessage := "could not get server info"
	return doRequestPtr[ServerInfoRepresentation](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet,
		makeURL(g.basePath, strings.TrimSuffix(g.Config.authAdminRealms, urlSeparator+"realms"), "serverinfo"), errMessage)
}

// GetUserInfo calls the OIDC UserInfo endpoint.
// Input: ctx [context.Context], accessToken [string] user token, realm [string].
// Output: [*UserInfo] user info, [error] if the call fails.
func (g *GoCloak) GetUserInfo(ctx context.Context, accessToken, realm string) (*UserInfo, error) {
	const errMessage = "could not get user info"

	return doRequestPtr[UserInfo](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getOIDCURL(realm, g.Config.openIDConnect, "userinfo"), errMessage)
}

// GetRawUserInfo calls the OIDC UserInfo endpoint and returns the raw JSON object.
// Input: ctx [context.Context], accessToken [string] user token, realm [string].
// Output: [map[string]any] user info claims, [error] if the call fails.
func (g *GoCloak) GetRawUserInfo(ctx context.Context, accessToken, realm string) (map[string]any, error) {
	const errMessage = "could not get user info"

	return doRequest[map[string]any](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getOIDCURL(realm, g.Config.openIDConnect, "userinfo"), errMessage)
}

// getNewCerts fetches the JWKS of realm [string] from Keycloak, without caching.
// Input: ctx [context.Context], realm [string]. Output: [*CertResponse] certs, [error] if the request fails.
func (g *GoCloak) getNewCerts(ctx context.Context, realm string) (*CertResponse, error) {
	const errMessage = "could not get newCerts"

	return doRequestPtr[CertResponse](g.GetRequest(ctx), http.MethodGet, g.getOIDCURL(realm, g.Config.openIDConnect, "certs"), errMessage)
}

// signingKey is a parsed JWK used to verify token signatures.
type signingKey struct {
	key crypto.PublicKey
	alg string // JWK alg, empty if Keycloak did not publish one
}

// certsState is one successful certs fetch of a realm. It is never modified after creation.
type certsState struct {
	certs     *CertResponse
	keys      map[string]signingKey // signature keys by kid, parsed once per fetch
	fetchedAt time.Time
}

// realmCerts is the certs cache of one realm.
type realmCerts struct {
	mu    sync.Mutex // guards all fields
	state *certsState
	err   error
	errAt time.Time
	call  *certsCall // running fetch, nil if none
}

// certsCall is a running certs fetch shared by all callers of a realm; done is closed once state and err are set.
type certsCall struct {
	done  chan struct{}
	state *certsState
	err   error
}

// finishedCertsCall returns a completed call holding state [*certsState] and err [error].
// Output: [*certsCall] with done already closed.
func finishedCertsCall(state *certsState, err error) *certsCall {
	call := &certsCall{done: make(chan struct{}), state: state, err: err}
	close(call.done)
	return call
}

// lockedFetchTimeout bounds cert fetches, so a hanging Keycloak cannot block waiting callers.
const lockedFetchTimeout = 10 * time.Second

// certsStaleFactor bounds how long stale certs are served while refreshing: up to factor * CertsInvalidateTime.
// After that, keys Keycloak may have revoked are no longer trusted and validation fails until a fetch succeeds.
const certsStaleFactor = 2

// certsMinRefreshInterval limits refetches caused by unknown key ids, so tokens with random kids cannot flood Keycloak.
var certsMinRefreshInterval = 30 * time.Second

// certsErrorRetryInterval is how long a failed fetch is reported to callers before Keycloak is asked again.
var certsErrorRetryInterval = 5 * time.Second

// realmCertsFor returns the cache of realm [string], creating it if needed.
// Entries of realms whose certs were never fetched successfully are removed after the failed fetch.
// Input: realm [string]. Output: [*realmCerts].
func (g *GoCloak) realmCertsFor(realm string) *realmCerts {
	value, _ := g.certsCache.LoadOrStore(realm, &realmCerts{})
	rc, _ := value.(*realmCerts)
	return rc
}

// load returns the current state and the error of a fetch that failed within certsErrorRetryInterval.
// Output: [*certsState] or nil, [error] recent fetch error or nil.
func (rc *realmCerts) load() (*certsState, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.err != nil && time.Since(rc.errAt) < certsErrorRetryInterval {
		return rc.state, rc.err
	}
	return rc.state, nil
}

// startCertsFetch returns the running fetch of realm [string] or starts one, unless the cache already holds a state
// younger than maxAge [time.Duration] or a fetch failed within certsErrorRetryInterval; then the returned call is
// already complete. One fetch per realm runs at a time, however many callers wait for it.
// Input: ctx [context.Context] (values only), realm [string], rc [*realmCerts], maxAge [time.Duration].
// Output: [*certsCall].
func (g *GoCloak) startCertsFetch(ctx context.Context, realm string, rc *realmCerts, maxAge time.Duration) *certsCall {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.call != nil {
		return rc.call
	}
	if rc.state != nil && time.Since(rc.state.fetchedAt) < maxAge {
		return finishedCertsCall(rc.state, nil)
	}
	if rc.err != nil && time.Since(rc.errAt) < certsErrorRetryInterval {
		return finishedCertsCall(nil, rc.err)
	}

	call := &certsCall{done: make(chan struct{})}
	rc.call = call
	go g.runCertsFetch(ctx, realm, rc, call)
	return call
}

// runCertsFetch fetches the certs of realm [string], stores the result (or the error) in rc [*realmCerts] and
// completes call [*certsCall]. The fetch is detached from the caller's cancellation, so a disconnected caller is not
// mistaken for a Keycloak outage, and bounded by lockedFetchTimeout.
// Input: ctx [context.Context] (values only), realm [string], rc [*realmCerts], call [*certsCall]. Output: none.
func (g *GoCloak) runCertsFetch(ctx context.Context, realm string, rc *realmCerts, call *certsCall) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), lockedFetchTimeout)
	defer cancel()

	certs, err := g.getNewCerts(ctx, realm)

	rc.mu.Lock()
	defer rc.mu.Unlock()
	if err != nil {
		rc.err, rc.errAt = err, time.Now()
		call.err = err
		if rc.state == nil {
			// never fetched successfully (e.g. unknown realm): drop the entry, so random realm names cannot fill the cache
			g.certsCache.CompareAndDelete(realm, rc)
		}
	} else {
		rc.state = &certsState{certs: certs, keys: parseSigningKeys(certs), fetchedAt: time.Now()}
		rc.err = nil
		call.state = rc.state
	}
	rc.call = nil
	close(call.done)
}

// parseSigningKeys parses the signature keys of certs [*CertResponse]; encryption keys and unparsable keys are skipped.
// Input: certs [*CertResponse]. Output: [map[string]signingKey] keys by kid.
func parseSigningKeys(certs *CertResponse) map[string]signingKey {
	keys := make(map[string]signingKey, len(certs.Keys))
	for _, k := range certs.Keys {
		if k.Kid == "" || (k.Use != "" && k.Use != "sig") {
			continue
		}
		key, err := jwx.DecodePublicKey(k.Kty, k.E, k.N, k.X, k.Y, k.Crv)
		if err != nil {
			continue
		}
		keys[k.Kid] = signingKey{key: key, alg: k.Alg}
	}
	return keys
}

// getCachedCerts returns the certs of realm [string], fetching them if the cached state is older than maxAge.
// With allowStale, a state younger than certsStaleFactor * CertsInvalidateTime is returned right away while one
// background refresh runs. A failed fetch is reported for certsErrorRetryInterval without asking Keycloak again.
// Callers waiting for a fetch return as soon as ctx is done; the fetch keeps running and fills the cache.
// Input: ctx [context.Context], realm [string], maxAge [time.Duration], allowStale [bool].
// Output: [*certsState], [error] if no usable state exists and fetching fails or ctx is done first.
func (g *GoCloak) getCachedCerts(ctx context.Context, realm string, maxAge time.Duration, allowStale bool) (*certsState, error) {
	const errMessage = "could not get certs"

	rc := g.realmCertsFor(realm)
	state, recentErr := rc.load()
	if state != nil {
		age := time.Since(state.fetchedAt)
		if age < maxAge {
			return state, nil
		}
		if allowStale && age < certsStaleFactor*g.Config.CertsInvalidateTime {
			if recentErr == nil {
				g.startCertsFetch(ctx, realm, rc, maxAge) // background refresh, joins a running one
			}
			return state, nil
		}
	}
	if recentErr != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, recentErr)
	}

	call := g.startCertsFetch(ctx, realm, rc, maxAge)
	select {
	case <-call.done:
		if call.err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, call.err)
		}
		return call.state, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("%s: %w", errMessage, ctx.Err())
	}
}

// GetCerts fetches certificates for the given realm from the public /open-id-connect/certs endpoint.
// Results are cached for Config.CertsInvalidateTime; afterwards the cached certs are served for up to
// another interval while they are refreshed in the background.
// Input: ctx [context.Context], realm [string].
// Output: [*CertResponse] certs, [error] if fetching fails.
func (g *GoCloak) GetCerts(ctx context.Context, realm string) (*CertResponse, error) {
	state, err := g.getCachedCerts(ctx, realm, g.Config.CertsInvalidateTime, true)
	if err != nil {
		return nil, err
	}
	// deep copy, so callers modifying the result cannot change the cached certs
	keys := slices.Clone(state.certs.Keys)
	for i := range keys {
		keys[i].KeyOps = slices.Clone(keys[i].KeyOps)
		keys[i].X5c = slices.Clone(keys[i].X5c)
	}
	return &CertResponse{Keys: keys}, nil
}

// GetIssuer fetches the issuer info of the given realm from the public realm endpoint.
// Input: ctx [context.Context], realm [string].
// Output: [*IssuerResponse] issuer info, [error] if the call fails.
func (g *GoCloak) GetIssuer(ctx context.Context, realm string) (*IssuerResponse, error) {
	const errMessage = "could not get issuer"

	return doRequestPtr[IssuerResponse](g.GetRequest(ctx), http.MethodGet, g.getRealmURL(realm), errMessage)
}

// RetrospectToken introspects a requesting party token via the OIDC token introspect endpoint.
// Input: ctx [context.Context], accessToken [string] token to introspect, clientID [string], clientSecret [string],
// realm [string].
// Output: [*IntroSpectTokenResult] introspection result, [error] if the call fails.
func (g *GoCloak) RetrospectToken(ctx context.Context, accessToken, clientID, clientSecret, realm string) (*IntroSpectTokenResult, error) {
	const errMessage = "could not introspect requesting party token"

	return doRequestPtr[IntroSpectTokenResult](g.GetRequestWithBasicAuth(ctx, clientID, clientSecret).
		SetFormData(map[string]string{
			"token_type_hint": "requesting_party_token",
			"token":           accessToken,
		}),
		http.MethodPost, g.getOIDCURL(realm, g.Config.tokenEndpoint, "introspect"), errMessage)
}

// decodeAccessTokenWithClaims verifies accessToken [string] with the signature key of realm [string] named by the
// token's kid and decodes it into claims [jwt.Claims]. An unknown kid triggers one rate-limited cert refetch.
// Input: ctx [context.Context], accessToken [string], realm [string], claims [jwt.Claims].
// Output: [*jwt.Token], [error] if no fitting key exists or verification fails.
func (g *GoCloak) decodeAccessTokenWithClaims(ctx context.Context, accessToken, realm string, claims jwt.Claims, opts ...jwt.ParserOption) (*jwt.Token, error) {
	const errMessage = "could not decode access token"
	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	decodedHeader, err := jwx.DecodeAccessTokenHeader(accessToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	state, err := g.getCachedCerts(ctx, realm, g.Config.CertsInvalidateTime, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	key, ok := state.keys[decodedHeader.Kid]
	if !ok {
		// unknown kid: Keycloak may have rotated its keys, refetch (rate limited)
		state, err = g.getCachedCerts(ctx, realm, certsMinRefreshInterval, false)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, err)
		}
		key, ok = state.keys[decodedHeader.Kid]
	}
	if !ok {
		return nil, fmt.Errorf("%s: cannot find a signature key to decode the token", errMessage)
	}

	return jwx.DecodeAccessTokenWithKey(accessToken, key.key, key.alg, claims, opts...)
}

// DecodeAccessToken verifies accessToken [string] against the signature keys of realm [string] and decodes its claims.
// Only signature, exp and nbf are checked. ID tokens of the realm are signed with the same keys and pass too, so
// callers must check iss, typ ("Bearer"), aud and azp themselves, or use ValidateAccessToken. Prefer a realm from
// trusted configuration: every successfully fetched realm stays in the certs cache.
// Input: ctx [context.Context], accessToken [string] optionally prefixed with "Bearer ", realm [string].
// Output: [*jwt.Token], [*jwt.MapClaims], [error] if verification fails.
func (g *GoCloak) DecodeAccessToken(ctx context.Context, accessToken, realm string) (*jwt.Token, *jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := g.decodeAccessTokenWithClaims(ctx, accessToken, realm, claims)
	if err != nil {
		return nil, nil, err
	}
	return token, &claims, nil
}

// DecodeAccessTokenCustomClaims works like DecodeAccessToken but decodes the claims into claims [jwt.Claims].
// The same limits apply: check iss, typ, aud and azp yourself, or use ValidateAccessTokenCustomClaims.
// Input: ctx [context.Context], accessToken [string], realm [string], claims [jwt.Claims].
// Output: [*jwt.Token], [error] if verification fails.
func (g *GoCloak) DecodeAccessTokenCustomClaims(ctx context.Context, accessToken, realm string, claims jwt.Claims) (*jwt.Token, error) {
	return g.decodeAccessTokenWithClaims(ctx, accessToken, realm, claims)
}

// ValidateAccessToken verifies accessToken [string] like DecodeAccessToken and additionally requires exp and checks
// iss, typ, aud and azp, so ID tokens, tokens without expiry, tokens of other realms and tokens issued to other
// clients are rejected. At least one of validation.Audience and
// validation.AuthorizedParty must be set.
// Input: ctx [context.Context], accessToken [string] optionally prefixed with "Bearer ", realm [string] trusted realm,
// validation [TokenValidation] expected claims. Output: [*jwt.Token], [*jwt.MapClaims], [error] if a check fails.
func (g *GoCloak) ValidateAccessToken(ctx context.Context, accessToken, realm string, validation TokenValidation) (*jwt.Token, *jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := g.ValidateAccessTokenCustomClaims(ctx, accessToken, realm, claims, validation)
	if err != nil {
		return nil, nil, err
	}
	return token, &claims, nil
}

// ValidateAccessTokenCustomClaims works like ValidateAccessToken but decodes the claims into claims [jwt.Claims].
// Input: ctx [context.Context], accessToken [string] optionally prefixed with "Bearer ", realm [string] trusted realm,
// claims [jwt.Claims] target, validation [TokenValidation] expected claims.
// Output: [*jwt.Token], [error] if verification or a check fails.
func (g *GoCloak) ValidateAccessTokenCustomClaims(ctx context.Context, accessToken, realm string, claims jwt.Claims, validation TokenValidation) (*jwt.Token, error) {
	const errMessage = "could not validate access token"
	if validation.Audience == "" && validation.AuthorizedParty == "" {
		return nil, fmt.Errorf("%s: audience or authorized party required", errMessage)
	}

	// Keycloak access tokens always carry exp; a token without one would never expire
	token, err := g.decodeAccessTokenWithClaims(ctx, accessToken, realm, claims,
		jwt.WithLeeway(validation.Leeway), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		// custom structs may lack typ or azp; read them from the payload whose signature was just verified
		mapClaims = jwt.MapClaims{}
		if _, _, err := jwt.NewParser().ParseUnverified(token.Raw, mapClaims); err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, err)
		}
	}
	if err := checkTokenIssuer(mapClaims, realm); err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	if err := checkTokenClaims(mapClaims, validation); err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	return token, nil
}

// checkTokenIssuer checks that the iss claim of claims [jwt.MapClaims] names realm [string]. Only the path is
// compared, because the public hostname in iss may differ from the configured base URL. This rejects tokens of other
// realms even if realms share a signing key.
// Input: claims [jwt.MapClaims], realm [string]. Output: [error] if iss is missing or names another realm.
func checkTokenIssuer(claims jwt.MapClaims, realm string) error {
	iss, _ := claims["iss"].(string)
	if strings.HasSuffix(iss, "/realms/"+realm) || strings.HasSuffix(iss, "/realms/"+url.PathEscape(realm)) {
		return nil
	}
	return fmt.Errorf("token issuer %q does not belong to realm %q", iss, realm)
}

// checkTokenClaims checks typ, aud and azp of claims [jwt.MapClaims] against validation [TokenValidation].
// Input: claims [jwt.MapClaims], validation [TokenValidation]. Output: [error] naming the first failed check.
func checkTokenClaims(claims jwt.MapClaims, validation TokenValidation) error {
	wantType := validation.Type
	if wantType == "" {
		wantType = "Bearer"
	}
	if typ, _ := claims["typ"].(string); typ != wantType {
		return fmt.Errorf("unexpected token type %q, want %q", typ, wantType)
	}

	if validation.Audience != "" {
		aud, err := claims.GetAudience()
		if err != nil || !slices.Contains(aud, validation.Audience) {
			return fmt.Errorf("token audience does not contain %q", validation.Audience)
		}
	}

	if validation.AuthorizedParty != "" {
		if azp, _ := claims["azp"].(string); azp != validation.AuthorizedParty {
			return fmt.Errorf("unexpected authorized party %q, want %q", azp, validation.AuthorizedParty)
		}
	}
	return nil
}

// GetToken fetches a token from the OIDC token endpoint using the given TokenOptions.
// Basic auth is used when a client secret is set, which then requires a client ID.
// Input: ctx [context.Context], realm [string], options [TokenOptions].
// Output: [*JWT] token, [error] if the call fails.
func (g *GoCloak) GetToken(ctx context.Context, realm string, options TokenOptions) (*JWT, error) {
	const errMessage = "could not get token"

	var req *resty.Request

	if options.ClientSecret != "" {
		if options.ClientID == "" {
			return nil, errors.New("client id required when a client secret is set")
		}
		req = g.GetRequestWithBasicAuth(ctx, options.ClientID, options.ClientSecret)
	} else {
		req = g.GetRequest(ctx)
	}

	return doRequestPtr[JWT](req.SetFormData(options.FormData()), http.MethodPost, g.getOIDCURL(realm, g.Config.tokenEndpoint), errMessage)
}

// GetRequestingPartyToken returns a requesting party token with the permissions granted by the server.
// Input: ctx [context.Context], token [string] access token, realm [string], options [RequestingPartyTokenOptions].
// Output: [*JWT] RPT, [error] if the call fails.
func (g *GoCloak) GetRequestingPartyToken(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) (*JWT, error) {
	const errMessage = "could not get requesting party token"

	return getRequestingParty[*JWT](ctx, g, token, realm, options, errMessage)
}

// GetRequestingPartyPermissions returns the requesting party permissions granted by the server.
// Input: ctx [context.Context], token [string] access token, realm [string], options [RequestingPartyTokenOptions].
// Output: [[]RequestingPartyPermission] granted permissions, [error] if the call fails.
func (g *GoCloak) GetRequestingPartyPermissions(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) ([]RequestingPartyPermission, error) {
	const errMessage = "could not get requesting party token"

	options.ResponseMode = "permissions"

	return getRequestingParty[[]RequestingPartyPermission](ctx, g, token, realm, options, errMessage)
}

// GetRequestingPartyPermissionDecision returns the requesting party permission decision of the server.
// Input: ctx [context.Context], token [string] access token, realm [string], options [RequestingPartyTokenOptions].
// Output: [*RequestingPartyPermissionDecision] decision, [error] if the call fails.
func (g *GoCloak) GetRequestingPartyPermissionDecision(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) (*RequestingPartyPermissionDecision, error) {
	const errMessage = "could not get requesting party token"

	options.ResponseMode = "decision"

	return getRequestingParty[*RequestingPartyPermissionDecision](ctx, g, token, realm, options, errMessage)
}

// RefreshToken exchanges the given refresh token for a new token.
// Input: ctx [context.Context], refreshToken [string], clientID [string], clientSecret [string], realm [string].
// Output: [*JWT] new token, [error] *APIError with details if the call fails.
func (g *GoCloak) RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*JWT, error) {
	return g.GetToken(ctx, realm, TokenOptions{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	})
}

// LoginAdmin performs a password login with the admin-cli client.
// Input: ctx [context.Context], username [string], password [string], realm [string].
// Output: [*JWT] token, [error] if the login fails.
func (g *GoCloak) LoginAdmin(ctx context.Context, username, password, realm string) (*JWT, error) {
	return g.GetToken(ctx, realm, TokenOptions{
		ClientID:  adminClientID,
		GrantType: "password",
		Username:  username,
		Password:  password,
	})
}

// LoginAdminOtp performs a password login with the admin-cli client and an OTP token.
// Input: ctx [context.Context], username [string], password [string], totp [string] OTP code, realm [string].
// Output: [*JWT] token, [error] if the login fails.
func (g *GoCloak) LoginAdminOtp(ctx context.Context, username, password, totp, realm string) (*JWT, error) {
	return g.GetToken(ctx, realm, TokenOptions{
		ClientID:  adminClientID,
		GrantType: "password",
		Username:  username,
		Password:  password,
		Totp:      totp,
	})
}

// LoginClient performs a login with client credentials.
// Input: ctx [context.Context], clientID [string], clientSecret [string], realm [string], scopes [...string] optional.
// Output: [*JWT] token, [error] if the login fails.
func (g *GoCloak) LoginClient(ctx context.Context, clientID, clientSecret, realm string, scopes ...string) (*JWT, error) {
	opts := TokenOptions{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "client_credentials",
		Scopes:       scopes,
	}

	return g.GetToken(ctx, realm, opts)
}

// LoginClientTokenExchange exchanges the presented access token for a token of the requesting client
// (standard token exchange, enabled per client via "standard.token.exchange.enabled").
// targetClient is sent as audience only when set; userID (requested_subject) is only supported by the
// legacy token exchange v1 preview feature.
// See https://www.keycloak.org/securing-apps/token-exchange
// Input: ctx [context.Context], clientID [string], token [string] subject access token, clientSecret [string],
// realm [string], targetClient [string] optional audience, userID [string] optional subject. Output: [*JWT], [error].
func (g *GoCloak) LoginClientTokenExchange(ctx context.Context, clientID, token, clientSecret, realm, targetClient, userID string) (*JWT, error) {
	tokenOptions := TokenOptions{ //nolint:gosec // URN constants, not credentials
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		SubjectToken:       token,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:access_token",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:refresh_token",
	}
	if targetClient != "" {
		tokenOptions.Audience = targetClient
	}
	if userID != "" {
		tokenOptions.RequestedSubject = userID
	}
	return g.GetToken(ctx, realm, tokenOptions)
}

// DirectNakedImpersonationTokenExchange performs "Direct Naked Impersonation" and returns a token for userID.
// Only available with the legacy token exchange v1 (preview feature "token-exchange"); standard token exchange
// in Keycloak 26.2+ rejects requests without subject_token.
// Input: ctx [context.Context], clientID [string], clientSecret [string], realm [string], userID [string].
// Output: [*JWT] token of the impersonated user, [error] if the exchange fails.
func (g *GoCloak) DirectNakedImpersonationTokenExchange(ctx context.Context, clientID, clientSecret, realm, userID string) (*JWT, error) {
	tokenOptions := TokenOptions{ //nolint:gosec // URN constants, not credentials
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:refresh_token",
		RequestedSubject:   userID,
	}
	return g.GetToken(ctx, realm, tokenOptions)
}

// getSignedJWTAudience returns the audience Keycloak expects in client assertions of realm [string]: its public
// issuer, which differs from basePath behind proxies or internal URLs. The result is cached per realm.
// Input: ctx [context.Context], realm [string]. Output: [string] audience, [error] if the issuer cannot be read.
func (g *GoCloak) getSignedJWTAudience(ctx context.Context, realm string) (string, error) {
	if cached, ok := g.issuerCache.Load(realm); ok {
		audience, _ := cached.(string)
		return audience, nil
	}

	issuer, err := g.GetIssuer(ctx, realm)
	if err != nil {
		return "", err
	}
	if issuer.TokenService == "" {
		return "", errors.New("could not get issuer: realm response has no token-service")
	}
	// shortcut: cached for the client's lifetime, a changed Keycloak hostname needs a new client.
	audience := strings.TrimSuffix(issuer.TokenService, "/protocol/openid-connect")
	g.issuerCache.Store(realm, audience)
	return audience, nil
}

// signedJWTDefaultLifetime is the client assertion lifetime used when the caller passes no expiry;
// it stays below Keycloak's default maximum of 60 seconds.
const signedJWTDefaultLifetime = 30 * time.Second

// LoginClientSignedJWT logs in clientID [string] with a client assertion signed by key [any] using signedMethod
// [jwt.SigningMethod]. A nil expiresAt [*jwt.NumericDate] defaults to now + 30 seconds.
// Input: ctx [context.Context], clientID [string], realm [string], key [any], signedMethod [jwt.SigningMethod],
// expiresAt [*jwt.NumericDate] or nil. Output: [*JWT], [error].
func (g *GoCloak) LoginClientSignedJWT(
	ctx context.Context,
	clientID,
	realm string,
	key any,
	signedMethod jwt.SigningMethod,
	expiresAt *jwt.NumericDate,
) (*JWT, error) {
	audience, err := g.getSignedJWTAudience(ctx, realm)
	if err != nil {
		return nil, err
	}
	if expiresAt == nil { // Keycloak rejects client assertions without exp
		expiresAt = jwt.NewNumericDate(time.Now().Add(signedJWTDefaultLifetime))
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: expiresAt,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    clientID,
		Subject:   clientID,
		ID:        rand.Text(),
		Audience:  jwt.ClaimStrings{audience},
	}
	assertion, err := jwx.SignClaims(claims, key, signedMethod)
	if err != nil {
		return nil, err
	}

	return g.GetToken(ctx, realm, TokenOptions{
		ClientID:            clientID,
		GrantType:           "client_credentials",
		ClientAssertionType: "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		ClientAssertion:     assertion,
	})
}

// Login performs a password login with user credentials and a client, requesting the "openid" scope.
// Input: ctx [context.Context], clientID [string], clientSecret [string], realm [string], username [string],
// password [string].
// Output: [*JWT] token, [error] if the login fails.
func (g *GoCloak) Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*JWT, error) {
	return g.GetToken(ctx, realm, TokenOptions{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "password",
		Username:     username,
		Password:     password,
		Scope:        "openid",
	})
}

// LoginOtp performs a password login with user credentials, a client and an OTP token.
// Input: ctx [context.Context], clientID [string], clientSecret [string], realm [string], username [string],
// password [string], totp [string] OTP code.
// Output: [*JWT] token, [error] if the login fails.
func (g *GoCloak) LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*JWT, error) {
	return g.GetToken(ctx, realm, TokenOptions{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "password",
		Username:     username,
		Password:     password,
		Totp:         totp,
	})
}

// Logout ends the session of the given refresh token via the OIDC logout endpoint, using client basic auth.
// Input: ctx [context.Context], clientID [string], clientSecret [string], realm [string], refreshToken [string].
// Output: [error] if the call fails.
func (g *GoCloak) Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error {
	const errMessage = "could not logout"

	return doNoResult(g.GetRequestWithBasicAuth(ctx, clientID, clientSecret).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"refresh_token": refreshToken,
		}),
		http.MethodPost, g.getOIDCURL(realm, g.Config.logoutEndpoint), errMessage)
}

// LogoutPublicClient ends the session of the given refresh token via the OIDC logout endpoint for a public client,
// authenticating with the access token.
// Input: ctx [context.Context], clientID [string], realm [string], accessToken [string], refreshToken [string].
// Output: [error] if the call fails.
func (g *GoCloak) LogoutPublicClient(ctx context.Context, clientID, realm, accessToken, refreshToken string) error {
	const errMessage = "could not logout public client"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, accessToken).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"refresh_token": refreshToken,
		}),
		http.MethodPost, g.getOIDCURL(realm, g.Config.logoutEndpoint), errMessage)
}

// LogoutAllSessions logs out all sessions of the user with the given ID.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], userID [string].
// Output: [error] if the call fails.
func (g *GoCloak) LogoutAllSessions(ctx context.Context, accessToken, realm, userID string) error {
	const errMessage = "could not logout"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodPost, g.getAdminRealmURL(realm, "users", userID, "logout"), errMessage)
}

// RevokeUserConsents revokes the consent and offline tokens the user granted to the given client.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], userID [string],
// clientID [string].
// Output: [error] if the call fails.
func (g *GoCloak) RevokeUserConsents(ctx context.Context, accessToken, realm, userID, clientID string) error {
	const errMessage = "could not revoke consents"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodDelete, g.getAdminRealmURL(realm, "users", userID, "consents", clientID), errMessage)
}

// LogoutUserSession deletes a single user session by its session ID.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], session [string] session ID.
// Output: [error] if the call fails.
func (g *GoCloak) LogoutUserSession(ctx context.Context, accessToken, realm, session string) error {
	const errMessage = "could not logout"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodDelete, g.getAdminRealmURL(realm, "sessions", session), errMessage)
}

// ExecuteActionsEmail sends the user an email with links to perform the given required actions.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [ExecuteActionsEmail]
// (UserID required).
// Output: [error] if the user ID is missing or the call fails.
func (g *GoCloak) ExecuteActionsEmail(ctx context.Context, token, realm string, params ExecuteActionsEmail) error {
	const errMessage = "could not execute actions email"

	if NilOrEmpty(params.UserID) {
		return errors.New("user id required")
	}

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return fmt.Errorf("%s: %w", errMessage, err)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(params.Actions).
		SetQueryParams(queryParams),
		http.MethodPut, g.getAdminRealmURL(realm, "users", *params.UserID, "execute-actions-email"), errMessage)
}

// SendVerifyEmail sends a verification e-mail to a user.
// Input: ctx [context.Context], token [string], realm [string], userID [string],
// params [...SendVerificationMailParams] optional, only the first is used. Output: [error].
func (g *GoCloak) SendVerifyEmail(ctx context.Context, token, realm, userID string, params ...SendVerificationMailParams) error {
	const errMessage = "failed to send verify email"

	queryParams := map[string]string{}
	if len(params) > 0 {
		if params[0].ClientID != nil {
			queryParams["client_id"] = *params[0].ClientID
		}

		if params[0].RedirectURI != nil {
			queryParams["redirect_uri"] = *params[0].RedirectURI
		}
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodPut, g.getAdminRealmURL(realm, "users", userID, "send-verify-email"), errMessage)
}

// CreateGroup creates a new top-level group.
// Input: ctx [context.Context], token [string] admin token, realm [string], group [Group].
// Output: [string] ID of the created group, [error] if the call fails.
func (g *GoCloak) CreateGroup(ctx context.Context, token, realm string, group Group) (string, error) {
	const errMessage = "could not create group"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(group),
		g.getAdminRealmURL(realm, "groups"), errMessage)
}

// CreateChildGroup creates a new child group below the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string] parent group,
// group [Group].
// Output: [string] ID of the created group, [error] if the call fails.
func (g *GoCloak) CreateChildGroup(ctx context.Context, token, realm, groupID string, group Group) (string, error) {
	const errMessage = "could not create child group"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(group),
		g.getAdminRealmURL(realm, "groups", groupID, "children"), errMessage)
}

// CreateComponent creates the given component.
// Input: ctx [context.Context], token [string] admin token, realm [string], component [Component].
// Output: [string] ID of the created component, [error] if the call fails.
func (g *GoCloak) CreateComponent(ctx context.Context, token, realm string, component Component) (string, error) {
	const errMessage = "could not create component"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(component),
		g.getAdminRealmURL(realm, "components"), errMessage)
}

// CreateClient creates the given client.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], newClient [Client].
// Output: [string] ID of the created client, [error] if the call fails.
func (g *GoCloak) CreateClient(ctx context.Context, accessToken, realm string, newClient Client) (string, error) {
	const errMessage = "could not create client"

	return doCreate(g.GetRequestWithBearerAuth(ctx, accessToken).
		SetBody(newClient),
		g.getAdminRealmURL(realm, "clients"), errMessage)
}

// CreateClientRepresentation registers a new client via the default client registration endpoint.
// Input: ctx [context.Context], token [string] initial access or bearer token, realm [string], newClient [Client].
// Output: [*Client] registered client, [error] if the call fails.
func (g *GoCloak) CreateClientRepresentation(ctx context.Context, token, realm string, newClient Client) (*Client, error) {
	const errMessage = "could not create client representation"

	return doRequestPtr[Client](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(newClient),
		http.MethodPost, g.getRealmURL(realm, "clients-registrations", "default"), errMessage)
}

// CreateClientRole creates a new role for a client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], role [Role].
// Output: [string] last segment of the Location header (the role name), [error] if the call fails.
func (g *GoCloak) CreateClientRole(ctx context.Context, token, realm, idOfClient string, role Role) (string, error) {
	const errMessage = "could not create client role"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(role),
		g.getAdminRealmURL(realm, "clients", idOfClient, "roles"), errMessage)
}

// CreateClientScope creates a new client scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], scope [ClientScope].
// Output: [string] ID of the created client scope, [error] if the call fails.
func (g *GoCloak) CreateClientScope(ctx context.Context, token, realm string, scope ClientScope) (string, error) {
	const errMessage = "could not create client scope"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(scope),
		g.getAdminRealmURL(realm, "client-scopes"), errMessage)
}

// CreateClientScopeProtocolMapper creates a new protocol mapper under the given client scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], scopeID [string],
// protocolMapper [ProtocolMappers].
// Output: [string] ID of the created protocol mapper, [error] if the call fails.
func (g *GoCloak) CreateClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID string, protocolMapper ProtocolMappers) (string, error) {
	const errMessage = "could not create client scope protocol mapper"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(protocolMapper),
		g.getAdminRealmURL(realm, "client-scopes", scopeID, "protocol-mappers", "models"), errMessage)
}

// UpdateGroup updates the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], updatedGroup [Group] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateGroup(ctx context.Context, token, realm string, updatedGroup Group) error {
	const errMessage = "could not update group"

	if NilOrEmpty(updatedGroup.ID) {
		return fmt.Errorf("%s: ID of a group required", errMessage)
	}
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(updatedGroup),
		http.MethodPut, g.getAdminRealmURL(realm, "groups", PString(updatedGroup.ID)), errMessage)
}

// UpdateGroupManagementPermissions updates the management permissions of the given group.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], idOfGroup [string],
// managementPermissions [ManagementPermissionRepresentation].
// Output: [*ManagementPermissionRepresentation] updated permissions, [error] if the call fails.
func (g *GoCloak) UpdateGroupManagementPermissions(ctx context.Context, accessToken, realm string, idOfGroup string, managementPermissions ManagementPermissionRepresentation) (*ManagementPermissionRepresentation, error) {
	const errMessage = "could not update group management permissions"

	return doRequestPtr[ManagementPermissionRepresentation](g.GetRequestWithBearerAuth(ctx, accessToken).
		SetBody(managementPermissions),
		http.MethodPut, g.getAdminRealmURL(realm, "groups", idOfGroup, "management", "permissions"), errMessage)
}

// UpdateClient updates the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], updatedClient [Client] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateClient(ctx context.Context, token, realm string, updatedClient Client) error {
	const errMessage = "could not update client"

	if NilOrEmpty(updatedClient.ID) {
		return fmt.Errorf("%s: ID of a client required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(updatedClient),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", PString(updatedClient.ID)), errMessage)
}

// UpdateClientRepresentation updates the given client via the default client registration endpoint.
// Input: ctx [context.Context], accessToken [string] registration token, realm [string],
// updatedClient [Client] (ID required; addressed by ClientID).
// Output: [*Client] updated client, [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateClientRepresentation(ctx context.Context, accessToken, realm string, updatedClient Client) (*Client, error) {
	const errMessage = "could not update client representation"

	if NilOrEmpty(updatedClient.ID) {
		return nil, fmt.Errorf("%s: ID of a client required", errMessage)
	}

	return doRequestPtr[Client](g.GetRequestWithBearerAuth(ctx, accessToken).
		SetBody(updatedClient),
		http.MethodPut, g.getRealmURL(realm, "clients-registrations", "default", PString(updatedClient.ClientID)), errMessage)
}

// UpdateClientManagementPermissions updates the management permissions of the given client.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], idOfClient [string],
// managementPermissions [ManagementPermissionRepresentation].
// Output: [*ManagementPermissionRepresentation] updated permissions, [error] if the call fails.
func (g *GoCloak) UpdateClientManagementPermissions(ctx context.Context, accessToken, realm string, idOfClient string, managementPermissions ManagementPermissionRepresentation) (*ManagementPermissionRepresentation, error) {
	const errMessage = "could not update client management permissions"

	return doRequestPtr[ManagementPermissionRepresentation](g.GetRequestWithBearerAuth(ctx, accessToken).
		SetBody(managementPermissions),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "management", "permissions"), errMessage)
}

// UpdateRole updates the given client role, addressed by its name.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], role [Role].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateRole(ctx context.Context, token, realm, idOfClient string, role Role) error {
	const errMessage = "could not update role"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(role),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "roles", PString(role.Name)), errMessage)
}

// UpdateClientScope updates the given client scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], scope [ClientScope] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateClientScope(ctx context.Context, token, realm string, scope ClientScope) error {
	const errMessage = "could not update client scope"

	if NilOrEmpty(scope.ID) {
		return fmt.Errorf("%s: ID of a client scope required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(scope),
		http.MethodPut, g.getAdminRealmURL(realm, "client-scopes", PString(scope.ID)), errMessage)
}

// UpdateClientScopeProtocolMapper updates the given protocol mapper of a client scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], scopeID [string],
// protocolMapper [ProtocolMappers] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID string, protocolMapper ProtocolMappers) error {
	const errMessage = "could not update client scope"

	if NilOrEmpty(protocolMapper.ID) {
		return fmt.Errorf("%s: ID of a protocol mapper required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(protocolMapper),
		http.MethodPut, g.getAdminRealmURL(realm, "client-scopes", scopeID, "protocol-mappers", "models", PString(protocolMapper.ID)), errMessage)
}

// DeleteGroup deletes the group with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteGroup(ctx context.Context, token, realm, groupID string) error {
	const errMessage = "could not delete group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "groups", groupID), errMessage)
}

// DeleteClient deletes the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClient(ctx context.Context, token, realm, idOfClient string) error {
	const errMessage = "could not delete client"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient), errMessage)
}

// DeleteComponent deletes the component with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], componentID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteComponent(ctx context.Context, token, realm, componentID string) error {
	const errMessage = "could not delete component"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "components", componentID), errMessage)
}

// DeleteClientRepresentation deletes the given client via the default client registration endpoint.
// Input: ctx [context.Context], accessToken [string] registration token, realm [string], clientID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientRepresentation(ctx context.Context, accessToken, realm, clientID string) error {
	const errMessage = "could not delete client representation"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodDelete, g.getRealmURL(realm, "clients-registrations", "default", clientID), errMessage)
}

// DeleteClientRole deletes the given client role.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], roleName [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientRole(ctx context.Context, token, realm, idOfClient, roleName string) error {
	const errMessage = "could not delete client role"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "roles", roleName), errMessage)
}

// DeleteClientScope deletes the client scope with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], scopeID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientScope(ctx context.Context, token, realm, scopeID string) error {
	const errMessage = "could not delete client scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "client-scopes", scopeID), errMessage)
}

// DeleteClientScopeProtocolMapper deletes the given protocol mapper from the client scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], scopeID [string],
// protocolMapperID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID, protocolMapperID string) error {
	const errMessage = "could not delete client scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "client-scopes", scopeID, "protocol-mappers", "models", protocolMapperID), errMessage)
}

// GetClient fetches the client with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [*Client] client, [error] if the call fails.
func (g *GoCloak) GetClient(ctx context.Context, token, realm, idOfClient string) (*Client, error) {
	const errMessage = "could not get client"

	return doRequestPtr[Client](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient), errMessage)
}

// GetClientRepresentation fetches a client via the default client registration endpoint.
// Input: ctx [context.Context], accessToken [string] registration token, realm [string], clientID [string].
// Output: [*Client] client, [error] if the call fails.
func (g *GoCloak) GetClientRepresentation(ctx context.Context, accessToken, realm, clientID string) (*Client, error) {
	const errMessage = "could not get client representation"

	return doRequestPtr[Client](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getRealmURL(realm, "clients-registrations", "default", clientID), errMessage)
}

// GetAdapterConfiguration fetches the installation adapter configuration of a client.
// Input: ctx [context.Context], accessToken [string] registration token, realm [string], clientID [string].
// Output: [*AdapterConfiguration] adapter configuration, [error] if the call fails.
func (g *GoCloak) GetAdapterConfiguration(ctx context.Context, accessToken, realm, clientID string) (*AdapterConfiguration, error) {
	const errMessage = "could not get adapter configuration"

	return doRequestPtr[AdapterConfiguration](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getRealmURL(realm, "clients-registrations", "install", clientID), errMessage)
}

// GetClientsDefaultScopes returns the default client scopes of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [[]*ClientScope] default scopes, [error] if the call fails.
func (g *GoCloak) GetClientsDefaultScopes(ctx context.Context, token, realm, idOfClient string) ([]*ClientScope, error) {
	const errMessage = "could not get clients default scopes"

	return doRequest[[]*ClientScope](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "default-client-scopes"), errMessage)
}

// AddDefaultScopeToClient adds a client scope to the default scopes of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], scopeID [string].
// Output: [error] if the call fails.
func (g *GoCloak) AddDefaultScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	const errMessage = "could not add default scope to client"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "default-client-scopes", scopeID), errMessage)
}

// RemoveDefaultScopeFromClient removes a client scope from the default scopes of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], scopeID [string].
// Output: [error] if the call fails.
func (g *GoCloak) RemoveDefaultScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	const errMessage = "could not remove default scope from client"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "default-client-scopes", scopeID), errMessage)
}

// GetClientsOptionalScopes returns the optional client scopes of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [[]*ClientScope] optional scopes, [error] if the call fails.
func (g *GoCloak) GetClientsOptionalScopes(ctx context.Context, token, realm, idOfClient string) ([]*ClientScope, error) {
	const errMessage = "could not get clients optional scopes"

	return doRequest[[]*ClientScope](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "optional-client-scopes"), errMessage)
}

// AddOptionalScopeToClient adds a client scope to the optional scopes of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], scopeID [string].
// Output: [error] if the call fails.
func (g *GoCloak) AddOptionalScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	const errMessage = "could not add optional scope to client"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "optional-client-scopes", scopeID), errMessage)
}

// RemoveOptionalScopeFromClient removes a client scope from the optional scopes of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], scopeID [string].
// Output: [error] if the call fails.
func (g *GoCloak) RemoveOptionalScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	const errMessage = "could not remove optional scope from client"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "optional-client-scopes", scopeID), errMessage)
}

// GetDefaultOptionalClientScopes returns the realm's default optional client scopes.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*ClientScope] default optional scopes, [error] if the call fails.
func (g *GoCloak) GetDefaultOptionalClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error) {
	const errMessage = "could not get default optional client scopes"

	return doRequest[[]*ClientScope](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "default-optional-client-scopes"), errMessage)
}

// GetDefaultDefaultClientScopes returns the realm-wide client scopes assigned to new clients as default scopes.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*ClientScope] default scopes, [error] if the call fails.
func (g *GoCloak) GetDefaultDefaultClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error) {
	const errMessage = "could not get default client scopes"

	return doRequest[[]*ClientScope](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "default-default-client-scopes"), errMessage)
}

// GetClientScope fetches the client scope with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], scopeID [string].
// Output: [*ClientScope] client scope, [error] if the call fails.
func (g *GoCloak) GetClientScope(ctx context.Context, token, realm, scopeID string) (*ClientScope, error) {
	const errMessage = "could not get client scope"

	return doRequestPtr[ClientScope](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes", scopeID), errMessage)
}

// GetClientScopes returns all client scopes of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*ClientScope] client scopes, [error] if the call fails.
func (g *GoCloak) GetClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error) {
	const errMessage = "could not get client scopes"

	return doRequest[[]*ClientScope](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes"), errMessage)
}

// GetClientScopeProtocolMappers returns all protocol mappers of a client scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], scopeID [string].
// Output: [[]*ProtocolMappers] protocol mappers, [error] if the call fails.
func (g *GoCloak) GetClientScopeProtocolMappers(ctx context.Context, token, realm, scopeID string) ([]*ProtocolMappers, error) {
	const errMessage = "could not get client scope protocol mappers"

	return doRequest[[]*ProtocolMappers](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes", scopeID, "protocol-mappers", "models"), errMessage)
}

// GetClientScopeProtocolMapper fetches one protocol mapper of a client scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], scopeID [string],
// protocolMapperID [string].
// Output: [*ProtocolMappers] protocol mapper, [error] if the call fails.
func (g *GoCloak) GetClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID, protocolMapperID string) (*ProtocolMappers, error) {
	const errMessage = "could not get client scope protocol mappers"

	return doRequestPtr[ProtocolMappers](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes", scopeID, "protocol-mappers", "models", protocolMapperID), errMessage)
}

// GetClientScopeMappings returns all realm and client scope mappings of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [*MappingsRepresentation] scope mappings, [error] if the call fails.
func (g *GoCloak) GetClientScopeMappings(ctx context.Context, token, realm, idOfClient string) (*MappingsRepresentation, error) {
	const errMessage = "could not get all scope mappings for the client"

	return doRequestPtr[MappingsRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings"), errMessage)
}

// GetRealmRoleGroups returns the groups that have the given realm role.
// Input: ctx [context.Context], token [string] admin token, roleName [string], realm [string].
// Output: [[]*Group] groups, [error] if the call fails.
func (g *GoCloak) GetRealmRoleGroups(ctx context.Context, token, roleName, realm string) ([]*Group, error) {
	const errMessage = "could not get groups by realm roleName"

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles", roleName, "groups"), errMessage)
}

// GetClientScopeMappingsRealmRoles returns the realm-level roles in the client's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [[]*Role] realm roles, [error] if the call fails.
func (g *GoCloak) GetClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string) ([]*Role, error) {
	const errMessage = "could not get realm-level roles with the client’s scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "realm"), errMessage)
}

// GetClientScopeMappingsRealmRolesAvailable returns the realm-level roles that can still be added to the client's
// scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [[]*Role] available realm roles, [error] if the call fails.
func (g *GoCloak) GetClientScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClient string) ([]*Role, error) {
	const errMessage = "could not get available realm-level roles with the client’s scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "realm", "available"), errMessage)
}

// CreateClientScopeMappingsRealmRoles adds realm-level roles to the client's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) CreateClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []Role) error {
	const errMessage = "could not create realm-level roles to the client’s scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "realm"), errMessage)
}

// DeleteClientScopeMappingsRealmRoles removes realm-level roles from the client's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []Role) error {
	const errMessage = "could not delete realm-level roles from the client’s scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "realm"), errMessage)
}

// GetClientScopeMappingsClientRoles returns the roles of the selected client in the client's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// idOfSelectedClient [string] client owning the roles.
// Output: [[]*Role] client roles, [error] if the call fails.
func (g *GoCloak) GetClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*Role, error) {
	const errMessage = "could not get roles associated with a client’s scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "clients", idOfSelectedClient), errMessage)
}

// GetClientScopeMappingsClientRolesAvailable returns the roles of the selected client that can still be added to
// the client's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// idOfSelectedClient [string] client owning the roles.
// Output: [[]*Role] available client roles, [error] if the call fails.
func (g *GoCloak) GetClientScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*Role, error) {
	const errMessage = "could not get available roles associated with a client’s scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "clients", idOfSelectedClient, "available"), errMessage)
}

// CreateClientScopeMappingsClientRoles adds roles of the selected client to the client's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// idOfSelectedClient [string] client owning the roles, roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) CreateClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []Role) error {
	const errMessage = "could not create client-level roles from the client’s scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "clients", idOfSelectedClient), errMessage)
}

// DeleteClientScopeMappingsClientRoles removes roles of the selected client from the client's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// idOfSelectedClient [string] client owning the roles, roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []Role) error {
	const errMessage = "could not delete client-level roles from the client’s scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "scope-mappings", "clients", idOfSelectedClient), errMessage)
}

// GetClientSecret fetches the secret of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [*CredentialRepresentation] client secret, [error] if the call fails.
func (g *GoCloak) GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*CredentialRepresentation, error) {
	const errMessage = "could not get client secret"

	return doRequestPtr[CredentialRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "client-secret"), errMessage)
}

// GetClientServiceAccount retrieves the service account [User] of the client identified by idOfClient [string].
// It uses the service-account-user endpoint (one call, needs view-clients only). Keycloak 27 removed that
// endpoint; on 404 it reads [Client.ServiceAccountUserID] and loads that user, which also needs view-users.
// Input: ctx [context.Context], token [string], realm [string], idOfClient [string].
// Output: [*User] service account user, [error] if the client has none or a request fails.
func (g *GoCloak) GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*User, error) {
	const errMessage = "could not get client service account"

	user, err := doRequestPtr[User](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "service-account-user"), errMessage)
	var apiErr *APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Code != http.StatusNotFound {
		return user, err
	}

	// shortcut: Keycloak 27+ costs a 404 plus two calls, switch the order once 26 is unsupported.
	client, err := g.GetClient(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	userID := PString(client.ServiceAccountUserID)
	if userID == "" {
		return nil, fmt.Errorf("%s: client has no service account", errMessage)
	}
	return g.GetUserByID(ctx, token, realm, userID)
}

// RegenerateClientSecret generates a new secret for the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [*CredentialRepresentation] new client secret, [error] if the call fails.
func (g *GoCloak) RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*CredentialRepresentation, error) {
	const errMessage = "could not regenerate client secret"

	return doRequestPtr[CredentialRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodPost, g.getAdminRealmURL(realm, "clients", idOfClient, "client-secret"), errMessage)
}

// GetClientOfflineSessions returns the offline sessions of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// params [...GetClientUserSessionsParams] optional, only the first is used.
// Output: [[]*UserSessionRepresentation] offline sessions, [error] if the call fails.
func (g *GoCloak) GetClientOfflineSessions(ctx context.Context, token, realm, idOfClient string, params ...GetClientUserSessionsParams) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get client offline sessions"

	queryParams := map[string]string{}
	if len(params) > 0 {
		var err error

		queryParams, err = GetQueryParams(params[0])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, err)
		}
	}

	return doRequest[[]*UserSessionRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "offline-sessions"), errMessage)
}

// GetClientUserSessions returns the user sessions of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// params [...GetClientUserSessionsParams] optional, only the first is used.
// Output: [[]*UserSessionRepresentation] user sessions, [error] if the call fails.
func (g *GoCloak) GetClientUserSessions(ctx context.Context, token, realm, idOfClient string, params ...GetClientUserSessionsParams) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get client user sessions"

	queryParams := map[string]string{}
	if len(params) > 0 {
		var err error

		queryParams, err = GetQueryParams(params[0])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, err)
		}
	}

	return doRequest[[]*UserSessionRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "user-sessions"), errMessage)
}

// CreateClientProtocolMapper creates a protocol mapper on the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// mapper [ProtocolMapperRepresentation].
// Output: [string] ID of the created protocol mapper, [error] if the call fails.
func (g *GoCloak) CreateClientProtocolMapper(ctx context.Context, token, realm, idOfClient string, mapper ProtocolMapperRepresentation) (string, error) {
	const errMessage = "could not create client protocol mapper"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(mapper),
		g.getAdminRealmURL(realm, "clients", idOfClient, "protocol-mappers", "models"), errMessage)
}

// UpdateClientProtocolMapper updates a protocol mapper of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], mapperID [string],
// mapper [ProtocolMapperRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string, mapper ProtocolMapperRepresentation) error {
	const errMessage = "could not update client protocol mapper"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(mapper),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "protocol-mappers", "models", mapperID), errMessage)
}

// DeleteClientProtocolMapper deletes a protocol mapper of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], mapperID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string) error {
	const errMessage = "could not delete client protocol mapper"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "protocol-mappers", "models", mapperID), errMessage)
}

// GetKeyStoreConfig returns the keys of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [*KeyStoreConfig] realm keys, [error] if the call fails.
func (g *GoCloak) GetKeyStoreConfig(ctx context.Context, token, realm string) (*KeyStoreConfig, error) {
	const errMessage = "could not get key store config"

	return doRequestPtr[KeyStoreConfig](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "keys"), errMessage)
}

// GetComponents returns all components of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*Component] components, [error] if the call fails.
func (g *GoCloak) GetComponents(ctx context.Context, token, realm string) ([]*Component, error) {
	const errMessage = "could not get components"

	return doRequest[[]*Component](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "components"), errMessage)
}

// GetComponentsWithParams returns the components of the realm filtered by the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetComponentsParams].
// Output: [[]*Component] components, [error] if the call fails.
func (g *GoCloak) GetComponentsWithParams(ctx context.Context, token, realm string, params GetComponentsParams) ([]*Component, error) {
	const errMessage = "could not get components"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	return doRequest[[]*Component](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "components"), errMessage)
}

// GetComponent fetches exactly one component by its ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], componentID [string].
// Output: [*Component] component, [error] if the call fails.
func (g *GoCloak) GetComponent(ctx context.Context, token, realm string, componentID string) (*Component, error) {
	const errMessage = "could not get components"

	return doRequestPtr[Component](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "components", componentID), errMessage)
}

// UpdateComponent updates the given component.
// Input: ctx [context.Context], token [string] admin token, realm [string], component [Component] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateComponent(ctx context.Context, token, realm string, component Component) error {
	const errMessage = "could not update component"

	if NilOrEmpty(component.ID) {
		return fmt.Errorf("%s: ID of a component required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(component),
		http.MethodPut, g.getAdminRealmURL(realm, "components", PString(component.ID)), errMessage)
}

// GetDefaultGroups returns the realm's default groups.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*Group] default groups, [error] if the call fails.
func (g *GoCloak) GetDefaultGroups(ctx context.Context, token, realm string) ([]*Group, error) {
	const errMessage = "could not get default groups"

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "default-groups"), errMessage)
}

// AddDefaultGroup adds a group to the realm's default groups.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [error] if the call fails.
func (g *GoCloak) AddDefaultGroup(ctx context.Context, token, realm, groupID string) error {
	const errMessage = "could not add default group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPut, g.getAdminRealmURL(realm, "default-groups", groupID), errMessage)
}

// RemoveDefaultGroup removes a group from the realm's default groups.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [error] if the call fails.
func (g *GoCloak) RemoveDefaultGroup(ctx context.Context, token, realm, groupID string) error {
	const errMessage = "could not remove default group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "default-groups", groupID), errMessage)
}

// getRoleMappings fetches the realm and client role mappings of a user or group.
// Input: ctx [context.Context], token [string] admin token, realm [string], path [string] "users" or "groups",
// objectID [string] user or group ID.
// Output: [*MappingsRepresentation] role mappings, [error] if the call fails.
func (g *GoCloak) getRoleMappings(ctx context.Context, token, realm, path, objectID string) (*MappingsRepresentation, error) {
	const errMessage = "could not get role mappings"

	return doRequestPtr[MappingsRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, path, objectID, "role-mappings"), errMessage)
}

// GetRoleMappingByGroupID returns the realm and client role mappings of the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [*MappingsRepresentation] role mappings, [error] if the call fails.
func (g *GoCloak) GetRoleMappingByGroupID(ctx context.Context, token, realm, groupID string) (*MappingsRepresentation, error) {
	return g.getRoleMappings(ctx, token, realm, "groups", groupID)
}

// GetRoleMappingByUserID returns the realm and client role mappings of the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [*MappingsRepresentation] role mappings, [error] if the call fails.
func (g *GoCloak) GetRoleMappingByUserID(ctx context.Context, token, realm, userID string) (*MappingsRepresentation, error) {
	return g.getRoleMappings(ctx, token, realm, "users", userID)
}

// GetGroup fetches the group with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [*Group] group, [error] if the call fails.
func (g *GoCloak) GetGroup(ctx context.Context, token, realm, groupID string) (*Group, error) {
	const errMessage = "could not get group"

	return doRequestPtr[Group](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID), errMessage)
}

// GetChildGroups returns the child groups of the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string],
// params [GetChildGroupsParams].
// Output: [[]*Group] child groups, [error] if the call fails.
func (g *GoCloak) GetChildGroups(ctx context.Context, token, realm, groupID string, params GetChildGroupsParams) ([]*Group, error) {
	const errMessage = "could not get child groups"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "children"), errMessage)
}

// GetGroupByPath fetches a group by its path (e.g. "/parent/child"); each group name is escaped.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupPath [string].
// Output: [*Group] group, [error] if the call fails.
func (g *GoCloak) GetGroupByPath(ctx context.Context, token, realm, groupPath string) (*Group, error) {
	const errMessage = "could not get group"

	return doRequestPtr[Group](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet,
		g.getAdminRealmURL(realm, append([]string{"group-by-path"}, groupPathSegments(groupPath)...)...), errMessage)
}

// groupPathSegments splits a group path like "/parent/child" into its group names, so each name is escaped as its
// own segment while "/" stays the separator Keycloak expects.
// Input: groupPath [string]. Output: [[]string] group names.
func groupPathSegments(groupPath string) []string {
	return strings.Split(strings.TrimLeft(groupPath, urlSeparator), urlSeparator)
}

// GetGroups returns the groups of the realm filtered by the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetGroupsParams].
// Output: [[]*Group] groups, [error] if the call fails.
func (g *GoCloak) GetGroups(ctx context.Context, token, realm string, params GetGroupsParams) ([]*Group, error) {
	const errMessage = "could not get groups"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "groups"), errMessage)
}

// GetGroupManagementPermissions returns whether group Authorization permissions have been initialized or not and a
// reference to the managed permissions.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfGroup [string].
// Output: [*ManagementPermissionRepresentation] permissions, [error] if the call fails.
func (g *GoCloak) GetGroupManagementPermissions(ctx context.Context, token, realm string, idOfGroup string) (*ManagementPermissionRepresentation, error) {
	const errMessage = "could not get management permissions"

	return doRequestPtr[ManagementPermissionRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", idOfGroup, "management", "permissions"), errMessage)
}

// GetGroupsByRole returns the groups that have the given realm role.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string].
// Output: [[]*Group] groups, [error] if the call fails.
func (g *GoCloak) GetGroupsByRole(ctx context.Context, token, realm string, roleName string) ([]*Group, error) {
	const errMessage = "could not get groups"

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles", roleName, "groups"), errMessage)
}

// GetGroupsByClientRole returns the groups that have the given client role.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string],
// clientID [string] ID of the client (not its clientId).
// Output: [[]*Group] groups, [error] if the call fails.
func (g *GoCloak) GetGroupsByClientRole(ctx context.Context, token, realm string, roleName string, clientID string) ([]*Group, error) {
	const errMessage = "could not get groups"

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", clientID, "roles", roleName, "groups"), errMessage)
}

// GetGroupsCount returns the number of groups in the realm matching the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetGroupsParams].
// Output: [int] group count, [error] if the call fails.
func (g *GoCloak) GetGroupsCount(ctx context.Context, token, realm string, params GetGroupsParams) (int, error) {
	const errMessage = "could not get groups count"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", errMessage, err)
	}

	result, err := doRequest[GroupsCount](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "groups", "count"), errMessage)

	return result.Count, err
}

// GetGroupMembers returns the users that are members of the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string],
// params [GetGroupsParams].
// Output: [[]*User] members, [error] if the call fails.
func (g *GoCloak) GetGroupMembers(ctx context.Context, token, realm, groupID string, params GetGroupsParams) ([]*User, error) {
	const errMessage = "could not get group members"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*User](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "members"), errMessage)
}

// GetClientRoles returns the roles of the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// params [GetRoleParams].
// Output: [[]*Role] client roles, [error] if the call fails.
func (g *GoCloak) GetClientRoles(ctx context.Context, token, realm, idOfClient string, params GetRoleParams) ([]*Role, error) {
	const errMessage = "could not get client roles"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "roles"), errMessage)
}

// GetClientRoleByID fetches a role by its ID via the roles-by-id endpoint.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleID [string].
// Output: [*Role] role, [error] if the call fails.
func (g *GoCloak) GetClientRoleByID(ctx context.Context, token, realm, roleID string) (*Role, error) {
	const errMessage = "could not get client role"

	return doRequestPtr[Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles-by-id", roleID), errMessage)
}

// GetClientRolesByUserID returns the roles of the given client that are directly assigned to the user.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], userID [string].
// Output: [[]*Role] client roles, [error] if the call fails.
func (g *GoCloak) GetClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*Role, error) {
	const errMessage = "could not get client roles by user id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", idOfClient), errMessage)
}

// GetClientRolesByGroupID returns the roles of the given client that are directly assigned to the group.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], groupID [string].
// Output: [[]*Role] client roles, [error] if the call fails.
func (g *GoCloak) GetClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*Role, error) {
	const errMessage = "could not get client roles by group id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", idOfClient), errMessage)
}

// GetCompositeClientRolesByRoleID returns the composite roles of the given client contained in the given role.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], roleID [string].
// Output: [[]*Role] composite client roles, [error] if the call fails.
func (g *GoCloak) GetCompositeClientRolesByRoleID(ctx context.Context, token, realm, idOfClient, roleID string) ([]*Role, error) {
	const errMessage = "could not get composite client roles by role id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles-by-id", roleID, "composites", "clients", idOfClient), errMessage)
}

// GetCompositeClientRolesByUserID returns the effective roles of the given client for the user, including
// composite roles.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], userID [string].
// Output: [[]*Role] effective client roles, [error] if the call fails.
func (g *GoCloak) GetCompositeClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*Role, error) {
	const errMessage = "could not get composite client roles by user id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", idOfClient, "composite"), errMessage)
}

// GetAvailableClientRolesByUserID returns the roles of the given client that can still be assigned to the user.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], userID [string].
// Output: [[]*Role] available client roles, [error] if the call fails.
func (g *GoCloak) GetAvailableClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*Role, error) {
	const errMessage = "could not get available client roles by user id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", idOfClient, "available"), errMessage)
}

// GetAvailableClientRolesByGroupID returns the roles of the given client that can still be assigned to the group.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], groupID [string].
// Output: [[]*Role] available client roles, [error] if the call fails.
func (g *GoCloak) GetAvailableClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*Role, error) {
	const errMessage = "could not get available client roles by user id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", idOfClient, "available"), errMessage)
}

// GetCompositeClientRolesByGroupID returns the effective roles of the given client for the group, including
// composite roles.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], groupID [string].
// Output: [[]*Role] effective client roles, [error] if the call fails.
func (g *GoCloak) GetCompositeClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*Role, error) {
	const errMessage = "could not get composite client roles by group id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", idOfClient, "composite"), errMessage)
}

// GetClientRole fetches a role of the given client by its name.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], roleName [string].
// Output: [*Role] client role, [error] if the call fails.
func (g *GoCloak) GetClientRole(ctx context.Context, token, realm, idOfClient, roleName string) (*Role, error) {
	const errMessage = "could not get client role"

	return doRequestPtr[Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "roles", roleName), errMessage)
}

// GetClients returns the clients of the realm filtered by the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetClientsParams].
// Output: [[]*Client] clients, [error] if the call fails.
func (g *GoCloak) GetClients(ctx context.Context, token, realm string, params GetClientsParams) ([]*Client, error) {
	const errMessage = "could not get clients"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	return doRequest[[]*Client](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients"), errMessage)
}

// GetClientManagementPermissions returns whether client Authorization permissions have been initialized or not and a
// reference to the managed permissions.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [*ManagementPermissionRepresentation] permissions, [error] if the call fails.
func (g *GoCloak) GetClientManagementPermissions(ctx context.Context, token, realm string, idOfClient string) (*ManagementPermissionRepresentation, error) {
	const errMessage = "could not get management permissions"

	return doRequestPtr[ManagementPermissionRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "management", "permissions"), errMessage)
}

// UserAttributeContains reports whether the given attribute contains the given value.
// Input: attributes [map[string][]string], attribute [string] attribute name, value [string].
// Output: [bool] true if the value is present.
func UserAttributeContains(attributes map[string][]string, attribute, value string) bool {
	return slices.Contains(attributes[attribute], value)
}

// -----------
// Realm Roles
// -----------

// CreateRealmRole creates a role in the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string], role [Role].
// Output: [string] last segment of the Location header (the role name), [error] if the call fails.
func (g *GoCloak) CreateRealmRole(ctx context.Context, token string, realm string, role Role) (string, error) {
	const errMessage = "could not create realm role"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(role),
		g.getAdminRealmURL(realm, "roles"), errMessage)
}

// GetRealmRole fetches a realm role by its name.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string].
// Output: [*Role] realm role, [error] if the call fails.
func (g *GoCloak) GetRealmRole(ctx context.Context, token, realm, roleName string) (*Role, error) {
	const errMessage = "could not get realm role"

	return doRequestPtr[Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles", roleName), errMessage)
}

// GetRealmRoleByID fetches a role by its ID via the roles-by-id endpoint.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleID [string].
// Output: [*Role] role, [error] if the call fails.
func (g *GoCloak) GetRealmRoleByID(ctx context.Context, token, realm, roleID string) (*Role, error) {
	const errMessage = "could not get realm role"

	return doRequestPtr[Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles-by-id", roleID), errMessage)
}

// GetRealmRoles returns the roles of the realm filtered by the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetRoleParams].
// Output: [[]*Role] realm roles, [error] if the call fails.
func (g *GoCloak) GetRealmRoles(ctx context.Context, token, realm string, params GetRoleParams) ([]*Role, error) {
	const errMessage = "could not get realm roles"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "roles"), errMessage)
}

// GetRealmRolesByUserID returns the realm roles directly assigned to the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [[]*Role] realm roles, [error] if the call fails.
func (g *GoCloak) GetRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*Role, error) {
	const errMessage = "could not get realm roles by user id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"), errMessage)
}

// GetRealmRolesByGroupID returns the realm roles directly assigned to the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [[]*Role] realm roles, [error] if the call fails.
func (g *GoCloak) GetRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*Role, error) {
	const errMessage = "could not get realm roles by group id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm"), errMessage)
}

// UpdateRealmRole updates a realm role addressed by its name.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string], role [Role].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateRealmRole(ctx context.Context, token, realm, roleName string, role Role) error {
	const errMessage = "could not update realm role"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(role),
		http.MethodPut, g.getAdminRealmURL(realm, "roles", roleName), errMessage)
}

// UpdateRealmRoleByID updates a role addressed by its ID via the roles-by-id endpoint.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleID [string], role [Role].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateRealmRoleByID(ctx context.Context, token, realm, roleID string, role Role) error {
	const errMessage = "could not update realm role"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(role),
		http.MethodPut, g.getAdminRealmURL(realm, "roles-by-id", roleID), errMessage)
}

// DeleteRealmRole deletes a realm role by its name.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteRealmRole(ctx context.Context, token, realm, roleName string) error {
	const errMessage = "could not delete realm role"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "roles", roleName), errMessage)
}

// AddRealmRoleToUser adds realm-level role mappings to the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) AddRealmRoleToUser(ctx context.Context, token, realm, userID string, roles []Role) error {
	const errMessage = "could not add realm role to user"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"), errMessage)
}

// DeleteRealmRoleFromUser removes realm-level role mappings from the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteRealmRoleFromUser(ctx context.Context, token, realm, userID string, roles []Role) error {
	const errMessage = "could not delete realm role from user"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"), errMessage)
}

// AddRealmRoleToGroup adds realm-level role mappings to the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) AddRealmRoleToGroup(ctx context.Context, token, realm, groupID string, roles []Role) error {
	const errMessage = "could not add realm role to group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm"), errMessage)
}

// DeleteRealmRoleFromGroup removes realm-level role mappings from the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteRealmRoleFromGroup(ctx context.Context, token, realm, groupID string, roles []Role) error {
	const errMessage = "could not delete realm role from group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm"), errMessage)
}

// AddRealmRoleComposite adds the given roles as composites of a realm role.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string] composite role,
// roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) AddRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []Role) error {
	const errMessage = "could not add realm role composite"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "roles", roleName, "composites"), errMessage)
}

// DeleteRealmRoleComposite removes the given roles from the composites of a realm role.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string] composite role,
// roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []Role) error {
	const errMessage = "could not delete realm role composite"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "roles", roleName, "composites"), errMessage)
}

// GetCompositeRealmRoles returns the composite roles of the given realm role.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string].
// Output: [[]*Role] composite roles, [error] if the call fails.
func (g *GoCloak) GetCompositeRealmRoles(ctx context.Context, token, realm, roleName string) ([]*Role, error) {
	const errMessage = "could not get composite realm roles by role"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles", roleName, "composites"), errMessage)
}

// GetCompositeRolesByRoleID returns all composite roles (realm and client) of the role with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleID [string].
// Output: [[]*Role] composite roles, [error] if the call fails.
func (g *GoCloak) GetCompositeRolesByRoleID(ctx context.Context, token, realm, roleID string) ([]*Role, error) {
	const errMessage = "could not get composite client roles by role id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles-by-id", roleID, "composites"), errMessage)
}

// GetCompositeRealmRolesByRoleID returns the realm composite roles of the role with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleID [string].
// Output: [[]*Role] composite realm roles, [error] if the call fails.
func (g *GoCloak) GetCompositeRealmRolesByRoleID(ctx context.Context, token, realm, roleID string) ([]*Role, error) {
	const errMessage = "could not get composite realm roles by role id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "roles-by-id", roleID, "composites", "realm"), errMessage)
}

// GetCompositeRealmRolesByUserID returns the effective realm roles of the user, including composite roles.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [[]*Role] effective realm roles, [error] if the call fails.
func (g *GoCloak) GetCompositeRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*Role, error) {
	const errMessage = "could not get composite realm roles by user id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm", "composite"), errMessage)
}

// GetCompositeRealmRolesByGroupID returns the effective realm roles of the group, including composite roles.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [[]*Role] effective realm roles, [error] if the call fails.
func (g *GoCloak) GetCompositeRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*Role, error) {
	const errMessage = "could not get composite realm roles by group id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm", "composite"), errMessage)
}

// GetAvailableRealmRolesByUserID returns the realm roles that can still be assigned to the user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [[]*Role] available realm roles, [error] if the call fails.
func (g *GoCloak) GetAvailableRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*Role, error) {
	const errMessage = "could not get available realm roles by user id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm", "available"), errMessage)
}

// GetAvailableRealmRolesByGroupID returns the realm roles that can still be assigned to the group.
// Input: ctx [context.Context], token [string] admin token, realm [string], groupID [string].
// Output: [[]*Role] available realm roles, [error] if the call fails.
func (g *GoCloak) GetAvailableRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*Role, error) {
	const errMessage = "could not get available realm roles by group id"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm", "available"), errMessage)
}

// -----
// Realm
// -----

// GetRealm returns the top-level representation of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [*RealmRepresentation] realm, [error] if realm is empty or the call fails.
func (g *GoCloak) GetRealm(ctx context.Context, token, realm string) (*RealmRepresentation, error) {
	const errMessage = "could not get realm"
	if realm == "" {
		return nil, errors.New("realm is empty")
	}
	return doRequestPtr[RealmRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm), errMessage)
}

// GetRealms returns the top-level representation of all realms.
// Input: ctx [context.Context], token [string] admin token.
// Output: [[]*RealmRepresentation] realms, [error] if the call fails.
func (g *GoCloak) GetRealms(ctx context.Context, token string) ([]*RealmRepresentation, error) {
	const errMessage = "could not get realms"

	return doRequest[[]*RealmRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(""), errMessage)
}

// CreateRealm creates the given realm.
// Input: ctx [context.Context], token [string] admin token, realm [RealmRepresentation].
// Output: [string] last segment of the Location header (the realm name), [error] if the call fails.
func (g *GoCloak) CreateRealm(ctx context.Context, token string, realm RealmRepresentation) (string, error) {
	const errMessage = "could not create realm"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(&realm),
		g.getAdminRealmURL(""), errMessage)
}

// UpdateRealm updates the given realm, addressed by its Realm name.
// Input: ctx [context.Context], token [string] admin token, realm [RealmRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateRealm(ctx context.Context, token string, realm RealmRepresentation) error {
	const errMessage = "could not update realm"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(realm),
		http.MethodPut, g.getAdminRealmURL(PString(realm.Realm)), errMessage)
}

// DeleteRealm deletes the given realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteRealm(ctx context.Context, token, realm string) error {
	const errMessage = "could not delete realm"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm), errMessage)
}

// ClearRealmCache clears the realm cache.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [error] if the call fails.
func (g *GoCloak) ClearRealmCache(ctx context.Context, token, realm string) error {
	const errMessage = "could not clear realm cache"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPost, g.getAdminRealmURL(realm, "clear-realm-cache"), errMessage)
}

// ClearUserCache clears the user cache of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [error] if the call fails.
func (g *GoCloak) ClearUserCache(ctx context.Context, token, realm string) error {
	const errMessage = "could not clear user cache"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPost, g.getAdminRealmURL(realm, "clear-user-cache"), errMessage)
}

// ClearKeysCache clears the keys cache of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [error] if the call fails.
func (g *GoCloak) ClearKeysCache(ctx context.Context, token, realm string) error {
	const errMessage = "could not clear keys cache"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPost, g.getAdminRealmURL(realm, "clear-keys-cache"), errMessage)
}

// GetAuthenticationFlows returns all authentication flows of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*AuthenticationFlowRepresentation] flows, [error] if the call fails.
func (g *GoCloak) GetAuthenticationFlows(ctx context.Context, token, realm string) ([]*AuthenticationFlowRepresentation, error) {
	const errMessage = "could not retrieve authentication flows"
	return doRequest[[]*AuthenticationFlowRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "flows"), errMessage)
}

// GetAuthenticationFlow fetches the authentication flow with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], authenticationFlowID [string].
// Output: [*AuthenticationFlowRepresentation] flow, [error] if the call fails.
func (g *GoCloak) GetAuthenticationFlow(ctx context.Context, token, realm string, authenticationFlowID string) (*AuthenticationFlowRepresentation, error) {
	const errMessage = "could not retrieve authentication flows"
	return doRequestPtr[AuthenticationFlowRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "flows", authenticationFlowID), errMessage)
}

// CreateAuthenticationFlow creates a new authentication flow in the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string], flow [AuthenticationFlowRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) CreateAuthenticationFlow(ctx context.Context, token, realm string, flow AuthenticationFlowRepresentation) error {
	const errMessage = "could not create authentication flows"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(flow),
		http.MethodPost, g.getAdminRealmURL(realm, "authentication", "flows"), errMessage)
}

// UpdateAuthenticationFlow updates the authentication flow with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], flow [AuthenticationFlowRepresentation],
// authenticationFlowID [string].
// Keycloak answers with 204 No Content; read the flow back with GetAuthenticationFlow if needed.
// Output: [error] if the call fails.
func (g *GoCloak) UpdateAuthenticationFlow(ctx context.Context, token, realm string, flow AuthenticationFlowRepresentation, authenticationFlowID string) error {
	const errMessage = "could not update authentication flow"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(flow),
		http.MethodPut, g.getAdminRealmURL(realm, "authentication", "flows", authenticationFlowID), errMessage)
}

// DeleteAuthenticationFlow deletes the authentication flow with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], flowID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteAuthenticationFlow(ctx context.Context, token, realm, flowID string) error {
	const errMessage = "could not delete authentication flows"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "authentication", "flows", flowID), errMessage)
}

// CreateAuthenticatorConfig creates a new authenticator configuration.
// Input: ctx [context.Context], token [string] admin token, realm [string],
// config [AuthenticatorConfigRepresentation].
// Output: [string] ID of the created configuration, [error] if the call fails.
func (g *GoCloak) CreateAuthenticatorConfig(ctx context.Context, token, realm string, config AuthenticatorConfigRepresentation) (string, error) {
	const errMessage = "could not create authenticator config"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).SetBody(config), g.getAdminRealmURL(realm, "authentication", "config"), errMessage)
}

// GetAuthenticatorConfigDescription fetches the configuration description of an authenticator provider.
// Input: ctx [context.Context], token [string] admin token, realm [string], providerID [string].
// Output: [*AuthenticatorConfigInfoRepresentation] config description, [error] if the call fails.
func (g *GoCloak) GetAuthenticatorConfigDescription(ctx context.Context, token, realm, providerID string) (*AuthenticatorConfigInfoRepresentation, error) {
	const errMessage = "could not retrieve authenticator config description"
	return doRequestPtr[AuthenticatorConfigInfoRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "config-description", providerID), errMessage)
}

// GetAuthenticatorConfig fetches an authenticator configuration by its ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], configID [string].
// Output: [*AuthenticatorConfigRepresentation] configuration, [error] if the call fails.
func (g *GoCloak) GetAuthenticatorConfig(ctx context.Context, token, realm, configID string) (*AuthenticatorConfigRepresentation, error) {
	const errMessage = "could not retrieve authenticator config"
	return doRequestPtr[AuthenticatorConfigRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "config", configID), errMessage)
}

// UpdateAuthenticatorConfig updates an authenticator configuration by its ID.
// Input: ctx [context.Context], token [string] admin token, realm [string],
// config [AuthenticatorConfigRepresentation], configID [string].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateAuthenticatorConfig(ctx context.Context, token, realm string, config AuthenticatorConfigRepresentation, configID string) error {
	const errMessage = "could not update authenticator config"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).SetBody(config), http.MethodPut, g.getAdminRealmURL(realm, "authentication", "config", configID), errMessage)
}

// DeleteAuthenticatorConfig deletes an authenticator configuration by its ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], configID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteAuthenticatorConfig(ctx context.Context, token, realm, configID string) error {
	const errMessage = "could not delete authenticator config"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "authentication", "config", configID), errMessage)
}

// GetAuthenticationExecutions returns all executions of the flow with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], flow [string] flow alias.
// Output: [[]*ModifyAuthenticationExecutionRepresentation] executions, [error] if the call fails.
func (g *GoCloak) GetAuthenticationExecutions(ctx context.Context, token, realm, flow string) ([]*ModifyAuthenticationExecutionRepresentation, error) {
	const errMessage = "could not retrieve authentication flows"
	return doRequest[[]*ModifyAuthenticationExecutionRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "flows", flow, "executions"), errMessage)
}

// CreateAuthenticationExecution adds a new authenticator execution to the flow with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], flow [string] flow alias,
// execution [CreateAuthenticationExecutionRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) CreateAuthenticationExecution(ctx context.Context, token, realm, flow string, execution CreateAuthenticationExecutionRepresentation) error {
	const errMessage = "could not create authentication execution"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).SetBody(execution), http.MethodPost, g.getAdminRealmURL(realm, "authentication", "flows", flow, "executions", "execution"), errMessage)
}

// UpdateAuthenticationExecution updates an execution of the flow with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], flow [string] flow alias,
// execution [ModifyAuthenticationExecutionRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateAuthenticationExecution(ctx context.Context, token, realm, flow string, execution ModifyAuthenticationExecutionRepresentation) error {
	const errMessage = "could not update authentication execution"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).SetBody(execution), http.MethodPut, g.getAdminRealmURL(realm, "authentication", "flows", flow, "executions"), errMessage)
}

// DeleteAuthenticationExecution deletes the execution with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], executionID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteAuthenticationExecution(ctx context.Context, token, realm, executionID string) error {
	const errMessage = "could not delete authentication execution"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "authentication", "executions", executionID), errMessage)
}

// CreateAuthenticationExecutionConfig creates a new configuration for an authentication execution.
// Input: ctx [context.Context], token [string] admin token, realm [string], executionID [string],
// config [AuthenticatorConfigRepresentation].
// Output: [string] ID of the created configuration, [error] if the call fails.
func (g *GoCloak) CreateAuthenticationExecutionConfig(ctx context.Context, token, realm, executionID string, config AuthenticatorConfigRepresentation) (string, error) {
	const errMessage = "could not create authentication execution config"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).SetBody(config), g.getAdminRealmURL(realm, "authentication", "executions", executionID, "config"), errMessage)
}

// GetAuthenticationExecutionConfig fetches a configuration of an authentication execution.
// Input: ctx [context.Context], token [string] admin token, realm [string], executionID [string], configID [string].
// Output: [*AuthenticatorConfigRepresentation] configuration, [error] if the call fails.
func (g *GoCloak) GetAuthenticationExecutionConfig(ctx context.Context, token, realm, executionID, configID string) (*AuthenticatorConfigRepresentation, error) {
	const errMessage = "could not retrieve authentication execution config"
	return doRequestPtr[AuthenticatorConfigRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "executions", executionID, "config", configID), errMessage)
}

// CreateAuthenticationExecutionFlow adds a new sub-flow execution to the flow with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], flow [string] flow alias,
// executionFlow [CreateAuthenticationExecutionFlowRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) CreateAuthenticationExecutionFlow(ctx context.Context, token, realm, flow string, executionFlow CreateAuthenticationExecutionFlowRepresentation) error {
	const errMessage = "could not create authentication execution flow"
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).SetBody(executionFlow), http.MethodPost, g.getAdminRealmURL(realm, "authentication", "flows", flow, "executions", "flow"), errMessage)
}

// -----
// Users
// -----

// CreateUser creates the given user in the given realm.
// Note: Keycloak has not documented what members of the User object are actually being accepted, when creating a user.
// Things like RealmRoles must be attached using followup calls to the respective functions.
// Input: ctx [context.Context], token [string] admin token, realm [string], user [User].
// Output: [string] ID of the created user, [error] if the call fails.
func (g *GoCloak) CreateUser(ctx context.Context, token, realm string, user User) (string, error) {
	const errMessage = "could not create user"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(user),
		g.getAdminRealmURL(realm, "users"), errMessage)
}

// DeleteUser deletes the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteUser(ctx context.Context, token, realm, userID string) error {
	const errMessage = "could not delete user"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "users", userID), errMessage)
}

// GetUserByID fetches the user with the given ID.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], userID [string].
// Output: [*User] user, [error] if userID is empty or the call fails.
func (g *GoCloak) GetUserByID(ctx context.Context, accessToken, realm, userID string) (*User, error) {
	const errMessage = "could not get user by id"

	if userID == "" {
		return nil, fmt.Errorf("%s: userID shall not be empty", errMessage)
	}

	return doRequestPtr[User](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getAdminRealmURL(realm, "users", userID), errMessage)
}

// GetUserCount returns the number of users in the realm matching the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetUsersParams].
// Output: [int] user count, [error] if the call fails.
func (g *GoCloak) GetUserCount(ctx context.Context, token string, realm string, params GetUsersParams) (int, error) {
	const errMessage = "could not get user count"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[int](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "users", "count"), errMessage)
}

// GetUserGroups returns the groups the given user is a member of.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string],
// params [GetGroupsParams].
// Output: [[]*Group] groups, [error] if the call fails.
func (g *GoCloak) GetUserGroups(ctx context.Context, token, realm, userID string, params GetGroupsParams) ([]*Group, error) {
	const errMessage = "could not get user groups"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "groups"), errMessage)
}

// GetUserProfileConfig fetches the user profile configuration of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [*UserProfileConfig] user profile configuration, [error] if the call fails.
func (g *GoCloak) GetUserProfileConfig(ctx context.Context, token, realm string) (*UserProfileConfig, error) {
	const errMessage = "could not get user profile configuration"

	return doRequestPtr[UserProfileConfig](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", "profile"), errMessage)
}

// GetUsers returns the users of the realm matching the given query params.
// Default number of results per page is 100, use GetUsersParams to specify it explicitly or to set offset for
// pagination.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetUsersParams].
// Output: [[]*User] users, [error] if the call fails.
func (g *GoCloak) GetUsers(ctx context.Context, token, realm string, params GetUsersParams) ([]*User, error) {
	const errMessage = "could not get users"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*User](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "users"), errMessage)
}

// GetUsersByRoleName returns the users that have the given realm role.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleName [string],
// params [GetUsersByRoleParams].
// Output: [[]*User] users, [error] if the call fails.
func (g *GoCloak) GetUsersByRoleName(ctx context.Context, token, realm, roleName string, params GetUsersByRoleParams) ([]*User, error) {
	const errMessage = "could not get users by role name"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*User](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "roles", roleName, "users"), errMessage)
}

// GetUsersByClientRoleName returns the users that have the given client role.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], roleName [string],
// params [GetUsersByRoleParams].
// Output: [[]*User] users, [error] if the call fails.
func (g *GoCloak) GetUsersByClientRoleName(ctx context.Context, token, realm, idOfClient, roleName string, params GetUsersByRoleParams) ([]*User, error) {
	const errMessage = "could not get users by client role name"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	return doRequest[[]*User](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "roles", roleName, "users"), errMessage)
}

// SetPassword sets a new password for the user with the given id. Needs elevated privileges.
// Input: ctx [context.Context], token [string], realm [string], userID [string], password [string],
// temporary [bool] user must change it on next login. Output: [error].
func (g *GoCloak) SetPassword(ctx context.Context, token, realm, userID, password string, temporary bool) error {
	const errMessage = "could not set password"

	requestBody := SetPasswordRequest{Password: password, Temporary: temporary, Type: "password"}
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(requestBody),
		http.MethodPut, g.getAdminRealmURL(realm, "users", userID, "reset-password"), errMessage)
}

// UpdateUser updates the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], user [User] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateUser(ctx context.Context, token, realm string, user User) error {
	const errMessage = "could not update user"

	if NilOrEmpty(user.ID) {
		return fmt.Errorf("%s: ID of a user required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(user),
		http.MethodPut, g.getAdminRealmURL(realm, "users", PString(user.ID)), errMessage)
}

// UpdateUserProfileConfig updates the user profile configuration of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string], config [UserProfileConfig].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateUserProfileConfig(ctx context.Context, token, realm string, config UserProfileConfig) error {
	const errMessage = "could not update user profile configuration"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(config),
		http.MethodPut, g.getAdminRealmURL(realm, "users", "profile"), errMessage)
}

// AddUserToGroup adds the given user to the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], groupID [string].
// Output: [error] if the call fails.
func (g *GoCloak) AddUserToGroup(ctx context.Context, token, realm, userID, groupID string) error {
	const errMessage = "could not add user to group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPut, g.getAdminRealmURL(realm, "users", userID, "groups", groupID), errMessage)
}

// DeleteUserFromGroup removes the given user from the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], groupID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteUserFromGroup(ctx context.Context, token, realm, userID, groupID string) error {
	const errMessage = "could not delete user from group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "users", userID, "groups", groupID), errMessage)
}

// GetUserSessions returns the sessions of the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [[]*UserSessionRepresentation] sessions, [error] if the call fails.
func (g *GoCloak) GetUserSessions(ctx context.Context, token, realm, userID string) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get user sessions"

	return doRequest[[]*UserSessionRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "sessions"), errMessage)
}

// GetUserOfflineSessionsForClient returns the offline sessions of the given user for the given client.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], idOfClient [string].
// Output: [[]*UserSessionRepresentation] offline sessions, [error] if the call fails.
func (g *GoCloak) GetUserOfflineSessionsForClient(ctx context.Context, token, realm, userID, idOfClient string) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get user offline sessions for client"

	return doRequest[[]*UserSessionRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "offline-sessions", idOfClient), errMessage)
}

// AddClientRolesToUser adds client-level role mappings to the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], userID [string],
// roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) AddClientRolesToUser(ctx context.Context, token, realm, idOfClient, userID string, roles []Role) error {
	const errMessage = "could not add client role to user"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", idOfClient), errMessage)
}

// AddClientRolesToGroup adds client-level role mappings to the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], groupID [string],
// roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) AddClientRolesToGroup(ctx context.Context, token, realm, idOfClient, groupID string, roles []Role) error {
	const errMessage = "could not add client role to group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", idOfClient), errMessage)
}

// DeleteClientRolesFromUser removes client-level role mappings from the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], userID [string],
// roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientRolesFromUser(ctx context.Context, token, realm, idOfClient, userID string, roles []Role) error {
	const errMessage = "could not delete client role from user"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", idOfClient), errMessage)
}

// DeleteClientRoleFromGroup removes client-level role mappings from the given group.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], groupID [string],
// roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientRoleFromGroup(ctx context.Context, token, realm, idOfClient, groupID string, roles []Role) error {
	const errMessage = "could not delete client role from group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", idOfClient), errMessage)
}

// AddClientRoleComposite adds the given roles as composites of the role with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) AddClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []Role) error {
	const errMessage = "could not add client role composite"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "roles-by-id", roleID, "composites"), errMessage)
}

// DeleteClientRoleComposite removes the given roles from the composites of the role with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], roleID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []Role) error {
	const errMessage = "could not delete client role composite"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "roles-by-id", roleID, "composites"), errMessage)
}

// GetUserFederatedIdentities returns the federated identities linked to the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [[]*FederatedIdentityRepresentation] federated identities, [error] if the call fails.
func (g *GoCloak) GetUserFederatedIdentities(ctx context.Context, token, realm, userID string) ([]*FederatedIdentityRepresentation, error) {
	const errMessage = "could not get user federated identities"

	return doRequest[[]*FederatedIdentityRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "federated-identity"), errMessage)
}

// CreateUserFederatedIdentity links the given user to an identity provider.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], providerID [string],
// federatedIdentityRep [FederatedIdentityRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) CreateUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string, federatedIdentityRep FederatedIdentityRepresentation) error {
	const errMessage = "could not create user federeated identity"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(federatedIdentityRep),
		http.MethodPost, g.getAdminRealmURL(realm, "users", userID, "federated-identity", providerID), errMessage)
}

// DeleteUserFederatedIdentity removes the link between the given user and an identity provider.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], providerID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string) error {
	const errMessage = "could not delete user federeated identity"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "users", userID, "federated-identity", providerID), errMessage)
}

// GetUserBruteForceDetectionStatus fetches the brute force protection status of the given user.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], userID [string].
// Output: [*BruteForceStatus] status, [error] if the call fails.
func (g *GoCloak) GetUserBruteForceDetectionStatus(ctx context.Context, accessToken, realm, userID string) (*BruteForceStatus, error) {
	const errMessage = "could not get brute force detection status"
	return doRequestPtr[BruteForceStatus](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getAttackDetectionURL(realm, "users", userID), errMessage)
}

// ------------------
// Identity Providers
// ------------------

// CreateIdentityProvider creates an identity provider in the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string],
// providerRep [IdentityProviderRepresentation].
// Output: [string] last segment of the Location header (the provider alias), [error] if the call fails.
func (g *GoCloak) CreateIdentityProvider(ctx context.Context, token string, realm string, providerRep IdentityProviderRepresentation) (string, error) {
	const errMessage = "could not create identity provider"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(providerRep),
		g.getAdminRealmURL(realm, "identity-provider", "instances"), errMessage)
}

// GetIdentityProviders returns all identity providers of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*IdentityProviderRepresentation] identity providers, [error] if the call fails.
func (g *GoCloak) GetIdentityProviders(ctx context.Context, token, realm string) ([]*IdentityProviderRepresentation, error) {
	const errMessage = "could not get identity providers"

	return doRequest[[]*IdentityProviderRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "identity-provider", "instances"), errMessage)
}

// GetIdentityProvider fetches the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string].
// Output: [*IdentityProviderRepresentation] identity provider, [error] if the call fails.
func (g *GoCloak) GetIdentityProvider(ctx context.Context, token, realm, alias string) (*IdentityProviderRepresentation, error) {
	const errMessage = "could not get identity provider"

	return doRequestPtr[IdentityProviderRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "identity-provider", "instances", alias), errMessage)
}

// UpdateIdentityProvider updates the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string],
// providerRep [IdentityProviderRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateIdentityProvider(ctx context.Context, token, realm, alias string, providerRep IdentityProviderRepresentation) error {
	const errMessage = "could not update identity provider"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(providerRep),
		http.MethodPut, g.getAdminRealmURL(realm, "identity-provider", "instances", alias), errMessage)
}

// DeleteIdentityProvider deletes the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteIdentityProvider(ctx context.Context, token, realm, alias string) error {
	const errMessage = "could not delete identity provider"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "identity-provider", "instances", alias), errMessage)
}

// ExportIDPPublicBrokerConfig exports the public broker configuration of the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string].
// Output: [*string] raw exported configuration (e.g. SAML XML), [error] if the call fails.
func (g *GoCloak) ExportIDPPublicBrokerConfig(ctx context.Context, token, realm, alias string) (*string, error) {
	const errMessage = "could not get public identity provider configuration"

	resp, err := g.GetRequestWithBearerAuthXMLHeader(ctx, token).
		Get(g.getAdminRealmURL(realm, "identity-provider", "instances", alias, "export"))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	result := resp.String()
	return &result, nil
}

// ImportIdentityProviderConfig lets Keycloak fetch and parse the identity provider metadata at the given URL.
// Input: ctx [context.Context], token [string] admin token, realm [string], fromURL [string] metadata URL,
// providerID [string] e.g. "oidc" or "saml".
// Output: [map[string]string] parsed provider config, [error] if the call fails.
func (g *GoCloak) ImportIdentityProviderConfig(ctx context.Context, token, realm, fromURL, providerID string) (map[string]string, error) {
	const errMessage = "could not import config"

	return doRequest[map[string]string](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(map[string]string{
			"fromUrl":    fromURL,
			"providerId": providerID,
		}),
		http.MethodPost, g.getAdminRealmURL(realm, "identity-provider", "import-config"), errMessage)
}

// ImportIdentityProviderConfigFromFile uploads an identity provider metadata file for Keycloak to parse.
// Input: ctx [context.Context], token [string] admin token, realm [string], providerID [string],
// fileName [string], fileBody [io.Reader] file content.
// Output: [map[string]string] parsed provider config, [error] if the call fails.
func (g *GoCloak) ImportIdentityProviderConfigFromFile(ctx context.Context, token, realm, providerID, fileName string, fileBody io.Reader) (map[string]string, error) {
	const errMessage = "could not import config"

	return doRequest[map[string]string](g.GetRequestWithBearerAuth(ctx, token).
		SetFileReader("file", fileName, fileBody).
		SetFormData(map[string]string{
			"providerId": providerID,
		}),
		http.MethodPost, g.getAdminRealmURL(realm, "identity-provider", "import-config"), errMessage)
}

// CreateIdentityProviderMapper creates a mapper for the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string],
// mapper [IdentityProviderMapper].
// Output: [string] ID of the created mapper, [error] if the call fails.
func (g *GoCloak) CreateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper IdentityProviderMapper) (string, error) {
	const errMessage = "could not create mapper for identity provider"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(mapper),
		g.getAdminRealmURL(realm, "identity-provider", "instances", alias, "mappers"), errMessage)
}

// DeleteIdentityProviderMapper deletes a mapper of the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string], mapperID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteIdentityProviderMapper(ctx context.Context, token, realm, alias, mapperID string) error {
	const errMessage = "could not delete mapper for identity provider"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "identity-provider", "instances", alias, "mappers", mapperID), errMessage)
}

// GetIdentityProviderMappers returns the mappers of the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string].
// Output: [[]*IdentityProviderMapper] mappers, [error] if the call fails.
func (g *GoCloak) GetIdentityProviderMappers(ctx context.Context, token, realm, alias string) ([]*IdentityProviderMapper, error) {
	const errMessage = "could not get identity provider mappers"

	return doRequest[[]*IdentityProviderMapper](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "identity-provider", "instances", alias, "mappers"), errMessage)
}

// GetIdentityProviderMapperByID gets the mapper of an identity provider.
// Input: ctx [context.Context], token [string], realm [string], alias [string] identity provider alias,
// mapperID [string]. Output: [*IdentityProviderMapper], [error].
func (g *GoCloak) GetIdentityProviderMapperByID(ctx context.Context, token, realm, alias, mapperID string) (*IdentityProviderMapper, error) {
	const errMessage = "could not get identity provider mappers"

	return doRequestPtr[IdentityProviderMapper](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "identity-provider", "instances", alias, "mappers", mapperID), errMessage)
}

// UpdateIdentityProviderMapper updates a mapper of the identity provider with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string],
// mapper [IdentityProviderMapper] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper IdentityProviderMapper) error {
	const errMessage = "could not update identity provider mapper"

	if NilOrEmpty(mapper.ID) {
		return fmt.Errorf("%s: ID of a mapper required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(mapper),
		http.MethodPut, g.getAdminRealmURL(realm, "identity-provider", "instances", alias, "mappers", PString(mapper.ID)), errMessage)
}

// ------------------
// Protection API
// ------------------

// GetResource fetches a resource of the client's resource server, using an admin token.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], resourceID [string].
// Output: [*ResourceRepresentation] resource, [error] if the call fails.
func (g *GoCloak) GetResource(ctx context.Context, token, realm, idOfClient, resourceID string) (*ResourceRepresentation, error) {
	const errMessage = "could not get resource"

	return doRequestPtr[ResourceRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "resource", resourceID), errMessage)
}

// GetResourceClient fetches a resource via the protection API, using a client token.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string], resourceID [string].
// Output: [*ResourceRepresentation] resource, [error] if the call fails.
func (g *GoCloak) GetResourceClient(ctx context.Context, token, realm, resourceID string) (*ResourceRepresentation, error) {
	const errMessage = "could not get resource"

	return doRequestPtr[ResourceRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getRealmURL(realm, "authz", "protection", "resource_set", resourceID), errMessage)
}

// GetResources returns the resources of the client's resource server, using an admin token.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// params [GetResourceParams].
// Output: [[]*ResourceRepresentation] resources, [error] if the call fails.
func (g *GoCloak) GetResources(ctx context.Context, token, realm, idOfClient string, params GetResourceParams) ([]*ResourceRepresentation, error) {
	const errMessage = "could not get resources"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	return doRequest[[]*ResourceRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "resource"), errMessage)
}

// GetResourcesClient returns the full resources via the protection API (deep=true), using a client token.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string], params [GetResourceParams].
// Output: [[]*ResourceRepresentation] resources, [error] if the call fails.
func (g *GoCloak) GetResourcesClient(ctx context.Context, token, realm string, params GetResourceParams) ([]*ResourceRepresentation, error) {
	const errMessage = "could not get resources"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	// deep=true returns the full resources in one call instead of only their IDs
	queryParams["deep"] = "true"

	return doRequest[[]*ResourceRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getRealmURL(realm, "authz", "protection", "resource_set"), errMessage)
}

// GetResourceServer returns the resource server settings of the given client.
// The access token must have the realm view_clients role on its service
// account to be allowed to call this endpoint.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string].
// Output: [*ResourceServerRepresentation] settings, [error] if the call fails.
func (g *GoCloak) GetResourceServer(ctx context.Context, token, realm, idOfClient string) (*ResourceServerRepresentation, error) {
	const errMessage = "could not get resource server settings"

	return doRequestPtr[ResourceServerRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "settings"), errMessage)
}

// UpdateResource updates a resource of the client's resource server, using an admin token.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// resource [ResourceRepresentation] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateResource(ctx context.Context, token, realm, idOfClient string, resource ResourceRepresentation) error {
	const errMessage = "could not update resource"

	if NilOrEmpty(resource.ID) {
		return errors.New("ID of a resource required")
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(resource),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "resource", *resource.ID), errMessage)
}

// UpdateResourceClient updates a resource via the protection API, using a client token.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string],
// resource [ResourceRepresentation] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateResourceClient(ctx context.Context, token, realm string, resource ResourceRepresentation) error {
	const errMessage = "could not update resource"

	if NilOrEmpty(resource.ID) {
		return errors.New("ID of a resource required")
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(resource),
		http.MethodPut, g.getRealmURL(realm, "authz", "protection", "resource_set", *resource.ID), errMessage)
}

// CreateResource creates a resource on the client's resource server, using an admin token.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// resource [ResourceRepresentation].
// Output: [*ResourceRepresentation] created resource, [error] if the call fails.
func (g *GoCloak) CreateResource(ctx context.Context, token, realm string, idOfClient string, resource ResourceRepresentation) (*ResourceRepresentation, error) {
	const errMessage = "could not create resource"

	return doRequestPtr[ResourceRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(resource),
		http.MethodPost, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "resource"), errMessage)
}

// CreateResourceClient creates a resource via the protection API, using a client token.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string], resource [ResourceRepresentation].
// Output: [*ResourceRepresentation] created resource, [error] if the call fails.
func (g *GoCloak) CreateResourceClient(ctx context.Context, token, realm string, resource ResourceRepresentation) (*ResourceRepresentation, error) {
	const errMessage = "could not create resource"

	return doRequestPtr[ResourceRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(resource),
		http.MethodPost, g.getRealmURL(realm, "authz", "protection", "resource_set"), errMessage)
}

// DeleteResource deletes a resource of the client's resource server, using an admin token.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], resourceID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteResource(ctx context.Context, token, realm, idOfClient, resourceID string) error {
	const errMessage = "could not delete resource"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "resource", resourceID), errMessage)
}

// DeleteResourceClient deletes a resource via the protection API, using a client token.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string], resourceID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteResourceClient(ctx context.Context, token, realm, resourceID string) error {
	const errMessage = "could not delete resource"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getRealmURL(realm, "authz", "protection", "resource_set", resourceID), errMessage)
}

// GetScope fetches an authorization scope of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], scopeID [string].
// Output: [*ScopeRepresentation] scope, [error] if the call fails.
func (g *GoCloak) GetScope(ctx context.Context, token, realm, idOfClient, scopeID string) (*ScopeRepresentation, error) {
	const errMessage = "could not get scope"

	return doRequestPtr[ScopeRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "scope", scopeID), errMessage)
}

// GetScopes returns the authorization scopes of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// params [GetScopeParams].
// Output: [[]*ScopeRepresentation] scopes, [error] if the call fails.
func (g *GoCloak) GetScopes(ctx context.Context, token, realm, idOfClient string, params GetScopeParams) ([]*ScopeRepresentation, error) {
	const errMessage = "could not get scopes"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}
	return doRequest[[]*ScopeRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "scope"), errMessage)
}

// CreateScope creates an authorization scope on the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// scope [ScopeRepresentation].
// Output: [*ScopeRepresentation] created scope, [error] if the call fails.
func (g *GoCloak) CreateScope(ctx context.Context, token, realm, idOfClient string, scope ScopeRepresentation) (*ScopeRepresentation, error) {
	const errMessage = "could not create scope"

	return doRequestPtr[ScopeRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(scope),
		http.MethodPost, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "scope"), errMessage)
}

// GetPermissionScope fetches a scope-based permission of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// idOfScope [string] permission ID.
// Output: [*PolicyRepresentation] permission, [error] if the call fails.
func (g *GoCloak) GetPermissionScope(ctx context.Context, token, realm, idOfClient string, idOfScope string) (*PolicyRepresentation, error) {
	const errMessage = "could not get permission scope"

	return doRequestPtr[PolicyRepresentation](g.GetRequestWithBearerAuth(ctx, token),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", "scope", idOfScope), errMessage)
}

// UpdatePermissionScope updates a scope-based permission of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// idOfScope [string] permission ID, policy [PolicyRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) UpdatePermissionScope(ctx context.Context, token, realm, idOfClient string, idOfScope string, policy PolicyRepresentation) error {
	const errMessage = "could not update permission scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(policy),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", "scope", idOfScope), errMessage)
}

// UpdateScope updates an authorization scope of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// scope [ScopeRepresentation] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateScope(ctx context.Context, token, realm, idOfClient string, scope ScopeRepresentation) error {
	const errMessage = "could not update scope"

	if NilOrEmpty(scope.ID) {
		return errors.New("ID of a scope required")
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(scope),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "scope", *scope.ID), errMessage)
}

// DeleteScope deletes an authorization scope of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], scopeID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteScope(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	const errMessage = "could not delete scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "scope", scopeID), errMessage)
}

// GetPolicy fetches a policy of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], policyID [string].
// Output: [*PolicyRepresentation] policy, [error] if the call fails.
func (g *GoCloak) GetPolicy(ctx context.Context, token, realm, idOfClient, policyID string) (*PolicyRepresentation, error) {
	const errMessage = "could not get policy"

	return doRequestPtr[PolicyRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "policy", policyID), errMessage)
}

// GetPolicies returns the policies of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// params [GetPolicyParams].
// Output: [[]*PolicyRepresentation] policies, [error] if the call fails.
func (g *GoCloak) GetPolicies(ctx context.Context, token, realm, idOfClient string, params GetPolicyParams) ([]*PolicyRepresentation, error) {
	const errMessage = "could not get policies"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*PolicyRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "policy"), errMessage)
}

// CreatePolicy creates a policy of the given type on the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// policy [PolicyRepresentation] (Type required).
// Output: [*PolicyRepresentation] created policy, [error] if the type is missing or the call fails.
func (g *GoCloak) CreatePolicy(ctx context.Context, token, realm, idOfClient string, policy PolicyRepresentation) (*PolicyRepresentation, error) {
	const errMessage = "could not create policy"

	if NilOrEmpty(policy.Type) {
		return nil, errors.New("type of a policy required")
	}

	path := []string{"clients", idOfClient, "authz", "resource-server", "policy", *policy.Type}

	return doRequestPtr[PolicyRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(policy),
		http.MethodPost, g.getAdminRealmURL(realm, path...), errMessage)
}

// UpdatePolicy updates a policy of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// policy [PolicyRepresentation] (ID and Type required).
// Output: [error] if ID or type is missing or the call fails.
func (g *GoCloak) UpdatePolicy(ctx context.Context, token, realm, idOfClient string, policy PolicyRepresentation) error {
	const errMessage = "could not update policy"

	if NilOrEmpty(policy.ID) {
		return errors.New("ID of a policy required")
	}

	if NilOrEmpty(policy.Type) {
		return errors.New("type of a policy required")
	}

	path := []string{"clients", idOfClient, "authz", "resource-server", "policy", *policy.Type, *policy.ID}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(policy),
		http.MethodPut, g.getAdminRealmURL(realm, path...), errMessage)
}

// DeletePolicy deletes a policy of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], policyID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeletePolicy(ctx context.Context, token, realm, idOfClient, policyID string) error {
	const errMessage = "could not delete policy"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "policy", policyID), errMessage)
}

// GetAuthorizationPolicyAssociatedPolicies returns the policies associated with the given policy.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], policyID [string].
// Output: [[]*PolicyRepresentation] associated policies, [error] if the call fails.
func (g *GoCloak) GetAuthorizationPolicyAssociatedPolicies(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PolicyRepresentation, error) {
	const errMessage = "could not get policy associated policies"

	return doRequest[[]*PolicyRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "policy", policyID, "associatedPolicies"), errMessage)
}

// GetAuthorizationPolicyResources returns the resources the given policy applies to.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], policyID [string].
// Output: [[]*PolicyResourceRepresentation] resources, [error] if the call fails.
func (g *GoCloak) GetAuthorizationPolicyResources(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PolicyResourceRepresentation, error) {
	const errMessage = "could not get policy resources"

	return doRequest[[]*PolicyResourceRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "policy", policyID, "resources"), errMessage)
}

// GetAuthorizationPolicyScopes returns the scopes the given policy applies to.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], policyID [string].
// Output: [[]*PolicyScopeRepresentation] scopes, [error] if the call fails.
func (g *GoCloak) GetAuthorizationPolicyScopes(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PolicyScopeRepresentation, error) {
	const errMessage = "could not get policy scopes"

	return doRequest[[]*PolicyScopeRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "policy", policyID, "scopes"), errMessage)
}

// GetResourcePolicy fetches a UMA permission (policy) via the protection API, using a token obtained by
// Resource Owner Password Credentials Grant or Token exchange.
// Input: ctx [context.Context], token [string] resource owner token, realm [string], permissionID [string].
// Output: [*ResourcePolicyRepresentation] permission, [error] if the call fails.
func (g *GoCloak) GetResourcePolicy(ctx context.Context, token, realm, permissionID string) (*ResourcePolicyRepresentation, error) {
	const errMessage = "could not get resource policy"

	return doRequestPtr[ResourcePolicyRepresentation](g.GetRequestWithBearerAuthNoCache(ctx, token), http.MethodGet, g.getRealmURL(realm, "authz", "protection", "uma-policy", permissionID), errMessage)
}

// GetResourcePolicies returns the UMA permissions (policies) via the protection API, using a token obtained by
// Resource Owner Password Credentials Grant or Token exchange.
// Input: ctx [context.Context], token [string] resource owner token, realm [string],
// params [GetResourcePoliciesParams].
// Output: [[]*ResourcePolicyRepresentation] permissions, [error] if the call fails.
func (g *GoCloak) GetResourcePolicies(ctx context.Context, token, realm string, params GetResourcePoliciesParams) ([]*ResourcePolicyRepresentation, error) {
	const errMessage = "could not get resource policies"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	return doRequest[[]*ResourcePolicyRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getRealmURL(realm, "authz", "protection", "uma-policy"), errMessage)
}

// CreateResourcePolicy associates a UMA permission with a specific resource, using a token obtained by
// Resource Owner Password Credentials Grant or Token exchange.
// Input: ctx [context.Context], token [string] resource owner token, realm [string], resourceID [string],
// policy [ResourcePolicyRepresentation].
// Output: [*ResourcePolicyRepresentation] created permission, [error] if the call fails.
func (g *GoCloak) CreateResourcePolicy(ctx context.Context, token, realm, resourceID string, policy ResourcePolicyRepresentation) (*ResourcePolicyRepresentation, error) {
	const errMessage = "could not create resource policy"

	return doRequestPtr[ResourcePolicyRepresentation](g.GetRequestWithBearerAuthNoCache(ctx, token).
		SetBody(policy),
		http.MethodPost, g.getRealmURL(realm, "authz", "protection", "uma-policy", resourceID), errMessage)
}

// UpdateResourcePolicy updates a UMA permission of a specific resource, using a token obtained by
// Resource Owner Password Credentials Grant or Token exchange.
// Input: ctx [context.Context], token [string] resource owner token, realm [string], permissionID [string],
// policy [ResourcePolicyRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateResourcePolicy(ctx context.Context, token, realm, permissionID string, policy ResourcePolicyRepresentation) error {
	const errMessage = "could not update resource policy"

	return doNoResult(g.GetRequestWithBearerAuthNoCache(ctx, token).
		SetBody(policy),
		http.MethodPut, g.getRealmURL(realm, "authz", "protection", "uma-policy", permissionID), errMessage)
}

// DeleteResourcePolicy deletes a UMA permission of a specific resource, using a token obtained by
// Resource Owner Password Credentials Grant or Token exchange.
// Input: ctx [context.Context], token [string] resource owner token, realm [string], permissionID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteResourcePolicy(ctx context.Context, token, realm, permissionID string) error {
	const errMessage = "could not  delete resource policy"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getRealmURL(realm, "authz", "protection", "uma-policy", permissionID), errMessage)
}

// GetPermission fetches a permission of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// permissionID [string].
// Output: [*PermissionRepresentation] permission, [error] if the call fails.
func (g *GoCloak) GetPermission(ctx context.Context, token, realm, idOfClient, permissionID string) (*PermissionRepresentation, error) {
	const errMessage = "could not get permission"

	return doRequestPtr[PermissionRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", permissionID), errMessage)
}

// GetDependentPermissions returns the permissions that depend on the given policy.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], policyID [string].
// Output: [[]*PermissionRepresentation] dependent permissions, [error] if the call fails.
func (g *GoCloak) GetDependentPermissions(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PermissionRepresentation, error) {
	const errMessage = "could not get permission"

	return doRequest[[]*PermissionRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "policy", policyID, "dependentPolicies"), errMessage)
}

// GetPermissionResources returns the resources attached to the given permission.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// permissionID [string].
// Output: [[]*PermissionResource] resources, [error] if the call fails.
func (g *GoCloak) GetPermissionResources(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*PermissionResource, error) {
	const errMessage = "could not get permission resource"

	return doRequest[[]*PermissionResource](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", permissionID, "resources"), errMessage)
}

// GetScopePermissions returns the permissions associated with the given authorization scope.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string], idOfScope [string].
// Output: [[]*PolicyRepresentation] permissions, [error] if the call fails.
func (g *GoCloak) GetScopePermissions(ctx context.Context, token, realm, idOfClient, idOfScope string) ([]*PolicyRepresentation, error) {
	const errMessage = "could not get scope permissions"

	return doRequest[[]*PolicyRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "scope", idOfScope, "permissions"), errMessage)
}

// GetPermissionScopes returns the scopes configured for the given permission.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// permissionID [string].
// Output: [[]*PermissionScope] scopes, [error] if the call fails.
func (g *GoCloak) GetPermissionScopes(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*PermissionScope, error) {
	const errMessage = "could not get permission scopes"

	return doRequest[[]*PermissionScope](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", permissionID, "scopes"), errMessage)
}

// GetPermissions returns the permissions of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// params [GetPermissionParams].
// Output: [[]*PermissionRepresentation] permissions, [error] if the call fails.
func (g *GoCloak) GetPermissions(ctx context.Context, token, realm, idOfClient string, params GetPermissionParams) ([]*PermissionRepresentation, error) {
	const errMessage = "could not get permissions"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*PermissionRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission"), errMessage)
}

// checkPermissionTicketParams checks that at least one ticket is given and each has a resource ID and scopes.
// Input: permissions [[]CreatePermissionTicketParams].
// Output: [error] describing the first missing field, nil if valid.
func checkPermissionTicketParams(permissions []CreatePermissionTicketParams) error {
	if len(permissions) == 0 {
		return errors.New("at least one permission ticket must be requested")
	}

	for _, pt := range permissions {

		if pt.ResourceID == "" {
			return errors.New("resourceID required for permission ticket")
		}
		if NilOrEmptyArray(pt.ResourceScopes) {
			return errors.New("at least one resourceScope required for permission ticket")
		}
	}

	return nil
}

// CreatePermissionTicket creates a permission ticket via the protection API, using a client token.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string],
// permissions [[]CreatePermissionTicketParams].
// Output: [*PermissionTicketResponseRepresentation] ticket, [error] if params are invalid or the call fails.
func (g *GoCloak) CreatePermissionTicket(ctx context.Context, token, realm string, permissions []CreatePermissionTicketParams) (*PermissionTicketResponseRepresentation, error) {
	const errMessage = "could not create permission ticket"

	err := checkPermissionTicketParams(permissions)
	if err != nil {
		return nil, err
	}

	return doRequestPtr[PermissionTicketResponseRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(permissions),
		http.MethodPost, g.getRealmURL(realm, "authz", "protection", "permission"), errMessage)
}

// checkPermissionGrantParams checks that requester ID, resource ID and scope name are set.
// Input: permission [PermissionGrantParams].
// Output: [error] describing the first missing field, nil if valid.
func checkPermissionGrantParams(permission PermissionGrantParams) error {
	if permission.RequesterID == "" {
		return errors.New("requesterID required to grant user permission")
	}

	if permission.ResourceID == "" {
		return errors.New("resourceID required to grant user permission")
	}

	if permission.ScopeName == "" {
		return errors.New("scopeName required to grant user permission")
	}

	return nil
}

// GrantUserPermission lets the resource owner grant a user permission on a resource scope.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string],
// permission [PermissionGrantParams] (RequesterID, ResourceID and ScopeName required).
// Output: [*PermissionGrantResponseRepresentation] granted ticket, [error] if params are invalid or the call fails.
func (g *GoCloak) GrantUserPermission(ctx context.Context, token, realm string, permission PermissionGrantParams) (*PermissionGrantResponseRepresentation, error) {
	const errMessage = "could not grant user permission"

	err := checkPermissionGrantParams(permission)
	if err != nil {
		return nil, err
	}

	permission.Granted = new(true)

	return doRequestPtr[PermissionGrantResponseRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(permission),
		http.MethodPost, g.getRealmURL(realm, "authz", "protection", "permission", "ticket"), errMessage)
}

// checkPermissionUpdateParams checks the grant params and that Granted is set.
// Input: permission [PermissionGrantParams].
// Output: [error] describing the first missing field, nil if valid.
func checkPermissionUpdateParams(permission PermissionGrantParams) error {
	err := checkPermissionGrantParams(permission)
	if err != nil {
		return err
	}

	if permission.Granted == nil {
		return errors.New("granted required to update user permission")
	}
	return nil
}

// UpdateUserPermission updates a permission ticket of a user.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string],
// permission [PermissionGrantParams] (grant fields and Granted required).
// Output: [*PermissionGrantResponseRepresentation] updated ticket, or nil if the permission was revoked
// (204 No Content), [error] if params are invalid or the call fails.
func (g *GoCloak) UpdateUserPermission(ctx context.Context, token, realm string, permission PermissionGrantParams) (*PermissionGrantResponseRepresentation, error) {
	const errMessage = "could not update user permission"

	err := checkPermissionUpdateParams(permission)
	if err != nil {
		return nil, err
	}

	var result PermissionGrantResponseRepresentation

	resp, err := g.GetRequestWithBearerAuth(ctx, token).
		SetResult(&result).
		SetBody(permission).
		Put(g.getRealmURL(realm, "authz", "protection", "permission", "ticket"))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusNoContent { // permission updated to 'not granted' removes permission
		return nil, nil
	}

	return &result, nil
}

// GetUserPermissions returns the permission tickets matching the given query params.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string],
// params [GetUserPermissionParams].
// Output: [[]*PermissionGrantResponseRepresentation] tickets, [error] if the call fails.
func (g *GoCloak) GetUserPermissions(ctx context.Context, token, realm string, params GetUserPermissionParams) ([]*PermissionGrantResponseRepresentation, error) {
	const errMessage = "could not get user permissions"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	return doRequest[[]*PermissionGrantResponseRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getRealmURL(realm, "authz", "protection", "permission", "ticket"), errMessage)
}

// DeleteUserPermission revokes a permission by deleting its permission ticket.
// Input: ctx [context.Context], token [string] client (PAT) token, realm [string], ticketID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteUserPermission(ctx context.Context, token, realm, ticketID string) error {
	const errMessage = "could not delete user permission"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getRealmURL(realm, "authz", "protection", "permission", "ticket", ticketID), errMessage)
}

// CreatePermission creates a permission of the given type on the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// permission [PermissionRepresentation] (Type required).
// Output: [*PermissionRepresentation] created permission, [error] if the type is missing or the call fails.
func (g *GoCloak) CreatePermission(ctx context.Context, token, realm, idOfClient string, permission PermissionRepresentation) (*PermissionRepresentation, error) {
	const errMessage = "could not create permission"

	if NilOrEmpty(permission.Type) {
		return nil, errors.New("type of a permission required")
	}

	return doRequestPtr[PermissionRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetBody(permission),
		http.MethodPost, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", *permission.Type), errMessage)
}

// UpdatePermission updates a permission of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// permission [PermissionRepresentation] (ID and Type required).
// Output: [error] if ID or type is missing or the call fails.
func (g *GoCloak) UpdatePermission(ctx context.Context, token, realm, idOfClient string, permission PermissionRepresentation) error {
	const errMessage = "could not update permission"

	if NilOrEmpty(permission.ID) {
		return errors.New("ID of a permission required")
	}
	if NilOrEmpty(permission.Type) {
		return errors.New("type of a permission required")
	}
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(permission),
		http.MethodPut, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", *permission.Type, *permission.ID), errMessage)
}

// DeletePermission deletes a permission of the client's resource server.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClient [string],
// permissionID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeletePermission(ctx context.Context, token, realm, idOfClient, permissionID string) error {
	const errMessage = "could not delete permission"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "clients", idOfClient, "authz", "resource-server", "permission", permissionID), errMessage)
}

// ---------------
// Credentials API
// ---------------

// GetCredentialRegistrators returns the credential registrators of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]string] registrator names, [error] if the call fails.
func (g *GoCloak) GetCredentialRegistrators(ctx context.Context, token, realm string) ([]string, error) {
	const errMessage = "could not get user credential registrators"

	return doRequest[[]string](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "credential-registrators"), errMessage)
}

// GetConfiguredUserStorageCredentialTypes returns the credential types provided by the user storage where the
// user is stored.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [[]string] credential types, [error] if the call fails.
func (g *GoCloak) GetConfiguredUserStorageCredentialTypes(ctx context.Context, token, realm, userID string) ([]string, error) {
	const errMessage = "could not get user credential registrators"

	return doRequest[[]string](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "configured-user-storage-credential-types"), errMessage)
}

// GetCredentials returns the credentials of the given user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string].
// Output: [[]*CredentialRepresentation] credentials, [error] if the call fails.
func (g *GoCloak) GetCredentials(ctx context.Context, token, realm, userID string) ([]*CredentialRepresentation, error) {
	const errMessage = "could not get user credentials"

	return doRequest[[]*CredentialRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "users", userID, "credentials"), errMessage)
}

// DeleteCredentials deletes the given credential of the user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], credentialID [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteCredentials(ctx context.Context, token, realm, userID, credentialID string) error {
	const errMessage = "could not delete user credentials"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "users", userID, "credentials", credentialID), errMessage)
}

// UpdateCredentialUserLabel sets the user label of the given credential.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], credentialID [string],
// userLabel [string].
// Output: [error] if the call fails.
func (g *GoCloak) UpdateCredentialUserLabel(ctx context.Context, token, realm, userID, credentialID, userLabel string) error {
	const errMessage = "could not update credential label for a user"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetHeader("Content-Type", "text/plain").
		SetBody(userLabel),
		http.MethodPut, g.getAdminRealmURL(realm, "users", userID, "credentials", credentialID, "userLabel"), errMessage)
}

// DisableAllCredentialsByType disables all credentials of the given types for the user.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string],
// types [[]string] credential types.
// Output: [error] if the call fails.
func (g *GoCloak) DisableAllCredentialsByType(ctx context.Context, token, realm, userID string, types []string) error {
	const errMessage = "could not update disable credentials"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(types),
		http.MethodPut, g.getAdminRealmURL(realm, "users", userID, "disable-credential-types"), errMessage)
}

// MoveCredentialBehind moves a credential to the position behind another credential.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], credentialID [string],
// newPreviousCredentialID [string].
// Output: [error] if the call fails.
func (g *GoCloak) MoveCredentialBehind(ctx context.Context, token, realm, userID, credentialID, newPreviousCredentialID string) error {
	const errMessage = "could not move credential"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPost, g.getAdminRealmURL(realm, "users", userID, "credentials", credentialID, "moveAfter", newPreviousCredentialID), errMessage)
}

// MoveCredentialToFirst moves a credential to the first position in the user's credential list.
// Input: ctx [context.Context], token [string] admin token, realm [string], userID [string], credentialID [string].
// Output: [error] if the call fails.
func (g *GoCloak) MoveCredentialToFirst(ctx context.Context, token, realm, userID, credentialID string) error {
	const errMessage = "could not move credential"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPost, g.getAdminRealmURL(realm, "users", userID, "credentials", credentialID, "moveToFirst"), errMessage)
}

// GetEvents returns the user events of the realm matching the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetEventsParams].
// Output: [[]*EventRepresentation] events, [error] if the call fails.
func (g *GoCloak) GetEvents(ctx context.Context, token string, realm string, params GetEventsParams) ([]*EventRepresentation, error) {
	const errMessage = "could not get events"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*EventRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams).
		SetQueryParamsFromValues(url.Values{"type": params.Type}),
		http.MethodGet, g.getAdminRealmURL(realm, "events"), errMessage)
}

// GetAdminEvents returns the admin events of the realm matching the given query params.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetAdminEventsParams].
// Output: [[]*AdminEventRepresentation] admin events, [error] if the call fails.
func (g *GoCloak) GetAdminEvents(ctx context.Context, token string, realm string, params GetAdminEventsParams) ([]*AdminEventRepresentation, error) {
	const errMessage = "could not get admin events"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*AdminEventRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams).
		SetQueryParamsFromValues(url.Values{"operationTypes": params.OperationTypes, "resourceTypes": params.ResourceTypes}),
		http.MethodGet, g.getAdminRealmURL(realm, "admin-events"), errMessage)
}

// GetClientScopesScopeMappingsRealmRolesAvailable returns the realm-level roles that can still be added to the
// client scope's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], clientScopeID [string].
// Output: [[]*Role] available realm roles, [error] if the call fails.
func (g *GoCloak) GetClientScopesScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, clientScopeID string) ([]*Role, error) {
	const errMessage = "could not get available realm-level roles with the client-scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes", clientScopeID, "scope-mappings", "realm", "available"), errMessage)
}

// GetClientScopesScopeMappingsRealmRoles returns the realm-level roles in the client scope's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], clientScopeID [string].
// Output: [[]*Role] realm roles, [error] if the call fails.
func (g *GoCloak) GetClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, clientScopeID string) ([]*Role, error) {
	const errMessage = "could not get realm-level roles with the client-scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes", clientScopeID, "scope-mappings", "realm"), errMessage)
}

// DeleteClientScopesScopeMappingsRealmRoles removes realm-level roles from the client scope's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], clientScopeID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, clientScopeID string, roles []Role) error {
	const errMessage = "could not delete realm-level roles from the client-scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "client-scopes", clientScopeID, "scope-mappings", "realm"), errMessage)
}

// CreateClientScopesScopeMappingsRealmRoles adds realm-level roles to the client scope's scope mappings.
// Input: ctx [context.Context], token [string] admin token, realm [string], clientScopeID [string], roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) CreateClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, clientScopeID string, roles []Role) error {
	const errMessage = "could not create realm-level roles to the client-scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "client-scopes", clientScopeID, "scope-mappings", "realm"), errMessage)
}

// RegisterRequiredAction registers a required action provider in the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string],
// requiredAction [RequiredActionProviderRepresentation].
// Output: [error] if the call fails.
func (g *GoCloak) RegisterRequiredAction(ctx context.Context, token string, realm string, requiredAction RequiredActionProviderRepresentation) error {
	const errMessage = "could not create required action"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(requiredAction),
		http.MethodPost, g.getAdminRealmURL(realm, "authentication", "register-required-action"), errMessage)
}

// GetUnregisteredRequiredActions returns the required action providers that are not yet registered in the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*UnregisteredRequiredActionProviderRepresentation] unregistered actions, [error] if the call fails.
func (g *GoCloak) GetUnregisteredRequiredActions(ctx context.Context, token string, realm string) ([]*UnregisteredRequiredActionProviderRepresentation, error) {
	const errMessage = "could not get unregistered required actions"

	return doRequest[[]*UnregisteredRequiredActionProviderRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "unregistered-required-actions"), errMessage)
}

// GetRequiredActions returns the registered required actions of the realm.
// Input: ctx [context.Context], token [string] admin token, realm [string].
// Output: [[]*RequiredActionProviderRepresentation] required actions, [error] if the call fails.
func (g *GoCloak) GetRequiredActions(ctx context.Context, token string, realm string) ([]*RequiredActionProviderRepresentation, error) {
	const errMessage = "could not get required actions"
	return doRequest[[]*RequiredActionProviderRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "required-actions"), errMessage)
}

// GetRequiredAction fetches the required action with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string].
// Output: [*RequiredActionProviderRepresentation] required action, [error] if alias is empty or the call fails.
func (g *GoCloak) GetRequiredAction(ctx context.Context, token string, realm string, alias string) (*RequiredActionProviderRepresentation, error) {
	const errMessage = "could not get required action"

	if alias == "" {
		return nil, errors.New("alias is required for getting a required action")
	}

	return doRequestPtr[RequiredActionProviderRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "authentication", "required-actions", alias), errMessage)
}

// UpdateRequiredAction updates a required action, addressed by its ProviderID.
// Input: ctx [context.Context], token [string] admin token, realm [string],
// requiredAction [RequiredActionProviderRepresentation] (ProviderID required).
// Output: [error] if the provider ID is missing or the call fails.
func (g *GoCloak) UpdateRequiredAction(ctx context.Context, token string, realm string, requiredAction RequiredActionProviderRepresentation) error {
	const errMessage = "could not update required action"

	if NilOrEmpty(requiredAction.ProviderID) {
		return errors.New("providerId is required for updating a required action")
	}
	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(requiredAction),
		http.MethodPut, g.getAdminRealmURL(realm, "authentication", "required-actions", *requiredAction.ProviderID), errMessage)
}

// DeleteRequiredAction deletes the required action with the given alias.
// Input: ctx [context.Context], token [string] admin token, realm [string], alias [string].
// Output: [error] if alias is empty or the call fails.
func (g *GoCloak) DeleteRequiredAction(ctx context.Context, token string, realm string, alias string) error {
	const errMessage = "could not delete required action"

	if alias == "" {
		return errors.New("alias is required for deleting a required action")
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "authentication", "required-actions", alias), errMessage)
}

// CreateClientScopesScopeMappingsClientRoles attaches client roles to a client scope (not client's scope).
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClientScope [string],
// idOfClient [string] client owning the roles, roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) CreateClientScopesScopeMappingsClientRoles(
	ctx context.Context, token, realm, idOfClientScope, idOfClient string, roles []Role,
) error {
	const errMessage = "could not create client-level roles to the client-scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodPost, g.getAdminRealmURL(realm, "client-scopes", idOfClientScope, "scope-mappings", "clients", idOfClient), errMessage)
}

// GetClientScopesScopeMappingsClientRolesAvailable returns available (i.e. not attached via
// CreateClientScopesScopeMappingsClientRoles) client roles for a specific client, for a client scope
// (not client's scope).
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClientScope [string],
// idOfClient [string] client owning the roles.
// Output: [[]*Role] available client roles, [error] if the call fails.
func (g *GoCloak) GetClientScopesScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*Role, error) {
	const errMessage = "could not get available client-level roles with the client-scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes", idOfClientScope, "scope-mappings", "clients", idOfClient, "available"), errMessage)
}

// GetClientScopesScopeMappingsClientRoles returns attached client roles for a specific client, for a client scope
// (not client's scope).
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClientScope [string],
// idOfClient [string] client owning the roles.
// Output: [[]*Role] attached client roles, [error] if the call fails.
func (g *GoCloak) GetClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*Role, error) {
	const errMessage = "could not get client-level roles with the client-scope"

	return doRequest[[]*Role](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "client-scopes", idOfClientScope, "scope-mappings", "clients", idOfClient), errMessage)
}

// DeleteClientScopesScopeMappingsClientRoles removes attachment of client roles from a client scope
// (not client's scope).
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfClientScope [string],
// idOfClient [string] client owning the roles, roles [[]Role].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string, roles []Role) error {
	const errMessage = "could not delete client-level roles from the client-scope"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(roles),
		http.MethodDelete, g.getAdminRealmURL(realm, "client-scopes", idOfClientScope, "scope-mappings", "clients", idOfClient), errMessage)
}

// RevokeToken revokes the passed token via the OIDC revoke endpoint. The token can either be an access or
// refresh token.
// Input: ctx [context.Context], realm [string], clientID [string], clientSecret [string],
// refreshToken [string] token to revoke.
// Output: [error] if the call fails.
func (g *GoCloak) RevokeToken(ctx context.Context, realm, clientID, clientSecret, refreshToken string) error {
	const errMessage = "could not revoke token"

	return doNoResult(g.GetRequestWithBasicAuth(ctx, clientID, clientSecret).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"token":         refreshToken,
		}),
		http.MethodPost, g.getOIDCURL(realm, g.Config.revokeEndpoint), errMessage)
}

// UpdateUsersManagementPermissions updates the realm's users management permissions.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string],
// managementPermissions [ManagementPermissionRepresentation].
// Output: [*ManagementPermissionRepresentation] updated permissions, [error] if the call fails.
func (g *GoCloak) UpdateUsersManagementPermissions(ctx context.Context, accessToken, realm string, managementPermissions ManagementPermissionRepresentation) (*ManagementPermissionRepresentation, error) {
	const errMessage = "could not update users management permissions"

	return doRequestPtr[ManagementPermissionRepresentation](g.GetRequestWithBearerAuth(ctx, accessToken).
		SetBody(managementPermissions),
		http.MethodPut, g.getAdminRealmURL(realm, "users-management-permissions"), errMessage)
}

// GetUsersManagementPermissions returns the realm's users management permissions.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string].
// Output: [*ManagementPermissionRepresentation] permissions, [error] if the call fails.
func (g *GoCloak) GetUsersManagementPermissions(ctx context.Context, accessToken, realm string) (*ManagementPermissionRepresentation, error) {
	const errMessage = "could not get users management permissions"

	return doRequestPtr[ManagementPermissionRepresentation](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getAdminRealmURL(realm, "users-management-permissions"), errMessage)
}

// CreateOrganization creates a new organization.
// Input: ctx [context.Context], token [string] admin token, realm [string],
// organization [OrganizationRepresentation].
// Output: [string] ID of the created organization, [error] if the call fails.
func (g *GoCloak) CreateOrganization(ctx context.Context, token, realm string, organization OrganizationRepresentation) (string, error) {
	const errMessage = "could not create organization"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(organization),
		g.getAdminRealmURL(realm, "organizations"), errMessage)
}

// GetOrganizations returns a paginated list of organizations filtered according to the specified parameters.
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetOrganizationsParams].
// Output: [[]*OrganizationRepresentation] organizations, [error] if the call fails.
func (g *GoCloak) GetOrganizations(ctx context.Context, token, realm string, params GetOrganizationsParams) ([]*OrganizationRepresentation, error) {
	const errMessage = "could not get organizations"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*OrganizationRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations"), errMessage)
}

// GetOrganizationByID fetches the organization with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string].
// Output: [*OrganizationRepresentation] organization, [error] if the call fails.
func (g *GoCloak) GetOrganizationByID(ctx context.Context, token, realm, idOfOrganization string) (*OrganizationRepresentation, error) {
	const errMessage = "could not find organization"
	return doRequestPtr[OrganizationRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization), errMessage)
}

// UpdateOrganization updates the given organization.
// Input: ctx [context.Context], token [string] admin token, realm [string],
// organization [OrganizationRepresentation] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateOrganization(ctx context.Context, token, realm string, organization OrganizationRepresentation) error {
	const errMessage = "could not update organization"

	if NilOrEmpty(organization.ID) {
		return fmt.Errorf("%s: ID of an organization required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(organization),
		http.MethodPut, g.getAdminRealmURL(realm, "organizations", PString(organization.ID)), errMessage)
}

// DeleteOrganization deletes the organization with the given ID.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteOrganization(ctx context.Context, token, realm, idOfOrganization string) error {
	const errMessage = "could not delete organization"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "organizations", idOfOrganization), errMessage)
}

// InviteUserToOrganization invites an existing user or sends a registration link to a new user, based on the
// provided e-mail address.
// If the user with the given e-mail address exists, it sends an invitation link, otherwise it sends a registration
// link. An invitation email will be sent to the user so SMTP settings are required in keycloak
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// user [OrganizationInviteUserParams] (Email required).
// Output: [error] if the email is missing or the call fails.
func (g *GoCloak) InviteUserToOrganization(ctx context.Context, token, realm string, idOfOrganization string, user OrganizationInviteUserParams) error {
	const errMessage = "could not invite organization user"

	err := checkOrganizationInviteUserParams(user)
	if err != nil {
		return err
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetFormData(user.FormData()),
		http.MethodPost, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members", "invite-user"), errMessage)
}

// checkOrganizationInviteUserParams checks that the invitation has an email address.
// Input: invitation [OrganizationInviteUserParams].
// Output: [error] if the email is missing, nil if valid.
func checkOrganizationInviteUserParams(invitation OrganizationInviteUserParams) error {
	if invitation.Email == "" {
		return errors.New("email required to invite organization user")
	}

	return nil
}

// InviteUserToOrganizationByID invites an existing user to the organization, using the specified user id.
// An invitation email will be sent to the user so SMTP settings are required in keycloak
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfUser [string].
// Output: [error] if the call fails.
func (g *GoCloak) InviteUserToOrganizationByID(ctx context.Context, token, realm, idOfOrganization, idOfUser string) error {
	const errMessage = "could not invite user to organization by id"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetFormData(map[string]string{
			"id": idOfUser,
		}),
		http.MethodPost, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members", "invite-existing-user"), errMessage)
}

// AddUserToOrganization adds the user with the specified id as a member of the organization.
// If no user is found, or if it is already associated with the organization, an error response is returned.
// No invitation email is sent to the user
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfUser [string].
// Output: [error] if the call fails.
func (g *GoCloak) AddUserToOrganization(ctx context.Context, token, realm, idOfOrganization, idOfUser string) error {
	const errMessage = "could not add user to organization"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(idOfUser),
		http.MethodPost, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members"), errMessage)
}

// GetOrganizationMemberCount returns the number of members in the organization.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string].
// Output: [int] member count, [error] if the call fails.
func (g *GoCloak) GetOrganizationMemberCount(ctx context.Context, token, realm, idOfOrganization string) (int, error) {
	const errMessage = "could not get organization members count"

	return doRequest[int](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members", "count"), errMessage)
}

// GetOrganizationMembers returns a paginated list of organization members filtered according to the specified
// parameters.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// params [GetMembersParams].
// Output: [[]*MemberRepresentation] members, [error] if the call fails.
func (g *GoCloak) GetOrganizationMembers(ctx context.Context, token, realm, idOfOrganization string, params GetMembersParams) ([]*MemberRepresentation, error) {
	const errMessage = "could not get organization members"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*MemberRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members"), errMessage)
}

// GetOrganizationMemberByID returns the member of the organization with the specified id.
// If the user is not a member of the organization, an error response with status NOT_FOUND is returned.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfUser [string].
// Output: [*MemberRepresentation] member, [error] if the call fails.
func (g *GoCloak) GetOrganizationMemberByID(ctx context.Context, token, realm, idOfOrganization, idOfUser string) (*MemberRepresentation, error) {
	const errMessage = "could not get organization member by ID"
	return doRequestPtr[MemberRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members", idOfUser), errMessage)
}

// GetMemberAssociatedOrganizations returns the organizations associated with the user that has the specified id.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfUser [string],
// briefRepresentation [bool].
// Output: [[]*OrganizationRepresentation] organizations, [error] if the call fails.
func (g *GoCloak) GetMemberAssociatedOrganizations(ctx context.Context, token, realm, idOfUser string, briefRepresentation bool) ([]*OrganizationRepresentation, error) {
	const errMessage = "could not get member's associated organizations"
	return doRequest[[]*OrganizationRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParam("briefRepresentation", strconv.FormatBool(briefRepresentation)),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", "members", idOfUser, "organizations"), errMessage)
}

// GetOrganizationMemberOrganizations returns the organizations of the given member, queried via the given
// organization.
// Input: ctx [context.Context], accessToken [string] admin token, realm [string], idOfOrganization [string],
// idOfUser [string].
// Output: [[]*OrganizationRepresentation] organizations, [error] if an ID is empty or the call fails.
func (g *GoCloak) GetOrganizationMemberOrganizations(ctx context.Context, accessToken, realm, idOfOrganization, idOfUser string) ([]*OrganizationRepresentation, error) {
	const errMessage = "could not get organization member organizations"

	if idOfOrganization == "" {
		return nil, fmt.Errorf("%s: organizationID shall not be empty", errMessage)
	}

	if idOfUser == "" {
		return nil, fmt.Errorf("%s: userID shall not be empty", errMessage)
	}

	return doRequest[[]*OrganizationRepresentation](g.GetRequestWithBearerAuth(ctx, accessToken), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members", idOfUser, "organizations"), errMessage)
}

// RemoveUserFromOrganization removes the user with the specified id from the organization.
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfUser [string].
// Output: [error] if the call fails.
func (g *GoCloak) RemoveUserFromOrganization(ctx context.Context, token, realm, idOfOrganization, idOfUser string) error {
	const errMessage = "could not remove user from organization"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members", idOfUser), errMessage)
}

// AddIdentityProviderToOrganization adds the identity provider with the specified alias to the organization.
// POST /admin/realms/{realm}/organizations/{id}/identity-providers
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// identityProviderAlias [string].
// Output: [error] if the call fails.
func (g *GoCloak) AddIdentityProviderToOrganization(ctx context.Context, token, realm string, idOfOrganization, identityProviderAlias string) error {
	const errMessage = "could not add identity provider to organization"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(identityProviderAlias),
		http.MethodPost, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "identity-providers"), errMessage)
}

// GetOrganizationIdentityProviders returns all identity providers associated with the organization.
// GET /admin/realms/{realm}/organizations/{id}/identity-providers
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string].
// Output: [[]*IdentityProviderRepresentation] identity providers, [error] if the call fails.
func (g *GoCloak) GetOrganizationIdentityProviders(ctx context.Context, token, realm, idOfOrganization string) ([]*IdentityProviderRepresentation, error) {
	const errMessage = "could not get organization identity providers"
	return doRequest[[]*IdentityProviderRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "identity-providers"), errMessage)
}

// GetOrganizationIdentityProvider returns the identity provider with the specified alias associated with the
// organization.
// GET /admin/realms/{realm}/organizations/{id}/identity-providers/{alias}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// alias [string].
// Output: [*IdentityProviderRepresentation] identity provider, [error] if the call fails.
func (g *GoCloak) GetOrganizationIdentityProvider(ctx context.Context, token, realm, idOfOrganization, alias string) (*IdentityProviderRepresentation, error) {
	const errMessage = "could not get organization identity provider"
	return doRequestPtr[IdentityProviderRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "identity-providers", alias), errMessage)
}

// RemoveIdentityProviderFromOrganization removes the identity provider with the specified alias from the
// organization.
// DELETE /admin/realms/{realm}/organizations/{id}/identity-providers/{alias}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// alias [string].
// Output: [error] if the call fails.
func (g *GoCloak) RemoveIdentityProviderFromOrganization(ctx context.Context, token, realm, idOfOrganization, alias string) error {
	const errMessage = "could not remove identity provider from organization"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "identity-providers", alias), errMessage)
}

// GetOrganizationCount returns the number of organizations in the realm.
// GET /admin/realms/{realm}/organizations/count
// Input: ctx [context.Context], token [string] admin token, realm [string], params [GetOrganizationCountParams].
// Output: [int64] organization count, [error] if the call fails.
func (g *GoCloak) GetOrganizationCount(ctx context.Context, token, realm string, params GetOrganizationCountParams) (int64, error) {
	const errMessage = "could not get organization count"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[int64](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", "count"), errMessage)
}

// GetOrganizationMemberGroups returns the groups the member with the specified id belongs to within the
// organization.
// GET /admin/realms/{realm}/organizations/{id}/members/{member-id}/groups
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfUser [string], briefRepresentation [bool].
// Output: [[]*Group] groups, [error] if the call fails.
func (g *GoCloak) GetOrganizationMemberGroups(ctx context.Context, token, realm, idOfOrganization, idOfUser string, briefRepresentation bool) ([]*Group, error) {
	const errMessage = "could not get organization member groups"
	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParam("briefRepresentation", strconv.FormatBool(briefRepresentation)),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "members", idOfUser, "groups"), errMessage)
}

// GetOrganizationInvitations returns a list of pending invitations for the organization.
// GET /admin/realms/{realm}/organizations/{id}/invitations
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// params [GetOrganizationInvitationsParams].
// Output: [[]*OrganizationInvitationRepresentation] invitations, [error] if the call fails.
func (g *GoCloak) GetOrganizationInvitations(ctx context.Context, token, realm, idOfOrganization string, params GetOrganizationInvitationsParams) ([]*OrganizationInvitationRepresentation, error) {
	const errMessage = "could not get organization invitations"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*OrganizationInvitationRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "invitations"), errMessage)
}

// GetOrganizationInvitationByID returns the invitation with the specified id for the organization.
// GET /admin/realms/{realm}/organizations/{id}/invitations/{invitation-id}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfInvitation [string].
// Output: [*OrganizationInvitationRepresentation] invitation, [error] if the call fails.
func (g *GoCloak) GetOrganizationInvitationByID(ctx context.Context, token, realm, idOfOrganization, idOfInvitation string) (*OrganizationInvitationRepresentation, error) {
	const errMessage = "could not get organization invitation"
	return doRequestPtr[OrganizationInvitationRepresentation](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "invitations", idOfInvitation), errMessage)
}

// DeleteOrganizationInvitation removes the invitation with the specified id from the organization.
// DELETE /admin/realms/{realm}/organizations/{id}/invitations/{invitation-id}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfInvitation [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteOrganizationInvitation(ctx context.Context, token, realm, idOfOrganization, idOfInvitation string) error {
	const errMessage = "could not delete organization invitation"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "invitations", idOfInvitation), errMessage)
}

// ResendOrganizationInvitation resends the invitation with the specified id.
// Keycloak replaces the invitation on resend: the old id is gone afterwards and the new invitation
// has to be looked up again, e.g. via GetOrganizationInvitations filtered by email.
// POST /admin/realms/{realm}/organizations/{id}/invitations/{invitation-id}/resend
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfInvitation [string].
// Output: [error] if the call fails.
func (g *GoCloak) ResendOrganizationInvitation(ctx context.Context, token, realm, idOfOrganization, idOfInvitation string) error {
	const errMessage = "could not resend organization invitation"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPost, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "invitations", idOfInvitation, "resend"), errMessage)
}

// AddOrganizationGroup creates a new top-level group in the organization.
// POST /admin/realms/{realm}/organizations/{id}/groups
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// group [Group].
// Output: [string] ID of the created group, [error] if the call fails.
func (g *GoCloak) AddOrganizationGroup(ctx context.Context, token, realm, idOfOrganization string, group Group) (string, error) {
	const errMessage = "could not add organization group"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(group),
		g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups"), errMessage)
}

// GetOrganizationGroups returns the groups associated with the organization.
// GET /admin/realms/{realm}/organizations/{id}/groups
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// params [GetOrganizationGroupsParams].
// Output: [[]*Group] groups, [error] if the call fails.
func (g *GoCloak) GetOrganizationGroups(ctx context.Context, token, realm, idOfOrganization string, params GetOrganizationGroupsParams) ([]*Group, error) {
	const errMessage = "could not get organization groups"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups"), errMessage)
}

// GetOrganizationGroupByPath returns the organization group matching the given path.
// GET /admin/realms/{realm}/organizations/{id}/groups/group-by-path/{path}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// groupPath [string], subGroupsCount [bool].
// Output: [*Group] group, [error] if the call fails.
func (g *GoCloak) GetOrganizationGroupByPath(ctx context.Context, token, realm, idOfOrganization, groupPath string, subGroupsCount bool) (*Group, error) {
	const errMessage = "could not get organization group by path"
	return doRequestPtr[Group](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParam("subGroupsCount", strconv.FormatBool(subGroupsCount)),
		http.MethodGet, g.getAdminRealmURL(realm, append([]string{"organizations", idOfOrganization, "groups", "group-by-path"},
			groupPathSegments(groupPath)...)...), errMessage)
}

// GetOrganizationGroupByID returns the organization group with the specified id.
// GET /admin/realms/{realm}/organizations/{id}/groups/{group-id}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfGroup [string].
// Output: [*Group] group, [error] if the call fails.
func (g *GoCloak) GetOrganizationGroupByID(ctx context.Context, token, realm, idOfOrganization, idOfGroup string) (*Group, error) {
	const errMessage = "could not get organization group"
	return doRequestPtr[Group](g.GetRequestWithBearerAuth(ctx, token), http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", idOfGroup), errMessage)
}

// UpdateOrganizationGroup updates the organization group with the specified id.
// PUT /admin/realms/{realm}/organizations/{id}/groups/{group-id}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// group [Group] (ID required).
// Output: [error] if the ID is missing or the call fails.
func (g *GoCloak) UpdateOrganizationGroup(ctx context.Context, token, realm, idOfOrganization string, group Group) error {
	const errMessage = "could not update organization group"

	if NilOrEmpty(group.ID) {
		return fmt.Errorf("%s: ID of a group required", errMessage)
	}

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(group),
		http.MethodPut, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", PString(group.ID)), errMessage)
}

// DeleteOrganizationGroup removes the organization group with the specified id.
// DELETE /admin/realms/{realm}/organizations/{id}/groups/{group-id}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfGroup [string].
// Output: [error] if the call fails.
func (g *GoCloak) DeleteOrganizationGroup(ctx context.Context, token, realm, idOfOrganization, idOfGroup string) error {
	const errMessage = "could not delete organization group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", idOfGroup), errMessage)
}

// GetOrganizationGroupSubgroups returns the subgroups of the specified organization group.
// GET /admin/realms/{realm}/organizations/{id}/groups/{group-id}/children
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfGroup [string], params [GetOrganizationGroupSubgroupsParams].
// Output: [[]*Group] subgroups, [error] if the call fails.
func (g *GoCloak) GetOrganizationGroupSubgroups(ctx context.Context, token, realm, idOfOrganization, idOfGroup string, params GetOrganizationGroupSubgroupsParams) ([]*Group, error) {
	const errMessage = "could not get organization group subgroups"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*Group](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", idOfGroup, "children"), errMessage)
}

// AddOrganizationSubgroup adds a subgroup to the specified organization group.
// POST /admin/realms/{realm}/organizations/{id}/groups/{group-id}/children
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfGroup [string] parent group, group [Group].
// Output: [string] ID of the created subgroup, [error] if the call fails.
func (g *GoCloak) AddOrganizationSubgroup(ctx context.Context, token, realm, idOfOrganization, idOfGroup string, group Group) (string, error) {
	const errMessage = "could not add organization subgroup"

	return doCreate(g.GetRequestWithBearerAuth(ctx, token).
		SetBody(group),
		g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", idOfGroup, "children"), errMessage)
}

// GetOrganizationGroupMembers returns the members of the specified organization group.
// GET /admin/realms/{realm}/organizations/{id}/groups/{group-id}/members
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfGroup [string], params [GetOrganizationMembersParams].
// Output: [[]*MemberRepresentation] members, [error] if the call fails.
func (g *GoCloak) GetOrganizationGroupMembers(ctx context.Context, token, realm, idOfOrganization, idOfGroup string, params GetOrganizationMembersParams) ([]*MemberRepresentation, error) {
	const errMessage = "could not get organization group members"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return doRequest[[]*MemberRepresentation](g.GetRequestWithBearerAuth(ctx, token).
		SetQueryParams(queryParams),
		http.MethodGet, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", idOfGroup, "members"), errMessage)
}

// AddMemberToOrganizationGroup adds the user with the specified id to the organization group.
// PUT /admin/realms/{realm}/organizations/{id}/groups/{group-id}/members/{user-id}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfGroup [string], idOfUser [string].
// Output: [error] if the call fails.
func (g *GoCloak) AddMemberToOrganizationGroup(ctx context.Context, token, realm, idOfOrganization, idOfGroup, idOfUser string) error {
	const errMessage = "could not add member to organization group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodPut, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", idOfGroup, "members", idOfUser), errMessage)
}

// RemoveMemberFromOrganizationGroup removes the user with the specified id from the organization group.
// DELETE /admin/realms/{realm}/organizations/{id}/groups/{group-id}/members/{user-id}
// Input: ctx [context.Context], token [string] admin token, realm [string], idOfOrganization [string],
// idOfGroup [string], idOfUser [string].
// Output: [error] if the call fails.
func (g *GoCloak) RemoveMemberFromOrganizationGroup(ctx context.Context, token, realm, idOfOrganization, idOfGroup, idOfUser string) error {
	const errMessage = "could not remove member from organization group"

	return doNoResult(g.GetRequestWithBearerAuth(ctx, token), http.MethodDelete, g.getAdminRealmURL(realm, "organizations", idOfOrganization, "groups", idOfGroup, "members", idOfUser), errMessage)
}
