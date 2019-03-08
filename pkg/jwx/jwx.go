package jwx

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

// DecodeAccessTokenHeader decodes the header of the accessToken
func DecodeAccessTokenHeader(token string) (*DecodedAccessTokenHeader, error) {
	token = strings.Replace(token, "Bearer ", "", 1)
	headerString := strings.Split(token, ".")
	decodedData, err := base64.RawStdEncoding.DecodeString(headerString[0])
	if err != nil {
		return nil, err
	}

	result := &DecodedAccessTokenHeader{}
	err = json.Unmarshal(decodedData, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func decodePublicKey(e, n string) (*rsa.PublicKey, error) {
	decN, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, err
	}
	nInt := big.NewInt(0)
	nInt.SetBytes(decN)

	decE, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	pKey := rsa.PublicKey{N: nInt, E: int(eInt)}
	return &pKey, nil
}

// DecodeAccessToken currently only supports RSA - sorry for that
func DecodeAccessToken(accessToken string, e string, n string) (*jwt.Token, *jwt.MapClaims, error) {
	rsaPublicKey, err := decodePublicKey(e, n)
	if err != nil {
		return nil, nil, err
	}

	claims := &jwt.MapClaims{}
	token2, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})
	if err != nil {
		return nil, nil, err
	}

	return token2, claims, nil
}

// DecodeAccessTokenCustomClaims currently only supports RSA - sorry for that
func DecodeAccessTokenCustomClaims(accessToken string, e string, n string, customClaims jwt.Claims) (*jwt.Token, error) {
	rsaPublicKey, err := decodePublicKey(e, n)
	if err != nil {
		return nil, err
	}

	token2, err := jwt.ParseWithClaims(accessToken, customClaims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})
	if err != nil {
		return nil, err
	}

	return token2, nil
}

func getRSAPublicKey(publicKey string) (*rsa.PublicKey, error) {
	var builder strings.Builder
	builder.WriteString("\n-----BEGIN PUBLIC KEY-----\n")
	builder.WriteString(publicKey)
	builder.WriteString("\n-----END PUBLIC KEY-----\n")

	block, _ := pem.Decode([]byte(builder.String()))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pkey, _ := x509.ParsePKIXPublicKey(block.Bytes)
	if pkey == nil {
		return nil, errors.New("failed to parse public key")
	}

	rsaPublicKey := pkey.(*rsa.PublicKey)
	return rsaPublicKey, nil
}
