// Package jwx provides jwt token utility function
package jwx

import (
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

// SignClaims signs the given claims using a given key and a method
func SignClaims(claims jwt.Claims, key any, method jwt.SigningMethod) (string, error) {
	token := jwt.NewWithClaims(method, claims)
	return token.SignedString(key)
}

// DecodeAccessTokenHeader decodes the header of the accessToken
func DecodeAccessTokenHeader(token string) (*DecodedAccessTokenHeader, error) {
	const errMessage = "could not decode access token header"
	token = strings.Replace(token, "Bearer ", "", 1)
	headerString := strings.Split(token, ".")
	decodedData, err := base64.RawStdEncoding.DecodeString(headerString[0])
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
// Input: x [*string], y [*string] base64url coordinates, crv [*string] curve name.
// Output: [*ecdsa.PublicKey], [error] if decoding fails or the point is invalid.
func decodeECDSAPublicKey(x, y, crv *string) (*ecdsa.PublicKey, error) {
	const errMessage = "could not decode public key"

	c, ok := curves[*crv]
	if !ok {
		return nil, fmt.Errorf("%s: unknown curve alg: %s", errMessage, *crv)
	}
	size := (c.Params().BitSize + 7) / 8

	point := make([]byte, 1+2*size)
	point[0] = 4 // uncompressed point marker
	for i, coordinate := range []*string{x, y} {
		decoded, err := base64.RawURLEncoding.DecodeString(*coordinate)
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

func decodeRSAPublicKey(e, n *string) (*rsa.PublicKey, error) {
	const errMessage = "could not decode public key"

	nInt, err := toBigInt(*n)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	decE, err := base64.RawURLEncoding.DecodeString(*e)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errMessage, err)
	}

	eInt := new(big.Int).SetBytes(decE)
	if !eInt.IsInt64() || eInt.Int64() > math.MaxInt32 {
		return nil, fmt.Errorf("%s: rsa exponent out of range", errMessage)
	}

	return &rsa.PublicKey{N: nInt, E: int(eInt.Int64())}, nil
}

// DecodeAccessTokenRSACustomClaims decodes string access token into jwt.Token
func DecodeAccessTokenRSACustomClaims(accessToken string, e, n *string, customClaims jwt.Claims) (*jwt.Token, error) {
	const errMessage = "could not decode accessToken with custom claims"
	accessToken = strings.Replace(accessToken, "Bearer ", "", 1)

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

// DecodeAccessTokenECDSACustomClaims decodes string access token into jwt.Token
func DecodeAccessTokenECDSACustomClaims(accessToken string, x, y, crv *string, customClaims jwt.Claims) (*jwt.Token, error) {
	const errMessage = "could not decode accessToken"
	accessToken = strings.Replace(accessToken, "Bearer ", "", 1)

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
