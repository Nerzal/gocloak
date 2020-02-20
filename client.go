package gocloak

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v4/pkg/jwx"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type gocloak struct {
	basePath    string
	certsCache  map[string]*CertResponse
	restyClient *resty.Client
	Config      struct {
		CertsInvalidateTime time.Duration
	}
	certsLock sync.Mutex
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

func (client *gocloak) getRequest() *resty.Request {
	var err HTTPErrorResponse
	return client.restyClient.R().SetError(&err)
}

func (client *gocloak) getRequestWithBearerAuth(token string) *resty.Request {
	return client.getRequest().
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json")
}

func (client *gocloak) getRequestWithBasicAuth(clientID string, clientSecret string) *resty.Request {
	req := client.getRequest().
		SetHeader("Content-Type", "application/x-www-form-urlencoded")
	// Public client doesn't require Basic Auth
	if len(clientID) > 0 && len(clientSecret) > 0 {
		httpBasicAuth := base64.URLEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
		req.SetHeader("Authorization", "Basic "+httpBasicAuth)
	}
	return req
}

func checkForError(resp *resty.Response, err error, errMessage string) error {
	if err != nil {
		return errors.Wrap(err, errMessage)
	}

	if resp == nil {
		return errors.New("empty response")
	}

	if resp.IsError() {
		var msg string

		e := resp.Error().(*HTTPErrorResponse)
		if e != nil && len(e.ErrorMessage) > 0 {
			msg = fmt.Sprintf("%s: %s", resp.Status(), e.ErrorMessage)
		} else if e != nil && len(e.Error) > 0 {
			msg = fmt.Sprintf("%s: %s", resp.Status(), e.Error)
		} else {
			msg = resp.Status()
		}

		return &APIError{
			Code:    resp.StatusCode(),
			Message: msg,
		}
	}

	return nil
}

func getID(resp *resty.Response) string {
	header := resp.Header().Get("Location")
	splittedPath := strings.Split(header, urlSeparator)
	return splittedPath[len(splittedPath)-1]
}

