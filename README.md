# gocloak
golang keycloak client

This client is based on : https://github.com/PhilippHeuer/go-keycloak/blob/master/client.go

## Features

```go
type Client interface {
    Login(username string, password string, realm string) (*models.JWT, error)

    DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*models.JWT, error)
    GetUsers(token *models.JWT, realm string) (*[]models.User, error)
    GetUserGroups(token *models.JWT, realm string, userID string) (*[]models.UserGroup, error)
    GetRoleMappingByGroupID(token *models.JWT, realm string, groupID string) (*[]models.RoleMapping, error)
    GetGroups(token *models.JWT, realm string) (*[]models.Group, error)
    GetRoles(token *models.JWT, realm string) (*[]models.Role, error)
    GetRolesByClientID(token *models.JWT, realm string, clientID string) (*[]models.Role, error)
    GetClients(token *models.JWT, realm string) (*[]models.RealmClient, error)
}
```

## developing & testing
As I was to lazy to add some environment variables. So i added a "super.secret.go" file, which holds some constants(username, password, realm), that are used for the tests.
