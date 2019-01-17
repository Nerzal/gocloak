package gocloak

// GoCloak holds all methods a client should fullfill
type GoCloak interface {
	Login(username string, password string, realm string, clientID string) (*JWT, error)
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	LoginAdmin(username, password, realm string) (*JWT, error)
	RefreshToken(refreshToken string, clientID, realm string) (*JWT, error)
	ValidateToken(token string, realm string) error

	DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error)

	CreateUser(token *JWT, realm string, user User) error
	CreateGroup(token *JWT, realm string, group Group) error
	CreateRole(token *JWT, realm string, clientID string, role Role) error
	CreateClient(token *JWT, realm string, clientID Client) error
	CreateClientScope(token *JWT, realm string, scope ClientScope) error
	CreateComponent(token *JWT, realm string, component Component) error

	UpdateUser(token *JWT, realm string, user User) error
	UpdateGroup(token *JWT, realm string, group Group) error
	UpdateRole(token *JWT, realm string, clientID string, role Role) error
	UpdateClient(token *JWT, realm string, clientID Client) error
	UpdateClientScope(token *JWT, realm string, scope ClientScope) error

	DeleteUser(token *JWT, realm, userID string) error
	DeleteComponent(token *JWT, realm, componentID string) error
	DeleteGroup(token *JWT, realm, groupID string) error
	DeleteRole(token *JWT, realm, clientID, roleName string) error
	DeleteClient(token *JWT, realm, clientID string) error
	DeleteClientScope(token *JWT, realm, scopeID string) error

	GetKeyStoreConfig(token *JWT, realm string) (*KeyStoreConfig, error)
	GetUser(token *JWT, realm, userID string) (*User, error)
	GetUserCount(token *JWT, realm string) (int, error)
	GetUsers(token *JWT, realm string) (*[]User, error)
	GetUserGroups(token *JWT, realm string, userID string) (*[]UserGroup, error)
	GetComponents(token *JWT, realm string) (*[]Component, error)

	GetGroups(token *JWT, realm string) (*[]Group, error)
	GetGroup(token *JWT, realm, groupID string) (*Group, error)
	GetRoles(token *JWT, realm string) (*[]Role, error)
	GetRoleMappingByGroupID(token *JWT, realm string, groupID string) (*[]RoleMapping, error)
	GetRolesByClientID(token *JWT, realm string, clientID string) (*[]Role, error)
	GetClients(token *JWT, realm string) (*[]Client, error)
}
