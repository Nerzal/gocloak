package jwx

import jwt "github.com/dgrijalva/jwt-go"

// DecodedAccessTokenHeader is the decoded header from the access token
type DecodedAccessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

// Claims served by keycloak inside the accessToken
type Claims struct {
	jwt.StandardClaims
	Typ            string   `json:"typ,omitempty"`
	Azp            string   `json:"azp,omitempty"`
	AuthTime       int      `json:"auth_time,omitempty"`
	SessionState   string   `json:"session_state,omitempty"`
	Acr            string   `json:"acr,omitempty"`
	AllowedOrigins []string `json:"allowed-origins,omitempty"`
	RealmAccess    struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess struct {
		RealmManagement struct {
			Roles []string `json:"roles"`
		} `json:"realm-management"`
		Account struct {
			Roles []string `json:"roles"`
		} `json:"account"`
	} `json:"resource_access"`
	Scope         string `json:"scope"`
	EmailVerified bool   `json:"email_verified"`
	Address       struct {
	} `json:"address"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}
