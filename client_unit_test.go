package gocloak

// Unit tests that run without a Keycloak instance, using httptest servers.

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// jwkFor returns the JWK representation of an RSA public key.
// Input: kid [string], key [*rsa.PublicKey]. Output: [CertResponseKey].
func jwkFor(kid string, key *rsa.PublicKey) CertResponseKey {
	return CertResponseKey{
		Kid: Ptr(kid),
		Kty: Ptr("RSA"),
		Alg: Ptr("RS256"),
		Use: Ptr("sig"),
		N:   Ptr(base64.RawURLEncoding.EncodeToString(key.N.Bytes())),
		E:   Ptr(base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())),
	}
}

// signToken signs a minimal token with the given key and kid.
// Input: t [*testing.T], kid [string], key [*rsa.PrivateKey]. Output: [string] signed token.
func signToken(t *testing.T, kid string, key *rsa.PrivateKey) string {
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
	require.True(t, errors.As(err, &apiErr), "expected *APIError, got %v", err)
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
		require.Equal(t, "/realms/r/protocol/openid-connect/certs", r.URL.Path)
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
	certsMinRefreshInterval = 0
	defer func() { certsMinRefreshInterval = 30 * time.Second }()
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
				result, err := client.compareVersions(context.Background(), "20.0.0", "token")
				require.NoError(t, err)
				require.Equal(t, 1, result)
			})
		}
		wg.Wait()
		require.Equal(t, int32(1), fetches.Load())
	})
}
