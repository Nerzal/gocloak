# gocloak
[![codebeat badge](https://codebeat.co/badges/c699bc56-aa5f-4cf5-893f-5cf564391b94)](https://codebeat.co/projects/github-com-nerzal-gocloak-master)
[![Go Report Card](https://goreportcard.com/badge/github.com/Nerzal/gocloak)](https://goreportcard.com/report/github.com/Nerzal/gocloak)
[![Go Doc](https://godoc.org/github.com/Nerzal/gocloak?status.svg)](https://godoc.org/github.com/Nerzal/gocloak)

golang keycloak client

This client is based on : [go-keycloak](https://github.com/PhilippHeuer/go-keycloak)

For Questions either raise an issue, or come to the [gopher-slack](https://invite.slack.golangbridge.org/) into the channel #gocloak

If u are using the echo framework have a look at [gocloak-echo](https://github.com/Nerzal/gocloak-echo)

### Keycloak Version < 4.8
If you are using a Keycloak Server version <4.8 please use the V1.0 release of keycloak.

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
	gocloak := gocloak.NewClient("https://mycool.keycloak.instance")
	token, err := gocloak.LoginAdmin("user", "password", "realmName")
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
	gocloak.CreateUser(token.AccessToken, "realm", user)
	if err != nil {
		panic("Oh no!, failed to create user :(")
	}
```

### Introspect Token
```go
	client := NewClient(hostname)
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
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	LoginAdmin(username, password, realm string) (*JWT, error)
	RequestPermission(clientID string, clientSecret string, realm string, username string, password string, permission string) (*JWT, error)
	RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error)
	DecodeAccessToken(accessToken string, realm string) (*jwt.Token, *jwt.MapClaims, error)
	DecodeAccessTokenCustomClaims(accessToken string, realm string, claims jwt.Claims) (*jwt.Token, error)
	RetrospectToken(accessToken string, clientID, clientSecret string, realm string) (*RetrospecTokenResult, error)

	GetIssuer(realm string) (*IssuerResponse, error)
	GetCerts(realm string) (*CertResponse, error)

	SetPassword(token string, userID string, realm string, password string, temporary bool) error
	CreateUser(token string, realm string, user User) (*string, error)
	CreateGroup(accessToken string, realm string, group Group) error
	CreateRole(accessToken string, realm string, clientID string, role Role) error
	CreateClient(accessToken string, realm string, clientID Client) error
	CreateClientScope(accessToken string, realm string, scope ClientScope) error
	CreateComponent(accessToken string, realm string, component Component) error

	UpdateUser(accessToken string, realm string, user User) error
	UpdateGroup(accessToken string, realm string, group Group) error
	UpdateRole(accessToken string, realm string, clientID string, role Role) error
	UpdateClient(accessToken string, realm string, clientID Client) error
	UpdateClientScope(accessToken string, realm string, scope ClientScope) error

	DeleteUser(accessToken string, realm, userID string) error
	DeleteComponent(accessToken string, realm, componentID string) error
	DeleteGroup(accessToken string, realm, groupID string) error
	DeleteRole(accessToken string, realm, clientID, roleName string) error
	DeleteClient(accessToken string, realm, clientID string) error
	DeleteClientScope(accessToken string, realm, scopeID string) error

	GetKeyStoreConfig(accessToken string, realm string) (*KeyStoreConfig, error)
	GetUserByID(accessToken string, realm string, userID string) (*User, error)
	GetUserCount(accessToken string, realm string) (int, error)
	GetUsers(accessToken string, realm string) (*[]User, error)
	GetUserGroups(accessToken string, realm string, userID string) (*[]UserGroup, error)
	GetComponents(accessToken string, realm string) (*[]Component, error)

	UserAttributeContains(attributes map[string][]string, attribute string, value string) bool

	GetGroups(accessToken string, realm string) (*[]Group, error)
	GetGroup(accessToken string, realm, groupID string) (*Group, error)
	GetRoles(accessToken string, realm string) (*[]Role, error)
	GetRoleMappingByGroupID(accessToken string, realm string, groupID string) (*[]RoleMapping, error)
	GetRoleMappingByUserID(accessToken string, realm string, userID string) (*[]RoleMapping, error)
	GetRolesByClientID(accessToken string, realm string, clientID string) (*[]Role, error)
	GetClients(accessToken string, realm string) (*[]Client, error)
	GetRealmRolesByUserID(accessToken string, realm string, userID string) (*[]Role, error)
	GetRealmRolesByGroupID(accessToken string, realm string, groupID string) (*[]Role, error)
	GetUsersByRoleName(token string, realm string, roleName string) (*[]User, error)
}
```

## developing & testing
I was to lazy to add some environment variables. So i added a "super.secret.go" file, which holds some constants(username, password, realm), that are used for the tests.
