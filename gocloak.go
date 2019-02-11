package gocloak

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// GoCloak holds all methods a client should fullfill
type GoCloak interface {
	// Login sends a request to the token endpoint using user and client credentials
	Login(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error)
	// LoginClient sends a request to the token endpoint using client credentials
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	// LoginAdmin login as admin
	LoginAdmin(username, password, realm string) (*JWT, error)
	// RequestPermisssion sends a request to the token endpoint with permission parameter
	RequestPermission(clientID string, clientSecret string, realm string, username string, password string, permission string) (*JWT, error)
	// RefreshToken used to refresh the token
	RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error)
	// DecodeAccessToken decodes the accessToken
	DecodeAccessToken(accessToken string, realm string) (*jwt.Token, *jwt.MapClaims, error)
	// DecodeAccessTokenCustomClaims decodes the accessToken and fills the given claims
	DecodeAccessTokenCustomClaims(accessToken string, realm string, claims jwt.Claims) (*jwt.Token, error)
	// DecodeAccessTokenCustomClaims calls the token introspection endpoint
	RetrospectToken(accessToken string, clientID, clientSecret string, realm string) (*RetrospecTokenResult, error)
	// GetIssuer calls the issuer endpoint for the given realm
	GetIssuer(realm string) (*IssuerResponse, error)
	// GetCerts gets the public keys for the given realm
	GetCerts(realm string) (*CertResponse, error)

	//SetPassword sets a new password for the user with the given id. Needs elevated priviliges
	SetPassword(token string, userID string, realm string, password string, temporary bool) error

	// CreateUser creates a new user
	CreateUser(token string, realm string, user User) (*string, error)
	// CreateGroup creates a new group
	CreateGroup(accessToken string, realm string, group Group) error
	// CreateRole creates a new role
	CreateRole(accessToken string, realm string, clientID string, role Role) error
	// CreateClient creates a new client
	CreateClient(accessToken string, realm string, clientID Client) error
	// CreateClientScope creates a new clientScope
	CreateClientScope(accessToken string, realm string, scope ClientScope) error
	// CreateComponent creates a new component
	CreateComponent(accessToken string, realm string, component Component) error

	// UpdateUser updates the given user
	UpdateUser(accessToken string, realm string, user User) error
	// UpdateGroup updates the given group
	UpdateGroup(accessToken string, realm string, group Group) error
	// UpdateRole updates the given role
	UpdateRole(accessToken string, realm string, clientID string, role Role) error
	// UpdateClient updates the given client
	UpdateClient(accessToken string, realm string, clientID Client) error
	// UpdateClientScope updates the given clientScope
	UpdateClientScope(accessToken string, realm string, scope ClientScope) error

	// DeleteUser deletes the given user
	DeleteUser(accessToken string, realm, userID string) error
	// DeleteComponent deletes the given component
	DeleteComponent(accessToken string, realm, componentID string) error
	// DeleteGroup deletes the given group
	DeleteGroup(accessToken string, realm, groupID string) error
	// DeleteRole deletes the given role
	DeleteRole(accessToken string, realm, clientID, roleName string) error
	// DeleteClient deletes the given client
	DeleteClient(accessToken string, realm, clientID string) error
	// DeleteClientScope
	DeleteClientScope(accessToken string, realm, scopeID string) error

	// GetKeyStoreConfig gets the keyStoreConfig
	GetKeyStoreConfig(accessToken string, realm string) (*KeyStoreConfig, error)
	// GetUserByID gets the user with the given id
	GetUserByID(accessToken string, realm string, userID string) (*User, error)
	// GetUser count returns the userCount of the given realm
	GetUserCount(accessToken string, realm string) (int, error)
	// GetUsers gets all users of the given realm
	GetUsers(accessToken string, realm string) (*[]User, error)
	// GetUserGroups gets the groups of the given user
	GetUserGroups(accessToken string, realm string, userID string) (*[]UserGroup, error)
	// GetComponents gets components of the given realm
	GetComponents(accessToken string, realm string) (*[]Component, error)
	// GetGroups gets all groups of the given realm
	GetGroups(accessToken string, realm string) (*[]Group, error)
	// GetGroup gets the given group
	GetGroup(accessToken string, realm, groupID string) (*Group, error)
	// GetRoles get all roles of the given realm
	GetRoles(accessToken string, realm string) (*[]Role, error)
	// GetRoleMappingByGroupID gets the rolemapping for the given group id
	GetRoleMappingByGroupID(accessToken string, realm string, groupID string) (*[]RoleMapping, error)
	// GetRoleMappingByUserID gets the rolemapping for the given user id
	GetRoleMappingByUserID(accessToken string, realm string, userID string) (*[]RoleMapping, error)
	// GetRolesByClientID gets roles for the given client
	GetRolesByClientID(accessToken string, realm string, clientID string) (*[]Role, error)
	// GetClients gets the clients in the realm
	GetClients(accessToken string, realm string) (*[]Client, error)
	// GetRealmRolesByUserID gets roles for the given uerID
	GetRealmRolesByUserID(accessToken string, realm string, userID string) (*[]Role, error)
	// GetRealmRolesByGroupID gets roles for given groupID
	GetRealmRolesByGroupID(accessToken string, realm string, groupID string) (*[]Role, error)
	// GetUsersByRoleName gets users for given roleName
	GetUsersByRoleName(token string, realm string, roleName string) (*[]User, error)

	// UserAttributeContains checks if the given attribute has the given value
	UserAttributeContains(attributes map[string][]string, attribute string, value string) bool
}
