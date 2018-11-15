package models

// User represents the Keycloak User Structure
type User struct {
	ID                         string        `json:"id"`
	CreatedTimestamp           int64         `json:"createdTimestamp"`
	Username                   string        `json:"username"`
	Enabled                    bool          `json:"enabled"`
	Totp                       bool          `json:"totp"`
	EmailVerified              bool          `json:"emailVerified"`
	FirstName                  string        `json:"firstName"`
	LastName                   string        `json:"lastName"`
	Email                      string        `json:"email"`
	FederationLink             string        `json:"federationLink"`
	Attributes                 Attributes    `json:"attributes"`
	DisableableCredentialTypes []interface{} `json:"disableableCredentialTypes"`
	RequiredActions            []interface{} `json:"requiredActions"`
	Access                     Access        `json:"access"`
}

// Attributes holds Attributes
type Attributes struct {
	LDAPENTRYDN []string `json:"LDAP_ENTRY_DN"`
	LDAPID      []string `json:"LDAP_ID"`
}

// Access represents access
type Access struct {
	ManageGroupMembership bool `json:"manageGroupMembership"`
	View                  bool `json:"view"`
	MapRoles              bool `json:"mapRoles"`
	Impersonate           bool `json:"impersonate"`
	Manage                bool `json:"manage"`
}

// UserGroup is a UserGroup
type UserGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// Group is a Group
type Group struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Path      string        `json:"path"`
	SubGroups []interface{} `json:"subGroups"`
}

// Role is a role
type Role struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ScopeParamRequired bool   `json:"scopeParamRequired"`
	Composite          bool   `json:"composite"`
	ClientRole         bool   `json:"clientRole"`
	ContainerID        string `json:"containerId"`
	Description        string `json:"description,omitempty"`
}

// RoleMapping is a role mapping
type RoleMapping struct {
	ID       string                  `json:"id"`
	Client   string                  `json:"client"`
	Mappings []ClientRoleMappingRole `json:"mappings"`
}

// ClientRoleMappingRole is a client role mapping role
type ClientRoleMappingRole struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	ScopeParamRequired bool   `json:"scopeParamRequired"`
	Composite          bool   `json:"composite"`
	ClientRole         bool   `json:"clientRole"`
	ContainerID        string `json:"containerId"`
}

// RealmClient is a realmClient
type RealmClient struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
}
