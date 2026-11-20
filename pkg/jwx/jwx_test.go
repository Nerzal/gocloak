package jwx

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

var claims = jwt.MapClaims{
	"testKey": "testValue",
}

// signTestToken signs the package test claims with the given key and method.
// Input: t [*testing.T], key [any] signing key, method [jwt.SigningMethod]. Output: [string] signed token.
func signTestToken(t *testing.T, key any, method jwt.SigningMethod) string {
	t.Helper()
	token, err := SignClaims(claims, key, method)
	require.NoError(t, err)
	return token
}

// b64 base64url-encodes data without padding.
// Input: data [[]byte]. Output: [string] encoded data.
func b64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func TestDecodeAccessTokenECDSACustomClaims(t *testing.T) {
	t.Parallel()
	tests := []struct {
		curveAlg   elliptic.Curve
		curveStr   string
		jwtSighAlg *jwt.SigningMethodECDSA
	}{
		{jwtSighAlg: jwt.SigningMethodES256, curveAlg: elliptic.P256(), curveStr: "P-256"},
		{jwtSighAlg: jwt.SigningMethodES384, curveAlg: elliptic.P384(), curveStr: "P-384"},
		{jwtSighAlg: jwt.SigningMethodES512, curveAlg: elliptic.P521(), curveStr: "P-521"},
	}

	for _, tc := range tests {
		t.Run(tc.curveStr, func(t *testing.T) {
			t.Parallel()
			pk, err := ecdsa.GenerateKey(tc.curveAlg, rand.Reader)
			require.NoError(t, err)
			token := signTestToken(t, pk, tc.jwtSighAlg)

			testClaims := jwt.MapClaims{}
			x := b64(pk.X.Bytes())
			y := b64(pk.Y.Bytes())
			_, err = DecodeAccessTokenECDSACustomClaims("Bearer "+token, x, y, tc.curveStr, testClaims)
			require.NoError(t, err)
			require.Equal(t, claims, testClaims)
		})
	}
}

func TestDecodeAccessTokenECDSACustomClaims_Errors(t *testing.T) {
	t.Parallel()
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	other, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	token := signTestToken(t, pk, jwt.SigningMethodES256)
	x, y, crv := b64(pk.X.Bytes()), b64(pk.Y.Bytes()), "P-256"
	one, tooLong, unknown, badB64 := b64([]byte{1}), b64(make([]byte, 33)), "P-999", "!!!"

	tests := []struct {
		name    string
		x, y, c string
		errText string
	}{
		{"empty x", "", y, crv, "missing x, y or crv"},
		{"empty y", x, "", crv, "missing x, y or crv"},
		{"empty crv", x, y, "", "missing x, y or crv"},
		{"unknown curve", x, y, unknown, "unknown curve"},
		{"bad base64", badB64, y, crv, "illegal base64"},
		{"coordinate too long", tooLong, y, crv, "coordinate too long"},
		{"point not on curve", one, one, crv, "could not decode public key"},
		{"wrong key", b64(other.X.Bytes()), b64(other.Y.Bytes()), crv, "verification error"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeAccessTokenECDSACustomClaims(token, tc.x, tc.y, tc.c, jwt.MapClaims{})
			require.ErrorContains(t, err, tc.errText)
		})
	}

	hsToken := signTestToken(t, []byte("secret"), jwt.SigningMethodHS256)
	_, err = DecodeAccessTokenECDSACustomClaims(hsToken, x, y, crv, jwt.MapClaims{})
	require.ErrorContains(t, err, "unexpected signing method")
}

func TestDecodeAccessTokenRSACustomClaims(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	e := b64(big.NewInt(int64(key.E)).Bytes())
	n := b64(key.N.Bytes())
	hugeE := b64(new(big.Int).Lsh(big.NewInt(1), 40).Bytes())
	badB64 := "!!!"

	rsaToken := signTestToken(t, key, jwt.SigningMethodRS256)
	hsToken := signTestToken(t, []byte("secret"), jwt.SigningMethodHS256)

	tests := []struct {
		name    string
		token   string
		e, n    string
		errText string // empty means success
	}{
		{name: "valid", token: rsaToken, e: e, n: n},
		{name: "bearer prefix", token: "Bearer " + rsaToken, e: e, n: n},
		{name: "exponent out of range", token: rsaToken, e: hugeE, n: n, errText: "rsa exponent out of range"},
		{name: "bad base64 n", token: rsaToken, e: e, n: badB64, errText: "illegal base64"},
		{name: "bad base64 e", token: rsaToken, e: badB64, n: n, errText: "illegal base64"},
		{name: "empty e", token: rsaToken, n: n, errText: "missing e or n"},
		{name: "empty n", token: rsaToken, e: e, errText: "missing e or n"},
		{name: "HS256 token rejected", token: hsToken, e: e, n: n, errText: "unexpected signing method"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testClaims := jwt.MapClaims{}
			_, err := DecodeAccessTokenRSACustomClaims(tc.token, tc.e, tc.n, testClaims)
			if tc.errText != "" {
				require.ErrorContains(t, err, tc.errText)
				return
			}
			require.NoError(t, err)
			require.Equal(t, claims, testClaims)
		})
	}
}

