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

	"github.com/Nerzal/gocloak/pkg/jwx"
	jwt "github.com/dgrijalva/jwt-go"
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

func (client *gocloak) GetCerts(realm string) (*CertResponse, error) {
	var result CertResponse
	resp, err := resty.R().
		SetResult(&result).
		Get(client.basePath + "/auth/realms/" + realm + "/protocol/openid-connect/certs")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}

	return &result, nil
}

func (client *gocloak) GetIssuer(realm string) (*IssuerResponse, error) {
	var result IssuerResponse
	resp, err := resty.R().
		SetResult(&result).
		Get(client.basePath + "/auth/realms/" + realm)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}

	return &result, nil
}

func (client *gocloak) RetrospectToken(accessToken string, clientID, clientSecret string, realm string) (*RetrospecTokenResult, error) {
	var result RetrospecTokenResult
	resp, err := resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", getBasicAuthForClient(clientID, clientSecret)).
		SetFormData(map[string]string{
			"token_type_hint": "requesting_party_token",
			"token":           accessToken,
		}).
		SetResult(&result).
		Post(client.basePath + "/auth/realms/" + realm + "/protocol/openid-connect/token/introspect")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}

	return &result, nil
}

func (client *gocloak) DecodeAccessToken(accessToken string, realm string) (*jwt.Token, *jwt.MapClaims, error) {
	decodedHeader, err := jwx.DecodeAccessTokenHeader(accessToken)
	if err != nil {
		return nil, nil, err
	}

	certResult, err := client.GetCerts(realm)
	if err != nil {
		return nil, nil, err
	}

	usedKey := findUsedKey(decodedHeader.Kid, certResult.Keys)
	return jwx.DecodeAccessToken(accessToken, usedKey.E, usedKey.N)
}

func (client *gocloak) DecodeAccessTokenCustomClaims(accessToken string, realm string, claims jwt.Claims) (*jwt.Token, error) {
	decodedHeader, err := jwx.DecodeAccessTokenHeader(accessToken)
	if err != nil {
		return nil, err
	}

	certResult, err := client.GetCerts(realm)
	if err != nil {
		return nil, err
	}

	usedKey := findUsedKey(decodedHeader.Kid, certResult.Keys)
	token, err := jwx.DecodeAccessTokenCustomClaims(accessToken, usedKey.E, usedKey.N, claims)
	return token, err
}

func findUsedKey(usedKeyID string, keys []CertResponseKey) *CertResponseKey {
	for _, key := range keys {
		if key.Kid != usedKeyID {
			continue
		}

		return &key
	}
	return nil
}

func (client *gocloak) RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error) {
	firstPart := "/auth/realms/"
	lastPart := "/protocol/openid-connect/token"
	loginPath := firstPart + realm + lastPart

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
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
		return nil, errors.New(res.Status)
	}

	jwt := &JWT{}
	err = json.Unmarshal(body, jwt)
	if err != nil {
		return nil, err
	}

	if jwt.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}
	return jwt, err
}

// LoginAdmin performs a login
func (client *gocloak) LoginAdmin(username, password, realm string) (*JWT, error) {
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

	jwt := &JWT{}
	err = json.Unmarshal(body, jwt)
	if err != nil {
		return nil, err
	}

	if jwt.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}
	return jwt, nil
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
	if err != nil {
		return nil, err
	}

	if jwt.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return jwt, nil
}

