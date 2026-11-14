package gocloak

// Unit tests that run without a Keycloak instance, using httptest servers.

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// jwkFor returns the JWK representation of an RSA public key.
// Input: kid [string], key [*rsa.PublicKey]. Output: [CertResponseKey].
func jwkFor(kid string, key *rsa.PublicKey) CertResponseKey {
	return CertResponseKey{
		Kid: new(kid),
		Kty: new("RSA"),
		Alg: new("RS256"),
		Use: new("sig"),
		N:   new(base64.RawURLEncoding.EncodeToString(key.N.Bytes())),
		E:   new(base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())),
	}
}

// signToken signs a minimal token with the given key and kid.
// Input: t [*testing.T], kid [string], key [*rsa.PrivateKey]. Output: [string] signed token.
func signToken(t *testing.T, kid string, key *rsa.PrivateKey) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "user"})
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func TestUnit_CheckForError_InvalidGrantType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Invalid user credentials"}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Login(context.Background(), "client", "secret", "r", "user", "wrong")

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr, "expected *APIError, got %v", err)
	require.Equal(t, http.StatusUnauthorized, apiErr.Code)
	require.Equal(t, APIErrTypeInvalidGrant, apiErr.Type)
}

func TestUnit_DecodeAccessToken_RefetchesCertsAfterKeyRotation(t *testing.T) {
	oldKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	newKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var fetches atomic.Int32
	var rotated atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/realms/r/protocol/openid-connect/certs" {
			http.NotFound(w, r)
			return
		}
		fetches.Add(1)
		key := jwkFor("old", &oldKey.PublicKey)
		if rotated.Load() {
			key = jwkFor("new", &newKey.PublicKey)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CertResponse{Keys: []CertResponseKey{key}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	_, _, err = client.DecodeAccessToken(ctx, signToken(t, "old", oldKey), "r")
	require.NoError(t, err)

	// within the refresh interval an unknown kid must not trigger another fetch
	rotated.Store(true)
	_, _, err = client.DecodeAccessToken(ctx, signToken(t, "new", newKey), "r")
	require.Error(t, err)
	require.Equal(t, int32(1), fetches.Load())

	// after the refresh interval the unknown kid triggers one refetch and succeeds
	prevInterval := certsMinRefreshInterval
	t.Cleanup(func() { certsMinRefreshInterval = prevInterval })
	certsMinRefreshInterval = 0
	_, _, err = client.DecodeAccessToken(ctx, signToken(t, "new", newKey), "r")
	require.NoError(t, err)
	require.Equal(t, int32(2), fetches.Load())
}

func TestUnit_URLSegmentsAreEscaped(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.EscapedPath())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	_, err := client.GetUserByID(ctx, "token", "r", "a/../../users?x=y")
	require.NoError(t, err)
	_, err = client.GetUserByID(ctx, "token", "r", "..")
	require.NoError(t, err)
	_, err = client.GetGroupByPath(ctx, "token", "r", "/parent/child")
	require.NoError(t, err)
	_, err = client.GetGroupByPath(ctx, "token", "r", "/parent/../x")
	require.NoError(t, err)

	require.Equal(t, []string{
		"/admin/realms/r/users/a%2F..%2F..%2Fusers%3Fx=y",
		"/admin/realms/r/users/%2E%2E",
		"/admin/realms/r/group-by-path/parent/child",
		"/admin/realms/r/group-by-path/parent/%2E%2E/x",
	}, paths)
}

func TestUnit_LoginClient_SendsAllScopes(t *testing.T) {
	var scope string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope = r.FormValue("scope")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"x"}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).LoginClient(context.Background(), "client", "secret", "r", "openid", "profile")
	require.NoError(t, err)
	require.Equal(t, "openid profile", scope)
}

func TestUnit_GetServerVersion(t *testing.T) {
	t.Run("missing systemInfo returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		_, err := NewClient(server.URL).GetServerVersion(context.Background(), "token")
		require.Error(t, err)
	})

	t.Run("concurrent calls fetch once", func(t *testing.T) {
		var fetches atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fetches.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"systemInfo":{"version":"26.0.0"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				// t.Errorf, not require: FailNow must not be called outside the test goroutine
				version, err := client.GetServerVersion(context.Background(), "token")
				if err != nil || version != "26.0.0" {
					t.Errorf("GetServerVersion = %q, %v; want 26.0.0, nil", version, err)
				}
			})
		}
		wg.Wait()
		require.Equal(t, int32(1), fetches.Load())
	})
}

// TestUnit_GetComponent_UsesSeparatePathSegments checks that the component ID is its own path segment
// and the slash is not escaped to %2F.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetComponent_UsesSeparatePathSegments(t *testing.T) {
	t.Parallel()
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	}))
	defer server.Close()

	component, err := NewClient(server.URL).GetComponent(context.Background(), "token", "realm", "abc")
	require.NoError(t, err)
	require.Equal(t, "/admin/realms/realm/components/abc", gotPath)
	require.Equal(t, "abc", PString(component.ID))
}

