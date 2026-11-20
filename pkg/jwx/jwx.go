// Package jwx provides jwt token utility function
package jwx

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// SignClaims signs the given claims with the given key and signing method.
// Input: claims [jwt.Claims], key [any] signing key, method [jwt.SigningMethod].
// Output: [string] signed JWT, [error] if signing fails.
func SignClaims(claims jwt.Claims, key any, method jwt.SigningMethod) (string, error) {
	token := jwt.NewWithClaims(method, claims)
	return token.SignedString(key)
}

// DecodeAccessTokenHeader decodes the base64url-encoded (RFC 7515) header of the accessToken.
// Input: token [string] JWT, optionally prefixed with "Bearer ". Output: [*DecodedAccessTokenHeader], [error] if decoding fails.
func DecodeAccessTokenHeader(token string) (*DecodedAccessTokenHeader, error) {
	const errMessage = "could not decode access token header"
	token = strings.TrimPrefix(token, "Bearer ")
	headerString := strings.Split(token, ".")
	decodedData, err := base64.RawURLEncoding.DecodeString(headerString[0])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	result := &DecodedAccessTokenHeader{}
	err = json.Unmarshal(decodedData, result)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	return result, nil
}

// toBigInt decodes a base64url (unpadded) string into a big integer.
// Input: v [string] base64url encoded bytes.
// Output: [*big.Int] decoded value, [error] if decoding fails.
func toBigInt(v string) (*big.Int, error) {
	decRes, err := base64.RawURLEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}

	res := big.NewInt(0)
	res.SetBytes(decRes)
	return res, nil
}

var curves = map[string]elliptic.Curve{
	"P-224": elliptic.P224(),
	"P-256": elliptic.P256(),
	"P-384": elliptic.P384(),
	"P-521": elliptic.P521(),
}

// decodeECDSAPublicKey builds an ECDSA public key from JWK coordinates and checks that the point is on the curve.
// Input: x [string], y [string] base64url coordinates, crv [string] curve name; empty means missing.
// Output: [*ecdsa.PublicKey], [error] if decoding fails or the point is invalid.
func decodeECDSAPublicKey(x, y, crv string) (*ecdsa.PublicKey, error) {
	const errMessage = "could not decode public key"
	if x == "" || y == "" || crv == "" {
		return nil, fmt.Errorf("%s: missing x, y or crv", errMessage)
	}

	c, ok := curves[crv]
	if !ok {
		return nil, fmt.Errorf("%s: unknown curve alg: %s", errMessage, crv)
	}
	size := (c.Params().BitSize + 7) / 8

	point := make([]byte, 1+2*size)
	point[0] = 4 // uncompressed point marker
	for i, coordinate := range []string{x, y} {
		decoded, err := base64.RawURLEncoding.DecodeString(coordinate)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errMessage, err)
		}
		if len(decoded) > size {
			return nil, fmt.Errorf("%s: coordinate too long", errMessage)
		}
		copy(point[1+(i+1)*size-len(decoded):], decoded) // left-pad to the curve size
	}

	key, err := ecdsa.ParseUncompressedPublicKey(c, point)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	return key, nil
}

// decodeRSAPublicKey builds an RSA public key from the base64url JWK exponent and modulus.
// Input: e [string] exponent, n [string] modulus; empty means missing. Output: [*rsa.PublicKey], [error] if a value is missing or invalid.
func decodeRSAPublicKey(e, n string) (*rsa.PublicKey, error) {
	const errMessage = "could not decode public key"
	if e == "" || n == "" {
		return nil, fmt.Errorf("%s: missing e or n", errMessage)
	}

	nInt, err := toBigInt(n)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	decE, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	eInt := new(big.Int).SetBytes(decE)
	if !eInt.IsInt64() || eInt.Int64() > math.MaxInt32 {
		return nil, fmt.Errorf("%s: rsa exponent out of range", errMessage)
	}

	return &rsa.PublicKey{N: nInt, E: int(eInt.Int64())}, nil
}

