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
	RealmAccess       RealmAccess    `json:"realm_access,omitempty"`
	ResourceAccess    ResourceAccess `json:"resource_access,omitempty"`
	Scope             string         `json:"scope,omitempty"`
	EmailVerified     bool           `json:"email_verified,omitempty"`
	Address           Address        `json:"address,omitempty"`
	Name              string         `json:"name,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty"`
	GivenName         string         `json:"given_name,omitempty"`
	FamilyName        string         `json:"family_name,omitempty"`
	Email             string         `json:"email,omitempty"`
}

// Address TODO what fields does any address have?
type Address struct {
}

// RealmAccess holds roles of the user
type RealmAccess struct {
	Roles []string `json:"roles,omitempty"`
}

type ResourceAccess struct {
	RealmManagement RealmManagement `json:"realm-management,omitempty"`
	Account         Account         `json:"account,omitempty"`
}

type RealmManagement struct {
	Roles []string `json:"roles,omitempty"`
}

type Account struct {
	Roles []string `json:"roles,omitempty"`
}
