package gocloak

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// NewPKCE creates a PKCE (RFC 7636) code verifier and its S256 code challenge.
// Send the challenge with code_challenge_method=S256 in the authorization request and the verifier as
// TokenOptions.CodeVerifier in the code exchange.
// Output: verifier [string] 43 characters from 32 random bytes, challenge [string] S256 challenge of the verifier.
func NewPKCE() (verifier, challenge string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // never fails since Go 1.24
	verifier = base64.RawURLEncoding.EncodeToString(b)
	return verifier, PKCEChallenge(verifier)
}

// PKCEChallenge returns the S256 code challenge of verifier [string]: base64url(SHA-256(verifier)) without padding.
// Output: [string] challenge.
func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
