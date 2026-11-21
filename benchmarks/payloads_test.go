package benchmarks

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v15"
	"github.com/golang-jwt/jwt/v5"
)

const (
	benchRealm = "gocloak"
	benchKid   = "bench-kid"
)

// payloadSet holds Keycloak-shaped response bodies and request bodies used by the offline benchmarks.
type payloadSet struct {
	key         *rsa.PrivateKey
	accessToken string
	token       []byte // token endpoint response
	certs       []byte // JWKS response
	user        []byte // single user
	users100    []byte // user list, 100 entries
	users1000   []byte // user list, 1000 entries
	groups      []byte // group tree with subgroups
	clients     []byte // clients of the real test realm export
	realm       []byte // full realm export (testdata/gocloak-realm.json)
	createUser  gocloak.User
}

var (
	payloadsOnce sync.Once
	payloadsVal  *payloadSet
)

// payloads builds the payload set once per process.
// Output: [*payloadSet] payloads.
func payloads() *payloadSet {
	payloadsOnce.Do(func() { payloadsVal = buildPayloads() })
	return payloadsVal
}

// buildPayloads generates deterministic-shaped payloads; only the RSA key and token signature are random.
// Output: [*payloadSet] payloads.
func buildPayloads() *payloadSet {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	must(err)
	p := &payloadSet{key: key}

	p.accessToken = signToken(key, 60*time.Minute)
	p.token = mustJSON(gocloak.JWT{
		AccessToken:      p.accessToken,
		IDToken:          signToken(key, 60*time.Minute),
		RefreshToken:     signToken(key, 30*time.Minute),
		ExpiresIn:        300,
		RefreshExpiresIn: 1800,
		TokenType:        "Bearer",
		SessionState:     "7c1ab4c4-9b4e-4d47-9a4f-8e2b8b5e1f00",
		Scope:            "openid profile email",
	})
	p.certs = mustJSON(gocloak.CertResponse{Keys: []gocloak.CertResponseKey{{
		Kid: benchKid, Kty: "RSA", Alg: "RS256", Use: "sig",
		N:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		X5c: []string{strings.Repeat("MIICnTCCAYUCBgGP", 50)},
		X5t: "hY2-3oNcUkfV8BBgqgWZtzYJxSk", X5tS256: "Zt4f1d0wq8y8LgqEo7nN2nX1uVQkqQ3c0b3m1p2Y8kA",
	}}})

	p.user = mustJSON(makeUser(0))
	p.users100 = mustJSON(makeUsers(100))
	p.users1000 = mustJSON(makeUsers(1000))
	p.groups = mustJSON(makeGroups(20, 5))
	p.createUser = makeUser(42)
	p.createUser.ID = nil

	realm, err := os.ReadFile("../testdata/gocloak-realm.json")
	must(err)
	p.realm = realm
	var export struct {
		Clients json.RawMessage `json:"clients"`
	}
	must(json.Unmarshal(realm, &export))
	p.clients = export.Clients
	return p
}

// signToken creates an RS256 token with claims shaped like a Keycloak access token.
// Input: key [*rsa.PrivateKey], ttl [time.Duration]. Output: [string] signed token.
func signToken(key *rsa.PrivateKey, ttl time.Duration) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"exp": now.Add(ttl).Unix(), "iat": now.Unix(), "jti": "onrtac:6f1c1b4e-1f7d-4a5c-9a0e-2d6f5b8c7e11",
		"iss": "http://localhost:8080/realms/gocloak", "aud": []string{"account", "gocloak"},
		"sub": "0f1e2d3c-4b5a-6978-8796-a5b4c3d2e1f0", "typ": "Bearer", "azp": "gocloak",
		"sid": "7c1ab4c4-9b4e-4d47-9a4f-8e2b8b5e1f00", "acr": "1",
		"allowed-origins":    []string{"https://app.example.com"},
		"realm_access":       map[string]any{"roles": []string{"offline_access", "uma_authorization", "default-roles-gocloak"}},
		"resource_access":    map[string]any{"account": map[string]any{"roles": []string{"manage-account", "manage-account-links", "view-profile"}}},
		"scope":              "openid profile email",
		"email_verified":     true,
		"name":               "Test User",
		"preferred_username": "test",
		"given_name":         "Test",
		"family_name":        "User",
		"email":              "test@example.com",
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = benchKid
	s, err := t.SignedString(key)
	must(err)
	return s
}

// makeUser builds a user like the Keycloak admin API returns it in a full representation.
// Input: i [int] index. Output: [gocloak.User].
func makeUser(i int) gocloak.User {
	return gocloak.User{
		ID:               new(fmt.Sprintf("0f1e2d3c-4b5a-6978-8796-%012d", i)),
		CreatedTimestamp: new(int64(1763700000000 + i)),
		Username:         new(fmt.Sprintf("user-%05d", i)),
		Enabled:          new(true),
		Totp:             new(false),
		EmailVerified:    new(i%2 == 0),
		FirstName:        new("Max"),
		LastName:         new(fmt.Sprintf("Mustermann-%d", i)),
		Email:            new(fmt.Sprintf("user-%05d@example.com", i)),
		Attributes: map[string][]string{
			"department": {"engineering"},
			"locale":     {"de"},
			"phone":      {"+49 30 1234567"},
		},
		DisableableCredentialTypes: []any{},
		RequiredActions:            []string{},
		Access: map[string]bool{
			"manageGroupMembership": true, "view": true, "mapRoles": true, "impersonate": false, "manage": true,
		},
	}
}

// makeUsers builds n users. Input: n [int]. Output: [[]gocloak.User].
func makeUsers(n int) []gocloak.User {
	users := make([]gocloak.User, n)
	for i := range users {
		users[i] = makeUser(i)
	}
	return users
}

// makeGroups builds a group tree. Input: n [int] top-level groups, sub [int] subgroups each. Output: [[]gocloak.Group].
func makeGroups(n, sub int) []gocloak.Group {
	groups := make([]gocloak.Group, n)
	for i := range groups {
		name := fmt.Sprintf("group-%02d", i)
		groups[i] = gocloak.Group{
			ID: new(fmt.Sprintf("a1b2c3d4-0000-0000-0000-%012d", i)), Name: new(name), Path: new("/" + name),
			Attributes: map[string][]string{"costCenter": {strconv.Itoa(1000 + i)}},
			RealmRoles: []string{"offline_access"},
		}
		for j := range sub {
			subName := fmt.Sprintf("sub-%02d", j)
			groups[i].SubGroups = append(groups[i].SubGroups, gocloak.Group{
				ID: new(fmt.Sprintf("a1b2c3d4-0000-0000-%04d-%012d", i, j)), Name: new(subName), Path: new("/" + name + "/" + subName),
			})
		}
	}
	return groups
}

// mustJSON marshals v or panics. Input: v [any]. Output: [[]byte].
func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	must(err)
	return b
}

// must panics on err. Input: err [error].
func must(err error) {
	if err != nil {
		panic(err)
	}
}
