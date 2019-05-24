# gocloak
[![codebeat badge](https://codebeat.co/badges/c699bc56-aa5f-4cf5-893f-5cf564391b94)](https://codebeat.co/projects/github-com-nerzal-gocloak-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nerzal/gocloak)](https://goreportcard.com/report/github.com/Nerzal/gocloak)
[![Go Doc](https://godoc.org/github.com/Nerzal/gocloak?status.svg)](https://godoc.org/github.com/Nerzal/gocloak)
[![Build Status](https://travis-ci.com/Nerzal/gocloak.svg?branch=master)](https://travis-ci.com/Nerzal/gocloak)
[![codecov](https://codecov.io/gh/Nerzal/gocloak/branch/master/graph/badge.svg)](https://codecov.io/gh/Nerzal/gocloak)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_shield)



golang keycloak client

This client is based on : [go-keycloak](https://github.com/PhilippHeuer/go-keycloak)

For Questions either raise an issue, or come to the [gopher-slack](https://invite.slack.golangbridge.org/) into the channel #gocloak

If u are using the echo framework have a look at [gocloak-echo](https://github.com/Nerzal/gocloak-echo)

### Keycloak Version < 4.8
If you are using a Keycloak Server version <4.8 please use the V1.0 release of gocloak.

```go
	go get gopkg.in/nerzal/gocloak.v1
``` 

```go
	import "gopkg.in/nerzal/gocloak.v1"
``` 

https://gopkg.in/nerzal/gocloak.v1

## Usage

### Create New User
```go
	client := gocloak.NewClient("https://mycool.keycloak.instance")
	token, err := client.LoginAdmin("user", "password", "realmName")
	if err != nil {
		panic("Something wrong with the credentials or url")
	}
	user := gocloak.User{
		FirstName: "Bob",
		LastName:  "Uncle",
		EMail:     "something@really.wrong",
		Enabled:   true,
		Username:  "CoolGuy",
	}
	client.CreateUser(token.AccessToken, "realm", user)
	if err != nil {
		panic("Oh no!, failed to create user :(")
	}
```

### Introspect Token
```go
	client := gocloak.NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		panic("Login failed:"+ err.Error())
	}

	rptResult, err := client.RetrospectToken(token.AccessToken, clientid, clientSecret, realm)
	if err != nil {
		panic("Inspection failed:"+ err.Error())
	}

	if !rptResult.Active {
		panic("Token is not active")
	}

	permissions := rptResult.Permissions
	//Do something with the permissions ;) 
```

## Features

```go
// GoCloak holds all methods a client should fullfill
type GoCloak interface {
	Login(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error)
	Logout(clientID, clientSecret, realm, refreshToken string) error
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	LoginAdmin(username, password, realm string) (*JWT, error)
	RequestPermission(clientID string, clientSecret string, realm string, username string, password string, permission string) (*JWT, error)
	RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error)
	DecodeAccessToken(accessToken string, realm string) (*jwt.Token, *jwt.MapClaims, error)
	DecodeAccessTokenCustomClaims(accessToken string, realm string, claims jwt.Claims) (*jwt.Token, error)
	RetrospectToken(accessToken string, clientID, clientSecret string, realm string) (*RetrospecTokenResult, error)
	GetIssuer(realm string) (*IssuerResponse, error)
	GetCerts(realm string) (*CertResponse, error)
	GetUserInfo(accessToken string, realm string) (*UserInfo, error)
	SetPassword(token string, userID string, realm string, password string, temporary bool) error
	ExecuteActionsEmail(token string, realm string, params ExecuteActionsEmail) error

	CreateUser(token string, realm string, user User) (*string, error)
	CreateGroup(accessToken string, realm string, group Group) error
	CreateClientRole(accessToken string, realm string, clientID string, role Role) error
	CreateClient(accessToken string, realm string, clientID Client) error
	CreateClientScope(accessToken string, realm string, scope ClientScope) error
	CreateComponent(accessToken string, realm string, component Component) error

	UpdateUser(accessToken string, realm string, user User) error
	UpdateGroup(accessToken string, realm string, updatedGroup Group) error
	UpdateRole(accessToken string, realm string, clientID string, role Role) error
	UpdateClient(accessToken string, realm string, clientID Client) error
	UpdateClientScope(accessToken string, realm string, scope ClientScope) error

	DeleteUser(accessToken string, realm, userID string) error
	DeleteComponent(accessToken string, realm, componentID string) error
	DeleteGroup(accessToken string, realm, groupID string) error
	DeleteClientRole(accessToken string, realm, clientID, roleName string) error
	DeleteClient(accessToken string, realm, clientID string) error
	DeleteClientScope(accessToken string, realm, scopeID string) error

	GetClient(accessToken string, realm string, clientID string) (*Client, error)
	GetClientSecret(token string, realm string, clientID string) (*CredentialRepresentation, error)
	GetKeyStoreConfig(accessToken string, realm string) (*KeyStoreConfig, error)
	GetUserByID(accessToken string, realm string, userID string) (*User, error)
	GetUserCount(accessToken string, realm string) (int, error)
	GetUsers(accessToken string, realm string, params GetUsersParams) (*[]User, error)
	GetUserGroups(accessToken string, realm string, userID string) (*[]UserGroup, error)
	GetComponents(accessToken string, realm string) (*[]Component, error)
	GetGroups(accessToken string, realm string, params GetGroupsParams) (*[]Group, error)
	GetGroup(accessToken string, realm, groupID string) (*Group, error)
	GetRoleMappingByGroupID(accessToken string, realm string, groupID string) (*MappingsRepresentation, error)
	GetRoleMappingByUserID(accessToken string, realm string, userID string) (*MappingsRepresentation, error)
	GetClientRoles(accessToken string, realm string, clientID string) (*[]Role, error)
	GetClientRole(token string, realm string, clientID string, roleName string) (*Role, error)
	GetClients(accessToken string, realm string, params GetClientsParams) (*[]Client, error)
	GetUsersByRoleName(token string, realm string, roleName string) (*[]User, error)
	UserAttributeContains(attributes map[string][]string, attribute string, value string) bool

	// *** Realm Roles ***

	CreateRealmRole(token string, realm string, role Role) error
	GetRealmRole(token string, realm string, roleName string) (*Role, error)
	GetRealmRoles(accessToken string, realm string) (*[]Role, error)
	GetRealmRolesByUserID(accessToken string, realm string, userID string) (*[]Role, error)
	GetRealmRolesByGroupID(accessToken string, realm string, groupID string) (*[]Role, error)
	UpdateRealmRole(token string, realm string, roleName string, role Role) error
	DeleteRealmRole(token string, realm string, roleName string) error
	AddRealmRoleToUser(token string, realm string, userID string, roles []Role) error
	DeleteRealmRoleFromUser(token string, realm string, userID string, roles []Role) error
	AddRealmRoleComposite(token string, realm string, roleName string, roles []Role) error
	DeleteRealmRoleComposite(token string, realm string, roleName string, roles []Role) error

	// *** Realm ***

	GetRealm(token string, realm string) (*RealmRepresentation, error)
	CreateRealm(token string, realm RealmRepresentation) error
	DeleteRealm(token string, realm string) error

	GetClientUserSessions(token, realm, clientID string) (*[]UserSessionRepresentation, error)
	GetClientOfflineSessions(token, realm, clientID string) (*[]UserSessionRepresentation, error)
	GetUserSessions(token, realm, userID string) (*[]UserSessionRepresentation, error)
	GetUserOfflineSessionsForClient(token, realm, userID, clientID string) (*[]UserSessionRepresentation, error)
}
```

## developing & testing
For local testing you need to start a docker container. Simply run following commands prior to starting the tests:

```bash
docker pull jboss/keycloak
docker run -d -e KEYCLOAK_USER=admin -e KEYCLOAK_PASSWORD=secret -e KEYCLOAK_IMPORT=/tmp/gocloak-realm.json -v `pwd`/testdata/gocloak-realm.json:/tmp/gocloak-realm.json -p 8080:8080 --name keycloak jboss/keycloak
go test
```

Or you can run the tests on you own keycloak:
```bash
export GOCLOAK_TEST_CONFIG=/path/to/gocloak/config.json
```

All resources created as a result of unit tests will be deleted, except for the test user defined in the configuration file.


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FNerzal%2Fgocloak.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FNerzal%2Fgocloak?ref=badge_large)