package jwx

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang-jwt/jwt/v4"
)

var claims = jwt.MapClaims{
	"testKey": "testValue",
}

func generateJWTToken(privKey *ecdsa.PrivateKey, sMethod *jwt.SigningMethodECDSA) (string, error) {
	token := jwt.NewWithClaims(sMethod, claims)
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		log.Fatal(err)
	}
	return tokenString, nil
}

func TestDecodeAccessTokenECDSACustomClaims(t *testing.T) {
	tests := []struct {
		curveAlg   elliptic.Curve
		curveStr   string
		jwtSighAlg *jwt.SigningMethodECDSA
	}{
		{
			jwtSighAlg: jwt.SigningMethodES256,
			curveAlg:   elliptic.P256(),
			curveStr:   "P-256",
		},
		{
			jwtSighAlg: jwt.SigningMethodES384,
			curveAlg:   elliptic.P384(),
			curveStr:   "P-384",
		},
		{
			jwtSighAlg: jwt.SigningMethodES512,
			curveAlg:   elliptic.P521(),
			curveStr:   "P-521",
		},
	}

	for _, tc := range tests {
		t.Run(tc.curveStr, func(t *testing.T) {
			pk, _ := ecdsa.GenerateKey(tc.curveAlg, rand.Reader)
			token, err := generateJWTToken(pk, tc.jwtSighAlg)
			require.NoError(t, err)

			testClaims := jwt.MapClaims{}
			x := base64.RawURLEncoding.EncodeToString(pk.X.Bytes())
			y := base64.RawURLEncoding.EncodeToString(pk.Y.Bytes())
			_, err = DecodeAccessTokenECDSACustomClaims(token, &x, &y, &tc.curveStr, testClaims)
			require.NoError(t, err)
			require.Equal(t, claims, testClaims)
		})
	}
}
