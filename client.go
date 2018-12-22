package gocloak

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Nerzal/gocloak/pkg/models"
	resty "gopkg.in/resty.v1"
)

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
const authRealm string = "/auth/admin/realms/"

// NewClient creates a new Client
func NewClient(basePath string) GoCloak {
	return &gocloak{
		basePath: basePath,
	}
}

func (client *gocloak) ValidateToken(token string, realm string) error {
	firstPart := "/auth/realms/"
	lastPart := "/protocol/openid-connect/userinfo"
	validationPath := firstPart + realm + lastPart

	req, _ := http.NewRequest("POST", client.basePath+validationPath, nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		log.Println(string(body))
		return models.APIError{
			Code:    res.StatusCode,
			Message: "Invalid or malformed token",
		}
	}

	return nil
}

func (client *gocloak) RefreshToken(refreshToken string, clientID, realm string) (*models.JWT, error) {
	firstPart := "/auth/realms/"
	lastPart := "/protocol/openid-connect/token"
	loginPath := firstPart + realm + lastPart

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", refreshToken)

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

	jwt := &models.JWT{}
	err = json.Unmarshal(body, jwt)
	return jwt, err
}

// Login performs a login
func (client *gocloak) LoginAdmin(username, password, realm string) (*models.JWT, error) {
	return client.Login(username, password, realm, adminClientID)
}

// Login performs a login
func (client *gocloak) Login(username, password, realm, clientID string) (*models.JWT, error) {
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

	jwt := &models.JWT{}
	err = json.Unmarshal(body, jwt)
	return jwt, err
}

// Login performs a login
func (client *gocloak) LoginClient(clientID, clientSecret, realm string) (*models.JWT, error) {
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

	jwt := &models.JWT{}
	err = json.Unmarshal(body, jwt)
	return jwt, err
}

// DirectGrantAuthentication like login, but with basic auth
func (client *gocloak) DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*models.JWT, error) {
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
		return &models.JWT{
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
func (client *gocloak) CreateUser(token *models.JWT, realm string, user models.User) error {
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + authRealm + realm + "/users")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateGroup(token *models.JWT, realm string, group models.Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + authRealm + realm + "/groups")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateComponent creates a new user
func (client *gocloak) CreateComponent(token *models.JWT, realm string, component models.Component) error {
	bytes, err := json.Marshal(component)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + authRealm + realm + "/components")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateClient(token *models.JWT, realm string, newClient models.Client) error {
	bytes, err := json.Marshal(newClient)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + authRealm + realm + "/clients")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateRole(token *models.JWT, realm string, clientID string, role models.Role) error {
	bytes, err := json.Marshal(role)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + authRealm + realm + "clients/" + clientID + "/roles")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateClientScope creates a new client scope
func (client *gocloak) CreateClientScope(token *models.JWT, realm string, scope models.ClientScope) error {
	bytes, err := json.Marshal(scope)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + authRealm + realm + "/client-scopes")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateUser(token *models.JWT, realm string, user models.User) error {
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + authRealm + realm + "/users/" + user.ID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateGroup(token *models.JWT, realm string, group models.Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + authRealm + realm + "/groups/" + group.ID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateClient(token *models.JWT, realm string, newClient models.Client) error {
	bytes, err := json.Marshal(newClient)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + authRealm + realm + "/clients")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateRole(token *models.JWT, realm string, clientID string, role models.Role) error {
	bytes, err := json.Marshal(role)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + authRealm + realm + "clients/" + clientID + "/roles/" + role.Name)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateClientScope creates a new client scope
func (client *gocloak) UpdateClientScope(token *models.JWT, realm string, scope models.ClientScope) error {
	bytes, err := json.Marshal(scope)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + authRealm + realm + "/client-scopes/" + scope.ID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteUser(token *models.JWT, realm, userID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + authRealm + realm + "/users/" + userID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteGroup(token *models.JWT, realm, groupID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + authRealm + realm + "/groups/" + groupID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteClient(token *models.JWT, realm, clientID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + authRealm + realm + "/clients/" + clientID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteComponent creates a new user
func (client *gocloak) DeleteComponent(token *models.JWT, realm, componentID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + authRealm + realm + "/components/" + componentID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteRole(token *models.JWT, realm, clientID, roleName string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Delete(client.basePath + authRealm + realm + "clients/" + clientID + "/roles/" + roleName)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// DeleteClientScope creates a new client scope
func (client *gocloak) DeleteClientScope(token *models.JWT, realm, scopeID string) error {
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		Put(client.basePath + authRealm + realm + "/client-scopes/" + scopeID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// GetKeyStoreConfig get keystoreconfig of the realm
func (client *gocloak) GetKeyStoreConfig(token *models.JWT, realm string) (*models.KeyStoreConfig, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/keys")
	if err != nil {
		return nil, err
	}

	var result models.KeyStoreConfig
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUser get all users inr ealm
func (client *gocloak) GetUser(token *models.JWT, realm, userID string) (*models.User, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/user/" + userID)
	if err != nil {
		return nil, err
	}

	var result models.User
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetComponents get all cimponents in realm
func (client *gocloak) GetComponents(token *models.JWT, realm string) (*[]models.Component, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/components")
	if err != nil {
		return nil, err
	}

	var result []models.Component
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsers get all users in realm
func (client *gocloak) GetUsers(token *models.JWT, realm string) (*[]models.User, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/users")
	if err != nil {
		return nil, err
	}

	var result []models.User
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserCount gets the user count in the realm
func (client *gocloak) GetUserCount(token *models.JWT, realm string) (int, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/users/count")
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
func (client *gocloak) GetUserGroups(token *models.JWT, realm string, userID string) (*[]models.UserGroup, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/users/" + userID + "/groups")
	if err != nil {
		return nil, err
	}

	var result []models.UserGroup
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoleMappingByGroupID gets the role mappings by group
func (client *gocloak) GetRoleMappingByGroupID(token *models.JWT, realm string, groupID string) (*[]models.RoleMapping, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/groups/" + groupID + "/role-mappings")
	if err != nil {
		return nil, err
	}

	var result []models.RoleMapping

	var f map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &f); err != nil {
		return nil, err
	}

	itemsMap := f["clientMappings"].(map[string]interface{})

	for _, v := range itemsMap {
		switch jsonObj := v.(type) {
		case interface{}:
			jsonClientMapping, _ := json.Marshal(jsonObj)
			var client models.RoleMapping
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
func (client *gocloak) GetGroup(token *models.JWT, realm, groupID string) (*models.Group, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/group/" + groupID)
	if err != nil {
		return nil, err
	}

	var result models.Group
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetGroups get all groups in realm
func (client *gocloak) GetGroups(token *models.JWT, realm string) (*[]models.Group, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/groups")
	if err != nil {
		return nil, err
	}

	var result []models.Group
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoles get all roles in realm
func (client *gocloak) GetRoles(token *models.JWT, realm string) (*[]models.Role, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/roles")
	if err != nil {
		return nil, err
	}

	var result []models.Role
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRolesByClientID get all roles for the given client in realm
func (client *gocloak) GetRolesByClientID(token *models.JWT, realm string, clientID string) (*[]models.Role, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/clients/" + clientID + "/roles")
	if err != nil {
		return nil, err
	}

	var result []models.Role
	ioutil.WriteFile("test.json", resp.Body(), 0644)
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetClients gets all clients in realm
func (client *gocloak) GetClients(token *models.JWT, realm string) (*[]models.Client, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/clients")
	if err != nil {
		return nil, err
	}

	var result []models.Client
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func getRequestWithHeader(token *models.JWT) *resty.Request {
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