func TestDecodeAccessTokenHeader(t *testing.T) {
	t.Parallel()
	// "?" and ">" encode to "/" and "+" in std base64 but "_" and "-" in base64url
	header := b64([]byte(`{"alg":"RS256","typ":"JWT","kid":"??????>>>>>>"}`))
	require.True(t, strings.ContainsAny(header, "-_"), "header must contain base64url-only characters")

	decoded, err := DecodeAccessTokenHeader("Bearer " + header + ".payload.signature")
	require.NoError(t, err)
	require.Equal(t, &DecodedAccessTokenHeader{Alg: "RS256", Typ: "JWT", Kid: "??????>>>>>>"}, decoded)

	_, err = DecodeAccessTokenHeader("!!!.payload.signature")
	require.ErrorContains(t, err, "could not decode access token header")

	_, err = DecodeAccessTokenHeader(b64([]byte("not json")) + ".x.y")
	require.ErrorContains(t, err, "could not decode access token header")
}

// TestDecodePublicKey checks that RSA and EC JWKs are decoded and that missing values and unknown key types fail.
// Input: t [*testing.T]. Output: none.
func TestDecodePublicKey(t *testing.T) {
	t.Parallel()
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	e, n := b64(big.NewInt(int64(rsaKey.E)).Bytes()), b64(rsaKey.N.Bytes())
	x, y := b64(ecKey.X.Bytes()), b64(ecKey.Y.Bytes())

	key, err := DecodePublicKey("RSA", e, n, "", "", "")
	require.NoError(t, err)
	require.True(t, rsaKey.PublicKey.Equal(key))

	key, err = DecodePublicKey("EC", "", "", x, y, "P-256")
	require.NoError(t, err)
	require.True(t, ecKey.PublicKey.Equal(key))

	_, err = DecodePublicKey("RSA", "", n, "", "", "")
	require.ErrorContains(t, err, "missing e or n")
	_, err = DecodePublicKey("EC", "", "", x, "", "P-256")
	require.ErrorContains(t, err, "missing x, y or crv")
	_, err = DecodePublicKey("oct", e, n, x, y, "P-256")
	require.ErrorContains(t, err, `unsupported key type "oct"`)
}

// TestDecodeAccessTokenWithKey checks signature verification with a parsed key, including the alg pinning and
// rejection of key confusion (HS256 with the RSA modulus), alg none and mismatched key types.
// Input: t [*testing.T]. Output: none.
func TestDecodeAccessTokenWithKey(t *testing.T) {
	t.Parallel()
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	otherRSA, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	rs256 := signTestToken(t, rsaKey, jwt.SigningMethodRS256)
	ps256 := signTestToken(t, rsaKey, jwt.SigningMethodPS256)
	es256 := signTestToken(t, ecKey, jwt.SigningMethodES256)
	// HS256 signed with the RSA modulus: the classic key confusion attack
	hs256 := signTestToken(t, rsaKey.N.Bytes(), jwt.SigningMethodHS256)
	none := signTestToken(t, jwt.UnsafeAllowNoneSignatureType, jwt.SigningMethodNone)

	tests := []struct {
		name    string
		token   string
		key     any
		alg     string
		errText string // empty means success
	}{
		{name: "RS256", token: rs256, key: &rsaKey.PublicKey},
		{name: "RS256 bearer prefix and alg", token: "Bearer " + rs256, key: &rsaKey.PublicKey, alg: "RS256"},
		{name: "PS256 with RSA key", token: ps256, key: &rsaKey.PublicKey},
		{name: "ES256", token: es256, key: &ecKey.PublicKey, alg: "ES256"},
		{name: "alg mismatch", token: ps256, key: &rsaKey.PublicKey, alg: "RS256", errText: "signing method PS256 is invalid"},
		{name: "wrong key", token: rs256, key: &otherRSA.PublicKey, errText: "verification error"},
		{name: "RSA token with EC key", token: rs256, key: &ecKey.PublicKey, errText: "unexpected signing method"},
		{name: "EC token with RSA key", token: es256, key: &rsaKey.PublicKey, errText: "unexpected signing method"},
		{name: "HS256 key confusion", token: hs256, key: &rsaKey.PublicKey, errText: "unexpected signing method"},
		{name: "alg none", token: none, key: &rsaKey.PublicKey, errText: "unexpected signing method"},
		{name: "unsupported key type", token: rs256, key: []byte("secret"), errText: "unexpected signing method"},
		{name: "malformed token", token: "not.a.token", key: &rsaKey.PublicKey, errText: "could not decode accessToken"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testClaims := jwt.MapClaims{}
			token, err := DecodeAccessTokenWithKey(tc.token, tc.key, tc.alg, testClaims)
			if tc.errText != "" {
				require.ErrorContains(t, err, tc.errText)
				require.Nil(t, token)
				return
			}
			require.NoError(t, err)
			require.True(t, token.Valid)
			require.Equal(t, claims, testClaims)
		})
	}
}

// TestDecodeAccessTokenWithKey_ParserOptions checks that parser options like jwt.WithLeeway are applied.
// Input: t [*testing.T]. Output: none.
func TestDecodeAccessTokenWithKey_ParserOptions(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	expired, err := SignClaims(jwt.MapClaims{"exp": time.Now().Add(-time.Minute).Unix()}, key, jwt.SigningMethodRS256)
	require.NoError(t, err)

	_, err = DecodeAccessTokenWithKey(expired, &key.PublicKey, "", jwt.MapClaims{})
	require.ErrorIs(t, err, jwt.ErrTokenExpired)

	_, err = DecodeAccessTokenWithKey(expired, &key.PublicKey, "", jwt.MapClaims{}, jwt.WithLeeway(2*time.Minute))
	require.NoError(t, err)
}
