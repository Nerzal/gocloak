package gocloak

import (
	"context"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"
)

// GoCloak holds all methods a client should fulfill
type GoCloak interface {
	// RestyClient returns a resty client that gocloak uses
	RestyClient() *resty.Client
	// Sets the resty Client that gocloak uses
	SetRestyClient(restyClient *resty.Client)

	// GetToken returns a token
	GetToken(ctx context.Context, realm string, options TokenOptions) (*JWT, error)
	// GetRequestingPartyToken returns a requesting party token with permissions granted by the server
	GetRequestingPartyToken(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) (*JWT, error)
	// GetRequestingPartyPermissions returns a permissions granted by the server to requesting party
	GetRequestingPartyPermissions(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) (*[]RequestingPartyPermission, error)
	// GetRequestingPartyPermissionDecision returns a permission decision granted by the server to requesting party
	GetRequestingPartyPermissionDecision(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) (*RequestingPartyPermissionDecision, error)
	// Login sends a request to the token endpoint using user and client credentials
	Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*JWT, error)
	// LoginOtp performs a login with user credentials and otp token
	LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*JWT, error)
	// Logout sends a request to the logout endpoint using refresh token
	Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error
	// LogoutPublicClient sends a request to the logout endpoint using refresh token
	LogoutPublicClient(ctx context.Context, idOfClient, realm, accessToken, refreshToken string) error
	// LogoutAllSessions logs out all sessions of a user given an id
	LogoutAllSessions(ctx context.Context, accessToken, realm, userID string) error
	// RevokeConsents revoke consent and offline tokens for particular client from user
	RevokeUserConsents(ctx context.Context, accessToken, realm, userID, clientID string) error
	// LogoutUserSessions logs out a single sessions of a user given a session id.
	// NOTE: this uses bearer token, but this token must belong to a user with proper privileges
	LogoutUserSession(ctx context.Context, accessToken, realm, session string) error
	// LoginClient sends a request to the token endpoint using client credentials
	LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*JWT, error)
	// LoginClientTokenExchange requests a login on a specified users behalf. Returning a user's tokens.
	LoginClientTokenExchange(ctx context.Context, clientID, token, clientSecret, realm, targetClient, userID string) (*JWT, error)
	// LoginClientSignedJWT performs a login with client credentials and signed jwt claims
	LoginClientSignedJWT(ctx context.Context, idOfClient, realm string, key interface{}, signedMethod jwt.SigningMethod, expiresAt *jwt.NumericDate) (*JWT, error)
	// LoginAdmin login as admin
	LoginAdmin(ctx context.Context, username, password, realm string) (*JWT, error)
	// RefreshToken used to refresh the token
	RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*JWT, error)
	// DecodeAccessToken decodes the accessToken
	DecodeAccessToken(ctx context.Context, accessToken, realm string) (*jwt.Token, *jwt.MapClaims, error)
	// DecodeAccessTokenCustomClaims decodes the accessToken and fills the given claims
	DecodeAccessTokenCustomClaims(ctx context.Context, accessToken, realm string, claims jwt.Claims) (*jwt.Token, error)
	// RetrospectToken calls the openid-connect introspect endpoint
	RetrospectToken(ctx context.Context, accessToken, clientID, clientSecret, realm string) (*RetrospecTokenResult, error)
	// GetIssuer calls the issuer endpoint for the given realm
	GetIssuer(ctx context.Context, realm string) (*IssuerResponse, error)
	// GetCerts gets the public keys for the given realm
	GetCerts(ctx context.Context, realm string) (*CertResponse, error)
	// GetServerInfo returns the server info
	GetServerInfo(ctx context.Context, accessToken string) (*ServerInfoRepesentation, error)
	// GetUserInfo gets the user info for the given realm
	GetUserInfo(ctx context.Context, accessToken, realm string) (*UserInfo, error)
	// GetRawUserInfo calls the UserInfo endpoint and returns a raw json object
	GetRawUserInfo(ctx context.Context, accessToken, realm string) (map[string]interface{}, error)

	// ExecuteActionsEmail executes an actions email
	ExecuteActionsEmail(ctx context.Context, token, realm string, params ExecuteActionsEmail) error

	// CreateGroup creates a new group
	CreateGroup(ctx context.Context, accessToken, realm string, group Group) (string, error)
	// CreateChildGroup creates a new child group
	CreateChildGroup(ctx context.Context, token, realm, groupID string, group Group) (string, error)
	// CreateClient creates a new client
	CreateClient(ctx context.Context, accessToken, realm string, newClient Client) (string, error)
	// CreateClientScope creates a new clientScope
	CreateClientScope(ctx context.Context, accessToken, realm string, scope ClientScope) (string, error)
	// CreateClientScopeProtocolMapper creates a new protocolMapper under the given client scope
	CreateClientScopeProtocolMapper(ctx context.Context, accessToken, realm, scopeID string, protocolMapper ProtocolMappers) (string, error)
	// CreateComponent creates a new component
	CreateComponent(ctx context.Context, accessToken, realm string, component Component) (string, error)
	// CreateClientScopeMappingsRealmRoles creates realm-level roles to the client’s scope
	CreateClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []Role) error
	// CreateClientScopeMappingsClientRoles creates client-level roles from the client’s scope
	CreateClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []Role) error
	// CreateClientScopesScopeMappingsRealmRoles creates realm-level roles to the client-scope
	CreateClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string, roles []Role) error
	// CreateClientScopesScopeMappingsClientRoles creates client-level roles to the client-scope
	CreateClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string, roles []Role) error
	// CreateClientRepresentation creates a new client representation
	CreateClientRepresentation(ctx context.Context, realm string) (*Client, error)

	// UpdateGroup updates the given group
	UpdateGroup(ctx context.Context, accessToken, realm string, updatedGroup Group) error
	// UpdateRole updates the given role
	UpdateRole(ctx context.Context, accessToken, realm, idOfClient string, role Role) error
	// UpdateClient updates the given client
	UpdateClient(ctx context.Context, accessToken, realm string, updatedClient Client) error
	// UpdateClientScope updates the given clientScope
	UpdateClientScope(ctx context.Context, accessToken, realm string, scope ClientScope) error
	// UpdateClientScopeProtocolMapper updates the given protocol mapper for a client scope
	UpdateClientScopeProtocolMapper(ctx context.Context, accessToken, realm, scopeID string, protocolMapper ProtocolMappers) error
	// UpdateClientRepresentation updates the given client representation
	UpdateClientRepresentation(ctx context.Context, accessToken, realm string, updatedClient Client) (*Client, error)

	// DeleteComponent deletes the given component
	DeleteComponent(ctx context.Context, accessToken, realm, componentID string) error
	// DeleteGroup deletes the given group
	DeleteGroup(ctx context.Context, accessToken, realm, groupID string) error
	// DeleteClient deletes the given client
	DeleteClient(ctx context.Context, accessToken, realm, idOfClient string) error
	// DeleteClientScope
	DeleteClientScope(ctx context.Context, accessToken, realm, scopeID string) error
	// DeleteClientScopeProtocolMapper deletes the given protocol mapper from the client scope
	DeleteClientScopeProtocolMapper(ctx context.Context, accessToken, realm, scopeID, protocolMapperID string) error
	// DeleteClientScopeMappingsRealmRoles deletes realm-level roles from the client’s scope
	DeleteClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []Role) error
	// DeleteClientScopeMappingsClientRoles deletes client-level roles from the client’s scope
	DeleteClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []Role) error
	// DeleteClientScopesScopeMappingsRealmRoles deletes realm-level roles from the client-scope
	DeleteClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string, roles []Role) error
	// DeleteClientScopesScopeMappingsClientRoles deletes client-level roles from the client-scope
	DeleteClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, ifOfClient string, roles []Role) error
	// DeleteClientRepresentation deletes a given client representation
	DeleteClientRepresentation(ctx context.Context, accessToken, realm, clientID string) error

	// GetClient returns a client
	GetClient(ctx context.Context, accessToken, realm, idOfClient string) (*Client, error)
	// GetClientsDefaultScopes returns a list of the client's default scopes
	GetClientsDefaultScopes(ctx context.Context, token, realm, idOfClient string) ([]*ClientScope, error)
	// AddDefaultScopeToClient adds a client scope to the list of client's default scopes
	AddDefaultScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	// RemoveDefaultScopeFromClient removes a client scope from the list of client's default scopes
	RemoveDefaultScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	// GetClientsOptionalScopes returns a list of the client's optional scopes
	GetClientsOptionalScopes(ctx context.Context, token, realm, idOfClient string) ([]*ClientScope, error)
	// AddOptionalScopeToClient adds a client scope to the list of client's optional scopes
	AddOptionalScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	// RemoveOptionalScopeFromClient deletes a client scope from the list of client's optional scopes
	RemoveOptionalScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	// GetDefaultOptionalClientScopes returns a list of default realm optional scopes
	GetDefaultOptionalClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error)
	// GetDefaultDefaultClientScopes returns a list of default realm default scopes
	GetDefaultDefaultClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error)
	// GetClientScope returns a clientscope
	GetClientScope(ctx context.Context, token, realm, scopeID string) (*ClientScope, error)
	// GetClientScopes returns all client scopes
	GetClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error)
	// GetClientScopeProtocolMappers returns all protocol mappers of a client scope
	GetClientScopeProtocolMappers(ctx context.Context, token, realm, scopeID string) ([]*ProtocolMappers, error)
	// GetClientScopeProtocolMapper returns a protocol mapper of a client scope
	GetClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID, protocolMapperID string) (*ProtocolMappers, error)
	// GetClientScopeMappings returns all scope mappings for the client
	GetClientScopeMappings(ctx context.Context, token, realm, idOfClient string) (*MappingsRepresentation, error)
	// GetClientScopeMappingsRealmRoles returns realm-level roles associated with the client’s scope
	GetClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string) ([]*Role, error)
	// GetClientScopeMappingsRealmRolesAvailable returns realm-level roles that are available to attach to this client’s scope
	GetClientScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClient string) ([]*Role, error)
	// GetClientScopesScopeMappingsRealmRolesAvailable returns realm-level roles that are available to attach to this client-scope
	GetClientScopesScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClientScope string) ([]*Role, error)
	// GetClientScopesScopeMappingsClientRolesAvailable returns client-level roles that are available to attach to this client-scope
	GetClientScopesScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*Role, error)
	// GetClientScopeMappingsClientRoles returns roles associated with a client’s scope
	GetClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*Role, error)
	// GetClientScopesScopeMappingsRealmRoles returns roles associated with a client-scope
	GetClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string) ([]*Role, error)
	// GetClientScopesScopeMappingsClientRoles returns client roles associated with a client-scope
	GetClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*Role, error)
	// GetClientScopeMappingsClientRolesAvailable returns available roles associated with a client’s scope
	GetClientScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*Role, error)
	// GetClientSecret returns a client's secret
	GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*CredentialRepresentation, error)
	// GetClientServiceAccount retrieves the service account "user" for a client if enabled
	GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*User, error)
	// RegenerateClientSecret creates a new client secret returning the updated CredentialRepresentation
	RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*CredentialRepresentation, error)
	// GetKeyStoreConfig gets the keyStoreConfig
	GetKeyStoreConfig(ctx context.Context, accessToken, realm string) (*KeyStoreConfig, error)
	// GetComponents gets components of the given realm
	GetComponents(ctx context.Context, accessToken, realm string) ([]*Component, error)
	// GetDefaultGroups returns a list of default groups
	GetDefaultGroups(ctx context.Context, accessToken, realm string) ([]*Group, error)
	// AddDefaultGroup adds group to the list of default groups
	AddDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
	// RemoveDefaultGroup removes group from the list of default groups
	RemoveDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
	// GetGroups gets all groups of the given realm
	GetGroups(ctx context.Context, accessToken, realm string, params GetGroupsParams) ([]*Group, error)
	// GetGroupsByRole gets groups with specified roles assigned of given realm
	GetGroupsByRole(ctx context.Context, accessToken, realm string, roleName string) ([]*Group, error)
	// GetGroupsByClientRole gets groups with specified roles assigned of given client within a realm
	GetGroupsByClientRole(ctx context.Context, accessToken, realm string, roleName string, clientId string) ([]*Group, error)
	// GetGroupsCount gets groups count of the given realm
	GetGroupsCount(ctx context.Context, token, realm string, params GetGroupsParams) (int, error)
	// GetGroup gets the given group
	GetGroup(ctx context.Context, accessToken, realm, groupID string) (*Group, error)
	// GetGroupMembers get a list of users of group with id in realm
	GetGroupMembers(ctx context.Context, accessToken, realm, groupID string, params GetGroupsParams) ([]*User, error)
	// GetRoleMappingByGroupID gets the rolemapping for the given group id
	GetRoleMappingByGroupID(ctx context.Context, accessToken, realm, groupID string) (*MappingsRepresentation, error)
	// GetRoleMappingByUserID gets the rolemapping for the given user id
	GetRoleMappingByUserID(ctx context.Context, accessToken, realm, userID string) (*MappingsRepresentation, error)
	// GetClients gets the clients in the realm
	GetClients(ctx context.Context, accessToken, realm string, params GetClientsParams) ([]*Client, error)
	// GetClientOfflineSessions returns offline sessions associated with the client
	GetClientOfflineSessions(ctx context.Context, token, realm, idOfClient string) ([]*UserSessionRepresentation, error)
	// GetClientUserSessions returns user sessions associated with the client
	GetClientUserSessions(ctx context.Context, token, realm, idOfClient string) ([]*UserSessionRepresentation, error)
	// CreateClientProtocolMapper creates a protocol mapper in client scope
	CreateClientProtocolMapper(ctx context.Context, token, realm, idOfClient string, mapper ProtocolMapperRepresentation) (string, error)
	// CreateClientProtocolMapper updates a protocol mapper in client scope
	UpdateClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string, mapper ProtocolMapperRepresentation) error
	// DeleteClientProtocolMapper deletes a protocol mapper in client scope
	DeleteClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string) error
	// GetClientRepresentation return a client representation
	GetClientRepresentation(ctx context.Context, accessToken, realm, clientID string) (*Client, error)
	// GetAdapterConfiguration returns a adapter configuration
	GetAdapterConfiguration(ctx context.Context, accessToken, realm, clientID string) (*AdapterConfiguration, error)

	// *** Realm Roles ***

	// CreateRealmRole creates a role in a realm
	CreateRealmRole(ctx context.Context, token, realm string, role Role) (string, error)
	// GetRealmRole returns a role from a realm by role's name
	GetRealmRole(ctx context.Context, token, realm, roleName string) (*Role, error)
	// GetRealmRoleByID returns a role from a realm by role's ID
	GetRealmRoleByID(ctx context.Context, token, realm, roleID string) (*Role, error)
	// GetRealmRoles get all roles of the given realm. It's an alias for the GetRoles function
	GetRealmRoles(ctx context.Context, accessToken, realm string, params GetRoleParams) ([]*Role, error)
	// GetRealmRolesByUserID returns all roles assigned to the given user
	GetRealmRolesByUserID(ctx context.Context, accessToken, realm, userID string) ([]*Role, error)
	// GetRealmRolesByGroupID returns all roles assigned to the given group
	GetRealmRolesByGroupID(ctx context.Context, accessToken, realm, groupID string) ([]*Role, error)
	// UpdateRealmRole updates a role in a realm
	UpdateRealmRole(ctx context.Context, token, realm, roleName string, role Role) error
	// UpdateRealmRoleByID updates a role in a realm by role's ID
	UpdateRealmRoleByID(ctx context.Context, token, realm, roleID string, role Role) error
	// DeleteRealmRole deletes a role in a realm by role's name
	DeleteRealmRole(ctx context.Context, token, realm, roleName string) error
	// AddRealmRoleToUser adds realm-level role mappings
	AddRealmRoleToUser(ctx context.Context, token, realm, userID string, roles []Role) error
	// DeleteRealmRoleFromUser deletes realm-level role mappings
	DeleteRealmRoleFromUser(ctx context.Context, token, realm, userID string, roles []Role) error
	// AddRealmRoleToGroup adds realm-level role mappings
	AddRealmRoleToGroup(ctx context.Context, token, realm, groupID string, roles []Role) error
	// DeleteRealmRoleFromGroup deletes realm-level role mappings
	DeleteRealmRoleFromGroup(ctx context.Context, token, realm, groupID string, roles []Role) error
	// AddRealmRoleComposite adds roles as composite
	AddRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []Role) error
	// AddRealmRoleComposite adds roles as composite
	DeleteRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []Role) error
	// GetCompositeRealmRoles returns all realm composite roles associated with the given realm role
	GetCompositeRealmRoles(ctx context.Context, token, realm, roleName string) ([]*Role, error)
	// GetCompositeRealmRolesByRoleID returns all realm composite roles associated with the given client role
	GetCompositeRealmRolesByRoleID(ctx context.Context, token, realm, roleID string) ([]*Role, error)
	// GetCompositeRealmRolesByUserID returns all realm roles and composite roles assigned to the given user
	GetCompositeRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*Role, error)
	// GetCompositeRealmRolesByGroupID returns all realm roles and composite roles assigned to the given group
	GetCompositeRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*Role, error)
	// GetAvailableRealmRolesByUserID returns all available realm roles to the given user
	GetAvailableRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*Role, error)
	// GetAvailableRealmRolesByGroupID returns all available realm roles to the given group
	GetAvailableRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*Role, error)

	// *** Client Roles ***

	// AddClientRoleToUser adds a client role to the user
	AddClientRoleToUser(ctx context.Context, token, realm, idOfClient, userID string, roles []Role) error
	// AddClientRoleToGroup adds a client role to the group
	AddClientRoleToGroup(ctx context.Context, token, realm, idOfClient, groupID string, roles []Role) error
	// CreateClientRole creates a new role for a client
	CreateClientRole(ctx context.Context, accessToken, realm, idOfClient string, role Role) (string, error)
	// DeleteClientRole deletes the given role
	DeleteClientRole(ctx context.Context, accessToken, realm, idOfClient, roleName string) error
	// DeleteClientRoleFromUser removes a client role from from the user
	DeleteClientRoleFromUser(ctx context.Context, token, realm, idOfClient, userID string, roles []Role) error
	// DeleteClientRoleFromGroup removes a client role from from the group
	DeleteClientRoleFromGroup(ctx context.Context, token, realm, idOfClient, groupID string, roles []Role) error
	// GetClientRoles gets roles for the given client
	GetClientRoles(ctx context.Context, accessToken, realm, idOfClient string, params GetRoleParams) ([]*Role, error)
	// GetClientRoleById gets role for the given client using role id
	GetClientRoleByID(ctx context.Context, accessToken, realm, roleID string) (*Role, error)
	// GetRealmRolesByUserID returns all client roles assigned to the given user
	GetClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*Role, error)
	// GetClientRolesByGroupID returns all client roles assigned to the given group
	GetClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*Role, error)
	// GetCompositeClientRolesByRoleID returns all client composite roles associated with the given client role
	GetCompositeClientRolesByRoleID(ctx context.Context, token, realm, idOfClient, roleID string) ([]*Role, error)
	// GetCompositeClientRolesByUserID returns all client roles and composite roles assigned to the given user
	GetCompositeClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*Role, error)
	// GetCompositeClientRolesByGroupID returns all client roles and composite roles assigned to the given group
	GetCompositeClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*Role, error)
	// GetAvailableClientRolesByUserID returns all available client roles to the given user
	GetAvailableClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*Role, error)
	// GetAvailableClientRolesByGroupID returns all available client roles to the given group
	GetAvailableClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*Role, error)

	// GetClientRole get a role for the given client in a realm by role name
	GetClientRole(ctx context.Context, token, realm, idOfClient, roleName string) (*Role, error)
	// AddClientRoleComposite adds roles as composite
	AddClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []Role) error
	// DeleteClientRoleComposite deletes composites from a role
	DeleteClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []Role) error

	// *** Realm ***

	// GetRealm returns top-level representation of the realm
	GetRealm(ctx context.Context, token, realm string) (*RealmRepresentation, error)
	// GetRealms returns top-level representation of all realms
	GetRealms(ctx context.Context, token string) ([]*RealmRepresentation, error)
	// CreateRealm creates a realm
	CreateRealm(ctx context.Context, token string, realm RealmRepresentation) (string, error)
	// UpdateRealm updates a given realm
	UpdateRealm(ctx context.Context, token string, realm RealmRepresentation) error
	// DeleteRealm removes a realm
	DeleteRealm(ctx context.Context, token, realm string) error
	// ClearRealmCache clears realm cache
	ClearRealmCache(ctx context.Context, token, realm string) error
	// ClearUserCache clears realm cache
	ClearUserCache(ctx context.Context, token, realm string) error
	// ClearKeysCache clears realm cache
	ClearKeysCache(ctx context.Context, token, realm string) error
	//GetAuthenticationFlows get all authentication flows from a realm
	GetAuthenticationFlows(ctx context.Context, token, realm string) ([]*AuthenticationFlowRepresentation, error)
	//Create a new Authentication flow in a realm
	CreateAuthenticationFlow(ctx context.Context, token, realm string, flow AuthenticationFlowRepresentation) error
	//DeleteAuthenticationFlow deletes a flow in a realm with the given ID
	DeleteAuthenticationFlow(ctx context.Context, token, realm, flowID string) error
	//GetAuthenticationExecutions retrieves all executions of a given flow
	GetAuthenticationExecutions(ctx context.Context, token, realm, flow string) ([]*ModifyAuthenticationExecutionRepresentation, error)
	//CreateAuthenticationExecution creates a new execution for the given flow name in the given realm
	CreateAuthenticationExecution(ctx context.Context, token, realm, flow string, execution CreateAuthenticationExecutionRepresentation) error
	//UpdateAuthenticationExecution updates an authentication execution for the given flow in the given realm
	UpdateAuthenticationExecution(ctx context.Context, token, realm, flow string, execution ModifyAuthenticationExecutionRepresentation) error
	// DeleteAuthenticationExecution delete a single execution with the given ID
	DeleteAuthenticationExecution(ctx context.Context, token, realm, executionID string) error

	//CreateAuthenticationExecutionFlow creates a new flow execution for the given flow name in the given realm
	CreateAuthenticationExecutionFlow(ctx context.Context, token, realm, flow string, execution CreateAuthenticationExecutionFlowRepresentation) error

	// *** Users ***
	// CreateUser creates a new user
	CreateUser(ctx context.Context, token, realm string, user User) (string, error)
	// DeleteUser deletes the given user
	DeleteUser(ctx context.Context, accessToken, realm, userID string) error
	// GetUserByID gets the user with the given id
	GetUserByID(ctx context.Context, accessToken, realm, userID string) (*User, error)
	// GetUser count returns the userCount of the given realm
	GetUserCount(ctx context.Context, accessToken, realm string, params GetUsersParams) (int, error)
	// GetUsers gets all users of the given realm
	GetUsers(ctx context.Context, accessToken, realm string, params GetUsersParams) ([]*User, error)
	// GetUserGroups gets the groups of the given user
	GetUserGroups(ctx context.Context, accessToken, realm, userID string, params GetGroupsParams) ([]*Group, error)
	// GetUsersByRoleName returns all users have a given role
	GetUsersByRoleName(ctx context.Context, token, realm, roleName string) ([]*User, error)
	// GetUsersByClientRoleName returns all users have a given client role
	GetUsersByClientRoleName(ctx context.Context, token, realm, idOfClient, roleName string, params GetUsersByRoleParams) ([]*User, error)
	// SetPassword sets a new password for the user with the given id. Needs elevated privileges
	SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error
	// UpdateUser updates the given user
	UpdateUser(ctx context.Context, accessToken, realm string, user User) error
	// AddUserToGroup puts given user to given group
	AddUserToGroup(ctx context.Context, token, realm, userID, groupID string) error
	// DeleteUserFromGroup deletes given user from given group
	DeleteUserFromGroup(ctx context.Context, token, realm, userID, groupID string) error
	// GetUserSessions returns user sessions associated with the user
	GetUserSessions(ctx context.Context, token, realm, userID string) ([]*UserSessionRepresentation, error)
	// GetUserOfflineSessionsForClient returns offline sessions associated with the user and client
	GetUserOfflineSessionsForClient(ctx context.Context, token, realm, userID, idOfClient string) ([]*UserSessionRepresentation, error)
	// GetUserFederatedIdentities gets all user federated identities
	GetUserFederatedIdentities(ctx context.Context, token, realm, userID string) ([]*FederatedIdentityRepresentation, error)
	// CreateUserFederatedIdentity creates an user federated identity
	CreateUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string, federatedIdentityRep FederatedIdentityRepresentation) error
	// DeleteUserFederatedIdentity deletes an user federated identity
	DeleteUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string) error

	// *** Identity Provider **
	// CreateIdentityProvider creates an identity provider in a realm
	CreateIdentityProvider(ctx context.Context, token, realm string, providerRep IdentityProviderRepresentation) (string, error)
	// GetIdentityProviders gets identity providers in a realm
	GetIdentityProviders(ctx context.Context, token, realm string) ([]*IdentityProviderRepresentation, error)
	// GetIdentityProvider gets the identity provider in a realm
	GetIdentityProvider(ctx context.Context, token, realm, alias string) (*IdentityProviderRepresentation, error)
	// UpdateIdentityProvider updates the identity provider in a realm
	UpdateIdentityProvider(ctx context.Context, token, realm, alias string, providerRep IdentityProviderRepresentation) error
	// DeleteIdentityProvider deletes the identity provider in a realm
	DeleteIdentityProvider(ctx context.Context, token, realm, alias string) error
	// ImportIdentityProviderConfig parses and returns the identity provider config at a given URL
	ImportIdentityProviderConfig(ctx context.Context, token, realm, fromURL, providerID string) (map[string]string, error)
	// ImportIdentityProviderConfigFromFile parses and returns the identity provider config from a given file
	ImportIdentityProviderConfigFromFile(ctx context.Context, token, realm, providerID, fileName string, fileBody io.Reader) (map[string]string, error)
	// ExportIDPPublicBrokerConfig exports the broker config for a given alias
	ExportIDPPublicBrokerConfig(ctx context.Context, token, realm, alias string) (*string, error)
	// CreateIdentityProviderMapper creates an instance of an identity provider mapper associated with the given alias
	CreateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper IdentityProviderMapper) (string, error)
	// GetIdentityProviderMapperByID gets the mapper of an identity provider
	GetIdentityProviderMapperByID(ctx context.Context, token, realm, alias, mapperID string) (*IdentityProviderMapper, error)
	// UpdateIdentityProviderMapper updates mapper of an identity provider
	UpdateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper IdentityProviderMapper) error
	// DeleteIdentityProviderMapper deletes an instance of an identity provider mapper associated with the given alias and mapper ID
	DeleteIdentityProviderMapper(ctx context.Context, token, realm, alias, mapperID string) error
	// GetIdentityProviderMappers returns list of mappers associated with an identity provider
	GetIdentityProviderMappers(ctx context.Context, token, realm, alias string) ([]*IdentityProviderMapper, error)

	// *** Protection API ***
	// GetResource returns a client's resource with the given id, using access token from client
	GetResourceClient(ctx context.Context, token, realm, resourceID string) (*ResourceRepresentation, error)
	// GetResources a returns resources associated with the client, using access token from client
	GetResourcesClient(ctx context.Context, token, realm string, params GetResourceParams) ([]*ResourceRepresentation, error)
	// CreateResource creates a resource associated with the client, using access token from client
	CreateResourceClient(ctx context.Context, token, realm string, resource ResourceRepresentation) (*ResourceRepresentation, error)
	// UpdateResource updates a resource associated with the client, using access token from client
	UpdateResourceClient(ctx context.Context, token, realm string, resource ResourceRepresentation) error
	// DeleteResource deletes a resource associated with the client, using access token from client
	DeleteResourceClient(ctx context.Context, token, realm, resourceID string) error

	// GetResource returns a client's resource with the given id, using access token from admin
	GetResource(ctx context.Context, token, realm, idOfClient, resourceID string) (*ResourceRepresentation, error)
	// GetResources a returns resources associated with the client, using access token from admin
	GetResources(ctx context.Context, token, realm, idOfClient string, params GetResourceParams) ([]*ResourceRepresentation, error)
	// CreateResource creates a resource associated with the client, using access token from admin
	CreateResource(ctx context.Context, token, realm, idOfClient string, resource ResourceRepresentation) (*ResourceRepresentation, error)
	// UpdateResource updates a resource associated with the client, using access token from admin
	UpdateResource(ctx context.Context, token, realm, idOfClient string, resource ResourceRepresentation) error
	// DeleteResource deletes a resource associated with the client, using access token from admin
	DeleteResource(ctx context.Context, token, realm, idOfClient, resourceID string) error

	// GetScope returns a client's scope with the given id, using access token from admin
	GetScope(ctx context.Context, token, realm, idOfClient, scopeID string) (*ScopeRepresentation, error)
	// GetScopes returns scopes associated with the client, using access token from admin
	GetScopes(ctx context.Context, token, realm, idOfClient string, params GetScopeParams) ([]*ScopeRepresentation, error)
	// CreateScope creates a scope associated with the client, using access token from admin
	CreateScope(ctx context.Context, token, realm, idOfClient string, scope ScopeRepresentation) (*ScopeRepresentation, error)
	// UpdateScope updates a scope associated with the client, using access token from admin
	UpdateScope(ctx context.Context, token, realm, idOfClient string, resource ScopeRepresentation) error
	// DeleteScope deletes a scope associated with the client, using access token from admin
	DeleteScope(ctx context.Context, token, realm, idOfClient, scopeID string) error

	// CreatePermissionTicket creates a permission ticket for a resource, using access token from client (typically a resource server)
	CreatePermissionTicket(ctx context.Context, token, realm string, permissions []CreatePermissionTicketParams) (*PermissionTicketResponseRepresentation, error)
	// GrantUserPermission lets resource owner grant permission for specific resource ID to specific user ID
	GrantUserPermission(ctx context.Context, token, realm string, permission PermissionGrantParams) (*PermissionGrantResponseRepresentation, error)
	// GrantPermission lets resource owner update permission for specific resource ID to specific user ID
	UpdateUserPermission(ctx context.Context, token, realm string, permission PermissionGrantParams) (*PermissionGrantResponseRepresentation, error)
	// GetUserPermission gets granted permissions according query parameters
	GetUserPermissions(ctx context.Context, token, realm string, params GetUserPermissionParams) ([]*PermissionGrantResponseRepresentation, error)
	// DeleteUserPermission lets resource owner delete permission for specific resource ID to specific user ID
	DeleteUserPermission(ctx context.Context, token, realm, ticketID string) error

	// GetPermission returns a client's permission with the given id
	GetPermission(ctx context.Context, token, realm, idOfClient, permissionID string) (*PermissionRepresentation, error)
	// GetPermissions returns permissions associated with the client
	GetPermissions(ctx context.Context, token, realm, idOfClient string, params GetPermissionParams) ([]*PermissionRepresentation, error)
	// CreatePermission creates a permission associated with the client
	CreatePermission(ctx context.Context, token, realm, idOfClient string, permission PermissionRepresentation) (*PermissionRepresentation, error)
	// UpdatePermission updates a permission associated with the client
	UpdatePermission(ctx context.Context, token, realm, idOfClient string, permission PermissionRepresentation) error
	// DeletePermission deletes a permission associated with the client
	DeletePermission(ctx context.Context, token, realm, idOfClient, permissionID string) error
	// GetDependentPermissions returns client's permissions dependent on the policy with given ID
	GetDependentPermissions(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PermissionRepresentation, error)
	GetPermissionResources(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*PermissionResource, error)
	GetPermissionScopes(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*PermissionScope, error)

	// GetPolicy returns a client's policy with the given id, using access token from admin
	GetPolicy(ctx context.Context, token, realm, idOfClient, policyID string) (*PolicyRepresentation, error)
	// GetPolicies returns policies associated with the client, using access token from admin
	GetPolicies(ctx context.Context, token, realm, idOfClient string, params GetPolicyParams) ([]*PolicyRepresentation, error)
	// CreatePolicy creates a policy associated with the client, using access token from admin
	CreatePolicy(ctx context.Context, token, realm, idOfClient string, policy PolicyRepresentation) (*PolicyRepresentation, error)
	// UpdatePolicy updates a policy associated with the client, using access token from admin
	UpdatePolicy(ctx context.Context, token, realm, idOfClient string, policy PolicyRepresentation) error
	// DeletePolicy deletes a policy associated with the client, using access token from admin
	DeletePolicy(ctx context.Context, token, realm, idOfClient, policyID string) error
	// GetPolicyAssociatedPolicies returns a client's policy associated policies with the given policy id, using access token from admin
	GetAuthorizationPolicyAssociatedPolicies(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PolicyRepresentation, error)
	// GetPolicyResources returns a client's resources of specific policy with the given policy id, using access token from admin
	GetAuthorizationPolicyResources(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PolicyResourceRepresentation, error)
	// GetPolicyScopes returns a client's scopes of specific policy with the given policy id, using access token from admin
	GetAuthorizationPolicyScopes(ctx context.Context, token, realm, idOfClient, policyID string) ([]*PolicyScopeRepresentation, error)

	// GetResourcePolicy updates a permission for a specifc resource, using token obtained by Resource Owner Password Credentials Grant or Token exchange
	GetResourcePolicy(ctx context.Context, token, realm, permissionID string) (*ResourcePolicyRepresentation, error)
	// GetResources returns resources associated with the client, using token obtained by Resource Owner Password Credentials Grant or Token exchange
	GetResourcePolicies(ctx context.Context, token, realm string, params GetResourcePoliciesParams) ([]*ResourcePolicyRepresentation, error)
	// GetResources returns all resources associated with the client, using token obtained by Resource Owner Password Credentials Grant or Token exchange
	CreateResourcePolicy(ctx context.Context, token, realm, resourceID string, policy ResourcePolicyRepresentation) (*ResourcePolicyRepresentation, error)
	// UpdateResourcePolicy updates a permission for a specifc resource, using token obtained by Resource Owner Password Credentials Grant or Token exchange
	UpdateResourcePolicy(ctx context.Context, token, realm, permissionID string, policy ResourcePolicyRepresentation) error
	// DeleteResourcePolicy deletes a permission for a specifc resource, using token obtained by Resource Owner Password Credentials Grant or Token exchange
	DeleteResourcePolicy(ctx context.Context, token, realm, permissionID string) error

	// ---------------
	// Credentials API
	// ---------------

	// GetCredentialRegistrators returns credentials registrators
	GetCredentialRegistrators(ctx context.Context, token, realm string) ([]string, error)
	// GetConfiguredUserStorageCredentialTypes returns credential types, which are provided by the user storage where user is stored
	GetConfiguredUserStorageCredentialTypes(ctx context.Context, token, realm, userID string) ([]string, error)

	// GetCredentials returns credentials available for a given user
	GetCredentials(ctx context.Context, token, realm, UserID string) ([]*CredentialRepresentation, error)
	// DeleteCredentials deletes the given credential for a given user
	DeleteCredentials(ctx context.Context, token, realm, UserID, CredentialID string) error
	// UpdateCredentialUserLabel updates label for the given credential for the given user
	UpdateCredentialUserLabel(ctx context.Context, token, realm, userID, credentialID, userLabel string) error
	// DisableAllCredentialsByType disables all credentials for a user of a specific type
	DisableAllCredentialsByType(ctx context.Context, token, realm, userID string, types []string) error
	// MoveCredentialBehind move a credential to a position behind another credential
	MoveCredentialBehind(ctx context.Context, token, realm, userID, credentialID, newPreviousCredentialID string) error
	// MoveCredentialToFirst move a credential to a first position in the credentials list of the user
	MoveCredentialToFirst(ctx context.Context, token, realm, userID, credentialID string) error

	// ---------------
	// Events API
	// ---------------

	// GetEvents returns events
	GetEvents(ctx context.Context, token string, realm string, params GetEventsParams) ([]*EventRepresentation, error)

	// -------------------
	// RequiredActions API
	// -------------------

	// UpdateRequiredAction updates a required action for a given realm
	RegisterRequiredAction(ctx context.Context, token string, realm string, requiredAction RequiredActionProviderRepresentation) error
	// UpdateRequiredAction updates a required action for a given realm
	UpdateRequiredAction(ctx context.Context, token string, realm string, requiredAction RequiredActionProviderRepresentation) error
	// UpdateRequiredAction updates a required action for a given realm
	GetRequiredAction(ctx context.Context, token string, realm string, alias string) (*RequiredActionProviderRepresentation, error)
	// UpdateRequiredAction updates a required action for a given realm
	GetRequiredActions(ctx context.Context, token string, realm string) ([]*RequiredActionProviderRepresentation, error)
	// UpdateRequiredAction updates a required action for a given realm
	DeleteRequiredAction(ctx context.Context, token string, realm string, alias string) error
}
