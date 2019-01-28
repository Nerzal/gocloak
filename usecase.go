package gocloak

import (
	"github.com/Nerzal/gocloak/pkg/jwx"
	jwt "github.com/dgrijalva/jwt-go"
)

// GoCloak holds all methods a client should fullfill
type GoCloak interface {
	Login(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error)
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	LoginAdmin(username, password, realm string) (*JWT, error)
	RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error)
	DecodeAccessToken(accessToken string, realm string) (*jwt.Token, *jwt.MapClaims, error)
	DecodeAccessTokenCustomClaims(accessToken string, realm string) (*jwt.Token, *jwx.Claims, error)

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
}
