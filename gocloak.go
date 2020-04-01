package gocloak

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/go-resty/resty/v2"
)

// GoCloak holds all methods a client should fulfill
type GoCloak interface {
	// RestyClient returns a resty client that gocloak uses
	RestyClient() *resty.Client
	// Sets the resty Client that gocloak uses
	SetRestyClient(restyClient *resty.Client)

	// GetToken returns a token
	GetToken(realm string, options TokenOptions) (*JWT, error)
	// GetRequestingPartyToken returns a requesting party token with permissions granted by the server
	GetRequestingPartyToken(token, realm string, options RequestingPartyTokenOptions) (*JWT, error)
	// Login sends a request to the token endpoint using user and client credentials
	Login(clientID, clientSecret, realm, username, password string) (*JWT, error)
	// Logout sends a request to the logout endpoint using refresh token
	Logout(clientID, clientSecret, realm, refreshToken string) error
	// LogoutPublicClient sends a request to the logout endpoint using refresh token
	LogoutPublicClient(clientID, realm, accessToken, refreshToken string) error
	// LoginClient sends a request to the token endpoint using client credentials
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	// LoginAdmin login as admin
	LoginAdmin(username, password, realm string) (*JWT, error)
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
	// GetServerInfo returns the server info
	GetServerInfo(accessToken string) (*ServerInfoRepesentation, error)
	// GetUserInfo gets the user info for the given realm
	GetUserInfo(accessToken string, realm string) (*UserInfo, error)

	// ExecuteActionsEmail executes an actions email
	ExecuteActionsEmail(token string, realm string, params ExecuteActionsEmail) error

	// CreateGroup creates a new group
	CreateGroup(accessToken, realm string, group Group) (string, error)
	// CreateChildGroup creates a new child group
	CreateChildGroup(token string, realm string, groupID string, group Group) (string, error)
	// CreateClient creates a new client
	CreateClient(accessToken, realm string, clientID Client) (string, error)
	// CreateClientScope creates a new clientScope
	CreateClientScope(accessToken, realm string, scope ClientScope) (string, error)
	// CreateComponent creates a new component
	CreateComponent(accessToken, realm string, component Component) (string, error)

	// UpdateGroup updates the given group
	UpdateGroup(accessToken string, realm string, updatedGroup Group) error
	// UpdateRole updates the given role
	UpdateRole(accessToken string, realm string, clientID string, role Role) error
	// UpdateClient updates the given client
	UpdateClient(accessToken string, realm string, updatedClient Client) error
	// UpdateClientScope updates the given clientScope
	UpdateClientScope(accessToken string, realm string, scope ClientScope) error

	// DeleteComponent deletes the given component
	DeleteComponent(accessToken string, realm, componentID string) error
	// DeleteGroup deletes the given group
	DeleteGroup(accessToken string, realm, groupID string) error
	// DeleteClient deletes the given client
	DeleteClient(accessToken string, realm, clientID string) error
	// DeleteClientScope
	DeleteClientScope(accessToken string, realm, scopeID string) error

	// GetClient returns a client
	GetClient(accessToken string, realm string, clientID string) (*Client, error)
	// GetClientsDefaultScopes returns a list of the client's default scopes
	GetClientsDefaultScopes(token string, realm string, clientID string) ([]*ClientScope, error)
	// AddDefaultScopeToClient adds a client scope to the list of client's default scopes
	AddDefaultScopeToClient(token string, realm string, clientID string, scopeID string) error
	// RemoveDefaultScopeFromClient removes a client scope from the list of client's default scopes
	RemoveDefaultScopeFromClient(token string, realm string, clientID string, scopeID string) error
	// GetClientsOptionalScopes returns a list of the client's optional scopes
	GetClientsOptionalScopes(token string, realm string, clientID string) ([]*ClientScope, error)
	// AddOptionalScopeToClient adds a client scope to the list of client's optional scopes
	AddOptionalScopeToClient(token string, realm string, clientID string, scopeID string) error
	// RemoveOptionalScopeFromClient deletes a client scope from the list of client's optional scopes
	RemoveOptionalScopeFromClient(token string, realm string, clientID string, scopeID string) error
	// GetDefaultOptionalClientScopes returns a list of default realm optional scopes
	GetDefaultOptionalClientScopes(token string, realm string) ([]*ClientScope, error)
	// GetDefaultDefaultClientScopes returns a list of default realm default scopes
	GetDefaultDefaultClientScopes(token string, realm string) ([]*ClientScope, error)
	// GetClientScope returns a clientscope
	GetClientScope(token string, realm string, scopeID string) (*ClientScope, error)
	// GetClientScopes returns all client scopes
	GetClientScopes(token string, realm string) ([]*ClientScope, error)
	// GetClientSecret returns a client's secret
	GetClientSecret(token string, realm string, clientID string) (*CredentialRepresentation, error)
	// GetClientServiceAccount retrieves the service account "user" for a client if enabled
	GetClientServiceAccount(token string, realm string, clientID string) (*User, error)
	// RegenerateClientSecret creates a new client secret returning the updated CredentialRepresentation
	RegenerateClientSecret(token string, realm string, clientID string) (*CredentialRepresentation, error)
	// GetKeyStoreConfig gets the keyStoreConfig
	GetKeyStoreConfig(accessToken string, realm string) (*KeyStoreConfig, error)
	// GetComponents gets components of the given realm
	GetComponents(accessToken string, realm string) ([]*Component, error)
	// GetDefaultGroups returns a list of default groups
	GetDefaultGroups(accessToken string, realm string) ([]*Group, error)
	// AddDefaultGroup adds group to the list of default groups
	AddDefaultGroup(accessToken string, realm string, groupID string) error
	// RemoveDefaultGroup removes group from the list of default groups
	RemoveDefaultGroup(accessToken string, realm string, groupID string) error
	// GetGroups gets all groups of the given realm
	GetGroups(accessToken string, realm string, params GetGroupsParams) ([]*Group, error)
	// GetGroup gets the given group
	GetGroup(accessToken string, realm, groupID string) (*Group, error)
	// GetGroupMembers get a list of users of group with id in realm
	GetGroupMembers(accessToken string, realm, groupID string, params GetGroupsParams) ([]*User, error)
	// GetRoleMappingByGroupID gets the rolemapping for the given group id
	GetRoleMappingByGroupID(accessToken string, realm string, groupID string) (*MappingsRepresentation, error)
	// GetRoleMappingByUserID gets the rolemapping for the given user id
	GetRoleMappingByUserID(accessToken string, realm string, userID string) (*MappingsRepresentation, error)
	// GetClients gets the clients in the realm
	GetClients(accessToken string, realm string, params GetClientsParams) ([]*Client, error)
	// GetClientOfflineSessions returns offline sessions associated with the client
	GetClientOfflineSessions(token, realm, clientID string) ([]*UserSessionRepresentation, error)
	// GetClientUserSessions returns user sessions associated with the client
	GetClientUserSessions(token, realm, clientID string) ([]*UserSessionRepresentation, error)
	// CreateClientProtocolMapper creates a protocol mapper in client scope
	CreateClientProtocolMapper(token, realm, clientID string, mapper ProtocolMapperRepresentation) (string, error)
	// CreateClientProtocolMapper updates a protocol mapper in client scope
	UpdateClientProtocolMapper(token, realm, clientID string, mapperID string, mapper ProtocolMapperRepresentation) error
	// DeleteClientProtocolMapper deletes a protocol mapper in client scope
	DeleteClientProtocolMapper(token, realm, clientID, mapperID string) error

	// UserAttributeContains checks if the given attribute has the given value
	UserAttributeContains(attributes map[string][]string, attribute string, value string) bool

	// *** Realm Roles ***

	// CreateRealmRole creates a role in a realm
	CreateRealmRole(token, realm string, role Role) (string, error)
	// GetRealmRole returns a role from a realm by role's name
	GetRealmRole(token string, realm string, roleName string) (*Role, error)
	// GetRealmRoles get all roles of the given realm. It's an alias for the GetRoles function
	GetRealmRoles(accessToken string, realm string) ([]*Role, error)
	// GetRealmRolesByUserID returns all roles assigned to the given user
	GetRealmRolesByUserID(accessToken string, realm string, userID string) ([]*Role, error)
	// GetRealmRolesByGroupID returns all roles assigned to the given group
	GetRealmRolesByGroupID(accessToken string, realm string, groupID string) ([]*Role, error)
	// UpdateRealmRole updates a role in a realm
	UpdateRealmRole(token string, realm string, roleName string, role Role) error
	// DeleteRealmRole deletes a role in a realm by role's name
	DeleteRealmRole(token string, realm string, roleName string) error
	// AddRealmRoleToUser adds realm-level role mappings
	AddRealmRoleToUser(token string, realm string, userID string, roles []Role) error
	// DeleteRealmRoleFromUser deletes realm-level role mappings
	DeleteRealmRoleFromUser(token string, realm string, userID string, roles []Role) error
	// AddRealmRoleToGroup adds realm-level role mappings
	AddRealmRoleToGroup(token string, realm string, groupID string, roles []Role) error
	// DeleteRealmRoleFromGroup deletes realm-level role mappings
	DeleteRealmRoleFromGroup(token string, realm string, groupID string, roles []Role) error
	// AddRealmRoleComposite adds roles as composite
	AddRealmRoleComposite(token string, realm string, roleName string, roles []Role) error
	// AddRealmRoleComposite adds roles as composite
	DeleteRealmRoleComposite(token string, realm string, roleName string, roles []Role) error

	// *** Client Roles ***

	// AddClientRoleToUser adds a client role to the user
	AddClientRoleToUser(token string, realm string, clientID string, userID string, roles []Role) error
	// AddClientRoleToGroup adds a client role to the group
	AddClientRoleToGroup(token string, realm string, clientID string, groupID string, roles []Role) error
	// CreateClientRole creates a new role for a client
	CreateClientRole(accessToken, realm, clientID string, role Role) (string, error)
	// DeleteClientRole deletes the given role
	DeleteClientRole(accessToken, realm, clientID, roleName string) error
	// DeleteClientRoleFromUser removes a client role from from the user
	DeleteClientRoleFromUser(token string, realm string, clientID string, userID string, roles []Role) error
	// DeleteClientRoleFromGroup removes a client role from from the group
	DeleteClientRoleFromGroup(token string, realm string, clientID string, groupID string, roles []Role) error
	// GetClientRoles gets roles for the given client
	GetClientRoles(accessToken string, realm string, clientID string) ([]*Role, error)
	// GetRealmRolesByUserID returns all client roles assigned to the given user
	GetClientRolesByUserID(token string, realm string, clientID string, userID string) ([]*Role, error)
	// GetClientRolesByGroupID returns all client roles assigned to the given group
	GetClientRolesByGroupID(token string, realm string, clientID string, groupID string) ([]*Role, error)
	// GetCompositeClientRolesByRoleID returns all client composite roles associated with the given client role
	GetCompositeClientRolesByRoleID(token string, realm string, clientID string, roleID string) ([]*Role, error)
	// GetCompositeClientRolesByUserID returns all client roles and composite roles assigned to the given user
	GetCompositeClientRolesByUserID(token string, realm string, clientID string, userID string) ([]*Role, error)
	// GetCompositeClientRolesByGroupID returns all client roles and composite roles assigned to the given group
	GetCompositeClientRolesByGroupID(token string, realm string, clientID string, groupID string) ([]*Role, error)
	// GetClientRole get a role for the given client in a realm by role name
	GetClientRole(token string, realm string, clientID string, roleName string) (*Role, error)
	// AddClientRoleComposite adds roles as composite
	AddClientRoleComposite(token string, realm string, roleID string, roles []Role) error
	// DeleteClientRoleComposite deletes composites from a role
	DeleteClientRoleComposite(token string, realm string, roleID string, roles []Role) error

	// *** Realm ***

	// GetRealm returns top-level representation of the realm
	GetRealm(token string, realm string) (*RealmRepresentation, error)
	// GetRealms returns top-level representation of all realms
	GetRealms(token string) ([]*RealmRepresentation, error)
	// CreateRealm creates a realm
	CreateRealm(token string, realm RealmRepresentation) (string, error)
	// UpdateRealm updates a given realm
	UpdateRealm(token string, realm RealmRepresentation) error
	// DeleteRealm removes a realm
	DeleteRealm(token string, realm string) error
	// ClearRealmCache clears realm cache
	ClearRealmCache(token string, realm string) error
	// ClearUserCache clears realm cache
	ClearUserCache(token string, realm string) error
	// ClearKeysCache clears realm cache
	ClearKeysCache(token string, realm string) error

	// *** Users ***
	// CreateUser creates a new user
	CreateUser(token string, realm string, user User) (string, error)
	// DeleteUser deletes the given user
	DeleteUser(accessToken string, realm, userID string) error
	// GetUserByID gets the user with the given id
	GetUserByID(accessToken string, realm string, userID string) (*User, error)
	// GetUser count returns the userCount of the given realm
	GetUserCount(accessToken string, realm string) (int, error)
	// GetUsers gets all users of the given realm
	GetUsers(accessToken string, realm string, params GetUsersParams) ([]*User, error)
	// GetUserGroups gets the groups of the given user
	GetUserGroups(accessToken string, realm string, userID string) ([]*UserGroup, error)
	// GetUsersByRoleName returns all users have a given role
	GetUsersByRoleName(token string, realm string, roleName string) ([]*User, error)
	// GetUsersByClientRoleName returns all users have a given client role
	GetUsersByClientRoleName(token string, realm string, clientID string, roleName string, params GetUsersByRoleParams) ([]*User, error)
	// SetPassword sets a new password for the user with the given id. Needs elevated privileges
	SetPassword(token string, userID string, realm string, password string, temporary bool) error
	// UpdateUser updates the given user
	UpdateUser(accessToken string, realm string, user User) error
	// AddUserToGroup puts given user to given group
	AddUserToGroup(token string, realm string, userID string, groupID string) error
	// DeleteUserFromGroup deletes given user from given group
	DeleteUserFromGroup(token string, realm string, userID string, groupID string) error
	// GetUserSessions returns user sessions associated with the user
	GetUserSessions(token, realm, userID string) ([]*UserSessionRepresentation, error)
	// GetUserOfflineSessionsForClient returns offline sessions associated with the user and client
	GetUserOfflineSessionsForClient(token, realm, userID, clientID string) ([]*UserSessionRepresentation, error)
	// GetUserFederatedIdentities gets all user federated identities
	GetUserFederatedIdentities(token, realm, userID string) ([]*FederatedIdentityRepresentation, error)
	// CreateUserFederatedIdentity creates an user federated identity
	CreateUserFederatedIdentity(token, realm, userID, providerID string, federatedIdentityRep FederatedIdentityRepresentation) error
	// DeleteUserFederatedIdentity deletes an user federated identity
	DeleteUserFederatedIdentity(token, realm, userID, providerID string) error

	// *** Identity Provider **
	// CreateIdentityProvider creates an identity provider in a realm
	CreateIdentityProvider(token string, realm string, providerRep IdentityProviderRepresentation) (string, error)
	// GetIdentityProviders gets identity providers in a realm
	GetIdentityProviders(token string, realm string) ([]*IdentityProviderRepresentation, error)
	// GetIdentityProvider gets the identity provider in a realm
	GetIdentityProvider(token string, realm string, alias string) (*IdentityProviderRepresentation, error)
	// UpdateIdentityProvider updates the identity provider in a realm
	UpdateIdentityProvider(token string, realm string, alias string, providerRep IdentityProviderRepresentation) error
	// DeleteIdentityProvider deletes the identity provider in a realm
	DeleteIdentityProvider(token string, realm string, alias string) error

	// *** Protection API ***
	// GetResource returns a client's resource with the given id
	GetResource(token string, realm string, clientID string, resourceID string) (*ResourceRepresentation, error)
	// GetResources a returns resources associated with the client
	GetResources(token string, realm string, clientID string, params GetResourceParams) ([]*ResourceRepresentation, error)
	// CreateResource creates a resource associated with the client
	CreateResource(token string, realm string, clientID string, resource ResourceRepresentation) (*ResourceRepresentation, error)
	// UpdateResource updates a resource associated with the client
	UpdateResource(token string, realm string, clientID string, resource ResourceRepresentation) error
	// DeleteResource deletes a resource associated with the client
	DeleteResource(token string, realm string, clientID string, resourceID string) error

	// GetScope returns a client's scope with the given id
	GetScope(token string, realm string, clientID string, scopeID string) (*ScopeRepresentation, error)
	// GetScopes returns scopes associated with the client
	GetScopes(token string, realm string, clientID string, params GetScopeParams) ([]*ScopeRepresentation, error)
	// CreateScope creates a scope associated with the client
	CreateScope(token string, realm string, clientID string, scope ScopeRepresentation) (*ScopeRepresentation, error)
	// UpdateScope updates a scope associated with the client
	UpdateScope(token string, realm string, clientID string, resource ScopeRepresentation) error
	// DeleteScope deletes a scope associated with the client
	DeleteScope(token string, realm string, clientID string, scopeID string) error

	// GetPolicy returns a client's policy with the given id
	GetPolicy(token string, realm string, clientID string, policyID string) (*PolicyRepresentation, error)
	// GetPolicies returns policies associated with the client
	GetPolicies(token string, realm string, clientID string, params GetPolicyParams) ([]*PolicyRepresentation, error)
	// CreatePolicy creates a policy associated with the client
	CreatePolicy(token string, realm string, clientID string, policy PolicyRepresentation) (*PolicyRepresentation, error)
	// UpdatePolicy updates a policy associated with the client
	UpdatePolicy(token string, realm string, clientID string, policy PolicyRepresentation) error
	// DeletePolicy deletes a policy associated with the client
	DeletePolicy(token string, realm string, clientID string, policyID string) error

	// GetPermission returns a client's permission with the given id
	GetPermission(token string, realm string, clientID string, permissionID string) (*PermissionRepresentation, error)
	// GetPermissions returns permissions associated with the client
	GetPermissions(token string, realm string, clientID string, params GetPermissionParams) ([]*PermissionRepresentation, error)
	// CreatePermission creates a permission associated with the client
	CreatePermission(token string, realm string, clientID string, permission PermissionRepresentation) (*PermissionRepresentation, error)
	// UpdatePermission updates a permission associated with the client
	UpdatePermission(token string, realm string, clientID string, permission PermissionRepresentation) error
	// DeletePermission deletes a permission associated with the client
	DeletePermission(token string, realm string, clientID string, permissionID string) error
}
