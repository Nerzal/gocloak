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

	"github.com/Nerzal/gocloak/models"
	resty "gopkg.in/resty.v1"
)

// Client holds all methods a client should fullfill
type Client interface {
	Login(username string, password string, realm string) (*models.JWT, error)

	DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*models.JWT, error)

	CreateUser(token *models.JWT, realm string, user models.User) error
	CreateGroup(token *models.JWT, realm string, group models.Group) error
	CreateRole(token *models.JWT, realm string, clientID string, role models.Role) error
	CreateClient(token *models.JWT, realm string, clientID models.Client) error

	GetUsers(token *models.JWT, realm string) (*[]models.User, error)
	GetUserGroups(token *models.JWT, realm string, userID string) (*[]models.UserGroup, error)
	GetRoleMappingByGroupID(token *models.JWT, realm string, groupID string) (*[]models.RoleMapping, error)
	GetGroups(token *models.JWT, realm string) (*[]models.Group, error)
	GetRoles(token *models.JWT, realm string) (*[]models.Role, error)
	GetRolesByClientID(token *models.JWT, realm string, clientID string) (*[]models.Role, error)
	GetClients(token *models.JWT, realm string) (*[]models.Client, error)
}

type client struct {
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
func NewClient(basePath string) Client {
	return &client{
		basePath: basePath,
	}
}

// Login performs a login
func (client *client) Login(username, password, realm string) (*models.JWT, error) {
	firstPart := "/auth/realms/"
	lastPart := "/protocol/openid-connect/token"
	loginPath := firstPart + realm + lastPart

	data := url.Values{}
	data.Set("client_id", adminClientID)
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

// DirectGrantAuthentication like login, but with basic auth
func (client *client) DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*models.JWT, error) {
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
func (client *client) CreateUser(token *models.JWT, realm string, user models.User) error {
	bytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/users")

	log.Println(string(resp.Body()))
	log.Println(resp.Status())

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *client) CreateGroup(token *models.JWT, realm string, group models.Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/groups")

	log.Println(string(resp.Body()))
	log.Println(resp.Status())

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *client) CreateClient(token *models.JWT, realm string, newClient models.Client) error {
	bytes, err := json.Marshal(newClient)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "/clients")

	log.Println(string(resp.Body()))
	log.Println(resp.Status())

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *client) CreateRole(token *models.JWT, realm string, clientID string, role models.Role) error {
	bytes, err := json.Marshal(role)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + "/auth/admin/realms/" + realm + "clients/" + clientID + "/roles")

	log.Println(string(resp.Body()))
	log.Println(resp.Status())

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 {
		return errors.New(resp.Status())
	}

	return nil
}

// GetUsers get all users inr ealm
func (client *client) GetUsers(token *models.JWT, realm string) (*[]models.User, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/users")
	if err != nil {
		return nil, err
	}

	var result []models.User
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsergroups get all groups for user
func (client *client) GetUserGroups(token *models.JWT, realm string, userID string) (*[]models.UserGroup, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/users/" + userID + "/groups")
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
func (client *client) GetRoleMappingByGroupID(token *models.JWT, realm string, groupID string) (*[]models.RoleMapping, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/groups/" + groupID + "/role-mappings")
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

// GetGroups get all groups in realm
func (client *client) GetGroups(token *models.JWT, realm string) (*[]models.Group, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/groups")
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
func (client *client) GetRoles(token *models.JWT, realm string) (*[]models.Role, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/roles")
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
func (client *client) GetRolesByClientID(token *models.JWT, realm string, clientID string) (*[]models.Role, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/clients/" + clientID + "/roles")
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
func (client *client) GetClients(token *models.JWT, realm string) (*[]models.Client, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/clients")
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
