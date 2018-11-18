# gocloak
golang keycloak client

This client is based on : https://github.com/PhilippHeuer/go-keycloak/blob/master/client.go

## Features

```go
type Client interface {
Login(username string, password string, realm string, clientID string) (*models.JWT, error)
	Login(username string, password string, realm string, clientID string) (*models.JWT, error)
	LoginAdmin(username, password, realm string) (*models.JWT, error)

	DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*models.JWT, error)

	CreateUser(token *models.JWT, realm string, user models.User) error
	CreateGroup(token *models.JWT, realm string, group models.Group) error
	CreateRole(token *models.JWT, realm string, clientID string, role models.Role) error
	CreateClient(token *models.JWT, realm string, clientID models.Client) error
	CreateClientScope(token *models.JWT, realm string, scope models.ClientScope) error
	CreateComponent(token *models.JWT, realm string, component models.Component) error

	UpdateUser(token *models.JWT, realm string, user models.User) error
	UpdateGroup(token *models.JWT, realm string, group models.Group) error
	UpdateRole(token *models.JWT, realm string, clientID string, role models.Role) error
	UpdateClient(token *models.JWT, realm string, clientID models.Client) error
	UpdateClientScope(token *models.JWT, realm string, scope models.ClientScope) error

	DeleteUser(token *models.JWT, realm, userID string) error
	DeleteComponent(token *models.JWT, realm, componentID string) error
	DeleteGroup(token *models.JWT, realm, groupID string) error
	DeleteRole(token *models.JWT, realm, clientID, roleName string) error
	DeleteClient(token *models.JWT, realm, clientID string) error
	DeleteClientScope(token *models.JWT, realm, scopeID string) error

	GetKeyStoreConfig(token *models.JWT, realm string) (*models.KeyStoreConfig, error)
	GetUser(token *models.JWT, realm, userID string) (*models.User, error)
	GetUserCount(token *models.JWT, realm string) (int, error)
	GetUsers(token *models.JWT, realm string) (*[]models.User, error)
	GetUserGroups(token *models.JWT, realm string, userID string) (*[]models.UserGroup, error)
	GetComponents(token *models.JWT, realm string) (*[]models.Component, error)

	GetGroups(token *models.JWT, realm string) (*[]models.Group, error)
	GetGroup(token *models.JWT, realm, groupID string) (*models.Group, error)
	GetRoles(token *models.JWT, realm string) (*[]models.Role, error)
	GetRoleMappingByGroupID(token *models.JWT, realm string, groupID string) (*[]models.RoleMapping, error)
	GetRolesByClientID(token *models.JWT, realm string, clientID string) (*[]models.Role, error)
	GetClients(token *models.JWT, realm string) (*[]models.Client, error)
}
```

## developing & testing
As I was to lazy to add some environment variables. So i added a "super.secret.go" file, which holds some constants(username, password, realm), that are used for the tests.