// TestUnit_UpdatePermission_RequiresType checks that a permission without a type is rejected
// before a request is sent instead of panicking.
// Input: t [*testing.T]. Output: none.
func TestUnit_UpdatePermission_RequiresType(t *testing.T) {
	t.Parallel()
	err := NewClient("http://127.0.0.1:0").UpdatePermission(context.Background(), "token", "realm", "client",
		PermissionRepresentation{ID: new("id")})
	require.EqualError(t, err, "type of a permission required")
}

// serverInfoServer returns a server answering /admin/serverinfo with the given version and counting the fetches.
// Input: t [*testing.T], version [string]. Output: [*httptest.Server], [*atomic.Int32] fetch counter.
func serverInfoServer(t *testing.T, version string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var fetches atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/serverinfo" {
			http.NotFound(w, r)
			return
		}
		fetches.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ServerInfoRepresentation{SystemInfo: &SystemInfoRepresentation{Version: new(version)}})
	}))
	t.Cleanup(server.Close)
	return server, &fetches
}

func TestUnit_GetServerVersion_Caching(t *testing.T) {
	t.Parallel()

	t.Run("SetServerVersion skips fetch", func(t *testing.T) {
		t.Parallel()
		server, fetches := serverInfoServer(t, "19.0.0")
		version, err := NewClient(server.URL, SetServerVersion("26.0.0")).GetServerVersion(context.Background(), "token")
		require.NoError(t, err)
		require.Equal(t, "26.0.0", version)
		require.Zero(t, fetches.Load())
	})

	t.Run("error is not cached", func(t *testing.T) {
		t.Parallel()
		var fetches atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if fetches.Add(1) == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"systemInfo":{"version":"26.0.0"}}`))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetServerVersion(context.Background(), "token")
		require.Error(t, err)
		version, err := client.GetServerVersion(context.Background(), "token")
		require.NoError(t, err)
		require.Equal(t, "26.0.0", version)
		require.Equal(t, int32(2), fetches.Load())
	})
}

// TestUnit_GetPolicies_SendsTypeAsQuery checks that the policy type is a query parameter and that no
// serverinfo call is needed, so admins of non-master realms can list policies.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetPolicies_SendsTypeAsQuery(t *testing.T) {
	t.Parallel()
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path+"?"+r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).GetPolicies(context.Background(), "token", "r", "c", GetPolicyParams{Type: new("role")})
	require.NoError(t, err)
	require.Equal(t, []string{"/admin/realms/r/clients/c/authz/resource-server/policy?type=role"}, requests)
}

// TestUnit_GetCerts_KeepsKnownKeysWhenRefreshFails checks that expired certs are still used while Keycloak is
// unreachable and that the next refresh waits for another interval.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetCerts_KeepsKnownKeysWhenRefreshFails(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var fetches atomic.Int32
	var down atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetches.Add(1)
		if down.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CertResponse{Keys: []CertResponseKey{jwkFor("k", &key.PublicKey)}})
	}))
	defer server.Close()

	client := NewClient(server.URL, SetCertCacheInvalidationTime(time.Hour))
	ctx := context.Background()
	token := signToken(t, "k", key)
	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.NoError(t, err)

	// let the cache entry expire, then take Keycloak down
	value, _ := client.certsCache.Load("r")
	entry, _ := value.(certsCacheEntry)
	entry.fetchedAt = time.Now().Add(-2 * time.Hour)
	client.certsCache.Store("r", entry)
	down.Store(true)

	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.NoError(t, err, "known keys must still validate while Keycloak is down")
	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.NoError(t, err)
	require.Equal(t, int32(2), fetches.Load(), "a failed refresh must not be retried on every call")
}

// TestUnit_RequiredFieldsDoNotPanic checks that missing required pointers return an error instead of panicking.
// Input: t [*testing.T]. Output: none.
func TestUnit_RequiredFieldsDoNotPanic(t *testing.T) {
	t.Parallel()
	client := NewClient("http://127.0.0.1:0")

	_, err := client.GetToken(context.Background(), "r", TokenOptions{ClientSecret: new("secret"), GrantType: new("client_credentials")})
	require.EqualError(t, err, "client id required when a client secret is set")

	err = client.ExecuteActionsEmail(context.Background(), "token", "r", ExecuteActionsEmail{Actions: []string{"UPDATE_PASSWORD"}})
	require.EqualError(t, err, "user id required")
}

// TestUnit_EventsSendRepeatedQueryParams checks that list filters are sent as repeated query parameters.
// Input: t [*testing.T]. Output: none.
func TestUnit_EventsSendRepeatedQueryParams(t *testing.T) {
	t.Parallel()
	var queries []url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()
	client := NewClient(server.URL)
	ctx := context.Background()

	_, err := client.GetEvents(ctx, "token", "r", GetEventsParams{Type: []string{"LOGIN", "LOGOUT"}, Max: new(int32(5))})
	require.NoError(t, err)
	_, err = client.GetAdminEvents(ctx, "token", "r", GetAdminEventsParams{
		OperationTypes: []string{"CREATE", "DELETE"}, ResourceTypes: []string{"USER"}, First: new(int32(10)),
	})
	require.NoError(t, err)

	require.Equal(t, url.Values{"type": {"LOGIN", "LOGOUT"}, "max": {"5"}}, queries[0])
	require.Equal(t, url.Values{"operationTypes": {"CREATE", "DELETE"}, "resourceTypes": {"USER"}, "first": {"10"}}, queries[1])
}

// TestUnit_CountFunctionsReturnAPIError checks that count functions return the *APIError itself, like all other methods.
// Input: t [*testing.T]. Output: none.
func TestUnit_CountFunctionsReturnAPIError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	client := NewClient(server.URL)
	ctx := context.Background()

	counts := map[string]func() (int, error){
		"GetGroupsCount": func() (int, error) { return client.GetGroupsCount(ctx, "token", "r", GetGroupsParams{}) },
		"GetUserCount":   func() (int, error) { return client.GetUserCount(ctx, "token", "r", GetUsersParams{}) },
		"GetOrganizationMemberCount": func() (int, error) {
			return client.GetOrganizationMemberCount(ctx, "token", "r", "org")
		},
	}
	for name, count := range counts {
		got, err := count()
		apiErr, ok := err.(*APIError) //nolint:errorlint // the type assertion is what callers rely on
		require.True(t, ok, "%s: want *APIError, got %T", name, err)
		require.Equal(t, http.StatusForbidden, apiErr.Code, name)
		require.Zero(t, got, name)
	}
}

func TestUnit_CheckForError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/empty":
			w.WriteHeader(http.StatusNotFound)
		case "/text":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		case "/badjson":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("not json"))
		case "/message":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"errorMessage":"User exists"}`))
		case "/description":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"bad credentials"}`))
		}
	}))
	t.Cleanup(server.Close) // defer would close it before the parallel subtests run
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()

	tests := []struct {
		name, url    string
		wantCode     int
		wantMessage  string // exact message, unless wantContains is set
		wantContains string
		wantType     APIErrType
	}{
		{name: "transport error", url: closed.URL, wantContains: "msg: ", wantType: APIErrTypeUnknown},
		{name: "4xx empty body", url: server.URL + "/empty", wantCode: 404, wantMessage: "404 Not Found", wantType: APIErrTypeUnknown},
		{name: "non-JSON body", url: server.URL + "/text", wantCode: 500, wantMessage: "500 Internal Server Error", wantType: APIErrTypeUnknown},
		{name: "invalid JSON body", url: server.URL + "/badjson", wantCode: 400, wantMessage: "400 Bad Request", wantType: APIErrTypeUnknown},
		{name: "errorMessage field", url: server.URL + "/message", wantCode: 409, wantMessage: "409 Conflict: User exists", wantType: APIErrTypeUnknown},
		{name: "error_description field", url: server.URL + "/description", wantCode: 401, wantMessage: "401 Unauthorized: invalid_grant: bad credentials", wantType: APIErrTypeInvalidGrant},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp, err := NewClient(server.URL).GetRequest(context.Background()).Get(tc.url)
			err = checkForError(resp, err, "msg")

			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			require.Equal(t, tc.wantCode, apiErr.Code)
			require.Equal(t, tc.wantType, apiErr.Type)
			if tc.wantContains != "" {
				require.Contains(t, apiErr.Message, tc.wantContains)
				return
			}
			require.Equal(t, tc.wantMessage, apiErr.Message)
		})
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		resp, err := NewClient(server.URL).GetRequest(context.Background()).Get(server.URL + "/ok")
		require.NoError(t, checkForError(resp, err, "msg"))
	})
}

func TestUnit_GetRequestWithBasicAuth(t *testing.T) {
	t.Parallel()
	var headers []string
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		headers = append(headers, r.Header.Get("Authorization"))
	}))
	defer server.Close()
	client := NewClient(server.URL)

	_, err := client.GetRequestWithBasicAuth(context.Background(), "my client", "p@ss:w/rd").Post(server.URL)
	require.NoError(t, err)
	// public client: no basic auth
	_, err = client.GetRequestWithBasicAuth(context.Background(), "my client", "").Post(server.URL)
	require.NoError(t, err)

	// RFC 6749 2.3.1: id and secret are form-urlencoded before base64
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("my+client:p%40ss%3Aw%2Frd"))
	require.Equal(t, []string{want, ""}, headers)
}

func TestUnit_GetID(t *testing.T) {
	t.Parallel()
	tests := []struct{ location, want string }{
		{"http://kc/admin/realms/r/users/abc-123", "abc-123"},
		{"abc", "abc"},
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.location, func(t *testing.T) {
			t.Parallel()
			resp := &resty.Response{RawResponse: &http.Response{Header: http.Header{"Location": {tc.location}}}}
			require.Equal(t, tc.want, getID(resp))
		})
	}
}

func TestUnit_FindUsedKey(t *testing.T) {
	t.Parallel()
	keys := []CertResponseKey{{Kid: nil}, {Kid: new("a"), Kty: new("RSA")}, {Kid: new("b"), Kty: new("EC")}}

	key := findUsedKey("b", keys)
	require.NotNil(t, key)
	require.Equal(t, "EC", PString(key.Kty))
	require.Nil(t, findUsedKey("c", keys))
	require.Nil(t, findUsedKey("", keys))
	require.Nil(t, findUsedKey("a", nil))
}

func TestUnit_MakeURL(t *testing.T) {
	t.Parallel()
	require.Equal(t, "http://kc/admin/realms", makeURL("http://kc", "admin", "realms"))
	require.Equal(t, "realms", makeURL("realms"))
	require.Empty(t, makeURL())
}

// TestUnit_CheckForError_UnwrapsTransportError checks that a canceled request keeps context.Canceled
// reachable through errors.Is.
// Input: t [*testing.T]. Output: none.
func TestUnit_CheckForError_UnwrapsTransportError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewClient("http://127.0.0.1:1").GetServerInfo(ctx, "token")
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Zero(t, apiErr.Code)
	require.ErrorIs(t, err, context.Canceled)
}

// TestUnit_GetResourcesClient_SingleDeepRequest checks that resources are fetched in one request with deep=true.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetResourcesClient_SingleDeepRequest(t *testing.T) {
	t.Parallel()
	var queries []url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"_id":"1","name":"a"},{"_id":"2","name":"b"}]`))
	}))
	defer server.Close()

	resources, err := NewClient(server.URL).GetResourcesClient(context.Background(), "token", "r", GetResourceParams{Name: new("a")})
	require.NoError(t, err)
	require.Equal(t, []url.Values{{"deep": {"true"}, "name": {"a"}}}, queries)
	require.Len(t, resources, 2)
	require.Equal(t, "2", PString(resources[1].ID))
}

