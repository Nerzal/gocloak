package gocloak

import (
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/pkg/jwx"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/resty.v1"
)

type gocloak struct {
	basePath    string
	certsCache  map[string]*CertResponse
	restyClient *resty.Client
	Config      struct {
		CertsInvalidateTime time.Duration
	}
}

type loginData struct {
	ClientID  string `json:"client_id"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}

const (
	adminClientID string = "admin-cli"
	urlSeparator  string = "/"
)

var authAdminRealms = makeURL("auth", "admin", "realms")
var authRealms = makeURL("auth", "realms")
var tokenEndpoint = makeURL("protocol", "openid-connect", "token")
var logoutEndpoint = makeURL("protocol", "openid-connect", "logout")
var openIDConnect = makeURL("protocol", "openid-connect")

func makeURL(path ...string) string {
	return strings.Join(path, urlSeparator)
}

func (client *gocloak) getRequestWithBearerAuth(token string) *resty.Request {
	return client.restyClient.R().
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json")
}

func (client *gocloak) getRequestWithBasicAuth(clientID string, clientSecret string) *resty.Request {
	var httpBasicAuth string
	if len(clientID) > 0 && len(clientSecret) > 0 {
		httpBasicAuth = base64.URLEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	}
	return client.restyClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", "Basic "+httpBasicAuth)
}

func checkForError(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}

	if resp.IsError() {
		if resp.StatusCode() == 409 {
			return &ObjectAlreadyExists{}
		}
		log.Printf("Error: Request returned a response with status '%s' and body: %s", resp.Status(), string(resp.Body()))
		return errors.New(resp.Status())
	}
	return nil
}

func findUsedKey(usedKeyID string, keys []CertResponseKey) *CertResponseKey {
	for _, key := range keys {
		if key.Kid == usedKeyID {
			return &key
		}
	}

	return nil
}

// ===============
// Keycloak client
// ===============

// NewClient creates a new Client
func NewClient(basePath string) GoCloak {

	c := gocloak{
		basePath:    strings.TrimRight(basePath, urlSeparator),
		certsCache:  make(map[string]*CertResponse),
		restyClient: resty.New(),
	}
	c.Config.CertsInvalidateTime = 10 * time.Minute

	return &c
}

func (client *gocloak) RestyClient() *resty.Client {
	return client.restyClient
}

func (client *gocloak) getRealmURL(realm string, path ...string) string {
	path = append([]string{client.basePath, authRealms, realm}, path...)
	return makeURL(path...)
}

func (client *gocloak) getAdminRealmURL(realm string, path ...string) string {
	path = append([]string{client.basePath, authAdminRealms, realm}, path...)
	return makeURL(path...)
}

// GetUserInfo calls the UserInfo endpoint
func (client *gocloak) GetUserInfo(accessToken string, realm string) (*UserInfo, error) {
	var result UserInfo
	resp, err := client.getRequestWithBearerAuth(accessToken).
		SetResult(&result).
		Get(client.getRealmURL(realm, openIDConnect, "userinfo"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *gocloak) GetUserInfoWithClaims(accessToken string, realm string, claims []string, output interface{}) error {
	resp, err := client.getRequestWithBearerAuth(accessToken).
		SetFormData(map[string]string{
			"claims": strings.Join(claims, ","),
		}).
		SetResult(output).
		Post(client.getRealmURL(realm, openIDConnect, "userinfo"))

	if err := checkForError(resp, err); err != nil {
		return err
	}

	return nil
}

func (client *gocloak) getNewCerts(realm string) (*CertResponse, error) {
	var result CertResponse
	resp, err := client.restyClient.R().
		SetResult(&result).
		Get(client.getRealmURL(realm, openIDConnect, "certs"))

	if err := checkForError(resp, err); err != nil {
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
		delete(client.certsCache, realm)
	}()
	return cert, nil
}

// GetIssuer gets the isser of the given realm
func (client *gocloak) GetIssuer(realm string) (*IssuerResponse, error) {
	var result IssuerResponse
	resp, err := client.restyClient.R().
		SetResult(&result).
		Get(client.getRealmURL(realm))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// RetrospectToken calls the openid-connect introspect endpoint
func (client *gocloak) RetrospectToken(accessToken string, clientID, clientSecret string, realm string) (*RetrospecTokenResult, error) {
	var result RetrospecTokenResult
	resp, err := client.getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"token_type_hint": "requesting_party_token",
			"token":           accessToken,
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint, "introspect"))

	if err := checkForError(resp, err); err != nil {
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

// RefreshToken refrehes the given token
func (client *gocloak) RefreshToken(refreshToken string, clientID, clientSecret, realm string) (*JWT, error) {
	var result JWT
	resp, err := client.restyClient.R().
		SetFormData(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

func (client *gocloak) ExchangeToken(authToken string, clientID, clientSecret, realm string) (*JWT, error) {
	var result JWT
	resp, err := client.restyClient.R().
		SetFormData(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type":    "urn:ietf:params:oauth:grant-type:token-exchange",
			"subject_token": authToken,
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// LoginAdmin performs a login
func (client *gocloak) LoginAdmin(username, password, realm string) (*JWT, error) {
	var result JWT
	resp, err := client.restyClient.R().
		SetFormData(map[string]string{
			"client_id":  adminClientID,
			"grant_type": "password",
			"username":   username,
			"password":   password,
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// Login performs a login
func (client *gocloak) LoginClient(clientID, clientSecret, realm string) (*JWT, error) {
	var result JWT
	resp, err := client.getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
			"grant_type":    "client_credentials",
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err); err != nil {
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
	resp, err := client.getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"grant_type": "password",
			"username":   username,
			"password":   password,
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// Logout logs out users with refresh token
func (client *gocloak) Logout(clientID string, clientSecret string, realm string, refreshToken string) error {
	resp, err := client.getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"refresh_token": refreshToken,
		}).
		Post(client.getRealmURL(realm, logoutEndpoint))

	return checkForError(resp, err)
}

// RequestPermission l
func (client *gocloak) RequestPermission(clientID string, clientSecret string, realm string, username string, password string, permission string) (*JWT, error) {
	var result JWT
	resp, err := client.getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"grant_type": "password",
			"username":   username,
			"password":   password,
			"permission": permission,
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, errors.New("Authentication Failed")
	}

	return &result, nil
}

// ExecuteActionsEmail executes an actions email
func (client *gocloak) ExecuteActionsEmail(token string, realm string, params ExecuteActionsEmail) error {
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return err
	}
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(params.Actions).
		SetQueryParams(queryParams).
		Put(client.getAdminRealmURL(realm, "users", params.UserID, "execute-actions-email"))

	return checkForError(resp, err)
}

// CreateUser creates a new user
func (client *gocloak) CreateGroup(token string, realm string, group Group) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(group).
		Post(client.getAdminRealmURL(realm, "groups"))

	return checkForError(resp, err)
}

// CreateComponent creates a new user
func (client *gocloak) CreateComponent(token string, realm string, component Component) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(component).
		Post(client.getAdminRealmURL(realm, "components"))

	return checkForError(resp, err)
}

// CreateUser creates a new user
func (client *gocloak) CreateClient(token string, realm string, newClient Client) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(newClient).
		Post(client.getAdminRealmURL(realm, "clients"))

	return checkForError(resp, err)
}

// CreateClientRole creates a new role for a client
func (client *gocloak) CreateClientRole(token string, realm string, clientID string, role Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "roles"))

	return checkForError(resp, err)
}

// CreateClientScope creates a new client scope
func (client *gocloak) CreateClientScope(token string, realm string, scope ClientScope) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(scope).
		Post(client.getAdminRealmURL(realm, "client-scopes"))

	return checkForError(resp, err)
}

// UpdateUser creates a new user
func (client *gocloak) UpdateGroup(token string, realm string, group Group) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(group).
		Put(client.getAdminRealmURL(realm, "groups", group.ID))

	return checkForError(resp, err)
}

// UpdateUser creates a new user
func (client *gocloak) UpdateClient(token string, realm string, newClient Client) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(newClient).
		Put(client.getAdminRealmURL(realm, "clients"))

	return checkForError(resp, err)
}

// UpdateUser creates a new user
func (client *gocloak) UpdateRole(token string, realm string, clientID string, role Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "roles", role.Name))

	return checkForError(resp, err)
}

// UpdateClientScope creates a new client scope
func (client *gocloak) UpdateClientScope(token string, realm string, scope ClientScope) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(scope).
		Put(client.getAdminRealmURL(realm, "client-scopes", scope.ID))

	return checkForError(resp, err)
}

// DeleteUser creates a new user
func (client *gocloak) DeleteGroup(token string, realm string, groupID string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "groups", groupID))

	return checkForError(resp, err)
}

// DeleteClient deletes a given client
func (client *gocloak) DeleteClient(token string, realm string, clientID string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID))

	return checkForError(resp, err)
}

// DeleteComponent creates a new user
func (client *gocloak) DeleteComponent(token string, realm string, componentID string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "components", componentID))

	return checkForError(resp, err)
}

// DeleteClientRole deletes a given role
func (client *gocloak) DeleteClientRole(token string, realm string, clientID, roleName string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "roles", roleName))

	return checkForError(resp, err)
}

// DeleteClientScope creates a new client scope
func (client *gocloak) DeleteClientScope(token string, realm string, scopeID string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Put(client.getAdminRealmURL(realm, "client-scopes", scopeID))

	return checkForError(resp, err)
}

// GetClient returns a client
func (client *gocloak) GetClient(token string, realm string, clientID string) (*Client, error) {
	var result Client

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetClientSecret returns a client's secret
func (client *gocloak) GetClientSecret(token string, realm string, clientID string) (*CredentialRepresentation, error) {
	var result CredentialRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "client-secret"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetKeyStoreConfig get keystoreconfig of the realm
func (client *gocloak) GetKeyStoreConfig(token string, realm string) (*KeyStoreConfig, error) {
	var result KeyStoreConfig
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "keys"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetComponents get all cimponents in realm
func (client *gocloak) GetComponents(token string, realm string) (*[]Component, error) {
	var result []Component
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "components"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

func (client *gocloak) getRoleMappings(token string, realm string, path string, objectID string) (*MappingsRepresentation, error) {
	var result MappingsRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, path, objectID, "role-mappings"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRoleMappingByGroupID gets the role mappings by group
func (client *gocloak) GetRoleMappingByGroupID(token string, realm string, groupID string) (*MappingsRepresentation, error) {
	return client.getRoleMappings(token, realm, "groups", groupID)
}

// GetRoleMappingByUserID gets the role mappings by user
func (client *gocloak) GetRoleMappingByUserID(token string, realm string, userID string) (*MappingsRepresentation, error) {
	return client.getRoleMappings(token, realm, "users", userID)
}

// GetGroup get group with id in realm
func (client *gocloak) GetGroup(token string, realm string, groupID string) (*Group, error) {
	var result Group
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "group", groupID))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetGroups get all groups in realm
func (client *gocloak) GetGroups(token string, realm string, params GetGroupsParams) (*[]Group, error) {
	var result []Group
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "groups"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetClientRoles get all roles for the given client in realm
func (client *gocloak) GetClientRoles(token string, realm string, clientID string) (*[]Role, error) {
	var result []Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "roles"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetClientRole get a role for the given client in a realm by role name
func (client *gocloak) GetClientRole(token string, realm string, clientID string, roleName string) (*Role, error) {
	var result Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "roles", roleName))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetClients gets all clients in realm
func (client *gocloak) GetClients(token string, realm string, params GetClientsParams) (*[]Client, error) {
	var result []Client
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "clients"))

	if err := checkForError(resp, err); err != nil {
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

// -----------
// Realm Roles
// -----------

// CreateRealmRole creates a role in a realm
func (client *gocloak) CreateRealmRole(token string, realm string, role Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Post(client.getAdminRealmURL(realm, "roles"))

	return checkForError(resp, err)
}

// GetRealmRole returns a role from a realm by role's name
func (client *gocloak) GetRealmRole(token string, realm string, roleName string) (*Role, error) {
	var result Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "roles", roleName))

	if err = checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRealmRoles get all roles of the given realm.
func (client *gocloak) GetRealmRoles(token string, realm string) (*[]Role, error) {
	var result []Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "roles"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRealmRolesByUserID returns all roles assigned to the given user
func (client *gocloak) GetRealmRolesByUserID(token string, realm string, userID string) (*[]Role, error) {
	var result []Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"))

	if err = checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRealmRolesByGroupID returns all roles assigned to the given group
func (client *gocloak) GetRealmRolesByGroupID(token string, realm string, groupID string) (*[]Role, error) {
	var result []Role
	resp, err := client.getRequestWithBearerAuth(token).
		Get(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm"))

	if err = checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateRealmRole updates a role in a realm
func (client *gocloak) UpdateRealmRole(token string, realm string, roleName string, role Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Put(client.getAdminRealmURL(realm, "roles", roleName))

	return checkForError(resp, err)
}

// DeleteRealmRole deletes a role in a realm by role's name
func (client *gocloak) DeleteRealmRole(token string, realm string, roleName string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "roles", roleName))

	return checkForError(resp, err)
}

// AddRealmRoleToUser adds realm-level role mappings
func (client *gocloak) AddRealmRoleToUser(token string, realm string, userID string, roles []Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"))

	return checkForError(resp, err)
}

// DeleteRealmRoleFromUser deletes realm-level role mappings
func (client *gocloak) DeleteRealmRoleFromUser(token string, realm string, userID string, roles []Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"))

	return checkForError(resp, err)
}

func (client *gocloak) AddRealmRoleComposite(token string, realm string, roleName string, roles []Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "roles", roleName, "composites"))

	return checkForError(resp, err)
}

func (client *gocloak) DeleteRealmRoleComposite(token string, realm string, roleName string, roles []Role) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "roles", roleName, "composites"))

	return checkForError(resp, err)
}

// -----
// Realm
// -----

// GetRealm returns top-level representation of the realm
func (client *gocloak) GetRealm(token string, realm string) (*RealmRepresentation, error) {
	var result RealmRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm))

	if err = checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateRealm creates a realm
func (client *gocloak) CreateRealm(token string, realm RealmRepresentation) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(&realm).
		Post(client.getAdminRealmURL(""))

	return checkForError(resp, err)
}

// DeleteRealm removes a realm
func (client *gocloak) DeleteRealm(token string, realm string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm))
	return checkForError(resp, err)
}

// -----
// Users
// -----

// CreateUser creates the given user in the given realm and returns it's userID
func (client *gocloak) CreateUser(token string, realm string, user User) (*string, error) {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(user).
		Post(client.getAdminRealmURL(realm, "users"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	userPath := resp.Header().Get("Location")
	splittedPath := strings.Split(userPath, urlSeparator)
	userID := splittedPath[len(splittedPath)-1]

	return &userID, nil
}

// DeleteUser delete a given user
func (client *gocloak) DeleteUser(token string, realm string, userID string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "users", userID))

	return checkForError(resp, err)
}

// GetUserByID fetches a user from the given realm with the given userID
func (client *gocloak) GetUserByID(accessToken string, realm string, userID string) (*User, error) {
	if userID == "" {
		return nil, errors.New("UserID shall not be empty")
	}

	var result User
	resp, err := client.getRequestWithBearerAuth(accessToken).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserCount gets the user count in the realm
func (client *gocloak) GetUserCount(token string, realm string) (int, error) {
	var result int
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", "count"))

	if err := checkForError(resp, err); err != nil {
		return -1, err
	}

	return result, nil
}

// GetUserGroups get all groups for user
func (client *gocloak) GetUserGroups(token string, realm string, userID string) (*[]UserGroup, error) {
	var result []UserGroup
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID, "groups"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsers get all users in realm
func (client *gocloak) GetUsers(token string, realm string, params GetUsersParams) (*[]User, error) {
	var result []User
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "users"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUsersByRoleName returns all users have a given role
func (client *gocloak) GetUsersByRoleName(token string, realm string, roleName string) (*[]User, error) {
	var result []User
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "roles", roleName, "users"))

	if err := checkForError(resp, err); err != nil {
		return nil, err
	}

	return &result, nil
}

// SetPassword sets a new password for the user with the given id. Needs elevated privileges
func (client *gocloak) SetPassword(token string, userID string, realm string, password string, temporary bool) error {
	requestBody := SetPasswordRequest{Password: password, Temporary: temporary, Type: "password"}
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(requestBody).
		Put(client.getAdminRealmURL(realm, "users", userID, "reset-password"))

	return checkForError(resp, err)
}

// UpdateUser creates a new user
func (client *gocloak) UpdateUser(token string, realm string, user User) error {
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(user).
		Put(client.getAdminRealmURL(realm, "users", user.ID))

	return checkForError(resp, err)
}

// AddUserToGroup puts given user to given group
func (client *gocloak) AddUserToGroup(token string, realm string, userID string, groupID string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Put(client.getAdminRealmURL(realm, "users", userID, "groups", groupID))

	return checkForError(resp, err)
}

// DeleteUserFromGroup deletes given user from given group
func (client *gocloak) DeleteUserFromGroup(token string, realm string, userID string, groupID string) error {
	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "users", userID, "groups", groupID))

	return checkForError(resp, err)
}
