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

type Admin interface {
	Login(username string, password string, realm string) (*models.JWT, error)
	GetUser(name string, token *models.JWT) error
}

type admin struct {
	basePath string
}

type loginData struct {
	ClientID  string `json:"client_id"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}

const adminClientID string = "admin-cli"

// NewAdminClient creates a new Client
func NewAdminClient(basePath string) Admin {
	return &admin{
		basePath: basePath,
	}
}

func (client *admin) GetUser(name string, token *models.JWT) error {
	lastPart := "/users/"
	path := "/" + realm + lastPart

	req, _ := http.NewRequest("GET", client.basePath+path, nil)
	req.Header.Add("Authorization", "bearer "+token.RefreshToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Println(string(body))

	return nil
}

func (client *admin) Login(username, password, realm string) (*models.JWT, error) {
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

/**
 * Direct Grant Authentication
 * -
 * This method directly gets you the OIDC Token from keycloak to use in your next requests
 */
func (client *admin) DirectGrantAuthentication(clientID string, clientSecret string, realm string, username string, password string) (*models.JWT, error) {
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

	// Here’s the actual decoding, and a check for associated errors.
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	// Check for Result
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

/**
 * User List
 */
func (client *admin) GetUserListInRealm(token *models.JWT, realm string) (*[]models.User, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/users")
	if err != nil {
		return nil, err
	}

	// Decode into struct
	var result []models.User
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

/**
 * Get Groups of UserId
 */
func (client *admin) GetUserGroupsInRealm(token *models.JWT, realm string, userID string) (*[]models.UserGroup, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/users/" + userID + "/groups")
	if err != nil {
		return nil, err
	}

	// Decode into struct
	var result []models.UserGroup
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

/**
 * Get Group Role Mapping
 */
func (client *admin) GetRoleMappingByGroupID(token *models.JWT, realm string, groupID string) (*[]models.RoleMapping, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/groups/" + groupID + "/role-mappings")
	if err != nil {
		return nil, err
	}

	var result []models.RoleMapping

	// Decode into struct
	var f map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &f); err != nil {
		return nil, err
	}

	// JSON object parses into a map with string keys
	itemsMap := f["clientMappings"].(map[string]interface{})

	// Loop through the Items; we're not interested in the key, just the values
	for _, v := range itemsMap {
		// Use type assertions to ensure that the value's a JSON object
		switch jsonObj := v.(type) {
		// The value is an Item, represented as a generic interface
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

/**
 * Group List
 */
func (client *admin) GetGroupListByRealm(token *models.JWT, realm string) (*[]models.Group, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/groups")
	if err != nil {
		return nil, err
	}

	// Decode into struct
	var result []models.Group
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

/**
 * Get Roles by Realm
 */
func (client *admin) GetRolesByRealm(token *models.JWT, realm string) (*[]models.Role, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/roles")
	if err != nil {
		return nil, err
	}

	// Decode into struct
	var result []models.Role
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

/**
 * Get Roles by Client and Realm
 */
func (client *admin) GetRolesByClientId(token *models.JWT, realm string, clientID string) (*[]models.Role, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/clients/" + clientID + "/roles")
	if err != nil {
		return nil, err
	}

	// Decode into struct
	var result []models.Role
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

/**
 * Get Clients by Realm
 */
func (client *admin) GetClientsInRealm(token *models.JWT, realm string) (*[]models.RealmClient, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token.AccessToken).
		Get(client.basePath + "/auth/admin/realms/" + realm + "/clients")
	if err != nil {
		return nil, err
	}

	// Decode into struct
	var result []models.RealmClient
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

/**
 * Function to build the HttpBasicAuth Base64 String
 */
func getBasicAuthForClient(clientID string, clientSecret string) string {
	var httpBasicAuth string
	if len(clientID) > 0 && len(clientSecret) > 0 {
		httpBasicAuth = base64.URLEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	}

	return "Basic " + httpBasicAuth
}
