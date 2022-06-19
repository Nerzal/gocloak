package jwx

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

// SignClaims signs the given claims using a given key and a method
func SignClaims(claims jwt.Claims, key interface{}, method jwt.SigningMethod) (string, error) {
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
		return nil, errors.Wrap(err, errMessage)
	}

	result := &DecodedAccessTokenHeader{}
	err = json.Unmarshal(decodedData, result)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

func decodeECDSAPublicKey(x, y, crv *string) (*ecdsa.PublicKey, error) {
	const errMessage = "could not decode public key"
	decX, err := base64.RawURLEncoding.DecodeString(*x)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	xInt := big.NewInt(0)
	xInt.SetBytes(decX)

	decY, err := base64.RawURLEncoding.DecodeString(*y)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	yInt := big.NewInt(0)
	yInt.SetBytes(decY)

	var c elliptic.Curve
	switch *crv {
	case "P-224":
		c = elliptic.P224()
	case "P-256":
		c = elliptic.P256()
	case "P-384":
		c = elliptic.P384()
	case "P-521":
		c = elliptic.P521()
	}

	pKey := &ecdsa.PublicKey{X: xInt, Y: yInt, Curve: c}
	return pKey, nil
}

func decodeRSAPublicKey(e, n *string) (*rsa.PublicKey, error) {
	const errMessage = "could not decode public key"

	decN, err := base64.RawURLEncoding.DecodeString(*n)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	nInt := big.NewInt(0)
	nInt.SetBytes(decN)

	decE, err := base64.RawURLEncoding.DecodeString(*e)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}

	eReader := bytes.NewReader(eBytes)
	var eInt uint64
	err = binary.Read(eReader, binary.BigEndian, &eInt)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	pKey := rsa.PublicKey{N: nInt, E: int(eInt)}
	return &pKey, nil
}

func DecodeAccessTokenECDSA(accessToken string, x, y, crv *string) (*jwt.Token, *jwt.MapClaims, error) {
	const errMessage = "could not decode accessToken"
	accessToken = strings.Replace(accessToken, "Bearer ", "", 1)

	publicKey, err := decodeECDSAPublicKey(x, y, crv)
	if err != nil {
		return nil, nil, errors.Wrap(err, errMessage)
	}

	claims := &jwt.MapClaims{}

	token2, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			fmt.Println()
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, errMessage)
	}

	return token2, claims, nil
}

func DecodeAccessTokenRSA(accessToken string, e, n *string) (*jwt.Token, *jwt.MapClaims, error) {
	const errMessage = "could not decode accessToken"
	accessToken = strings.Replace(accessToken, "Bearer ", "", 1)

	rsaPublicKey, err := decodeRSAPublicKey(e, n)
	if err != nil {
		return nil, nil, errors.Wrap(err, errMessage)
	}

	claims := &jwt.MapClaims{}

	token2, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, errMessage)
	}

	return token2, claims, nil
}

// DecodeAccessTokenCustomClaims currently only supports RSA - sorry for that
func DecodeAccessTokenCustomClaims(accessToken string, e, n *string, customClaims jwt.Claims) (*jwt.Token, error) {
	const errMessage = "could not decode accessToken with custom claims"
	accessToken = strings.Replace(accessToken, "Bearer ", "", 1)

	rsaPublicKey, err := decodeRSAPublicKey(e, n)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	token2, err := jwt.ParseWithClaims(accessToken, customClaims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})

	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return token2, nil
}
