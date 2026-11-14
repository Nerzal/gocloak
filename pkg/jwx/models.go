package jwx

import jwt "github.com/golang-jwt/jwt/v5"

// DecodedAccessTokenHeader is the decoded header from the access token
type DecodedAccessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

// Claims served by keycloak inside the accessToken
type Claims struct {
	jwt.RegisteredClaims
	Typ               string         `json:"typ,omitempty"`
	Azp               string         `json:"azp,omitempty"`
	AuthTime          int            `json:"auth_time,omitempty"`
	SessionState      string         `json:"session_state,omitempty"`
	Acr               string         `json:"acr,omitempty"`
	AllowedOrigins    []string       `json:"allowed-origins,omitempty"`
	RealmAccess       RealmAccess    `json:"realm_access,omitzero"`
	ResourceAccess    ResourceAccess `json:"resource_access,omitzero"`
	Scope             string         `json:"scope,omitempty"`
	EmailVerified     bool           `json:"email_verified,omitempty"`
	Address           Address        `json:"address,omitzero"`
	Name              string         `json:"name,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty"`
	GivenName         string         `json:"given_name,omitempty"`
	FamilyName        string         `json:"family_name,omitempty"`
	Email             string         `json:"email,omitempty"`
	ClientID          string         `json:"clientId,omitempty"`
	ClientHost        string         `json:"clientHost,omitempty"`
	ClientIP          string         `json:"clientAddress,omitempty"`
}

// Address TODO what fields does any address have?
type Address struct{}

// RealmAccess holds roles of the user
type RealmAccess struct {
	Roles []string `json:"roles,omitempty"`
}

// ResourceAccess holds TODO: What does it hold?
type ResourceAccess struct {
	RealmManagement RealmManagement `json:"realm-management,omitzero"`
	Account         Account         `json:"account,omitzero"`
}

// RealmManagement holds TODO: What does it hold?
type RealmManagement struct {
	Roles []string `json:"roles,omitempty"`
}

// Account holds TODO: What does it hold?
type Account struct {
	Roles []string `json:"roles,omitempty"`
}