func findUsedKey(usedKeyID string, keys []*CertResponseKey) *CertResponseKey {
	for _, key := range keys {
		if *(key.Kid) == usedKeyID {
			return key
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

func (client *gocloak) SetRestyClient(restyClient *resty.Client) {
	client.restyClient = restyClient
}

func (client *gocloak) getRealmURL(realm string, path ...string) string {
	path = append([]string{client.basePath, authRealms, realm}, path...)
	return makeURL(path...)
}

func (client *gocloak) getAdminRealmURL(realm string, path ...string) string {
	path = append([]string{client.basePath, authAdminRealms, realm}, path...)
	return makeURL(path...)
}

func (client *gocloak) GetServerInfo(accessToken string) (*ServerInfoRepesentation, error) {
	var errMessage = "could not get server info"
	var result ServerInfoRepesentation

	resp, err := client.getRequestWithBearerAuth(accessToken).
		SetResult(&result).
		Get(makeURL(client.basePath, "auth", "admin", "serverinfo"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetUserInfo calls the UserInfo endpoint
func (client *gocloak) GetUserInfo(accessToken string, realm string) (*UserInfo, error) {
	const errMessage = "could not get user info"

	var result UserInfo
	resp, err := client.getRequestWithBearerAuth(accessToken).
		SetResult(&result).
		Get(client.getRealmURL(realm, openIDConnect, "userinfo"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

func (client *gocloak) getNewCerts(realm string) (*CertResponse, error) {
	const errMessage = "could not get newCerts"

	var result CertResponse
	resp, err := client.getRequest().
		SetResult(&result).
		Get(client.getRealmURL(realm, openIDConnect, "certs"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetCerts fetches certificates for the given realm from the public /open-id-connect/certs endpoint
func (client *gocloak) GetCerts(realm string) (*CertResponse, error) {
	const errMessage = "could not get certs"

	client.certsLock.Lock()
	defer client.certsLock.Unlock()

	if cert, ok := client.certsCache[realm]; ok {
		return cert, nil
	}

	cert, err := client.getNewCerts(realm)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	client.certsCache[realm] = cert

	timer := time.NewTimer(client.Config.CertsInvalidateTime)
	go func() {
		<-timer.C
		client.certsLock.Lock()
		delete(client.certsCache, realm)
		client.certsLock.Unlock()
	}()

	return cert, nil
}

// GetIssuer gets the issuer of the given realm
func (client *gocloak) GetIssuer(realm string) (*IssuerResponse, error) {
	const errMessage = "could not get issuer"

	var result IssuerResponse
	resp, err := client.getRequest().
		SetResult(&result).
		Get(client.getRealmURL(realm))

	if err := checkForError(resp, err, err.Error()); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// RetrospectToken calls the openid-connect introspect endpoint
func (client *gocloak) RetrospectToken(accessToken string, clientID, clientSecret string, realm string) (*RetrospecTokenResult, error) {
	const errMessage = "could not introspect requesting party token"

	var result RetrospecTokenResult
	resp, err := client.getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"token_type_hint": "requesting_party_token",
			"token":           accessToken,
		}).
		SetResult(&result).
		Post(client.getRealmURL(realm, tokenEndpoint, "introspect"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// DecodeAccessToken decodes the accessToken
func (client *gocloak) DecodeAccessToken(accessToken, realm string) (*jwt.Token, *jwt.MapClaims, error) {
	const errMessage = "could not decode access token"

	decodedHeader, err := jwx.DecodeAccessTokenHeader(accessToken)
	if err != nil {
		return nil, nil, errors.Wrap(err, errMessage)
	}

	certResult, err := client.GetCerts(realm)
	if err != nil {
		return nil, nil, errors.Wrap(err, errMessage)
	}

	usedKey := findUsedKey(decodedHeader.Kid, certResult.Keys)
	if usedKey == nil {
		return nil, nil, errors.Wrap(errors.New("cannot find a key to decode the token"), errMessage)
	}

	return jwx.DecodeAccessToken(accessToken, usedKey.E, usedKey.N)
}

// DecodeAccessTokenCustomClaims decodes the accessToken and writes claims into the given claims
func (client *gocloak) DecodeAccessTokenCustomClaims(accessToken string, realm string, claims jwt.Claims) (*jwt.Token, error) {
	const errMessage = "could not decode access token with custom claims"

	decodedHeader, err := jwx.DecodeAccessTokenHeader(accessToken)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	certResult, err := client.GetCerts(realm)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	usedKey := findUsedKey(decodedHeader.Kid, certResult.Keys)
	if usedKey == nil {
		return nil, errors.Wrap(errors.New("cannot find a key to decode the token"), errMessage)
	}

	return jwx.DecodeAccessTokenCustomClaims(accessToken, usedKey.E, usedKey.N, claims)
}

func (client *gocloak) GetToken(realm string, options TokenOptions) (*JWT, error) {
	const errMessage = "could not get token"

	var token JWT
	var req *resty.Request

	if !NilOrEmpty(options.ClientSecret) {
		req = client.getRequestWithBasicAuth(*(options.ClientID), *(options.ClientSecret))
	} else {
		req = client.getRequest()
	}

	resp, err := req.SetFormData(options.FormData()).
		SetResult(&token).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &token, nil
}

// GetRequestingPartyToken returns a requesting party token with permissions granted by the server
func (client *gocloak) GetRequestingPartyToken(token, realm string, options RequestingPartyTokenOptions) (*JWT, error) {
	const errMessage = "could not get requesting party token"

	var res JWT

	resp, err := client.getRequestWithBearerAuth(token).
		SetFormData(options.FormData()).
		SetFormDataFromValues(url.Values{"permission": options.Permissions}).
		SetResult(&res).
		Post(client.getRealmURL(realm, tokenEndpoint))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &res, nil
}

// RefreshToken refreshes the given token
func (client *gocloak) RefreshToken(refreshToken, clientID, clientSecret, realm string) (*JWT, error) {
	return client.GetToken(realm, TokenOptions{
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
		GrantType:    StringP("refresh_token"),
		RefreshToken: &refreshToken,
	})
}

// LoginAdmin performs a login with Admin client
func (client *gocloak) LoginAdmin(username, password, realm string) (*JWT, error) {
	return client.GetToken(realm, TokenOptions{
		ClientID:  StringP(adminClientID),
		GrantType: StringP("password"),
		Username:  &username,
		Password:  &password,
	})
}

// Login performs a login with client credentials
func (client *gocloak) LoginClient(clientID, clientSecret, realm string) (*JWT, error) {
	return client.GetToken(realm, TokenOptions{
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
		GrantType:    StringP("client_credentials"),
	})
}

// Login performs a login with user credentials and a client
func (client *gocloak) Login(clientID, clientSecret, realm, username, password string) (*JWT, error) {
	return client.GetToken(realm, TokenOptions{
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
		GrantType:    StringP("password"),
		Username:     &username,
		Password:     &password,
	})
}

// Logout logs out users with refresh token
func (client *gocloak) Logout(clientID, clientSecret, realm, refreshToken string) error {
	const errMessage = "could not logout"

	resp, err := client.getRequestWithBasicAuth(clientID, clientSecret).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"refresh_token": refreshToken,
		}).
		Post(client.getRealmURL(realm, logoutEndpoint))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) LogoutPublicClient(clientID, realm, accessToken, refreshToken string) error {
	const errMessage = "could not logout public client"

	resp, err := client.getRequestWithBearerAuth(accessToken).
		SetFormData(map[string]string{
			"client_id":     clientID,
			"refresh_token": refreshToken,
		}).
		Post(client.getRealmURL(realm, logoutEndpoint))

	return checkForError(resp, err, errMessage)
}

// ExecuteActionsEmail executes an actions email
func (client *gocloak) ExecuteActionsEmail(token, realm string, params ExecuteActionsEmail) error {
	const errMessage = "could not execute actions email"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return errors.Wrap(err, errMessage)
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(params.Actions).
		SetQueryParams(queryParams).
		Put(client.getAdminRealmURL(realm, "users", *(params.UserID), "execute-actions-email"))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) CreateGroup(token, realm string, group Group) (string, error) {
	const errMessage = "could not create group"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(group).
		Post(client.getAdminRealmURL(realm, "groups"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}
	return getID(resp), nil
}

// CreateChildGroup creates a new child group
func (client *gocloak) CreateChildGroup(token string, realm string, groupID string, group Group) (string, error) {
	const errMessage = "could not create child group"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(group).
		Post(client.getAdminRealmURL(realm, "groups", groupID, "children"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

func (client *gocloak) CreateComponent(token, realm string, component Component) (string, error) {
	const errMessage = "could not create component"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(component).
		Post(client.getAdminRealmURL(realm, "components"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

func (client *gocloak) CreateClient(token, realm string, newClient Client) (string, error) {
	const errMessage = "could not create client"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(newClient).
		Post(client.getAdminRealmURL(realm, "clients"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

// CreateClientRole creates a new role for a client
func (client *gocloak) CreateClientRole(token, realm, clientID string, role Role) (string, error) {
	const errMessage = "could not create client role"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "roles"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

// CreateClientScope creates a new client scope
func (client *gocloak) CreateClientScope(token, realm string, scope ClientScope) (string, error) {
	const errMessage = "could not create client scope"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(scope).
		Post(client.getAdminRealmURL(realm, "client-scopes"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

func (client *gocloak) UpdateGroup(token, realm string, updatedGroup Group) error {
	const errMessage = "could not update group"

	if NilOrEmpty(updatedGroup.ID) {
		return errors.Wrap(errors.New("ID of a group required"), errMessage)
	}
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(updatedGroup).
		Put(client.getAdminRealmURL(realm, "groups", PString(updatedGroup.ID)))

	return checkForError(resp, err, errMessage)
}

// UpdateClient updates the given Client
func (client *gocloak) UpdateClient(token, realm string, updatedClient Client) error {
	const errMessage = "could not update client"

	if NilOrEmpty(updatedClient.ID) {
		return errors.Wrap(errors.New("ID of a client required"), errMessage)
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(updatedClient).
		Put(client.getAdminRealmURL(realm, "clients", PString(updatedClient.ID)))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) UpdateRole(token, realm, clientID string, role Role) error {
	const errMessage = "could not update role"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "roles", PString(role.Name)))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) UpdateClientScope(token string, realm string, scope ClientScope) error {
	const errMessage = "could not update client scope"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(scope).
		Put(client.getAdminRealmURL(realm, "client-scopes", PString(scope.ID)))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) DeleteGroup(token string, realm string, groupID string) error {
	const errMessage = "could not delete group"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "groups", groupID))

	return checkForError(resp, err, errMessage)
}

// DeleteClient deletes a given client
func (client *gocloak) DeleteClient(token string, realm string, clientID string) error {
	const errMessage = "could not delete client"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) DeleteComponent(token string, realm string, componentID string) error {
	const errMessage = "could not delete component"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "components", componentID))

	return checkForError(resp, err, errMessage)
}

// DeleteClientRole deletes a given role
func (client *gocloak) DeleteClientRole(token, realm, clientID, roleName string) error {
	const errMessage = "could not delete client role"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "roles", roleName))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) DeleteClientScope(token string, realm string, scopeID string) error {
	const errMessage = "could not delete client scope"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "client-scopes", scopeID))

	return checkForError(resp, err, errMessage)
}

// GetClient returns a client
func (client *gocloak) GetClient(token string, realm string, clientID string) (*Client, error) {
	const errMessage = "could not get client"

	var result Client

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetClientsDefaultScopes returns a list of the client's default scopes
func (client *gocloak) GetClientsDefaultScopes(token string, realm string, clientID string) ([]*ClientScope, error) {
	const errMessage = "could not get clients default scopes"

	var result []*ClientScope

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "default-client-scopes"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// AddDefaultScopeToClient adds a client scope to the list of client's default scopes
func (client *gocloak) AddDefaultScopeToClient(token string, realm string, clientID string, scopeID string) error {
	const errMessage = "could not add default scope to client"

	resp, err := client.getRequestWithBearerAuth(token).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "default-client-scopes", scopeID))

	return checkForError(resp, err, errMessage)
}

// RemoveDefaultScopeFromClient removes a client scope from the list of client's default scopes
func (client *gocloak) RemoveDefaultScopeFromClient(token string, realm string, clientID string, scopeID string) error {
	const errMessage = "could not remove default scope from client"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "default-client-scopes", scopeID))

	return checkForError(resp, err, errMessage)
}

// GetClientsOptionalScopes returns a list of the client's optional scopes
func (client *gocloak) GetClientsOptionalScopes(token string, realm string, clientID string) ([]*ClientScope, error) {
	const errMessage = "could not get clients optional scopes"

	var result []*ClientScope

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "optional-client-scopes"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// AddOptionalScopeToClient adds a client scope to the list of client's optional scopes
func (client *gocloak) AddOptionalScopeToClient(token string, realm string, clientID string, scopeID string) error {
	const errMessage = "could not add optional scope to client"

	resp, err := client.getRequestWithBearerAuth(token).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "optional-client-scopes", scopeID))

	return checkForError(resp, err, errMessage)
}

// RemoveOptionalScopeFromClient deletes a client scope from the list of client's optional scopes
func (client *gocloak) RemoveOptionalScopeFromClient(token string, realm string, clientID string, scopeID string) error {
	const errMessage = "could not remove optional scope from client"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "optional-client-scopes", scopeID))

	return checkForError(resp, err, errMessage)
}

// GetDefaultOptionalClientScopes returns a list of default realm optional scopes
func (client *gocloak) GetDefaultOptionalClientScopes(token string, realm string) ([]*ClientScope, error) {
	const errMessage = "could not get default optional client scopes"

	var result []*ClientScope

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "default-optional-client-scopes"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetDefaultDefaultClientScopes returns a list of default realm default scopes
func (client *gocloak) GetDefaultDefaultClientScopes(token string, realm string) ([]*ClientScope, error) {
	const errMessage = "could not get default client scopes"

	var result []*ClientScope

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "default-default-client-scopes"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetClientScope returns a clientscope
func (client *gocloak) GetClientScope(token string, realm string, scopeID string) (*ClientScope, error) {
	const errMessage = "could not get client scope"

	var result ClientScope

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "client-scopes", scopeID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetClientScopes returns all client scopes
func (client *gocloak) GetClientScopes(token string, realm string) ([]*ClientScope, error) {
	const errMessage = "could not get client scopes"

	var result []*ClientScope

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "client-scopes"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetClientSecret returns a client's secret
func (client *gocloak) GetClientSecret(token string, realm string, clientID string) (*CredentialRepresentation, error) {
	const errMessage = "could not get client secret"

	var result CredentialRepresentation

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "client-secret"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetClientServiceAccount retrieves the service account "user" for a client if enabled
func (client *gocloak) GetClientServiceAccount(token string, realm string, clientID string) (*User, error) {
	const errMessage = "could not get client service account"

	var result User
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "service-account-user"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

func (client *gocloak) RegenerateClientSecret(token string, realm string, clientID string) (*CredentialRepresentation, error) {
	const errMessage = "could not regenerate client secret"

	var result CredentialRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "client-secret"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetClientOfflineSessions returns offline sessions associated with the client
func (client *gocloak) GetClientOfflineSessions(token, realm, clientID string) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get client offline sessions"

	var res []*UserSessionRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&res).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "offline-sessions"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return res, nil
}

// GetClientUserSessions returns user sessions associated with the client
func (client *gocloak) GetClientUserSessions(token, realm, clientID string) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get client user sessions"

	var res []*UserSessionRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&res).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "user-sessions"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return res, nil
}

// CreateClientProtocolMapper creates a protocol mapper in client scope
func (client *gocloak) CreateClientProtocolMapper(token, realm, clientID string, mapper ProtocolMapperRepresentation) (string, error) {
	const errMessage = "could not create client protocol mapper"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(mapper).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "protocol-mappers", "models"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

// UpdateClientProtocolMapper updates a protocol mapper in client scope
func (client *gocloak) UpdateClientProtocolMapper(token, realm, clientID string, mapperID string, mapper ProtocolMapperRepresentation) error {
	const errMessage = "could not update client protocol mapper"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(mapper).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "protocol-mappers", "models", mapperID))

	return checkForError(resp, err, errMessage)
}

// DeleteClientProtocolMapper deletes a protocol mapper in client scope
func (client *gocloak) DeleteClientProtocolMapper(token, realm, clientID, mapperID string) error {
	const errMessage = "could not delete client protocol mapper"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "protocol-mappers", "models", mapperID))

	return checkForError(resp, err, errMessage)
}

// GetKeyStoreConfig get keystoreconfig of the realm
func (client *gocloak) GetKeyStoreConfig(token string, realm string) (*KeyStoreConfig, error) {
	const errMessage = "could not get key store config"

	var result KeyStoreConfig
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "keys"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetComponents get all components in realm
func (client *gocloak) GetComponents(token string, realm string) ([]*Component, error) {
	const errMessage = "could not get components"

	var result []*Component
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "components"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetDefaultGroups returns a list of default groups
func (client *gocloak) GetDefaultGroups(token string, realm string) ([]*Group, error) {
	const errMessage = "could not get default groups"

	var result []*Group

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "default-groups"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// AddDefaultGroup adds group to the list of default groups
func (client *gocloak) AddDefaultGroup(token string, realm string, groupID string) error {
	const errMessage = "could not add default group"

	resp, err := client.getRequestWithBearerAuth(token).
		Put(client.getAdminRealmURL(realm, "default-groups", groupID))

	return checkForError(resp, err, errMessage)
}

// RemoveDefaultGroup removes group from the list of default groups
func (client *gocloak) RemoveDefaultGroup(token string, realm string, groupID string) error {
	const errMessage = "could not remove default group"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "default-groups", groupID))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) getRoleMappings(token string, realm string, path string, objectID string) (*MappingsRepresentation, error) {
	const errMessage = "could not get role mappings"

	var result MappingsRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, path, objectID, "role-mappings"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
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
	const errMessage = "could not get group"

	var result Group

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "groups", groupID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetGroups get all groups in realm
func (client *gocloak) GetGroups(token string, realm string, params GetGroupsParams) ([]*Group, error) {
	const errMessage = "could not get groups"

	var result []*Group
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "groups"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetGroupMembers get a list of users of group with id in realm
func (client *gocloak) GetGroupMembers(token string, realm string, groupID string, params GetGroupsParams) ([]*User, error) {
	const errMessage = "could not get group members"

	var result []*User
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "groups", groupID, "members"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetClientRoles get all roles for the given client in realm
func (client *gocloak) GetClientRoles(token string, realm string, clientID string) ([]*Role, error) {
	const errMessage = "could not get client roles"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "roles"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetRealmRolesByUserID returns all client roles assigned to the given user
func (client *gocloak) GetClientRolesByUserID(token string, realm string, clientID string, userID string) ([]*Role, error) {
	const errMessage = "could not client roles by user id"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", clientID))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// GetClientRolesByGroupID returns all client roles assigned to the given group
func (client *gocloak) GetClientRolesByGroupID(token string, realm string, clientID string, groupID string) ([]*Role, error) {
	const errMessage = "could not get client roles by group id"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", clientID))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// GetCompositeClientRolesByRoleID returns all client composite roles associated with the given client role
func (client *gocloak) GetCompositeClientRolesByRoleID(token string, realm string, clientID string, roleID string) ([]*Role, error) {
	const errMessage = "could not get composite client roles by role id"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "roles-by-id", roleID, "composites", "clients", clientID))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// GetCompositeClientRolesByUserID returns all client roles and composite roles assigned to the given user
func (client *gocloak) GetCompositeClientRolesByUserID(token string, realm string, clientID string, userID string) ([]*Role, error) {
	const errMessage = "could not get composite client roles by user id"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", clientID, "composite"))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// GetCompositeClientRolesByGroupID returns all client roles and composite roles assigned to the given group
func (client *gocloak) GetCompositeClientRolesByGroupID(token string, realm string, clientID string, groupID string) ([]*Role, error) {
	const errMessage = "could not get composite client roles by group id"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", clientID, "composite"))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// GetClientRole get a role for the given client in a realm by role name
func (client *gocloak) GetClientRole(token string, realm string, clientID string, roleName string) (*Role, error) {
	const errMessage = "could not get client role"

	var result Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "roles", roleName))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetClients gets all clients in realm
func (client *gocloak) GetClients(token string, realm string, params GetClientsParams) ([]*Client, error) {
	const errMessage = "could not get clients"

	var result []*Client
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "clients"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
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
func (client *gocloak) CreateRealmRole(token string, realm string, role Role) (string, error) {
	const errMessage = "could not create realm role"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Post(client.getAdminRealmURL(realm, "roles"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

// GetRealmRole returns a role from a realm by role's name
func (client *gocloak) GetRealmRole(token string, realm string, roleName string) (*Role, error) {
	const errMessage = "could not get realm role"

	var result Role

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "roles", roleName))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetRealmRoles get all roles of the given realm.
func (client *gocloak) GetRealmRoles(token string, realm string) ([]*Role, error) {
	const errMessage = "could not get realm roles"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "roles"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetRealmRolesByUserID returns all roles assigned to the given user
func (client *gocloak) GetRealmRolesByUserID(token string, realm string, userID string) ([]*Role, error) {
	const errMessage = "could not get realm roles by user id"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetRealmRolesByGroupID returns all roles assigned to the given group
func (client *gocloak) GetRealmRolesByGroupID(token string, realm string, groupID string) ([]*Role, error) {
	const errMessage = "could not get realm roles by group id"

	var result []*Role
	resp, err := client.getRequestWithBearerAuth(token).
		Get(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm"))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// UpdateRealmRole updates a role in a realm
func (client *gocloak) UpdateRealmRole(token string, realm string, roleName string, role Role) error {
	const errMessage = "could not update realm role"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(role).
		Put(client.getAdminRealmURL(realm, "roles", roleName))

	return checkForError(resp, err, errMessage)
}

// DeleteRealmRole deletes a role in a realm by role's name
func (client *gocloak) DeleteRealmRole(token string, realm string, roleName string) error {
	const errMessage = "could not delete realm role"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "roles", roleName))

	return checkForError(resp, err, errMessage)
}

// AddRealmRoleToUser adds realm-level role mappings
func (client *gocloak) AddRealmRoleToUser(token string, realm string, userID string, roles []Role) error {
	const errMessage = "could not add realm role to user"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"))

	return checkForError(resp, err, errMessage)
}

// DeleteRealmRoleFromUser deletes realm-level role mappings
func (client *gocloak) DeleteRealmRoleFromUser(token string, realm string, userID string, roles []Role) error {
	const errMessage = "could not delete realm role from user"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "realm"))

	return checkForError(resp, err, errMessage)
}

// AddRealmRoleToGroup adds realm-level role mappings
func (client *gocloak) AddRealmRoleToGroup(token string, realm string, groupID string, roles []Role) error {
	const errMessage = "could not add realm role to group"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm"))

	return checkForError(resp, err, errMessage)
}

// DeleteRealmRoleFromGroup deletes realm-level role mappings
func (client *gocloak) DeleteRealmRoleFromGroup(token string, realm string, groupID string, roles []Role) error {
	const errMessage = "could not delete realm role from group"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "realm"))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) AddRealmRoleComposite(token string, realm string, roleName string, roles []Role) error {
	const errMessage = "could not add realm role composite"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "roles", roleName, "composites"))

	return checkForError(resp, err, errMessage)
}

func (client *gocloak) DeleteRealmRoleComposite(token string, realm string, roleName string, roles []Role) error {
	const errMessage = "could not delete realm role composite"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "roles", roleName, "composites"))

	return checkForError(resp, err, errMessage)
}

// -----
// Realm
// -----

// GetRealm returns top-level representation of the realm
func (client *gocloak) GetRealm(token string, realm string) (*RealmRepresentation, error) {
	const errMessage = "could not get realm"

	var result RealmRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetRealms returns top-level representation of all realms
func (client *gocloak) GetRealms(token string) ([]*RealmRepresentation, error) {
	const errMessage = "could not get realms"

	var result []*RealmRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(""))

	if err = checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// CreateRealm creates a realm
func (client *gocloak) CreateRealm(token string, realm RealmRepresentation) (string, error) {
	const errMessage = "could not create realm"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(&realm).
		Post(client.getAdminRealmURL(""))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}
	return getID(resp), nil
}

// UpdateRealm updates a given realm
func (client *gocloak) UpdateRealm(token string, realm RealmRepresentation) error {
	const errMessage = "could not update realm"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(realm).
		Put(client.getAdminRealmURL(PString(realm.Realm)))

	return checkForError(resp, err, errMessage)
}

// DeleteRealm removes a realm
func (client *gocloak) DeleteRealm(token string, realm string) error {
	const errMessage = "could not delete realm"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm))

	return checkForError(resp, err, errMessage)
}

// ClearRealmCache clears realm cache
func (client *gocloak) ClearRealmCache(token string, realm string) error {
	const errMessage = "could not clear realm cache"

	resp, err := client.getRequestWithBearerAuth(token).
		Post(client.getAdminRealmURL(realm, "clear-realm-cache"))

	return checkForError(resp, err, errMessage)
}

// ClearUserCache clears realm cache
func (client *gocloak) ClearUserCache(token string, realm string) error {
	const errMessage = "could not clear user cache"

	resp, err := client.getRequestWithBearerAuth(token).
		Post(client.getAdminRealmURL(realm, "clear-user-cache"))

	return checkForError(resp, err, errMessage)
}

// ClearKeysCache clears realm cache
func (client *gocloak) ClearKeysCache(token string, realm string) error {
	const errMessage = "could not clear keys cache"

	resp, err := client.getRequestWithBearerAuth(token).
		Post(client.getAdminRealmURL(realm, "clear-keys-cache"))

	return checkForError(resp, err, errMessage)
}

// -----
// Users
// -----

// CreateUser creates the given user in the given realm and returns it's userID
func (client *gocloak) CreateUser(token string, realm string, user User) (string, error) {
	const errMessage = "could not create user"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(user).
		Post(client.getAdminRealmURL(realm, "users"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", errors.Wrap(err, errMessage)
	}

	return getID(resp), nil
}

// DeleteUser delete a given user
func (client *gocloak) DeleteUser(token string, realm string, userID string) error {
	const errMessage = "could not delete user"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "users", userID))

	return checkForError(resp, err, errMessage)
}

// GetUserByID fetches a user from the given realm with the given userID
func (client *gocloak) GetUserByID(accessToken string, realm string, userID string) (*User, error) {
	const errMessage = "could not get user by id"

	if userID == "" {
		return nil, errors.Wrap(errors.New("userID shall not be empty"), errMessage)
	}

	var result User
	resp, err := client.getRequestWithBearerAuth(accessToken).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return &result, nil
}

// GetUserCount gets the user count in the realm
func (client *gocloak) GetUserCount(token string, realm string) (int, error) {
	const errMessage = "could not get user count"

	var result int
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", "count"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return -1, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetUserGroups get all groups for user
func (client *gocloak) GetUserGroups(token string, realm string, userID string) ([]*UserGroup, error) {
	const errMessage = "could not get user groups"

	var result []*UserGroup
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "users", userID, "groups"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetUsers get all users in realm
func (client *gocloak) GetUsers(token string, realm string, params GetUsersParams) ([]*User, error) {
	const errMessage = "could not get users"

	var result []*User
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "users"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetUsersByRoleName returns all users have a given role
func (client *gocloak) GetUsersByRoleName(token string, realm string, roleName string) ([]*User, error) {
	const errMessage = "could not get users by role name"

	var result []*User
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "roles", roleName, "users"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return result, nil
}

// GetUsersByClientRoleName returns all users have a given client role
func (client *gocloak) GetUsersByClientRoleName(token string, realm string, clientID string, roleName string, params GetUsersByRoleParams) ([]*User, error) {
	const errMessage = "could not get users by client role name"

	var result []*User
	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "roles", roleName, "users"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// SetPassword sets a new password for the user with the given id. Needs elevated privileges
func (client *gocloak) SetPassword(token string, userID string, realm string, password string, temporary bool) error {
	const errMessage = "could not set password"

	requestBody := SetPasswordRequest{Password: &password, Temporary: &temporary, Type: StringP("password")}
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(requestBody).
		Put(client.getAdminRealmURL(realm, "users", userID, "reset-password"))

	return checkForError(resp, err, errMessage)
}

// UpdateUser updates a given user
func (client *gocloak) UpdateUser(token string, realm string, user User) error {
	const errMessage = "could not update user"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(user).
		Put(client.getAdminRealmURL(realm, "users", PString(user.ID)))

	return checkForError(resp, err, errMessage)
}

// AddUserToGroup puts given user to given group
func (client *gocloak) AddUserToGroup(token string, realm string, userID string, groupID string) error {
	const errMessage = "could not add user to group"

	resp, err := client.getRequestWithBearerAuth(token).
		Put(client.getAdminRealmURL(realm, "users", userID, "groups", groupID))

	return checkForError(resp, err, errMessage)
}

// DeleteUserFromGroup deletes given user from given group
func (client *gocloak) DeleteUserFromGroup(token string, realm string, userID string, groupID string) error {
	const errMessage = "could not delete user from group"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "users", userID, "groups", groupID))

	return checkForError(resp, err, errMessage)
}

// GetUserSessions returns user sessions associated with the user
func (client *gocloak) GetUserSessions(token, realm, userID string) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get user sessions"

	var res []*UserSessionRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&res).
		Get(client.getAdminRealmURL(realm, "users", userID, "sessions"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return res, nil
}

// GetUserOfflineSessionsForClient returns offline sessions associated with the user and client
func (client *gocloak) GetUserOfflineSessionsForClient(token, realm, userID, clientID string) ([]*UserSessionRepresentation, error) {
	const errMessage = "could not get user offline sessions for client"

	var res []*UserSessionRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&res).
		Get(client.getAdminRealmURL(realm, "users", userID, "offline-sessions", clientID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	return res, nil
}

// AddClientRoleToUser adds client-level role mappings
func (client *gocloak) AddClientRoleToUser(token string, realm string, clientID string, userID string, roles []Role) error {
	const errMessage = "could not add client role to user"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", clientID))

	return checkForError(resp, err, errMessage)
}

// AddClientRoleToGroup adds a client role to the group
func (client *gocloak) AddClientRoleToGroup(token string, realm string, clientID string, groupID string, roles []Role) error {
	const errMessage = "could not add client role to group"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", clientID))

	return checkForError(resp, err, errMessage)
}

// DeleteClientRoleFromUser adds client-level role mappings
func (client *gocloak) DeleteClientRoleFromUser(token string, realm string, clientID string, userID string, roles []Role) error {
	const errMessage = "could not delete client role from user"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "users", userID, "role-mappings", "clients", clientID))

	return checkForError(resp, err, errMessage)
}

// DeleteClientRoleFromGroup removes a client role from from the group
func (client *gocloak) DeleteClientRoleFromGroup(token string, realm string, clientID string, groupID string, roles []Role) error {
	const errMessage = "could not client role from group"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "groups", groupID, "role-mappings", "clients", clientID))

	return checkForError(resp, err, errMessage)
}

// AddClientRoleComposite adds roles as composite
func (client *gocloak) AddClientRoleComposite(token string, realm string, roleID string, roles []Role) error {
	const errMessage = "could not add client role composite"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Post(client.getAdminRealmURL(realm, "roles-by-id", roleID, "composites"))

	return checkForError(resp, err, errMessage)
}

// DeleteClientRoleComposite deletes composites from a role
func (client *gocloak) DeleteClientRoleComposite(token string, realm string, roleID string, roles []Role) error {
	const errMessage = "could not delete client role composite"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(roles).
		Delete(client.getAdminRealmURL(realm, "roles-by-id", roleID, "composites"))

	return checkForError(resp, err, errMessage)
}

// ------------------
// Identity Providers
// ------------------

// CreateIdentityProvider creates an identity provider in a realm
func (client *gocloak) CreateIdentityProvider(token string, realm string, providerRep IdentityProviderRepresentation) (string, error) {
	const errMessage = "could not create identity provider"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(providerRep).
		Post(client.getAdminRealmURL(realm, "identity-provider", "instances"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return "", err
	}

	return getID(resp), nil
}

// GetIdentityProviders returns list of identity providers in a realm
func (client *gocloak) GetIdentityProviders(token string, realm string) ([]*IdentityProviderRepresentation, error) {
	const errMessage = "could not get identity providers"

	result := []*IdentityProviderRepresentation{}
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "identity-provider", "instances"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// GetIdentityProvider gets the identity provider in a realm
func (client *gocloak) GetIdentityProvider(token string, realm string, alias string) (*IdentityProviderRepresentation, error) {
	const errMessage = "could not get identity provider"

	result := IdentityProviderRepresentation{}
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "identity-provider", "instances", alias))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateIdentityProvider updates the identity provider in a realm
func (client *gocloak) UpdateIdentityProvider(token string, realm string, alias string, providerRep IdentityProviderRepresentation) error {
	const errMessage = "could not update identity provider"

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(providerRep).
		Put(client.getAdminRealmURL(realm, "identity-provider", "instances", alias))

	return checkForError(resp, err, errMessage)
}

// DeleteIdentityProvider deletes the identity provider in a realm
func (client *gocloak) DeleteIdentityProvider(token string, realm string, alias string) error {
	const errMessage = "could not delete identity provider"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "identity-provider", "instances", alias))

	return checkForError(resp, err, errMessage)
}

// GetResource returns a client's resource with the given id
func (client *gocloak) GetResource(token string, realm string, clientID string, resourceID string) (*ResourceRepresentation, error) {
	const errMessage = "could not get resource"

	var result ResourceRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "resource", resourceID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetResources returns resources associated with the client
func (client *gocloak) GetResources(token string, realm string, clientID string, params GetResourceParams) ([]*ResourceRepresentation, error) {
	const errMessage = "could not get resources"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}

	var result []*ResourceRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "resource"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateResource creates a resource associated with the client
func (client *gocloak) CreateResource(token, realm string, clientID string, resource ResourceRepresentation) (*ResourceRepresentation, error) {
	const errMessage = "could not create resource"

	var result ResourceRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetBody(resource).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "resource"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateResource updates a resource associated with the client
func (client *gocloak) UpdateResource(token string, realm string, clientID string, resource ResourceRepresentation) error {
	const errMessage = "could not update resource"

	if NilOrEmpty(resource.ID) {
		return errors.New("ID of a resource required")
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(resource).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "resource", *(resource.ID)))

	return checkForError(resp, err, errMessage)
}

// DeleteResource deletes a resource associated with the client
func (client *gocloak) DeleteResource(token string, realm string, clientID string, resourceID string) error {
	const errMessage = "could not delete resource"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "resource", resourceID))

	return checkForError(resp, err, errMessage)
}

// GetScope returns a client's scope with the given id
func (client *gocloak) GetScope(token string, realm string, clientID string, scopeID string) (*ScopeRepresentation, error) {
	const errMessage = "could not get scope"

	var result ScopeRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "scope", scopeID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetScopes returns scopes associated with the client
func (client *gocloak) GetScopes(token string, realm string, clientID string, params GetScopeParams) ([]*ScopeRepresentation, error) {
	const errMessage = "could not get scopes"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, err
	}
	var result []*ScopeRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "scope"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateScope creates a scope associated with the client
func (client *gocloak) CreateScope(token string, realm string, clientID string, scope ScopeRepresentation) (*ScopeRepresentation, error) {
	const errMessage = "could not create scope"

	var result ScopeRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetBody(scope).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "scope"))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateScope updates a scope associated with the client
func (client *gocloak) UpdateScope(token string, realm string, clientID string, scope ScopeRepresentation) error {
	const errMessage = "could not update scope"

	if NilOrEmpty(scope.ID) {
		return errors.New("ID of a scope required")
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(scope).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "scope", *(scope.ID)))

	return checkForError(resp, err, errMessage)
}

// DeleteScope deletes a scope associated with the client
func (client *gocloak) DeleteScope(token string, realm string, clientID string, scopeID string) error {
	const errMessage = "could not delete scope"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "scope", scopeID))

	return checkForError(resp, err, errMessage)
}

// GetPolicy returns a client's policy with the given id
func (client *gocloak) GetPolicy(token string, realm string, clientID string, policyID string) (*PolicyRepresentation, error) {
	const errMessage = "could not get policy"

	var result PolicyRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "policy", policyID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPolicies returns policies associated with the client
func (client *gocloak) GetPolicies(token string, realm string, clientID string, params GetPolicyParams) ([]*PolicyRepresentation, error) {
	const errMessage = "could not get policies"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	var adminURL string
	if NilOrEmpty(params.Type) {
		adminURL = client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "policy")
	} else {
		client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "policy", *(params.Type))
	}

	var result []*PolicyRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(adminURL)

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// CreatePolicy creates a policy associated with the client
func (client *gocloak) CreatePolicy(token string, realm string, clientID string, policy PolicyRepresentation) (*PolicyRepresentation, error) {
	const errMessage = "could not create policy"

	if NilOrEmpty(policy.Type) {
		return nil, errors.New("type of a policy required")
	}

	var result PolicyRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetBody(policy).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "policy", *(policy.Type)))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdatePolicy updates a policy associated with the client
func (client *gocloak) UpdatePolicy(token string, realm string, clientID string, policy PolicyRepresentation) error {
	const errMessage = "could not update policy"

	if NilOrEmpty(policy.ID) {
		return errors.New("ID of a policy required")
	}

	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(policy).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "policy", *(policy.ID)))

	return checkForError(resp, err, errMessage)
}

// DeletePolicy deletes a policy associated with the client
func (client *gocloak) DeletePolicy(token string, realm string, clientID string, policyID string) error {
	const errMessage = "could not delete policy"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "policy", policyID))

	return checkForError(resp, err, errMessage)
}

// GetPermission returns a client's permission with the given id
func (client *gocloak) GetPermission(token string, realm string, clientID string, permissionID string) (*PermissionRepresentation, error) {
	const errMessage = "could not get permission"

	var result PermissionRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		Get(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "permission", permissionID))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPolicies returns permissions associated with the client
func (client *gocloak) GetPermissions(token string, realm string, clientID string, params GetPermissionParams) ([]*PermissionRepresentation, error) {
	const errMessage = "could not get permissions"

	queryParams, err := GetQueryParams(params)
	if err != nil {
		return nil, errors.Wrap(err, errMessage)
	}

	var adminURL string
	if NilOrEmpty(params.Type) {
		adminURL = client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "permission")
	} else {
		client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "permission", *(params.Type))
	}

	var result []*PermissionRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetQueryParams(queryParams).
		Get(adminURL)

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return result, nil
}

