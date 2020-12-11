# gocloak
[![codebeat badge](https://codebeat.co/badges/c699bc56-aa5f-4cf5-893f-5cf564391b94)](https://codebeat.co/projects/github-com-nerzal-gocloak-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nerzal/gocloak)](https://goreportcard.com/report/github.com/Nerzal/gocloak)
[![Go Doc](https://godoc.org/github.com/Nerzal/gocloak?status.svg)](https://godoc.org/github.com/Nerzal/gocloak)
[![Build Status](https://github.com/Nerzal/gocloak/workflows/Tests/badge.svg)](https://github.com/Nerzal/gocloak/actions?query=branch%3Amaster+event%3Apush)
[![GitHub release](https://img.shields.io/github/tag/Nerzal/gocloak.svg)](https://GitHub.com/Nerzal/gocloak/releases/)
[![codecov](https://codecov.io/gh/Nerzal/gocloak/branch/master/graph/badge.svg)](https://codecov.io/gh/Nerzal/gocloak)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_shield)


Golang Keycloak API Package

This client is based on: [go-keycloak](https://github.com/PhilippHeuer/go-keycloak)

For Questions either raise an issue, or come to the [gopher-slack](https://invite.slack.golangbridge.org/) into the channel [#gocloak](https://gophers.slack.com/app_redirect?channel=gocloak)

If u are using the echo framework have a look at [gocloak-echo](https://github.com/Nerzal/gocloak-echo)

Benchmarks: https://nerzal.github.io/gocloak/dev/bench/


## Contribution
(WIP) https://github.com/Nerzal/gocloak/wiki/Contribute

## Changelog

### v7:
Breaking Change
 * Added support for array values in aud claim
 * When decoding an access Token, it is now needed to provide the audience to check
 * Add member "MatchingURI" to GetResourceParams
 * Add resource policy functions (thanks to timdrysdale)
 * Add type field to APIError 
 * Most of the protection API should now be implemented (thanks to timdrysdale)

### v6:
There are several backward incompatible changes
* all client functions now take `context.Context` as first argument.
* `UserAttributeContains` was moved from client method to package function.
* all structures now use pointers for the array types ([]string -> *[]string)

### v5:
There is only one change, but it's backward incompatible:
* Wrap Errors and use APIError struct to also provide the httpstatus code. ([#146](https://github.com/Nerzal/gocloak/pull/146))

### v4:
There are a lot of backward incompatible changes:
* all functions what create an object now return an ID of the created object. The return statement of those functions has been changed from (error) to (string, error)
* All structures now use pointers instead of general types (bool -> *bool, string -> *string). It has been done to properly use omitempty tag, otherwise it was impossible to set a false value for any of the bool propertires.

## Usage

### Installation

```shell
go get github.com/Nerzal/gocloak/v7
```

### Importing

```go
	import "github.com/Nerzal/gocloak/v7"
```

### Create New User
```go
	client := gocloak.NewClient("https://mycool.keycloak.instance")
	ctx := context.Background()
	token, err := client.LoginAdmin(ctx, "user", "password", "realmName")
	if err != nil {
		panic("Something wrong with the credentials or url")
	}
	
	user := gocloak.User{
		FirstName: gocloak.StringP(""Bob"),
		LastName:  gocloak.StringP(""Uncle"),
		Email:     gocloak.StringP(""something@really.wrong"),
		Enabled:   gocloak.BoolP(true),
		Username:  gocloak.StringP("CoolGuy"),
	}

	_, err = client.CreateUser(ctx, token.AccessToken, "realm", user)
	if err != nil {
		panic("Oh no!, failed to create user :(")
	}
```

### Introspect Token
```go
	client := gocloak.NewClient(hostname)
	ctx := context.Background()
	token, err := client.LoginClient(ctx, clientid, clientSecret, realm)
	if err != nil {
		panic("Login failed:"+ err.Error())
	}

	rptResult, err := client.RetrospectToken(ctx, token.AccessToken, clientid, clientSecret, realm)
	if err != nil {
		panic("Inspection failed:"+ err.Error())
	}

	if !rptResult.Active {
		panic("Token is not active")
	}

	permissions := rptResult.Permissions
	// Do something with the permissions ;)
```

## Features

```go
// GoCloak holds all methods a client should fulfill
type GoCloak interface {

	RestyClient() *resty.Client
	SetRestyClient(restyClient *resty.Client)

	GetToken(ctx context.Context, realm string, options TokenOptions) (*JWT, error)
	GetRequestingPartyToken(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) (*JWT, error)
	GetRequestingPartyPermissions(ctx context.Context, token, realm string, options RequestingPartyTokenOptions) (*[]RequestingPartyPermission, error)

	Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*JWT, error)
	LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*JWT, error)
	Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error
	LogoutPublicClient(ctx context.Context, clientID, realm, accessToken, refreshToken string) error
	LogoutAllSessions(ctx context.Context, accessToken, realm, userID string) error
	LogoutUserSession(ctx context.Context, accessToken, realm, session string) error
	LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*JWT, error)
	LoginClientSignedJWT(ctx context.Context, clientID, realm string, key interface{}, signedMethod jwt.SigningMethod, expiresAt *jwt.Time) (*JWT, error)
	LoginAdmin(ctx context.Context, username, password, realm string) (*JWT, error)
	RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*JWT, error)
	DecodeAccessToken(ctx context.Context, accessToken, realm, expectedAudience string) (*jwt.Token, *jwt.MapClaims, error)
	DecodeAccessTokenCustomClaims(ctx context.Context, accessToken, realm, expectedAudience string, claims jwt.Claims) (*jwt.Token, error)
	RetrospectToken(ctx context.Context, accessToken, clientID, clientSecret, realm string) (*RetrospecTokenResult, error)
	GetIssuer(ctx context.Context, realm string) (*IssuerResponse, error)
	GetCerts(ctx context.Context, realm string) (*CertResponse, error)
	GetServerInfo(ctx context.Context, accessToken string) (*ServerInfoRepesentation, error)
	GetUserInfo(ctx context.Context, accessToken, realm string) (*UserInfo, error)
	GetRawUserInfo(ctx context.Context, accessToken, realm string) (map[string]interface{}, error)
	SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error
	ExecuteActionsEmail(ctx context.Context, token, realm string, params ExecuteActionsEmail) error

	CreateUser(ctx context.Context, token, realm string, user User) (string, error)
	CreateGroup(ctx context.Context, accessToken, realm string, group Group) (string, error)
	CreateChildGroup(ctx context.Context, token, realm, groupID string, group Group) (string, error)
	CreateClientRole(ctx context.Context, accessToken, realm, clientID string, role Role) (string, error)
	CreateClient(ctx context.Context, accessToken, realm string, clientID Client) (string, error)
	CreateClientScope(ctx context.Context, accessToken, realm string, scope ClientScope) (string, error)
	CreateComponent(ctx context.Context, accessToken, realm string, component Component) (string, error)
	CreateClientScopeMappingsRealmRoles(ctx context.Context, token, realm, clientID string, roles []Role) error
	CreateClientScopeMappingsClientRoles(ctx context.Context, token, realm, clientID, clientsID string, roles []Role) error

	UpdateUser(ctx context.Context, accessToken, realm string, user User) error
	UpdateGroup(ctx context.Context, accessToken, realm string, updatedGroup Group) error
	UpdateRole(ctx context.Context, accessToken, realm, clientID string, role Role) error
	UpdateClient(ctx context.Context, accessToken, realm string, updatedClient Client) error
	UpdateClientScope(ctx context.Context, accessToken, realm string, scope ClientScope) error

	DeleteUser(ctx context.Context, accessToken, realm, userID string) error
	DeleteComponent(ctx context.Context, accessToken, realm, componentID string) error
	DeleteGroup(ctx context.Context, accessToken, realm, groupID string) error
	DeleteClientRole(ctx context.Context, accessToken, realm, clientID, roleName string) error
	DeleteClientRoleFromUser(ctx context.Context, token, realm, clientID, userID string, roles []Role) error
	DeleteClient(ctx context.Context, accessToken, realm, clientID string) error
	DeleteClientScope(ctx context.Context, accessToken, realm, scopeID string) error
	DeleteClientScopeMappingsRealmRoles(ctx context.Context, token, realm, clientID string, roles []Role) error
	DeleteClientScopeMappingsClientRoles(ctx context.Context, token, realm, clientID, clientsID string, roles []Role) error

	GetClient(ctx context.Context, accessToken, realm, clientID string) (*Client, error)
	GetClientsDefaultScopes(ctx context.Context, token, realm, clientID string) ([]*ClientScope, error)
	AddDefaultScopeToClient(ctx context.Context, token, realm, clientID, scopeID string) error
	RemoveDefaultScopeFromClient(ctx context.Context, token, realm, clientID, scopeID string) error
	GetClientsOptionalScopes(ctx context.Context, token, realm, clientID string) ([]*ClientScope, error)
	AddOptionalScopeToClient(ctx context.Context, token, realm, clientID, scopeID string) error
	RemoveOptionalScopeFromClient(ctx context.Context, token, realm, clientID, scopeID string) error
	GetDefaultOptionalClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error)
	GetDefaultDefaultClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error)
	GetClientScope(ctx context.Context, token, realm, scopeID string) (*ClientScope, error)
	GetClientScopes(ctx context.Context, token, realm string) ([]*ClientScope, error)
	GetClientScopeMappings(ctx context.Context, token, realm, clientID string) (*MappingsRepresentation, error)
	GetClientScopeMappingsRealmRoles(ctx context.Context, token, realm, clientID string) ([]*Role, error)
	GetClientScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, clientID string) ([]*Role, error)
	GetClientScopeMappingsClientRoles(ctx context.Context, token, realm, clientID, clientsID string) ([]*Role, error)
	GetClientScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, clientID, clientsID string) ([]*Role, error)
	GetClientSecret(ctx context.Context, token, realm, clientID string) (*CredentialRepresentation, error)
	GetClientServiceAccount(ctx context.Context, token, realm, clientID string) (*User, error)
	RegenerateClientSecret(ctx context.Context, token, realm, clientID string) (*CredentialRepresentation, error)
	GetKeyStoreConfig(ctx context.Context, accessToken, realm string) (*KeyStoreConfig, error)
	GetUserByID(ctx context.Context, accessToken, realm, userID string) (*User, error)
	GetUserCount(ctx context.Context, accessToken, realm string, params GetUsersParams) (int, error)
	GetUsers(ctx context.Context, accessToken, realm string, params GetUsersParams) ([]*User, error)
	GetUserGroups(ctx context.Context, accessToken, realm, userID string, params GetGroupsParams) ([]*UserGroup, error)
	AddUserToGroup(ctx context.Context, token, realm, userID, groupID string) error
	DeleteUserFromGroup(ctx context.Context, token, realm, userID, groupID string) error
	GetComponents(ctx context.Context, accessToken, realm string) ([]*Component, error)
	GetGroups(ctx context.Context, accessToken, realm string, params GetGroupsParams) ([]*Group, error)
	GetGroupsCount(ctx context.Context, token, realm string, params GetGroupsParams) (int, error)
	GetGroup(ctx context.Context, accessToken, realm, groupID string) (*Group, error)
	GetDefaultGroups(ctx context.Context, accessToken, realm string) ([]*Group, error)
	AddDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
	RemoveDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
	GetGroupMembers(ctx context.Context, accessToken, realm, groupID string, params GetGroupsParams) ([]*User, error)
	GetRoleMappingByGroupID(ctx context.Context, accessToken, realm, groupID string) (*MappingsRepresentation, error)
	GetRoleMappingByUserID(ctx context.Context, accessToken, realm, userID string) (*MappingsRepresentation, error)
	GetClientRoles(ctx context.Context, accessToken, realm, clientID string) ([]*Role, error)
	GetClientRole(ctx context.Context, token, realm, clientID, roleName string) (*Role, error)
	GetClientRoleByID(ctx context.Context, accessToken, realm, roleID string) (*Role, error)
	GetClients(ctx context.Context, accessToken, realm string, params GetClientsParams) ([]*Client, error)
	AddClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []Role) error
	DeleteClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []Role) error
	GetUsersByRoleName(ctx context.Context, token, realm, roleName string) ([]*User, error)
	GetUsersByClientRoleName(ctx context.Context, token, realm, clientID, roleName string, params GetUsersByRoleParams) ([]*User, error)
	CreateClientProtocolMapper(ctx context.Context, token, realm, clientID string, mapper ProtocolMapperRepresentation) (string, error)
	UpdateClientProtocolMapper(ctx context.Context, token, realm, clientID, mapperID string, mapper ProtocolMapperRepresentation) error
	DeleteClientProtocolMapper(ctx context.Context, token, realm, clientID, mapperID string) error

	// *** Realm Roles ***

	CreateRealmRole(ctx context.Context, token, realm string, role Role) (string, error)
	GetRealmRole(ctx context.Context, token, realm, roleName string) (*Role, error)
	GetRealmRoles(ctx context.Context, accessToken, realm string) ([]*Role, error)
	GetRealmRolesByUserID(ctx context.Context, accessToken, realm, userID string) ([]*Role, error)
	GetRealmRolesByGroupID(ctx context.Context, accessToken, realm, groupID string) ([]*Role, error)
	UpdateRealmRole(ctx context.Context, token, realm, roleName string, role Role) error
	DeleteRealmRole(ctx context.Context, token, realm, roleName string) error
	AddRealmRoleToUser(ctx context.Context, token, realm, userID string, roles []Role) error
	DeleteRealmRoleFromUser(ctx context.Context, token, realm, userID string, roles []Role) error
	AddRealmRoleToGroup(ctx context.Context, token, realm, groupID string, roles []Role) error
	DeleteRealmRoleFromGroup(ctx context.Context, token, realm, groupID string, roles []Role) error
	AddRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []Role) error
	DeleteRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []Role) error
	GetCompositeRealmRolesByRoleID(ctx context.Context, token, realm, roleID string) ([]*Role, error)
	GetCompositeRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*Role, error)
	GetCompositeRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*Role, error)
	GetAvailableRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*Role, error)
	GetAvailableRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*Role, error)

	// *** Client Roles ***

	AddClientRoleToUser(ctx context.Context, token, realm, clientID, userID string, roles []Role) error
	AddClientRoleToGroup(ctx context.Context, token, realm, clientID, groupID string, roles []Role) error
	DeleteClientRoleFromGroup(ctx context.Context, token, realm, clientID, groupID string, roles []Role) error
	GetCompositeClientRolesByRoleID(ctx context.Context, token, realm, clientID, roleID string) ([]*Role, error)
	GetClientRolesByUserID(ctx context.Context, token, realm, clientID, userID string) ([]*Role, error)
	GetClientRolesByGroupID(ctx context.Context, token, realm, clientID, groupID string) ([]*Role, error)
	GetCompositeClientRolesByUserID(ctx context.Context, token, realm, clientID, userID string) ([]*Role, error)
	GetCompositeClientRolesByGroupID(ctx context.Context, token, realm, clientID, groupID string) ([]*Role, error)
	GetAvailableClientRolesByUserID(ctx context.Context, token, realm, clientID, userID string) ([]*Role, error)
	GetAvailableClientRolesByGroupID(ctx context.Context, token, realm, clientID, groupID string) ([]*Role, error)

	// *** Realm ***

	GetRealm(ctx context.Context, token, realm string) (*RealmRepresentation, error)
	GetRealms(ctx context.Context, token string) ([]*RealmRepresentation, error)
	CreateRealm(ctx context.Context, token string, realm RealmRepresentation) (string, error)
	UpdateRealm(ctx context.Context, token string, realm RealmRepresentation) error
	DeleteRealm(ctx context.Context, token, realm string) error
	ClearRealmCache(ctx context.Context, token, realm string) error
	ClearUserCache(ctx context.Context, token, realm string) error
	ClearKeysCache(ctx context.Context, token, realm string) error

	GetClientUserSessions(ctx context.Context, token, realm, clientID string) ([]*UserSessionRepresentation, error)
	GetClientOfflineSessions(ctx context.Context, token, realm, clientID string) ([]*UserSessionRepresentation, error)
	GetUserSessions(ctx context.Context, token, realm, userID string) ([]*UserSessionRepresentation, error)
	GetUserOfflineSessionsForClient(ctx context.Context, token, realm, userID, clientID string) ([]*UserSessionRepresentation, error)

	// *** Protection API ***
	GetResource(ctx context.Context, token, realm, clientID, resourceID string) (*ResourceRepresentation, error)
	GetResources(ctx context.Context, token, realm, clientID string, params GetResourceParams) ([]*ResourceRepresentation, error)
	CreateResource(ctx context.Context, token, realm, clientID string, resource ResourceRepresentation) (*ResourceRepresentation, error)
	UpdateResource(ctx context.Context, token, realm, clientID string, resource ResourceRepresentation) error
	DeleteResource(ctx context.Context, token, realm, clientID, resourceID string) error

	GetResourceClient(ctx context.Context, token, realm, resourceID string) (*ResourceRepresentation, error)
	GetResourcesClient(ctx context.Context, token, realm string, params GetResourceParams) ([]*ResourceRepresentation, error)
	CreateResourceClient(ctx context.Context, token, realm string, resource ResourceRepresentation) (*ResourceRepresentation, error)
	UpdateResourceClient(ctx context.Context, token, realm string, resource ResourceRepresentation) error
	DeleteResourceClient(ctx context.Context, token, realm, resourceID string) error

	GetScope(ctx context.Context, token, realm, clientID, scopeID string) (*ScopeRepresentation, error)
	GetScopes(ctx context.Context, token, realm, clientID string, params GetScopeParams) ([]*ScopeRepresentation, error)
	CreateScope(ctx context.Context, token, realm, clientID string, scope ScopeRepresentation) (*ScopeRepresentation, error)
	UpdateScope(ctx context.Context, token, realm, clientID string, resource ScopeRepresentation) error
	DeleteScope(ctx context.Context, token, realm, clientID, scopeID string) error

	GetPolicy(ctx context.Context, token, realm, clientID, policyID string) (*PolicyRepresentation, error)
	GetPolicies(ctx context.Context, token, realm, clientID string, params GetPolicyParams) ([]*PolicyRepresentation, error)
	CreatePolicy(ctx context.Context, token, realm, clientID string, policy PolicyRepresentation) (*PolicyRepresentation, error)
	UpdatePolicy(ctx context.Context, token, realm, clientID string, policy PolicyRepresentation) error
	DeletePolicy(ctx context.Context, token, realm, clientID, policyID string) error

	GetResourcePolicy(ctx context.Context, token, realm, permissionID string) (*ResourcePolicyRepresentation, error) 
	GetResourcePolicies(ctx context.Context, token, realm string, params GetResourcePoliciesParams) ([]*ResourcePolicyRepresentation, error) 
	CreateResourcePolicy(ctx context.Context, token, realm, resourceID string, policy ResourcePolicyRepresentation) (*ResourcePolicyRepresentation, error) 
	UpdateResourcePolicy(ctx context.Context, token, realm, permissionID string, policy ResourcePolicyRepresentation) error 
	DeleteResourcePolicy(ctx context.Context, token, realm, permissionID string) error 

	GetPermission(ctx context.Context, token, realm, clientID, permissionID string) (*PermissionRepresentation, error)
	GetPermissions(ctx context.Context, token, realm, clientID string, params GetPermissionParams) ([]*PermissionRepresentation, error)
	GetPermissionResources(ctx context.Context, token, realm, clientID, permissionID string) ([]*PermissionResource, error)
	GetPermissionScopes(ctx context.Context, token, realm, clientID, permissionID string) ([]*PermissionScope, error)
	GetDependentPermissions(ctx context.Context, token, realm, clientID, policyID string) ([]*PermissionRepresentation, error)
	CreatePermission(ctx context.Context, token, realm, clientID string, permission PermissionRepresentation) (*PermissionRepresentation, error)
	UpdatePermission(ctx context.Context, token, realm, clientID string, permission PermissionRepresentation) error
	DeletePermission(ctx context.Context, token, realm, clientID, permissionID string) error

	CreatePermissionTicket(ctx context.Context, token, realm string, permissions []CreatePermissionTicketParams) (*PermissionTicketResponseRepresentation, error)
	GrantUserPermission(ctx context.Context, token, realm string, permission PermissionGrantParams) (*PermissionGrantResponseRepresentation, error)
	UpdateUserPermission(ctx context.Context, token, realm string, permission PermissionGrantParams) (*PermissionGrantResponseRepresentation, error)
	GetUserPermissions(ctx context.Context, token, realm string, params GetUserPermissionParams) ([]*PermissionGrantResponseRepresentation, error)
	DeleteUserPermission(ctx context.Context, token, realm, ticketID string) error

	// *** Credentials API ***

	GetCredentialRegistrators(ctx context.Context, token, realm string) ([]string, error)
	GetConfiguredUserStorageCredentialTypes(ctx context.Context, token, realm, userID string) ([]string, error)
	GetCredentials(ctx context.Context, token, realm, UserID string) ([]*CredentialRepresentation, error)
	DeleteCredentials(ctx context.Context, token, realm, UserID, CredentialID string) error
	UpdateCredentialUserLabel(ctx context.Context, token, realm, userID, credentialID, userLabel string) error
	DisableAllCredentialsByType(ctx context.Context, token, realm, userID string, types []string) error
	MoveCredentialBehind(ctx context.Context, token, realm, userID, credentialID, newPreviousCredentialID string) error
	MoveCredentialToFirst(ctx context.Context, token, realm, userID, credentialID string) error

	// *** Identity Providers ***

	CreateIdentityProvider(ctx context.Context, token, realm string, providerRep IdentityProviderRepresentation) (string, error)
	GetIdentityProvider(ctx context.Context, token, realm, alias string) (*IdentityProviderRepresentation, error)
	GetIdentityProviders(ctx context.Context, token, realm string) ([]*IdentityProviderRepresentation, error)
	UpdateIdentityProvider(ctx context.Context, token, realm, alias string, providerRep IdentityProviderRepresentation) error
	DeleteIdentityProvider(ctx context.Context, token, realm, alias string) error

	CreateUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string, federatedIdentityRep FederatedIdentityRepresentation) error
	GetUserFederatedIdentities(ctx context.Context, token, realm, userID string) ([]*FederatedIdentityRepresentation, error)
	DeleteUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string) error

}
```

## Configure gocloak to skip TLS Insecure Verification

```go
    client := gocloak.NewClient(serverURL)
    restyClient := client.RestyClient()
    restyClient.SetDebug(true)
    restyClient.SetTLSClientConfig(&tls.Config{ InsecureSkipVerify: true }
```

## developing & testing
For local testing you need to start a docker container. Simply run following commands prior to starting the tests:

```shell
docker pull quay.io/keycloak/keycloak
docker run -d \
	-e KEYCLOAK_USER=admin \
	-e KEYCLOAK_PASSWORD=secret \
	-e KEYCLOAK_IMPORT=/tmp/gocloak-realm.json \
	-v "`pwd`/testdata/gocloak-realm.json:/tmp/gocloak-realm.json" \
	-p 8080:8080 \
	--name gocloak-test \
	quay.io/keycloak/keycloak:latest -Dkeycloak.profile.feature.upload_scripts=enabled

go test
```

Or you can run with docker compose using the run-tests script
```shell
./run-tests.sh
```
or
```shell
./run-tests.sh <TestCase>
```


Or you can run the tests on you own keycloak:
```shell
export GOCLOAK_TEST_CONFIG=/path/to/gocloak/config.json
```

All resources created as a result of unit tests will be deleted, except for the test user defined in the configuration file.

To remove running docker container after completion of tests:

```shell
docker stop gocloak-test
docker rm gocloak-test
```

### Inspecting custom types

The custom types contain many pointers, so printing them yields mostly pointer values, which aren't much help when debugging your application. For example

```
someRealmRepresentation := gocloak.RealmRepresentation{
   <snip>
}

fmt.Println(someRealmRepresentation)

```
yields a large set of pointer values
```
{<nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> 0xc00000e960 <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> 0xc000093cf0 <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> null <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil> <nil>}
```
For convenience, the ```String()``` interface has been added so you can easily see the contents, even for nested custom types. For example,

```
fmt.Println(someRealmRepresentation.String())
```
yields

```
{
	"clients": [
		{
			"name": "someClient",
			"protocolMappers": [
				{
					"config": {
						"bar": "foo",
						"ping": "pong"
					},
					"name": "someMapper"
				}
			]
		},
		{
			"name": "AnotherClient"
		}
	],
	"displayName": "someRealm"
}
```
Note that empty parameters are not included, because of the use of ```omitempty``` in the type definitions.


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_large)
