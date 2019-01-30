# gocloak
golang keycloak client

This client is based on : https://github.com/PhilippHeuer/go-keycloak/blob/master/client.go

Use this together with the keycloak client [gocloak-echo](https://github.com/Nerzal/gocloak-echo)

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

## Features

```go
// GoCloak holds all methods a client should fullfill
type GoCloak interface {
	Login(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error)
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	LoginAdmin(username, password, realm string) (*JWT, error)
	RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error)
    DecodeAccessToken(accessToken string, adminAccessToken string, realm string) (*jwt.Token, *jwt.MapClaims, error)
    DecodeAccessTokenCustomClaims(accessToken string, adminAccessToken string, realm string) (*jwt.Token, *jwx.Claims, error)

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
	GetUser(accessToken string, realm, userID string) (*User, error)
	GetUserByID(accessToken string, realm string, userID string) (*User, error)
	GetUserCount(accessToken string, realm string) (int, error)
	GetUsers(accessToken string, realm string) (*[]User, error)
	GetUserGroups(accessToken string, realm string, userID string) (*[]UserGroup, error)
	GetComponents(accessToken string, realm string) (*[]Component, error)

	GetGroups(accessToken string, realm string) (*[]Group, error)
	GetGroup(accessToken string, realm, groupID string) (*Group, error)
	GetRoles(accessToken string, realm string) (*[]Role, error)
	GetRoleMappingByGroupID(accessToken string, realm string, groupID string) (*[]RoleMapping, error)
	GetRolesByClientID(accessToken string, realm string, clientID string) (*[]Role, error)
	GetClients(accessToken string, realm string) (*[]Client, error)

	UserAttributeContains(attributes map[string][]string, attribute string, value string) bool
}

```

## developing & testing
I was to lazy to add some environment variables. So i added a "super.secret.go" file, which holds some constants(username, password, realm), that are used for the tests.