// CreatePermission creates a permission associated with the client
func (client *gocloak) CreatePermission(token string, realm string, clientID string, permission PermissionRepresentation) (*PermissionRepresentation, error) {
	const errMessage = "could not craete permission"

	if NilOrEmpty(permission.Type) {
		return nil, errors.New("type of a permission required")
	}

	var result PermissionRepresentation
	resp, err := client.getRequestWithBearerAuth(token).
		SetResult(&result).
		SetBody(permission).
		Post(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "permission", *(permission.Type)))

	if err := checkForError(resp, err, errMessage); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdatePermission updates a permission associated with the client
func (client *gocloak) UpdatePermission(token string, realm string, clientID string, permission PermissionRepresentation) error {
	const errMessage = "could not update permission"

	if NilOrEmpty(permission.ID) {
		return errors.New("ID of a permission required")
	}
	resp, err := client.getRequestWithBearerAuth(token).
		SetBody(permission).
		Put(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "permission", *(permission.ID)))

	return checkForError(resp, err, errMessage)
}

// DeletePermission deletes a policy associated with the client
func (client *gocloak) DeletePermission(token string, realm string, clientID string, permissionID string) error {
	const errMessage = "could not delete permission"

	resp, err := client.getRequestWithBearerAuth(token).
		Delete(client.getAdminRealmURL(realm, "clients", clientID, "authz", "resource-server", "permission", permissionID))

	return checkForError(resp, err, errMessage)
}
