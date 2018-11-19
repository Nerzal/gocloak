package core

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	resty "gopkg.in/resty.v1"
)

// GoCloak holds all methods a client should fullfill
type GoCloak interface {
	Login(username string, password string, realm string, clientID string) (*JWT, error)
	LoginClient(clientID, clientSecret, realm string) (*JWT, error)
	LoginAdmin(username, password, realm string) (*JWT, error)

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

type gocloak struct {
	basePath string
}

type loginData struct {
	ClientID  string `json:"client_id"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}

const adminClientID string = "admin-cli"

// NewClient creates a new Client
func NewClient(basePath string) GoCloak {
	return &gocloak{
		basePath: basePath,
	}
}

// Login performs a login
func (client *gocloak) LoginAdmin(username, password, realm string) (*JWT, error) {
	return client.Login(username, password, realm, adminClientID)
}

// Login performs a login
func (client *gocloak) Login(username, password, realm, clientID string) (*JWT, error) {
	firstPart := "/auth/realms/"
	lastPart := "/protocol/openid-connect/token"
	loginPath := firstPart + realm + lastPart

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Add("grant_type", "password")
	data.Add("username", username)
	data.Add("password", password)

	req, _ := http.NewRequest("POST", client.basePath+loginPath, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		log.Println(string(body))
	}

	jwt := &JWT{}
	err = json.Unmarshal(body, jwt)
	return jwt, err
}

// Login performs a login
func (client *gocloak) LoginClient(clientID, clientSecret, realm string) (*JWT, error) {
	firstPart := "/auth/realms/"
	lastPart := "/protocol/openid-connect/token"
	loginPath := firstPart + realm + lastPart

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Add("grant_type", "client_credentials")
	data.Add("client_secret", clientSecret)

	req, _ := http.NewRequest("POST", client.basePath+loginPath, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		log.Println(string(body))
	}

	jwt := &JWT{}
	err = json.Unmarshal(body, jwt)
	return jwt, err
}

// DirectGrantAuthentication like login, but with basic auth
func (client *gocloak) DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", getBasicAuthForClient(clientID, clientSecret)).
		SetFormData(map[string]string{
			"grant_type": "password",
			"username":   username,
			"password":   password,
		}).Post(client.basePath + "/auth/realms/" + realm + "/protocol/openid-connect/token")
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	if val, ok := result["access_token"]; ok {
		_ = val
		return &JWT{
			AccessToken:      result["access_token"].(string),
			ExpiresIn:        result["expires_in"].(int),
			RefreshExpiresIn: result["refresh_expires_in"].(int),
			RefreshToken:     result["refresh_token"].(string),
			TokenType:        result["token_type"].(string),
		}, nil
	}

	return nil, errors.New("Authentication failed")
}

// CreateUser creates a new user
func (client *gocloak) CreateUser(token *JWT, realm string, user User) error {
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/users")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateGroup(token *JWT, realm string, group Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/groups")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateComponent creates a new user
func (client *gocloak) CreateComponent(token *JWT, realm string, component Component) error {
	bytes, err := json.Marshal(component)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/components")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateClient(token *JWT, realm string, newClient Client) error {
	bytes, err := json.Marshal(newClient)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/clients")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateRole(token *JWT, realm string, clientID string, role Role) error {
	bytes, err := json.Marshal(role)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "clients/" + clientID + "/roles")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateClientScope creates a new client scope
func (client *gocloak) CreateClientScope(token *JWT, realm string, scope ClientScope) error {
	bytes, err := json.Marshal(scope)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/client-scopes")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateUser(token *JWT, realm string, user User) error {
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + "/auth/admin/realms/" + realm + "/users/" + user.ID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateGroup(token *JWT, realm string, group Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + "/auth/admin/realms/" + realm + "/groups/" + group.ID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateClient(token *JWT, realm string, newClient Client) error {
	bytes, err := json.Marshal(newClient)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + "/auth/admin/realms/" + realm + "/clients")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateRole(token *JWT, realm string, clientID string, role Role) error {
	bytes, err := json.Marshal(role)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + "/auth/admin/realms/" + realm + "clients/" + clientID + "/roles/" + role.Name)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateClientScope creates a new client scope
func (client *gocloak) UpdateClientScope(token *JWT, realm string, scope ClientScope) error {
	bytes, err := json.Marshal(scope)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + "/auth/admin/realms/" + realm + "/client-scopes/" + scope.ID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteUser(token *JWT, realm, userID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + "/auth/admin/realms/" + realm + "/users/" + userID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteGroup(token *JWT, realm, groupID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + "/auth/admin/realms/" + realm + "/groups/" + groupID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteClient(token *JWT, realm, clientID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + "/auth/admin/realms/" + realm + "/clients/" + clientID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteComponent creates a new user
func (client *gocloak) DeleteComponent(token *JWT, realm, componentID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + "/auth/admin/realms/" + realm + "/components/" + componentID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteRole(token *JWT, realm, clientID, roleName string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + "/auth/admin/realms/" + realm + "clients/" + clientID + "/roles/" + roleName)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteClientScope creates a new client scope
func (client *gocloak) DeleteClientScope(token *JWT, realm, scopeID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Put(client.basePath + "/auth/admin/realms/" + realm + "/client-scopes/" + scopeID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// GetKeyStoreConfig get keystoreconfig of the realm
func (client *gocloak) GetKeyStoreConfig(token *JWT, realm string) (*KeyStoreConfig, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/keys")
	if err != nil {
		return nil, err
	}

	var result KeyStoreConfig
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUser get all users inr ealm
func (client *gocloak) GetUser(token *JWT, realm, userID string) (*User, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/user/" + userID)
	if err != nil {
		return nil, err
	}

	var result User
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetComponents get all cimponents in realm
func (client *gocloak) GetComponents(token *JWT, realm string) (*[]Component, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/components")
	if err != nil {
		return nil, err
	}

	var result []Component
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsers get all users in realm
func (client *gocloak) GetUsers(token *JWT, realm string) (*[]User, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/users")
	if err != nil {
		return nil, err
	}

	var result []User
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserCount gets the user count in the realm
func (client *gocloak) GetUserCount(token *JWT, realm string) (int, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/users/count")
	if err != nil {
		return -1, err
	}

	var result int
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return -1, err
	}

	return result, nil
}

// GetUsergroups get all groups for user
func (client *gocloak) GetUserGroups(token *JWT, realm string, userID string) (*[]UserGroup, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/users/" + userID + "/groups")
	if err != nil {
		return nil, err
	}

	var result []UserGroup
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoleMappingByGroupID gets the role mappings by group
func (client *gocloak) GetRoleMappingByGroupID(token *JWT, realm string, groupID string) (*[]RoleMapping, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/groups/" + groupID + "/role-mappings")
	if err != nil {
		return nil, err
	}

	var result []RoleMapping

	var f map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &f); err != nil {
		return nil, err
	}

	itemsMap := f["clientMappings"].(map[string]interface{})

	for _, v := range itemsMap {
		switch jsonObj := v.(type) {
		case interface{}:
			jsonClientMapping, _ := json.Marshal(jsonObj)
			var client RoleMapping
			if err := json.Unmarshal(jsonClientMapping, &client); err != nil {
				return nil, err
			}
			result = append(result, client)
		default:
			return nil, errors.New("Expecting a JSON object; got something else")
		}
	}

	return &result, nil
}

// GetGroup get group with id in realm
func (client *gocloak) GetGroup(token *JWT, realm, groupID string) (*Group, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/group/" + groupID)
	if err != nil {
		return nil, err
	}

	var result Group
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetGroups get all groups in realm
func (client *gocloak) GetGroups(token *JWT, realm string) (*[]Group, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/groups")
	if err != nil {
		return nil, err
	}

	var result []Group
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoles get all roles in realm
func (client *gocloak) GetRoles(token *JWT, realm string) (*[]Role, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/roles")
	if err != nil {
		return nil, err
	}

	var result []Role
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRolesByClientID get all roles for the given client in realm
func (client *gocloak) GetRolesByClientID(token *JWT, realm string, clientID string) (*[]Role, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/clients/" + clientID + "/roles")
	if err != nil {
		return nil, err
	}

	var result []Role
	ioutil.WriteFile("test.json", resp.Body(), 0644)
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetClients gets all clients in realm
func (client *gocloak) GetClients(token *JWT, realm string) (*[]Client, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/clients")
	if err != nil {
		return nil, err
	}

	var result []Client
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func getRequestWithHeader(token *JWT) *resty.Request {
	return resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken)
}

func getBasicAuthForClient(clientID string, clientSecret string) string {
	var httpBasicAuth string
	if len(clientID) > 0 && len(clientSecret) > 0 {
		httpBasicAuth = base64.URLEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	}

	return "Basic " + httpBasicAuth
}
