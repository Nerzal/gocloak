package gocloak

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/pkg/jwx"
	jwt "github.com/dgrijalva/jwt-go"
	resty "gopkg.in/resty.v1"
)

type gocloak struct {
	basePath   string
	certsCache map[string]*CertResponse
	Config     struct {
		CertsInvalidateTime time.Duration
	}
}

type loginData struct {
	ClientID  string `json:"client_id"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}

const adminClientID string = "admin-cli"
const authRealm string = "/auth/admin/realms/"
const authRealms string = "/auth/realms/"
const tokenEndpoint string = "/protocol/openid-connect/token"
const openIDConnect string = "/protocol/openid-connect"

// NewClient creates a new Client
func NewClient(basePath string) GoCloak {
	c := gocloak{
		basePath:   basePath,
		certsCache: make(map[string]*CertResponse),
	}
	c.Config.CertsInvalidateTime = 10 * time.Minute

	return &c
}

func (client *gocloak) getNewCerts(realm string) (*CertResponse, error) {
	var result CertResponse
	resp, err := resty.R().
		SetResult(&result).
		Get(client.basePath + authRealms + realm + openIDConnect + "/certs")
	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCerts fetches certificates for the given realm from the public /open-id-connect/certs endpoint
func (client *gocloak) GetCerts(realm string) (*CertResponse, error) {
	if cert, ok := client.certsCache[realm]; ok {
		return cert, nil
	}
	cert, err := client.getNewCerts(realm)
	if err != nil {
		return nil, err
	}
	client.certsCache[realm] = cert
	timer := time.NewTimer(client.Config.CertsInvalidateTime)
	go func() {
		<-timer.C
		client.certsCache[realm] = cert
	}()
	return cert, nil
}

// GetIssuer gets the isser of the given realm
func (client *gocloak) GetIssuer(realm string) (*IssuerResponse, error) {
	var result IssuerResponse
	resp, err := resty.R().
		SetResult(&result).
		Get(client.basePath + authRealms + realm)
	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// RetrospectToken calls the openid-connect introspect endpoint
func (client *gocloak) RetrospectToken(accessToken string, clientID, clientSecret string, realm string) (*RetrospecTokenResult, error) {
	var result RetrospecTokenResult
	resp, err := getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"token_type_hint": "requesting_party_token",
			"token":           accessToken,
		}).
		SetResult(&result).
		Post(client.basePath + authRealms + realm + tokenEndpoint + "/introspect")
	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeAccessToken decodes the accessToken
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
	if usedKey == nil {
		return nil, nil, errors.New("Cannot find a key to decode the token")
	}

	return jwx.DecodeAccessToken(accessToken, usedKey.E, usedKey.N)
}

// DecodeAccesTokenCustomClaims decodes the accessToken and writes claims into the given claims
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
		if key.Kid == usedKeyID {
			return &key
		}
	}

	return nil
}

// RefreshToken refrehes the given token
func (client *gocloak) RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error) {
	refreshPath := authRealms + realm + tokenEndpoint

	var result JWT
	resp, err := resty.R().
		SetFormData(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
		}).
		SetResult(&result).
		Post(client.basePath + refreshPath)
	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// LoginAdmin performs a login
func (client *gocloak) LoginAdmin(username, password, realm string) (*JWT, error) {
	loginPath := authRealms + realm + tokenEndpoint

	var result JWT
	resp, err := resty.R().
		SetFormData(map[string]string{
			"client_id":  adminClientID,
			"grant_type": "password",
			"username":   username,
			"password":   password,
		}).
		SetResult(&result).
		Post(client.basePath + loginPath)

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// Login performs a login
func (client *gocloak) LoginClient(clientID, clientSecret, realm string) (*JWT, error) {
	loginPath := authRealms + realm + tokenEndpoint

	var result JWT
	resp, err := getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type":    "client_credentials",
		}).
		SetResult(&result).
		Post(client.basePath + loginPath)
	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// Login like login, but with basic auth
func (client *gocloak) Login(clientID string, clientSecret string, realm string, username string, password string) (*JWT, error) {
	var result JWT
	resp, err := getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"grant_type": "password",
			"username":   username,
			"password":   password,
		}).
		SetResult(&result).
		Post(client.basePath + authRealms + realm + tokenEndpoint)
	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// RequestPermission l
func (client *gocloak) RequestPermission(clientID string, clientSecret string, realm string, username string, password string, permission string) (*JWT, error) {
	var result JWT
	resp, err := getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"grant_type": "password",
			"username":   username,
			"password":   password,
			"permission": permission,
		}).
		SetResult(&result).
		Post(client.basePath + authRealms + realm + tokenEndpoint)
	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// SetPassword sets a new password
func (client *gocloak) SetPassword(token string, userID string, realm string, password string, temporary bool) error {
	requestBody := SetPasswordRequest{Password: password, Temporary: temporary, Type: "password"}
	resp, err := getRequestWithBearerAuth(token).
		SetBody(requestBody).
		Put(client.basePath + authRealm + realm + "/users/" + userID + "/reset-password")
	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// CreateUser tries to create the given user in the given realm and returns it's userID
func (client *gocloak) CreateUser(token string, realm string, user User) (*string, error) {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(user).
		Post(client.basePath + authRealm + realm + "/users")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	userPath := resp.Header().Get("Location")
	splittedPath := strings.Split(userPath, "/")
	userID := splittedPath[len(splittedPath)-1]

	return &userID, nil
}

// CreateUser creates a new user
func (client *gocloak) CreateGroup(token string, realm string, group Group) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(group).
		Post(client.basePath + authRealm + realm + "/groups")

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// CreateComponent creates a new user
func (client *gocloak) CreateComponent(token string, realm string, component Component) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(component).
		Post(client.basePath + authRealm + realm + "/components")

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateClient(token string, realm string, newClient Client) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(newClient).
		Post(client.basePath + authRealm + realm + "/clients")

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// CreateUser creates a new user
func (client *gocloak) CreateRole(token string, realm string, clientID string, role Role) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(role).
		Post(client.basePath + authRealm + realm + "/clients/" + clientID + "/roles")

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// CreateClientScope creates a new client scope
func (client *gocloak) CreateClientScope(token string, realm string, scope ClientScope) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(scope).
		Post(client.basePath + authRealm + realm + "/client-scopes")

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateUser(token string, realm string, user User) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(user).
		Put(client.basePath + authRealm + realm + "/users/" + user.ID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateGroup(token string, realm string, group Group) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(group).
		Put(client.basePath + authRealm + realm + "/groups/" + group.ID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateClient(token string, realm string, newClient Client) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(newClient).
		Put(client.basePath + authRealm + realm + "/clients")

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser creates a new user
func (client *gocloak) UpdateRole(token string, realm string, clientID string, role Role) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(role).
		Put(client.basePath + authRealm + realm + "/clients/" + clientID + "/roles/" + role.Name)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// UpdateClientScope creates a new client scope
func (client *gocloak) UpdateClientScope(token string, realm string, scope ClientScope) error {
	resp, err := getRequestWithBearerAuth(token).
		SetBody(scope).
		Put(client.basePath + authRealm + realm + "/client-scopes/" + scope.ID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteUser(token string, realm string, userID string) error {
	resp, err := getRequestWithBearerAuth(token).
		Delete(client.basePath + authRealm + realm + "/users/" + userID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteGroup(token string, realm string, groupID string) error {
	resp, err := getRequestWithBearerAuth(token).
		Delete(client.basePath + authRealm + realm + "/groups/" + groupID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteClient(token string, realm string, clientID string) error {
	resp, err := getRequestWithBearerAuth(token).
		Delete(client.basePath + authRealm + realm + "/clients/" + clientID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// DeleteComponent creates a new user
func (client *gocloak) DeleteComponent(token string, realm string, componentID string) error {
	resp, err := getRequestWithBearerAuth(token).
		Delete(client.basePath + authRealm + realm + "/components/" + componentID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser creates a new user
func (client *gocloak) DeleteRole(token string, realm string, clientID, roleName string) error {
	resp, err := getRequestWithBearerAuth(token).
		Delete(client.basePath + authRealm + realm + "/clients/" + clientID + "/roles/" + roleName)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// DeleteClientScope creates a new client scope
func (client *gocloak) DeleteClientScope(token string, realm string, scopeID string) error {
	resp, err := getRequestWithBearerAuth(token).
		Put(client.basePath + authRealm + realm + "/client-scopes/" + scopeID)

	err = checkForError(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// GetKeyStoreConfig get keystoreconfig of the realm
func (client *gocloak) GetKeyStoreConfig(token string, realm string) (*KeyStoreConfig, error) {
	var result KeyStoreConfig
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/keys")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserByID fetches a user from the given realm witht he given userID
func (client *gocloak) GetUserByID(accessToken string, realm string, userID string) (*User, error) {
	if userID == "" {
		return nil, errors.New("UserID shall not be empty")
	}

	var result User
	resp, err := getRequestWithBearerAuth(accessToken).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users/" + userID)

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetComponents get all cimponents in realm
func (client *gocloak) GetComponents(token string, realm string) (*[]Component, error) {
	var result []Component
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/components")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsers get all users in realm
func (client *gocloak) GetUsers(token string, realm string) (*[]User, error) {
	var result []User
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserCount gets the user count in the realm
func (client *gocloak) GetUserCount(token string, realm string) (int, error) {
	var result int
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users/count")

	err = checkForError(resp, err)
	if err != nil {
		return -1, err
	}

	return result, nil
}

// GetUsergroups get all groups for user
func (client *gocloak) GetUserGroups(token string, realm string, userID string) (*[]UserGroup, error) {
	var result []UserGroup
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users/" + userID + "/groups")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoleMappingByGroupID gets the role mappings by group
func (client *gocloak) GetRoleMappingByGroupID(token string, realm string, groupID string) (*[]RoleMapping, error) {
	resp, err := getRequestWithBearerAuth(token).
		Get(client.basePath + authRealm + realm + "/groups/" + groupID + "/role-mappings")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	var result []RoleMapping

	var f map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &f); err != nil {
		return nil, err
	}

	if _, ok := f["clientMappings"]; ok {
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
	}

	return &result, nil
}

// GetRoleMappingByUserID gets the role mappings by user
func (client *gocloak) GetRoleMappingByUserID(token string, realm string, userID string) (*[]RoleMapping, error) {
	resp, err := getRequestWithBearerAuth(token).
		Get(client.basePath + authRealm + realm + "/users/" + userID + "/role-mappings")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	var result []RoleMapping

	var f map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &f); err != nil {
		return nil, err
	}

	if _, ok := f["clientMappings"]; ok {
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
	}

	return &result, nil
}

// GetGroup get group with id in realm
func (client *gocloak) GetGroup(token string, realm string, groupID string) (*Group, error) {
	var result Group
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/group/" + groupID)

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetGroups get all groups in realm
func (client *gocloak) GetGroups(token string, realm string) (*[]Group, error) {
	var result []Group
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/groups")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoles get all roles in realm
func (client *gocloak) GetRoles(token string, realm string) (*[]Role, error) {
	var result []Role
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/roles")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRolesByClientID get all roles for the given client in realm
func (client *gocloak) GetRolesByClientID(token string, realm string, clientID string) (*[]Role, error) {
	var result []Role
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/clients/" + clientID + "/roles")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetClients gets all clients in realm
func (client *gocloak) GetClients(token string, realm string) (*[]Client, error) {
	var result []Client
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/clients")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UserAttributeContains checks if the given attribute value is set
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

func getRequestWithBearerAuth(token string) *resty.Request {
	return resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+token)
}

func getRequestWithBasicAuth(clientID string, clientSecret string) *resty.Request {
	var httpBasicAuth string
	if len(clientID) > 0 && len(clientSecret) > 0 {
		httpBasicAuth = base64.URLEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	}
	return resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", "Basic "+httpBasicAuth)
}

// GetRealmRolesByUserID gets the roles by user
func (client *gocloak) GetRealmRolesByUserID(token string, realm string, userID string) (*[]Role, error) {
	var result []Role
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/users/" + userID + "/role-mappings/realm")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRealmRolesByGroupID gets the roles by group
func (client *gocloak) GetRealmRolesByGroupID(token string, realm string, groupID string) (*[]Role, error) {
	var result []Role
	resp, err := getRequestWithBearerAuth(token).
		Get(client.basePath + authRealm + realm + "/groups/" + groupID + "/role-mappings/realm")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsersByRoleName returns Users by a Role Name
func (client *gocloak) GetUsersByRoleName(token string, realm string, roleName string) (*[]User, error) {
	var result []User
	resp, err := getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.basePath + authRealm + realm + "/roles/" + roleName + "/users")

	err = checkForError(resp, err)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func checkForError(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}

	if resp.IsError() {
		if resp.StatusCode() == 409 {
			return &ObjectAllreadyExists{}
		}
		log.Printf("Error: Request returned a response with status '%s' and body: %s", resp.Status(), string(resp.Body()))
		return errors.New(resp.Status())
	}
	return nil
}