// Login like login, but with basic auth
func (client *gocloak) Login(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error) {
	var result JWT
	resp, err := resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", getBasicAuthForClient(clientID, clientSecret)).
		SetFormData(map[string]string{
			"grant_type": "password",
			"username":   username,
			"password":   password,
		}).
		SetResult(&result).
		Post(client.basePath + "/auth/realms/" + realm + "/protocol/openid-connect/token")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		log.Println(string(resp.Body()))
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// RequestPermission l
func (client *gocloak) RequestPermission(clientID string, clientSecret string, realm string, username string, password string, permission string) (*JWT, error) {
	var result JWT
	resp, err := resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", getBasicAuthForClient(clientID, clientSecret)).
		SetFormData(map[string]string{
			"grant_type": "password",
			"username":   username,
			"password":   password,
			"permission": permission,
		}).
		SetResult(&result).
		Post(client.basePath + "/auth/realms/" + realm + "/protocol/openid-connect/token")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		log.Println(string(resp.Body()))
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// SetPassword sets a new password
func (client *gocloak) SetPassword(token string, userID string, realm string, password string, temporary bool) error {
	requestBody := SetPasswordRequest{Password: password, Temporary: temporary, Type: "password"}
	bytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Put(client.basePath + authRealm + realm + "/users/" + userID + "/reset-password")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 201 && resp.StatusCode() != 204 && resp.StatusCode() != 409 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser tries to create the given user in the given realm and returns it's userID
func (client *gocloak) CreateUser(token string, realm string, user User) (*string, error) {
	bytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	resp, err := getRequestWithHeader(token).
		SetHeader("Content-Type", "application/json").
		SetBody(string(bytes)).
		Post(client.basePath + authRealm + realm + "/users")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 201 && resp.StatusCode() != 204 && resp.StatusCode() != 409 {
		return nil, errors.New(resp.Status())
	}

	userPath := resp.Header().Get("Location")
	splittedPath := strings.Split(userPath, "/")
	userID := splittedPath[len(splittedPath)-1]

	return &userID, nil
}

// CreateUser creates a new user
func (client *gocloak) CreateGroup(token string, realm string, group Group) error {
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

	if resp.StatusCode() != 201 && resp.StatusCode() != 409 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateComponent creates a new user
func (client *gocloak) CreateComponent(token string, realm string, component Component) error {
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

	if resp.StatusCode() != 201 && resp.StatusCode() != 409 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateClient(token string, realm string, newClient Client) error {
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

	if resp.StatusCode() != 201 && resp.StatusCode() != 409 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateRole(token string, realm string, clientID string, role Role) error {
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

	if resp.StatusCode() != 201 && resp.StatusCode() != 409 {
		return errors.New(resp.Status())
	}

	return nil
}

// CreateClientScope creates a new client scope
func (client *gocloak) CreateClientScope(token string, realm string, scope ClientScope) error {
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

	if resp.StatusCode() != 201 && resp.StatusCode() != 409 {
		return errors.New(resp.Status())
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateUser(token string, realm string, user User) error {
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
func (client *gocloak) UpdateGroup(token string, realm string, group Group) error {
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
func (client *gocloak) UpdateClient(token string, realm string, newClient Client) error {
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
func (client *gocloak) UpdateRole(token string, realm string, clientID string, role Role) error {
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
func (client *gocloak) UpdateClientScope(token string, realm string, scope ClientScope) error {
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
func (client *gocloak) DeleteUser(token string, realm string, userID string) error {
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
func (client *gocloak) DeleteGroup(token string, realm string, groupID string) error {
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
func (client *gocloak) DeleteClient(token string, realm string, clientID string) error {
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
func (client *gocloak) DeleteComponent(token string, realm string, componentID string) error {
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
func (client *gocloak) DeleteRole(token string, realm string, clientID, roleName string) error {
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
func (client *gocloak) DeleteClientScope(token string, realm string, scopeID string) error {
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
func (client *gocloak) GetKeyStoreConfig(token string, realm string) (*KeyStoreConfig, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/keys")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}

	var result KeyStoreConfig
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *gocloak) GetUserByID(accessToken string, realm string, userID string) (*User, error) {
	if userID == "" {
		return nil, errors.New("UserID shall not be empty")
	}

	var result User
	_, err := getRequestWithHeader(accessToken).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users/" + userID)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetComponents get all cimponents in realm
func (client *gocloak) GetComponents(token string, realm string) (*[]Component, error) {
	var result []Component
	_, err := getRequestWithHeader(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/components")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsers get all users in realm
func (client *gocloak) GetUsers(token string, realm string) (*[]User, error) {
	var result []User
	_, err := getRequestWithHeader(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserCount gets the user count in the realm
func (client *gocloak) GetUserCount(token string, realm string) (int, error) {
	var result int
	_, err := getRequestWithHeader(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users/count")
	if err != nil {
		return -1, err
	}

	return result, nil
}

// GetUsergroups get all groups for user
func (client *gocloak) GetUserGroups(token string, realm string, userID string) (*[]UserGroup, error) {
	var result []UserGroup
	_, err := getRequestWithHeader(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users/" + userID + "/groups")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoleMappingByGroupID gets the role mappings by group
func (client *gocloak) GetRoleMappingByGroupID(token string, realm string, groupID string) (*[]RoleMapping, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/groups/" + groupID + "/role-mappings")
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
func (client *gocloak) GetGroup(token string, realm string, groupID string) (*Group, error) {
	var result Group
	_, err := getRequestWithHeader(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/group/" + groupID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetGroups get all groups in realm
func (client *gocloak) GetGroups(token string, realm string) (*[]Group, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/groups")
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
func (client *gocloak) GetRoles(token string, realm string) (*[]Role, error) {
	resp, err := getRequestWithHeader(token).
		Get(client.basePath + authRealm + realm + "/roles")
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
func (client *gocloak) GetRolesByClientID(token string, realm string, clientID string) (*[]Role, error) {
	var result []Role
	_, err := getRequestWithHeader(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/clients/" + clientID + "/roles")
	if err != nil {
		return nil, err
	}
	// ioutil.WriteFile("test.json", resp.Body(), 0644)

	return &result, nil
}

// GetClients gets all clients in realm
func (client *gocloak) GetClients(token string, realm string) (*[]Client, error) {
	var result []Client
	_, err := getRequestWithHeader(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/clients")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *gocloak) UserAttributeContains(attributes map[string][]string, attribute string, value string) bool {
	if val, ok := attributes[attribute]; ok {
		for _, item := range val {
			if item == value {
				return true
			}
		}
	}
	return false
}

func getRequestWithHeader(token string) *resty.Request {
	return resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token)
}

func getBasicAuthForClient(clientID string, clientSecret string) string {
	var httpBasicAuth string
	if len(clientID) > 0 && len(clientSecret) > 0 {
		httpBasicAuth = base64.URLEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	}

	return "Basic " + httpBasicAuth
}
