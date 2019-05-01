package gocloak

import (
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/resty.v1"
)

// GoCloak holds all methods a client should fulfill
type GoCloak interface {
	// RestyClient returns a resty client that gocloak uses
	RestyClient() *resty.Client

	// Login sends a request to the token endpoint using user and client credentials
	Login(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error)
	// Logout sends a request to the logout endpoint using refresh token
	Logout(clientID, clientSecret, realm, refreshToken string) error
	// LoginClient sends a request to the token endpoint using client credentials
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	// LoginAdmin login as admin
	LoginAdmin(username, password, realm string) (*JWT, error)
	// RequestPermisssion sends a request to the token endpoint with permission parameter
	RequestPermission(clientID string, clientSecret string, realm string, username string, password string, permission string) (*JWT, error)
	// RefreshToken used to refresh the token
	RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error)
	// Exchanges a token for another client for one that can be used by this client
	ExchangeToken(authToken string, clientID, clientSecret, realm string) (*JWT, error)
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
	// GetUserInfo gets the user info for the given realm
	GetUserInfo(accessToken string, realm string) (*UserInfo, error)
	// GetUserInfoWithClaims gets the user info, adding any additional claims you want to request
	// Output object should have json mappings for the return object
	GetUserInfoWithClaims(accessToken string, realm string, claims []string, output interface{}) error

	// ExecuteActionsEmail executes an actions email
	ExecuteActionsEmail(token string, realm string, params ExecuteActionsEmail) error

	// CreateGroup creates a new group
	CreateGroup(accessToken string, realm string, group Group) error
	// CreateClientRole creates a new role for a client
	CreateClientRole(accessToken string, realm string, clientID string, role Role) error
	// CreateClient creates a new client
	CreateClient(accessToken string, realm string, clientID Client) error
	// CreateClientScope creates a new clientScope
	CreateClientScope(accessToken string, realm string, scope ClientScope) error
	// CreateComponent creates a new component
	CreateComponent(accessToken string, realm string, component Component) error

	// UpdateGroup updates the given group
	UpdateGroup(accessToken string, realm string, group Group) error
	// UpdateRole updates the given role
	UpdateRole(accessToken string, realm string, clientID string, role Role) error
	// UpdateClient updates the given client
	UpdateClient(accessToken string, realm string, clientID Client) error
	// UpdateClientScope updates the given clientScope
	UpdateClientScope(accessToken string, realm string, scope ClientScope) error

	// DeleteComponent deletes the given component
	DeleteComponent(accessToken string, realm, componentID string) error
	// DeleteGroup deletes the given group
	DeleteGroup(accessToken string, realm, groupID string) error
	// DeleteClientRole deletes the given role
	DeleteClientRole(accessToken string, realm, clientID, roleName string) error
	// DeleteClient deletes the given client
	DeleteClient(accessToken string, realm, clientID string) error
	// DeleteClientScope
	DeleteClientScope(accessToken string, realm, scopeID string) error

	// GetClient returns a client
	GetClient(accessToken string, realm string, clientID string) (*Client, error)
	// GetClientSecret returns a client's secret
	GetClientSecret(token string, realm string, clientID string) (*CredentialRepresentation, error)
	// GetKeyStoreConfig gets the keyStoreConfig
	GetKeyStoreConfig(accessToken string, realm string) (*KeyStoreConfig, error)
	// GetComponents gets components of the given realm
	GetComponents(accessToken string, realm string) (*[]Component, error)
	// GetGroups gets all groups of the given realm
	GetGroups(accessToken string, realm string, params GetGroupsParams) (*[]Group, error)
	// GetGroup gets the given group
	GetGroup(accessToken string, realm, groupID string) (*Group, error)
	// GetRoleMappingByGroupID gets the rolemapping for the given group id
	GetRoleMappingByGroupID(accessToken string, realm string, groupID string) (*MappingsRepresentation, error)
	// GetRoleMappingByUserID gets the rolemapping for the given user id
	GetRoleMappingByUserID(accessToken string, realm string, userID string) (*MappingsRepresentation, error)
	// GetClientRoles gets roles for the given client
	GetClientRoles(accessToken string, realm string, clientID string) (*[]Role, error)
	// GetClientRole get a role for the given client in a realm by role name
	GetClientRole(token string, realm string, clientID string, roleName string) (*Role, error)
	// GetClients gets the clients in the realm
	GetClients(accessToken string, realm string, params GetClientsParams) (*[]Client, error)

	// UserAttributeContains checks if the given attribute has the given value
	UserAttributeContains(attributes map[string][]string, attribute string, value string) bool

	// *** Realm Roles ***

	// CreateRealmRole creates a role in a realm
	CreateRealmRole(token string, realm string, role Role) error
	// GetRealmRole returns a role from a realm by role's name
	GetRealmRole(token string, realm string, roleName string) (*Role, error)
	// GetRealmRoles get all roles of the given realm. It's an alias for the GetRoles function
	GetRealmRoles(accessToken string, realm string) (*[]Role, error)
	// GetRealmRolesByUserID returns all roles assigned to the given user
	GetRealmRolesByUserID(accessToken string, realm string, userID string) (*[]Role, error)
	// GetRealmRolesByGroupID returns all roles assigned to the given group
	GetRealmRolesByGroupID(accessToken string, realm string, groupID string) (*[]Role, error)
	// UpdateRealmRole updates a role in a realm
	UpdateRealmRole(token string, realm string, roleName string, role Role) error
	// DeleteRealmRole deletes a role in a realm by role's name
	DeleteRealmRole(token string, realm string, roleName string) error
	// AddRealmRoleToUser adds realm-level role mappings
	AddRealmRoleToUser(token string, realm string, userID string, roles []Role) error
	// DeleteRealmRoleFromUser deletes realm-level role mappings
	DeleteRealmRoleFromUser(token string, realm string, userID string, roles []Role) error
	// AddRealmRoleComposite adds roles as composite
	AddRealmRoleComposite(token string, realm string, roleName string, roles []Role) error
	// AddRealmRoleComposite adds roles as composite
	DeleteRealmRoleComposite(token string, realm string, roleName string, roles []Role) error

	// *** Realm ***

	// GetRealm returns top-level representation of the realm
	GetRealm(token string, realm string) (*RealmRepresentation, error)
	// CreateRealm creates a realm
	CreateRealm(token string, realm RealmRepresentation) error
	// DeleteRealm removes a realm
	DeleteRealm(token string, realm string) error

	// *** Users ***
	// CreateUser creates a new user
	CreateUser(token string, realm string, user User) (*string, error)
	// DeleteUser deletes the given user
	DeleteUser(accessToken string, realm, userID string) error
	// GetUserByID gets the user with the given id
	GetUserByID(accessToken string, realm string, userID string) (*User, error)
	// GetUser count returns the userCount of the given realm
	GetUserCount(accessToken string, realm string) (int, error)
	// GetUsers gets all users of the given realm
	GetUsers(accessToken string, realm string, params GetUsersParams) (*[]User, error)
	// GetUserGroups gets the groups of the given user
	GetUserGroups(accessToken string, realm string, userID string) (*[]UserGroup, error)
	// GetUsersByRoleName returns all users have a given role
	GetUsersByRoleName(token string, realm string, roleName string) (*[]User, error)
	// SetPassword sets a new password for the user with the given id. Needs elevated privileges
	SetPassword(token string, userID string, realm string, password string, temporary bool) error
	// UpdateUser updates the given user
	UpdateUser(accessToken string, realm string, user User) error
	// AddUserToGroup puts given user to given group
	AddUserToGroup(token string, realm string, userID string, groupID string) error
	// DeleteUserFromGroup deletes given user from given group
	DeleteUserFromGroup(token string, realm string, userID string, groupID string) error
}
