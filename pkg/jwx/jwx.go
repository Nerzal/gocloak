package jwx

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
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

// DecodeAccessToken currently only supports RSA - sorry for that
func DecodeAccessToken(accessToken string, publicKey string) (*jwt.Token, *jwt.MapClaims, error) {
	rsaPublicKey, err := getRSAPublicKey(publicKey)
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
func DecodeAccessTokenCustomClaims(accessToken string, publicKey string, customClaims jwt.Claims) (*jwt.Token, error) {
	rsaPublicKey, err := getRSAPublicKey(publicKey)
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
