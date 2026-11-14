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
			_, err = DecodeAccessTokenECDSACustomClaims("Bearer "+token, &x, &y, &tc.curveStr, testClaims)
			require.NoError(t, err)
			require.Equal(t, claims, testClaims)
		})
	}
}

func TestDecodeAccessTokenECDSACustomClaims_Errors(t *testing.T) {
	t.Parallel()
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	token := signTestToken(t, pk, jwt.SigningMethodES256)
	x, y, crv := b64(pk.X.Bytes()), b64(pk.Y.Bytes()), "P-256"
	one, tooLong, unknown, badB64 := b64([]byte{1}), b64(make([]byte, 33)), "P-999", "!!!"

	tests := []struct {
		name    string
		x, y, c *string
		errText string
	}{
		{"nil x", nil, &y, &crv, "missing x, y or crv"},
		{"nil y", &x, nil, &crv, "missing x, y or crv"},
		{"nil crv", &x, &y, nil, "missing x, y or crv"},
		{"unknown curve", &x, &y, &unknown, "unknown curve"},
		{"bad base64", &badB64, &y, &crv, "illegal base64"},
		{"coordinate too long", &tooLong, &y, &crv, "coordinate too long"},
		{"point not on curve", &one, &one, &crv, "could not decode public key"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeAccessTokenECDSACustomClaims(token, tc.x, tc.y, tc.c, jwt.MapClaims{})
			require.ErrorContains(t, err, tc.errText)
		})
	}
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
		e, n    *string
		errText string // empty means success
	}{
		{name: "valid", token: rsaToken, e: &e, n: &n},
		{name: "bearer prefix", token: "Bearer " + rsaToken, e: &e, n: &n},
		{name: "exponent out of range", token: rsaToken, e: &hugeE, n: &n, errText: "rsa exponent out of range"},
		{name: "bad base64 n", token: rsaToken, e: &e, n: &badB64, errText: "illegal base64"},
		{name: "bad base64 e", token: rsaToken, e: &badB64, n: &n, errText: "illegal base64"},
		{name: "nil e", token: rsaToken, n: &n, errText: "missing e or n"},
		{name: "nil n", token: rsaToken, e: &e, errText: "missing e or n"},
		{name: "HS256 token rejected", token: hsToken, e: &e, n: &n, errText: "unexpected signing method"},
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