// DecodeAccessTokenRSACustomClaims verifies an RS* signed access token with the RSA key (e, n) and decodes it into
// the given custom claims. A "Bearer " prefix is stripped.
// Input: accessToken [string], e [string] base64url exponent, n [string] base64url modulus,
// customClaims [jwt.Claims] target claims.
// Output: [*jwt.Token] parsed token, [error] if the key is invalid or verification fails.
func DecodeAccessTokenRSACustomClaims(accessToken string, e, n string, customClaims jwt.Claims) (*jwt.Token, error) {
	const errMessage = "could not decode accessToken with custom claims"
	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	rsaPublicKey, err := decodeRSAPublicKey(e, n)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	token2, err := jwt.ParseWithClaims(accessToken, customClaims, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	return token2, nil
}

// DecodeAccessTokenECDSACustomClaims verifies an ES* signed access token with the EC key (x, y, crv) and decodes it
// into the given custom claims. A "Bearer " prefix is stripped.
// Input: accessToken [string], x [string] base64url x coordinate, y [string] base64url y coordinate,
// crv [string] curve name, customClaims [jwt.Claims] target claims.
// Output: [*jwt.Token] parsed token, [error] if the key is invalid or verification fails.
func DecodeAccessTokenECDSACustomClaims(accessToken string, x, y, crv string, customClaims jwt.Claims) (*jwt.Token, error) {
	const errMessage = "could not decode accessToken"
	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	publicKey, err := decodeECDSAPublicKey(x, y, crv)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	token2, err := jwt.ParseWithClaims(accessToken, customClaims, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	return token2, nil
}

// DecodePublicKey builds the public key of a JWK with key type kty [string] ("RSA" or "EC").
// Input: kty [string], e [string], n [string] RSA exponent and modulus, x [string], y [string], crv [string] EC values.
// Output: [crypto.PublicKey] *rsa.PublicKey or *ecdsa.PublicKey, [error] if the key type is unsupported or a value is invalid.
func DecodePublicKey(kty string, e, n, x, y, crv string) (crypto.PublicKey, error) {
	switch kty {
	case "RSA":
		return decodeRSAPublicKey(e, n)
	case "EC":
		return decodeECDSAPublicKey(x, y, crv)
	default:
		return nil, fmt.Errorf("could not decode public key: unsupported key type %q", kty)
	}
}

// DecodeAccessTokenWithKey verifies accessToken [string] with an already parsed public key and decodes it into customClaims.
// The signing method must fit the key type; if alg [string] is not empty, the token must use exactly that algorithm.
// Input: accessToken [string], key [crypto.PublicKey] *rsa.PublicKey or *ecdsa.PublicKey, alg [string] expected alg or "",
// customClaims [jwt.Claims], opts [...jwt.ParserOption] extra parser options such as jwt.WithLeeway.
// Output: [*jwt.Token], [error] if verification fails.
func DecodeAccessTokenWithKey(accessToken string, key crypto.PublicKey, alg string, customClaims jwt.Claims, opts ...jwt.ParserOption) (*jwt.Token, error) {
	const errMessage = "could not decode accessToken"
	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	if alg != "" {
		opts = append(opts, jwt.WithValidMethods([]string{alg}))
	}

	token, err := jwt.ParseWithClaims(accessToken, customClaims, func(token *jwt.Token) (any, error) {
		if !methodFitsKey(token.Method, key) {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}
	return token, nil
}

// methodFitsKey reports whether the signing method [jwt.SigningMethod] can be verified with key [crypto.PublicKey].
// Input: method [jwt.SigningMethod], key [crypto.PublicKey]. Output: [bool].
func methodFitsKey(method jwt.SigningMethod, key crypto.PublicKey) bool {
	switch key.(type) {
	case *rsa.PublicKey:
		_, rsaOK := method.(*jwt.SigningMethodRSA)
		_, pssOK := method.(*jwt.SigningMethodRSAPSS)
		return rsaOK || pssOK
	case *ecdsa.PublicKey:
		_, ok := method.(*jwt.SigningMethodECDSA)
		return ok
	default:
		return false
	}
}
