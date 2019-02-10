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
	Typ               string         `json:"typ,omitempty"`
	Azp               string         `json:"azp,omitempty"`
	AuthTime          int            `json:"auth_time,omitempty"`
	SessionState      string         `json:"session_state,omitempty"`
	Acr               string         `json:"acr,omitempty"`
	AllowedOrigins    []string       `json:"allowed-origins,omitempty"`
	RealmAccess       RealmAccess    `json:"realm_access"`
	ResourceAccess    ResourceAccess `json:"resource_access"`
	Scope             string         `json:"scope"`
	EmailVerified     bool           `json:"email_verified"`
	Address           Address        `json:"address"`
	Name              string         `json:"name"`
	PreferredUsername string         `json:"preferred_username"`
	GivenName         string         `json:"given_name"`
	FamilyName        string         `json:"family_name"`
	Email             string         `json:"email"`
}

// Address TODO what fields does any address have?
type Address struct {
}

// RealmAccess holds roles of the user
type RealmAccess struct {
	Roles []string `json:"roles"`
}

type ResourceAccess struct {
	RealmManagement RealmManagement `json:"realm-management"`
	Account         Account         `json:"account"`
}

type RealmManagement struct {
	Roles []string `json:"roles"`
}

type Account struct {
	Roles []string `json:"roles"`
}
