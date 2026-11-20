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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jwkFor returns the JWK representation of an RSA public key.
// Input: kid [string], key [*rsa.PublicKey]. Output: [CertResponseKey].
func jwkFor(kid string, key *rsa.PublicKey) CertResponseKey {
	return CertResponseKey{
		Kid: kid,
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
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
	_, err = client.GetOrganizationGroupByPath(ctx, "token", "r", "org", "/parent/child name", false)
	require.NoError(t, err)

	require.Equal(t, []string{
		"/admin/realms/r/users/a%2F..%2F..%2Fusers%3Fx=y",
		"/admin/realms/r/users/%2E%2E",
		"/admin/realms/r/group-by-path/parent/child",
		"/admin/realms/r/group-by-path/parent/%2E%2E/x",
		"/admin/realms/r/organizations/org/groups/group-by-path/parent/child%20name",
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
		_ = json.NewEncoder(w).Encode(ServerInfoRepresentation{SystemInfo: SystemInfoRepresentation{Version: version}})
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

// ageCerts moves the cached certs of realm back in time and forgets a recent fetch error.
// Input: t [*testing.T], client [*GoCloak], realm [string], age [time.Duration]. Output: none.
func ageCerts(t *testing.T, client *GoCloak, realm string, age time.Duration) {
	t.Helper()
	rc := client.realmCertsFor(realm)
	rc.mu.Lock()
	defer rc.mu.Unlock()
	require.NotNil(t, rc.state)
	state := *rc.state
	state.fetchedAt = time.Now().Add(-age)
	rc.state = &state
	rc.errAt = time.Time{}
}

// TestUnit_GetCerts_StaleKeysAreBounded checks that expired certs are served while a background refresh runs,
// that a failed refresh is not retried on every call, and that the old keys stop being trusted after
// certsStaleFactor * CertsInvalidateTime, so a revoked key cannot stay valid while Keycloak is unreachable.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetCerts_StaleKeysAreBounded(t *testing.T) {
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

	// expired but within the stale window: served at once, refreshed in the background
	ageCerts(t, client, "r", 90*time.Minute)
	down.Store(true)
	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.NoError(t, err, "known keys must still validate during the stale window")
	require.Eventually(t, func() bool {
		_, recentErr := client.realmCertsFor("r").load()
		return recentErr != nil
	}, 5*time.Second, 10*time.Millisecond, "background refresh must run and fail")
	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.NoError(t, err)
	require.Equal(t, int32(2), fetches.Load(), "a failed refresh must not be retried on every call")

	// beyond the stale window the old keys are no longer trusted
	ageCerts(t, client, "r", 3*time.Hour)
	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.Error(t, err)
	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.Error(t, err)
	require.Equal(t, int32(3), fetches.Load(), "a failed fetch must be reported to waiting callers without refetching")

	// Keycloak back: the next fetch after the retry interval succeeds
	down.Store(false)
	ageCerts(t, client, "r", 3*time.Hour)
	_, _, err = client.DecodeAccessToken(ctx, token, "r")
	require.NoError(t, err)
}

// TestUnit_GetCerts_CanceledCallerStillFillsCache checks that a canceled caller returns at once while the detached
// fetch still fills the cache, so one disconnected client cannot be mistaken for a Keycloak outage.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetCerts_CanceledCallerStillFillsCache(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CertResponse{Keys: []CertResponseKey{jwkFor("k", &key.PublicKey)}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.GetCerts(ctx, "r")
	require.ErrorIs(t, err, context.Canceled) // the canceled caller leaves at once on a cold cache

	// the detached fetch still fills the cache, its result is not a stored error
	certs, err := client.GetCerts(context.Background(), "r")
	require.NoError(t, err)
	require.Len(t, certs.Keys, 1)
}

// TestUnit_GetCerts_RespectsCallerDeadline checks that a caller with a short deadline is not blocked by a hanging
// Keycloak for the full cert fetch timeout.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetCerts_RespectsCallerDeadline(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := NewClient(server.URL).GetCerts(ctx, "r")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(start), time.Second)
}

// TestUnit_ValidateAccessToken checks that typ, aud and azp are enforced on top of the signature.
// Input: t [*testing.T]. Output: none.
func TestUnit_ValidateAccessToken(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CertResponse{Keys: []CertResponseKey{jwkFor("k", &key.PublicKey)}})
	}))
	defer server.Close()
	client := NewClient(server.URL)

	exp := time.Now().Add(time.Minute).Unix()
	sign := func(claims jwt.MapClaims) string {
		if v, ok := claims["exp"]; !ok {
			claims["exp"] = exp
		} else if v == nil {
			delete(claims, "exp")
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "k"
		signed, err := token.SignedString(key)
		require.NoError(t, err)
		return signed
	}
	valid := jwt.MapClaims{"typ": "Bearer", "aud": []string{"account", "api"}, "azp": "web"}
	want := TokenValidation{Audience: "api", AuthorizedParty: "web"}

	tests := []struct {
		name       string
		claims     jwt.MapClaims
		validation TokenValidation
		wantErr    bool
	}{
		{"valid", valid, want, false},
		{"string audience", jwt.MapClaims{"typ": "Bearer", "aud": "api"}, TokenValidation{Audience: "api"}, false},
		{"id token", jwt.MapClaims{"typ": "ID", "aud": "api", "azp": "web"}, want, true},
		{"missing typ", jwt.MapClaims{"aud": "api", "azp": "web"}, want, true},
		{"wrong audience", jwt.MapClaims{"typ": "Bearer", "aud": "other", "azp": "web"}, want, true},
		{"wrong azp", jwt.MapClaims{"typ": "Bearer", "aud": "api", "azp": "evil"}, want, true},
		{"custom type", jwt.MapClaims{"typ": "Custom", "azp": "web"}, TokenValidation{AuthorizedParty: "web", Type: "Custom"}, false},
		{"nothing to check", valid, TokenValidation{}, true},
		{"no expiry", jwt.MapClaims{"typ": "Bearer", "aud": "api", "azp": "web", "exp": nil}, want, true},
		{"expired", jwt.MapClaims{"typ": "Bearer", "aud": "api", "azp": "web", "exp": time.Now().Add(-time.Minute).Unix()}, want, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, claims, err := client.ValidateAccessToken(context.Background(), "Bearer "+sign(tc.claims), "r", tc.validation)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, claims)
		})
	}

	_, _, err = client.ValidateAccessToken(context.Background(), sign(jwt.MapClaims{"typ": "Bearer", "azp": "web", "exp": nil}), "r", want)
	require.ErrorIs(t, err, jwt.ErrTokenRequiredClaimMissing)
}