// TestUnit_LoginClientSignedJWT_AudienceIsPublicIssuer checks that the client assertion audience is the issuer
// Keycloak publishes, not the base URL the client was created with.
// Input: t [*testing.T]. Output: none.
func TestUnit_LoginClientSignedJWT_AudienceIsPublicIssuer(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var audience []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/realms/r":
			_, _ = w.Write([]byte(`{"realm":"r","token-service":"https://auth.example.com/realms/r/protocol/openid-connect"}`))
		case "/realms/r/protocol/openid-connect/token":
			claims := jwt.RegisteredClaims{}
			_, _, err := jwt.NewParser().ParseUnverified(r.FormValue("client_assertion"), &claims)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			audience = claims.Audience
			_, _ = w.Write([]byte(`{"access_token":"x"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, err = NewClient(server.URL).LoginClientSignedJWT(context.Background(), "client", "r", key, jwt.SigningMethodRS256,
		jwt.NewNumericDate(time.Now().Add(time.Minute)))
	require.NoError(t, err)
	require.Equal(t, []string{"https://auth.example.com/realms/r"}, audience)
}

// TestUnit_UpdateRequiresID checks that update methods reject a missing ID before sending a request.
// Input: t [*testing.T]. Output: none.
func TestUnit_UpdateRequiresID(t *testing.T) {
	t.Parallel()
	client := NewClient("http://127.0.0.1:0")
	ctx := context.Background()

	updates := map[string]error{
		"UpdateUser":                      client.UpdateUser(ctx, "token", "r", User{Username: new("bob")}),
		"UpdateClientScope":               client.UpdateClientScope(ctx, "token", "r", ClientScope{}),
		"UpdateComponent":                 client.UpdateComponent(ctx, "token", "r", Component{}),
		"UpdateIdentityProviderMapper":    client.UpdateIdentityProviderMapper(ctx, "token", "r", "alias", IdentityProviderMapper{}),
		"UpdateClientScopeProtocolMapper": client.UpdateClientScopeProtocolMapper(ctx, "token", "r", "scope", ProtocolMappers{}),
	}
	for name, err := range updates {
		require.ErrorContains(t, err, "required", name)
	}
}