// TestUnit_FormDataKeepsOptions checks that FormData does not modify the caller's options.
// Input: t [*testing.T]. Output: none.
func TestUnit_FormDataKeepsOptions(t *testing.T) {
	t.Parallel()
	options := TokenOptions{Scopes: []string{"a", "b"}}
	require.Equal(t, "a b", options.FormData()["scope"])
	require.Empty(t, options.Scope)
	require.Empty(t, options.ResponseType)

	rpt := RequestingPartyTokenOptions{}
	require.Equal(t, "true", rpt.FormData()["response_include_resource_name"])
	require.Empty(t, rpt.GrantType)
	require.Nil(t, rpt.ResponseIncludeResourceName)
}

// TestUnit_EmptySlicesAreSent checks that empty, non-nil slices and maps are sent, so lists can be cleared on update,
// while nil ones are omitted.
// Input: t [*testing.T]. Output: none.
func TestUnit_EmptySlicesAreSent(t *testing.T) {
	t.Parallel()
	out, err := json.Marshal(User{RequiredActions: []string{}, Attributes: map[string][]string{}})
	require.NoError(t, err)
	require.JSONEq(t, `{"requiredActions":[],"attributes":{}}`, string(out))

	out, err = json.Marshal(User{})
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(out))
}

// TestUnit_RequiredFieldsDoNotPanic checks that missing required pointers return an error instead of panicking.
// Input: t [*testing.T]. Output: none.
func TestUnit_RequiredFieldsDoNotPanic(t *testing.T) {
	t.Parallel()
	client := NewClient("http://127.0.0.1:0")

	_, err := client.GetToken(context.Background(), "r", TokenOptions{ClientSecret: "secret", GrantType: "client_credentials"})
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
		{"http://kc/admin/realms/r/roles/my%20role", "my role"},
		{"http://kc/admin/realms/r/roles/a%2Fb", "a/b"},
		{"http://kc/admin/realms/r/roles/bad%zz", "bad%zz"}, // invalid escape is returned as is
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

// TestUnit_ParseSigningKeys checks that only parsable signature keys with a kid are kept, so encryption keys
// published in the same JWKS can never verify a token signature.
// Input: t [*testing.T]. Output: none.
func TestUnit_ParseSigningKeys(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	sig := jwkFor("sig", &key.PublicKey)
	noUse := jwkFor("no-use", &key.PublicKey)
	noUse.Use = ""
	enc := jwkFor("enc", &key.PublicKey)
	enc.Use = "enc"
	noKid := jwkFor("", &key.PublicKey)
	noKid.Kid = ""
	broken := CertResponseKey{Kid: "broken", Kty: "RSA"}

	keys := parseSigningKeys(&CertResponse{Keys: []CertResponseKey{sig, noUse, enc, noKid, broken}})
	require.Len(t, keys, 2)
	require.Contains(t, keys, "sig")
	require.Contains(t, keys, "no-use")
	require.Equal(t, "RS256", keys["sig"].alg)
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
	var expiry *jwt.NumericDate
	var issuerFetches atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/realms/r":
			issuerFetches.Add(1)
			_, _ = w.Write([]byte(`{"realm":"r","token-service":"https://auth.example.com/realms/r/protocol/openid-connect"}`))
		case "/realms/empty":
			_, _ = w.Write([]byte(`{"realm":"empty"}`))
		case "/realms/r/protocol/openid-connect/token":
			claims := jwt.RegisteredClaims{}
			_, _, err := jwt.NewParser().ParseUnverified(r.FormValue("client_assertion"), &claims)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			audience, expiry = claims.Audience, claims.ExpiresAt
			_, _ = w.Write([]byte(`{"access_token":"x"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Minute))
	for range 2 {
		_, err = client.LoginClientSignedJWT(context.Background(), "client", "r", key, jwt.SigningMethodRS256, expiresAt)
		require.NoError(t, err)
	}
	require.Equal(t, []string{"https://auth.example.com/realms/r"}, audience)
	require.Equal(t, int32(1), issuerFetches.Load(), "the issuer must be cached per realm")

	// without an expiry the assertion still carries exp, Keycloak rejects it otherwise
	_, err = client.LoginClientSignedJWT(context.Background(), "client", "r", key, jwt.SigningMethodRS256, nil)
	require.NoError(t, err)
	require.NotNil(t, expiry)
	require.WithinDuration(t, time.Now().Add(signedJWTDefaultLifetime), expiry.Time, 5*time.Second)

	_, err = client.LoginClientSignedJWT(context.Background(), "client", "empty", key, jwt.SigningMethodRS256, expiresAt)
	require.ErrorContains(t, err, "no token-service")
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

// TestUnit_GetClientServiceAccount_UsesServiceAccountUserID checks that the service account user is
// loaded via serviceAccountUserId from the client representation (Keycloak 27+) and that a failing
// client lookup is reported as an error.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetClientServiceAccount_UsesServiceAccountUserID(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/admin/realms/realm/clients/client-id":
			_, _ = w.Write([]byte(`{"id":"client-id","serviceAccountUserId":"sa-id"}`))
		case "/admin/realms/legacy/clients/client-id/service-account-user":
			_, _ = w.Write([]byte(`{"id":"legacy-sa-id"}`))
		case "/admin/realms/realm/users/sa-id":
			_, _ = w.Write([]byte(`{"id":"sa-id","username":"service-account-client"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	client := NewClient(server.URL)

	user, err := client.GetClientServiceAccount(context.Background(), "token", "realm", "client-id")
	require.NoError(t, err)
	require.Equal(t, "sa-id", PString(user.ID))
	require.Equal(t, "service-account-client", PString(user.Username))

	_, err = client.GetClientServiceAccount(context.Background(), "token", "realm", "missing")
	require.Error(t, err)

	// Keycloak < 27: the legacy endpoint answers in one call
	user, err = client.GetClientServiceAccount(context.Background(), "token", "legacy", "client-id")
	require.NoError(t, err)
	require.Equal(t, "legacy-sa-id", PString(user.ID))
}

// TestUnit_GetCerts_SharesOneFetch checks that concurrent callers on a cold cache share one cert fetch.
// Input: t [*testing.T]. Output: none.
func TestUnit_GetCerts_SharesOneFetch(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var fetches atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetches.Add(1)
		time.Sleep(50 * time.Millisecond) // keep the fetch running while the others arrive
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CertResponse{Keys: []CertResponseKey{jwkFor("k", &key.PublicKey)}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			_, err := client.GetCerts(context.Background(), "r")
			assert.NoError(t, err)
		})
	}
	wg.Wait()
	require.Equal(t, int32(1), fetches.Load())
}

// TestUnit_ValidateAccessToken_LeewayAndCustomClaims checks the clock skew leeway and that typ, aud and azp are
// checked for custom claim structs too.
// Input: t [*testing.T]. Output: none.
func TestUnit_ValidateAccessToken_LeewayAndCustomClaims(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CertResponse{Keys: []CertResponseKey{jwkFor("k", &key.PublicKey)}})
	}))
	defer server.Close()
	client := NewClient(server.URL)

	sign := func(claims jwt.MapClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "k"
		signed, err := token.SignedString(key)
		require.NoError(t, err)
		return signed
	}
	ctx := context.Background()
	exp := time.Now().Add(time.Minute).Unix()

	// expired 5s ago: rejected without leeway, accepted with 10s
	expired := sign(jwt.MapClaims{"typ": "Bearer", "azp": "web", "exp": time.Now().Add(-5 * time.Second).Unix()})
	_, _, err = client.ValidateAccessToken(ctx, expired, "r", TokenValidation{AuthorizedParty: "web"})
	require.ErrorIs(t, err, jwt.ErrTokenExpired)
	_, _, err = client.ValidateAccessToken(ctx, expired, "r", TokenValidation{AuthorizedParty: "web", Leeway: 10 * time.Second})
	require.NoError(t, err)

	// custom claims without typ/azp fields: the checks still read them from the verified payload
	var claims jwt.RegisteredClaims
	_, err = client.ValidateAccessTokenCustomClaims(ctx, sign(jwt.MapClaims{"typ": "ID", "azp": "web", "sub": "u", "exp": exp}), "r",
		&claims, TokenValidation{AuthorizedParty: "web"})
	require.ErrorContains(t, err, "unexpected token type")
	_, err = client.ValidateAccessTokenCustomClaims(ctx, sign(jwt.MapClaims{"typ": "Bearer", "azp": "web", "sub": "u", "exp": exp}), "r",
		&claims, TokenValidation{AuthorizedParty: "web"})
	require.NoError(t, err)
	require.Equal(t, "u", claims.Subject)
}

// TestUnit_PKCE checks the S256 challenge against RFC 7636 Appendix B, the verifier format and that the verifier is
// sent in the code exchange only when set.
// Input: t [*testing.T]. Output: none.
func TestUnit_PKCE(t *testing.T) {
	t.Parallel()
	require.Equal(t, "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", PKCEChallenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"))

	verifier, challenge := NewPKCE()
	require.Len(t, verifier, 43)
	require.Regexp(t, `^[A-Za-z0-9\-._~]+$`, verifier)
	require.Equal(t, PKCEChallenge(verifier), challenge)
	other, _ := NewPKCE()
	require.NotEqual(t, verifier, other)

	require.Equal(t, "x", (&TokenOptions{CodeVerifier: "x"}).FormData()["code_verifier"])
	require.NotContains(t, (&TokenOptions{}).FormData(), "code_verifier")
}
