package gocloak

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type configAdmin struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	Realm    string `json:"realm"`
}

type configGoCloak struct {
	UserName     string `json:"username"`
	Password     string `json:"password"`
	Realm        string `json:"realm"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Config struct {
	HostName string        `json:"hostname"`
	Proxy    string        `json:"proxy,omitempty"`
	Admin    configAdmin   `json:"admin"`
	GoCloak  configGoCloak `json:"gocloak"`
}

var (
	config     *Config
	configOnce sync.Once
	setupOnce  sync.Once
	testUserID string
)

const (
	gocloakClientID = "60be66a5-e007-464c-9b74-0e3c2e69e478"
)

func GetConfig(t testing.TB) *Config {
	configOnce.Do(func() {
		rand.Seed(time.Now().UTC().UnixNano())
		configFileName, ok := os.LookupEnv("GOCLOAK_TEST_CONFIG")
		if !ok {
			configFileName = filepath.Join("testdata", "config.json")
		}
		configFile, err := os.Open(configFileName)
		assert.NoError(t, err, "cannot open config.json")
		defer func() {
			err := configFile.Close()
			assert.NoError(t, err, "cannot close config file")
		}()
		data, err := ioutil.ReadAll(configFile)
		assert.NoError(t, err, "cannot read config.json")
		config = &Config{}
		err = json.Unmarshal(data, config)
		assert.NoError(t, err, "cannot parse config.json")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		if len(config.Proxy) != 0 {
			proxy, err := url.Parse(config.Proxy)
			assert.NoError(t, err, "incorrect proxy url: "+config.Proxy)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(proxy)
		}
		if config.GoCloak.UserName == "" {
			config.GoCloak.UserName = "test_user"
		}
	})
	return config
}

func GetClientToken(t *testing.T, client GoCloak) *JWT {
	cfg := GetConfig(t)
	token, err := client.LoginClient(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	assert.NoError(t, err, "Login failed")
	return token
}

func GetUserToken(t *testing.T, client GoCloak) *JWT {
	SetUpTestUser(t, client)
	cfg := GetConfig(t)
	token, err := client.Login(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password)
	assert.NoError(t, err, "Login failed")
	return token
}

func GetAdminToken(t testing.TB, client GoCloak) *JWT {
	cfg := GetConfig(t)
	token, err := client.LoginAdmin(
		cfg.Admin.UserName,
		cfg.Admin.Password,
		cfg.Admin.Realm)
	assert.NoError(t, err, "Login failed")
	return token
}

func GetRandomName(name string) string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	randomNumber := r1.Intn(100000)
	return name + strconv.Itoa(randomNumber)
}

func GetRandomNameP(name string) *string {
	r := GetRandomName(name)
	return &r
}

func GetClientByClientID(t *testing.T, client GoCloak, clientID string) *Client {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	clients, err := client.GetClients(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetClientsParams{
			ClientID: &clientID,
		})
	assert.NoError(t, err, "GetClients failed")
	for _, fetchedClient := range clients {
		if fetchedClient.ClientID == nil {
			continue
		}
		if *(fetchedClient.ClientID) == clientID {
			return fetchedClient
		}
	}
	t.Fatal("Client not found")
	return nil
}

func CreateGroup(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	group := Group{
		Name: GetRandomNameP("GroupName"),
		Attributes: map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
	}
	groupID, err := client.CreateGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		group)
	assert.NoError(t, err, "CreateGroup failed")
	t.Logf("Created Group ID: %s ", groupID)

	tearDown := func() {
		err := client.DeleteGroup(
			token.AccessToken,
			cfg.GoCloak.Realm,
			groupID)
		assert.NoError(t, err, "DeleteGroup failed")
	}
	return tearDown, groupID
}

func CreateResource(t *testing.T, client GoCloak, clientID string) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	resource := ResourceRepresentation{
		Name:        GetRandomNameP("ResourceName"),
		DisplayName: StringP("Resource Display Name"),
		Type:        StringP("urn:gocloak:resources:test"),
		IconURI:     StringP("/resource/test/icon"),
		Attributes: map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
		URIs: []string{
			"/resource/1",
			"/resource/2",
		},
		OwnerManagedAccess: BoolP(true),
	}
	createdResource, err := client.CreateResource(
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		resource)
	assert.NoError(t, err, "CreateResource failed")
	t.Logf("Created Resource ID: %s ", *(createdResource.ID))

	tearDown := func() {
		err := client.DeleteResource(
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*(createdResource.ID))
		assert.NoError(t, err, "DeleteResource failed")
	}
	return tearDown, *(createdResource.ID)
}

func CreateScope(t *testing.T, client GoCloak, clientID string) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	scope := ScopeRepresentation{
		Name:        GetRandomNameP("ScopeName"),
		DisplayName: StringP("Scope Display Name"),
		IconURI:     StringP("/scope/test/icon"),
	}
	createdScope, err := client.CreateScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		scope)
	assert.NoError(t, err, "CreateScope failed")
	t.Logf("Created Scope ID: %s ", *(createdScope.ID))

	tearDown := func() {
		err := client.DeleteScope(
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*(createdScope.ID))
		assert.NoError(t, err, "DeleteScope failed")
	}
	return tearDown, *(createdScope.ID)
}

func CreatePolicy(t *testing.T, client GoCloak, clientID string, policy PolicyRepresentation) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	createdPolicy, err := client.CreatePolicy(
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		policy)
	assert.NoError(t, err, "CreatePolicy failed")
	t.Logf("Created Policy ID: %s ", *(createdPolicy.ID))

	tearDown := func() {
		err := client.DeletePolicy(
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*(createdPolicy.ID))
		assert.NoError(t, err, "DeletePolicy failed")
	}
	return tearDown, *(createdPolicy.ID)
}

func CreatePermission(t *testing.T, client GoCloak, clientID string, permission PermissionRepresentation) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	createdPermission, err := client.CreatePermission(
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		permission)
	assert.NoError(t, err, "CreatePermission failed")
	t.Logf("Created Permission ID: %s ", *(createdPermission.ID))

	tearDown := func() {
		err := client.DeletePermission(
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*(createdPermission.ID))
		assert.NoError(t, err, "DeletePermission failed")
	}
	return tearDown, *(createdPermission.ID)
}

func SetUpTestUser(t testing.TB, client GoCloak) {
	setupOnce.Do(func() {
		cfg := GetConfig(t)
		token := GetAdminToken(t, client)

		user := User{
			Username:      StringP(cfg.GoCloak.UserName),
			Email:         StringP(cfg.GoCloak.UserName + "@localhost"),
			EmailVerified: BoolP(true),
			Enabled:       BoolP(true),
		}

		createdUserID, err := client.CreateUser(
			token.AccessToken,
			cfg.GoCloak.Realm,
			user,
		)
		if IsObjectAlreadyExists(err) {
			users, err := client.GetUsers(
				token.AccessToken,
				cfg.GoCloak.Realm,
				GetUsersParams{
					Username: StringP(cfg.GoCloak.UserName),
				})
			assert.NoError(t, err, "GetUsers failed")
			for _, user := range users {
				if PString(user.Username) == cfg.GoCloak.UserName {
					testUserID = PString(user.ID)
					break
				}
			}
		} else {
			assert.NoError(t, err, "CreateUser failed")
			testUserID = createdUserID
		}

		err = client.SetPassword(
			token.AccessToken,
			testUserID,
			cfg.GoCloak.Realm,
			cfg.GoCloak.Password,
			false)
		assert.NoError(t, err, "SetPassword failed")
	})
}

type RestyLogWriter struct {
	io.Writer
	t testing.TB
}

func (w *RestyLogWriter) Errorf(format string, v ...interface{}) {
	w.write("[ERROR] "+format, v...)
}

func (w *RestyLogWriter) Warnf(format string, v ...interface{}) {
	w.write("[WARN] "+format, v...)
}

func (w *RestyLogWriter) Debugf(format string, v ...interface{}) {
	w.write("[DEBUG] "+format, v...)
}

func (w *RestyLogWriter) write(format string, v ...interface{}) {
	w.t.Logf(format, v...)
}

func NewClientWithDebug(t testing.TB) GoCloak {
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	restyClient := client.RestyClient()
	restyClient.SetDebug(true)
	restyClient.SetLogger(&RestyLogWriter{
		t: t,
	})

	cond := func(resp *resty.Response, err error) bool {
		if resp != nil && resp.IsError() {
			e := resp.Error().(*HTTPErrorResponse)
			if e != nil {
				var msg string
				if len(e.ErrorMessage) > 0 {
					msg = e.ErrorMessage
				} else if len(e.Error) > 0 {
					msg = e.Error
				}
				return strings.HasPrefix(msg, "Cached clientScope not found")
			}
		}
		return false
	}
	restyClient.AddRetryCondition(cond)
	restyClient.SetRetryCount(10)

	return client
}

// FailRequest fails requests and returns an error
//   err - returned error or nil to return the default error
//   failN - number of requests to be failed
//   skipN = number of requests to be executed and not failed by this function
func FailRequest(client GoCloak, err error, failN, skipN int) GoCloak {
	client.RestyClient().OnBeforeRequest(
		func(c *resty.Client, r *resty.Request) error {
			if skipN > 0 {
				skipN--
				return nil
			}
			if failN == 0 {
				return nil
			}
			failN--
			if err == nil {
				err = fmt.Errorf("an error for request: %+v", r)
			}
			return err
		},
	)
	return client
}

func ClearRealmCache(t testing.TB, client GoCloak, realm ...string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	if len(realm) == 0 {
		realm = append(realm, cfg.Admin.Realm, cfg.GoCloak.Realm)
	}
	for _, r := range realm {
		err := client.ClearRealmCache(token.AccessToken, r)
		assert.NoError(t, err, "ClearRealmCache failed for a realm: %s", r)
		err = client.ClearUserCache(token.AccessToken, r)
		assert.NoError(t, err, "ClearUserCache failed for a realm: %s", r)
		err = client.ClearKeysCache(token.AccessToken, r)
		assert.NoError(t, err, "ClearKeysCache failed for a realm: %s", r)
	}
}

// -----
// Tests
// -----

func TestGocloak_RestyClient(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	restyClient := client.RestyClient()
	assert.NotEqual(t, restyClient, resty.New())
}

func TestGocloak_SetRestyClient(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	newRestyClient := resty.New()
	client.SetRestyClient(newRestyClient)
	restyClient := client.RestyClient()
	assert.Equal(t, newRestyClient, restyClient)
}

func TestGocloak_checkForError(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	FailRequest(client, nil, 1, 0)
	_, err := client.Login("", "", "", "", "")
	assert.Error(t, err, "All requests must fail with NewClientWithError")
	t.Logf("Error: %s", err.Error())
}

// ---------
// API tests
// ---------

func TestGocloak_GetServerInfo(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	serverInfo, err := client.GetServerInfo(
		token.AccessToken,
	)
	assert.NoError(t, err, "Failed to fetch server info")
	t.Logf("Server Info: %+v", serverInfo)

	FailRequest(client, nil, 1, 0)
	_, err = client.GetServerInfo(
		token.AccessToken,
	)
	assert.Error(t, err)
}

func TestGocloak_GetUserInfo(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)
	userInfo, err := client.GetUserInfo(
		token.AccessToken,
		cfg.GoCloak.Realm)
	assert.NoError(t, err, "Failed to fetch userinfo")
	t.Log(userInfo)
	FailRequest(client, nil, 1, 0)
	_, err = client.GetUserInfo(
		token.AccessToken,
		cfg.GoCloak.Realm)
	assert.Error(t, err, "")
}

func TestGocloak_RequestPermission(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	token, err := client.Login(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password)
	assert.NoError(t, err, "login failed")

	rpt, err := client.GetRequestingPartyToken(token.AccessToken, cfg.GoCloak.Realm, RequestingPartyTokenOptions{
		Audience: StringP(cfg.GoCloak.ClientID),
		Permissions: []string{
			"Fake Resource",
		},
	})
	assert.Error(t, err, "GetRequestingPartyToken failed")
	assert.Nil(t, rpt)

	rpt, err = client.GetRequestingPartyToken(token.AccessToken, cfg.GoCloak.Realm, RequestingPartyTokenOptions{
		Audience: StringP(cfg.GoCloak.ClientID),
		Permissions: []string{
			"Default Resource",
		},
	})
	assert.NoError(t, err, "GetRequestingPartyToken failed")
	assert.NotNil(t, rpt)

	rptResult, err := client.RetrospectToken(
		rpt.AccessToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(rptResult)
	assert.NoError(t, err, "inspection failed")
	assert.True(t, PBool(rptResult.Active), "Inactive Token oO")
	assert.Equal(t, 1, len(rptResult.Permissions), "GetRequestingPartyToken failed")
	assert.Equal(t, "Default Resource", *(rptResult.Permissions[0].RSName), "GetRequestingPartyToken failed")
}

func TestGocloak_GetCerts(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	certs, err := client.GetCerts(cfg.GoCloak.Realm)
	assert.NoError(t, err, "get certs")
	t.Log(certs)
}

func TestGocloak_LoginClient_UnknownRealm(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginClient(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		"ThisRealmDoesNotExist")
	assert.Error(t, err, "Login shouldn't be successful")
	assert.EqualError(t, err, "404 Not Found: Realm does not exist")
}

func TestGocloak_GetIssuer(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	issuer, err := client.GetIssuer(cfg.GoCloak.Realm)
	t.Log(issuer)
	assert.NoError(t, err, "get issuer")
}

func TestGocloak_RetrospectToken_InactiveToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)

	rptResult, err := client.RetrospectToken(
		"foobar",
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(rptResult)
	assert.NoError(t, err, "inspection failed")
	assert.False(t, PBool(rptResult.Active), "That should never happen. Token is active")
}

func TestGocloak_RetrospectToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	rptResult, err := client.RetrospectToken(
		token.AccessToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(rptResult)
	assert.NoError(t, err, "Inspection failed")
	assert.False(t, !PBool(rptResult.Active), "Inactive Token oO")
}

func TestGocloak_DecodeAccessToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	resultToken, claims, err := client.DecodeAccessToken(
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	assert.NoError(t, err)
	t.Log(resultToken)
	t.Log(claims)
}

func TestGocloak_DecodeAccessTokenCustomClaims(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	claims := jwt.MapClaims{}
	resultToken, err := client.DecodeAccessTokenCustomClaims(
		token.AccessToken,
		cfg.GoCloak.Realm,
		claims,
	)
	assert.NoError(t, err)
	t.Log(resultToken)
	t.Log(claims)
}

func TestGocloak_RefreshToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	token, err := client.RefreshToken(
		token.RefreshToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(token)
	assert.NoError(t, err, "RefreshToken failed")
}

func TestGocloak_UserAttributeContains(t *testing.T) {
	t.Parallel()

	attributes := map[string][]string{}
	attributes["foo"] = []string{"bar", "alice", "bob", "roflcopter"}
	attributes["bar"] = []string{"baz"}

	client := NewClientWithDebug(t)
	ok := client.UserAttributeContains(attributes, "foo", "alice")
	assert.False(t, !ok, "UserAttributeContains")
}

func TestGocloak_GetKeyStoreConfig(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	config, err := client.GetKeyStoreConfig(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Log(config)
	assert.NoError(t, err, "GetKeyStoreConfig")
}

func TestGocloak_Login(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.Login(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password)
	assert.NoError(t, err, "Login failed")
}

func TestGocloak_GetToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	newToken, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:      &cfg.GoCloak.ClientID,
			ClientSecret:  &cfg.GoCloak.ClientSecret,
			Username:      &cfg.GoCloak.UserName,
			Password:      &cfg.GoCloak.Password,
			GrantType:     StringP("password"),
			ResponseTypes: []string{"token", "id_token"},
			Scopes:        []string{"openid", "offline_access"},
		},
	)
	assert.NoError(t, err, "Login failed")
	t.Logf("New token: %+v", *newToken)
	assert.Equal(t, newToken.RefreshExpiresIn, 0, "Got a refresh token instead of offline")
	assert.NotEmpty(t, newToken.IDToken, "Got an empty if token")
}

func TestGocloak_GetRequestingPartyToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	newToken, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:      &cfg.GoCloak.ClientID,
			ClientSecret:  &cfg.GoCloak.ClientSecret,
			Username:      &cfg.GoCloak.UserName,
			Password:      &cfg.GoCloak.Password,
			GrantType:     StringP("password"),
			ResponseTypes: []string{"token", "id_token"},
			Scopes:        []string{"openid"},
		},
	)
	assert.NoError(t, err, "Login failed")
	t.Logf("New token: %+v", *newToken)
	assert.NotEmpty(t, newToken.IDToken, "Got an empty id token")

	rpt, err := client.GetRequestingPartyToken(
		newToken.AccessToken,
		cfg.GoCloak.Realm,
		RequestingPartyTokenOptions{
			Audience: &cfg.GoCloak.ClientID,
		},
	)
	assert.NoError(t, err, "Get requesting party token failed")
	t.Logf("New RPT: %+v", *rpt)

	_, err = client.RetrospectToken(rpt.AccessToken, cfg.GoCloak.ClientID, cfg.GoCloak.ClientSecret, cfg.GoCloak.Realm)
	assert.NoError(t, err, "RetrospectToken failed")
}

func TestGocloak_LoginClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginClient(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	assert.NoError(t, err, "LoginClient failed")
}

func TestGocloak_LoginAdmin(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginAdmin(
		cfg.Admin.UserName,
		cfg.Admin.Password,
		cfg.Admin.Realm)
	assert.NoError(t, err, "LoginAdmin failed")
}

func TestGocloak_SetPassword(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	err := client.SetPassword(
		token.AccessToken,
		userID,
		cfg.GoCloak.Realm,
		"passwort1234!",
		false)
	assert.NoError(t, err, "Failed to set password")
}

func TestGocloak_CreateListGetUpdateDeleteGetChildGroup(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Create
	tearDown, groupID := CreateGroup(t, client)
	// Delete
	defer tearDown()

	// List
	createdGroup, err := client.GetGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	assert.NoError(t, err, "GetGroup failed")
	t.Logf("Created Group: %+v", createdGroup)
	assert.Equal(t, groupID, *(createdGroup.ID))

	err = client.UpdateGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Group{},
	)
	assert.Error(t, err, "Should fail because of missing ID of the group")

	createdGroup.Name = GetRandomNameP("GroupName")
	err = client.UpdateGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*createdGroup,
	)
	assert.NoError(t, err, "UpdateGroup failed")

	updatedGroup, err := client.GetGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	assert.NoError(t, err, "GetGroup failed")
	assert.Equal(t, *(createdGroup.Name), *(updatedGroup.Name))

	childGroupID, err := client.CreateChildGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		Group{
			Name: GetRandomNameP("GroupName"),
		},
	)
	assert.NoError(t, err, "CreateChildGroup failed")

	_, err = client.GetGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		childGroupID,
	)
	assert.NoError(t, err, "GetGroup failed")
}

func CreateClientRole(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	roleName := GetRandomName("Role")
	t.Logf("Creating Client Role: %s", roleName)
	clientRoleID, err := client.CreateClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		Role{
			Name: &roleName,
		})
	t.Logf("Created Client Role ID: %s", clientRoleID)
	assert.Equal(t, roleName, clientRoleID)

	assert.NoError(t, err, "CreateClientRole failed")
	tearDown := func() {
		err := client.DeleteClientRole(
			token.AccessToken,
			cfg.GoCloak.Realm,
			gocloakClientID,
			roleName)
		assert.NoError(t, err, "DeleteClientRole failed")
	}
	return tearDown, roleName
}

func TestGocloak_CreateClientRole(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	tearDown, _ := CreateClientRole(t, client)
	tearDown()
}

func TestGocloak_GetClientRole(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	tearDown, roleName := CreateClientRole(t, client)
	defer tearDown()
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	role, err := client.GetClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName,
	)
	assert.NoError(t, err, "GetClientRoleI failed")
	assert.NotNil(t, role)
	token = GetAdminToken(t, client)
	role, err = client.GetClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		"Fake-Role-Name",
	)
	assert.Error(t, err)
	assert.Nil(t, role)
}

func CreateClientScope(t *testing.T, client GoCloak, scope *ClientScope) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	if scope == nil {
		scope = &ClientScope{
			ID:   GetRandomNameP("client-scope-id-"),
			Name: GetRandomNameP("client-scope-name-"),
		}
	}

	t.Logf("Creating Client Scope: %+v", scope)
	clientScopeID, err := client.CreateClientScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*scope,
	)
	if !NilOrEmpty(scope.ID) {
		assert.Equal(t, clientScopeID, *(scope.ID))
	}
	assert.NoError(t, err, "CreateClientScope failed")
	tearDown := func() {
		err := client.DeleteClientScope(
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientScopeID,
		)
		assert.NoError(t, err, "DeleteClientScope failed")
	}
	return tearDown, clientScopeID
}

func TestGocloak_CreateClientScope_DeleteClientScope(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	tearDown, _ := CreateClientScope(t, client, nil)
	tearDown()
}

func TestGocloak_ListAddRemoveDefaultClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	scope := ClientScope{
		ID:       GetRandomNameP("client-scope-id-"),
		Name:     GetRandomNameP("client-scope-name-"),
		Protocol: StringP("openid-connect"),
		ClientScopeAttributes: &ClientScopeAttributes{
			IncludeInTokenScope: StringP("true"),
		},
	}

	tearDown, scopeID := CreateClientScope(t, client, &scope)
	defer tearDown()

	scopesBeforeAdding, err := client.GetClientsDefaultScopes(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	assert.NoError(t, err, "GetClientsDefaultScopes failed")

	err = client.AddDefaultScopeToClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	assert.NoError(t, err, "AddDefaultScopeToClient failed")

	scopesAfterAdding, err := client.GetClientsDefaultScopes(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	assert.NoError(t, err, "GetClientsDefaultScopes failed")

	assert.NotEqual(t, len(scopesBeforeAdding), len(scopesAfterAdding), "scope should have been added")

	err = client.RemoveDefaultScopeFromClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	assert.NoError(t, err, "RemoveDefaultScopeFromClient failed")

	scopesAfterRemoving, err := client.GetClientsDefaultScopes(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	assert.NoError(t, err, "GetClientsDefaultScopes failed")

	assert.Equal(t, len(scopesAfterRemoving), len(scopesBeforeAdding), "scope should have been removed")
}

func TestGocloak_ListAddRemoveOptionalClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	scope := ClientScope{
		ID:       GetRandomNameP("client-scope-id-"),
		Name:     GetRandomNameP("client-scope-name-"),
		Protocol: StringP("openid-connect"),
		ClientScopeAttributes: &ClientScopeAttributes{
			IncludeInTokenScope: StringP("true"),
		},
	}
	tearDown, scopeID := CreateClientScope(t, client, &scope)
	defer tearDown()

	scopesBeforeAdding, err := client.GetClientsOptionalScopes(token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID)
	assert.NoError(t, err, "GetClientsOptionalScopes failed")

	err = client.AddOptionalScopeToClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID)
	assert.NoError(t, err, "AddOptionalScopeToClient failed")

	scopesAfterAdding, err := client.GetClientsOptionalScopes(token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID)
	assert.NoError(t, err, "GetClientsOptionalScopes failed")

	assert.NotEqual(t, len(scopesAfterAdding), len(scopesBeforeAdding), "scope should have been added")

	err = client.RemoveOptionalScopeFromClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID)
	assert.NoError(t, err, "RemoveOptionalScopeFromClient failed")

	scopesAfterRemoving, err := client.GetClientsOptionalScopes(token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID)
	assert.NoError(t, err, "GetClientsOptionalScopes failed")

	assert.Equal(t, len(scopesBeforeAdding), len(scopesAfterRemoving), "scope should have been removed")
}

func TestGocloak_GetDefaultOptionalClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	scopes, err := client.GetDefaultOptionalClientScopes(
		token.AccessToken,
		cfg.GoCloak.Realm)

	assert.NoError(t, err, "GetDefaultOptionalClientScopes failed")

	assert.NotEqual(t, 0, len(scopes), "there should be default optional client scopes")
}

func TestGocloak_GetDefaultDefaultClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	scopes, err := client.GetDefaultDefaultClientScopes(
		token.AccessToken,
		cfg.GoCloak.Realm)

	assert.NoError(t, err, "GetDefaultDefaultClientScopes failed")

	assert.NotEqual(t, 0, len(scopes), "there should be default default client scopes")
}

func TestGocloak_GetClientScope(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	tearDown, scopeID := CreateClientScope(t, client, nil)
	defer tearDown()

	// Getting exact client scope
	createdClientScope, err := client.GetClientScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		scopeID,
	)
	assert.NoError(t, err, "GetClientScope failed")
	// Checking that GetClientScope returns same client scope
	assert.NotNil(t, createdClientScope.ID)
	assert.Equal(t, scopeID, *(createdClientScope.ID))
}

func TestGocloak_GetClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Getting client scopes
	scopes, err := client.GetClientScopes(
		token.AccessToken,
		cfg.GoCloak.Realm)
	assert.NoError(t, err, "GetClientScopes failed")
	// Checking that GetClientScopes returns scopes
	assert.NotZero(t, len(scopes), "there should be client scopes")
}

func TestGocloak_CreateListGetUpdateDeleteClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	clientID := GetRandomNameP("ClientID")
	t.Logf("Client ID: %s", *clientID)

	// Creating a client
	createdClientID, err := client.CreateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Client{
			ClientID: clientID,
			Name:     GetRandomNameP("Name"),
			BaseURL:  StringP("http://example.com"),
		},
	)
	assert.NoError(t, err, "CreateClient failed")

	// Looking for a created client
	clients, err := client.GetClients(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetClientsParams{
			ClientID: clientID,
		},
	)
	assert.NoError(t, err, "CreateClients failed")
	assert.Len(t, clients, 1, "GetClients should return exact 1 client")
	assert.Equal(t, createdClientID, *(clients[0].ID))
	t.Logf("Clients: %+v", clients)

	// Getting exact client
	createdClient, err := client.GetClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		createdClientID,
	)
	assert.NoError(t, err, "GetClient failed")
	t.Logf("Created client: %+v", createdClient)
	// Checking that GetClient returns same client
	assert.Equal(t, clients[0], createdClient)

	// Updating the client

	// Should fail
	err = client.UpdateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Client{},
	)
	assert.Error(t, err, "Should fail because of missing ID of the client")

	// Update existing client
	createdClient.Name = GetRandomNameP("Name")
	err = client.UpdateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*createdClient,
	)
	assert.NoError(t, err, "GetClient failed")

	// Getting updated client
	updatedClient, err := client.GetClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		createdClientID,
	)
	assert.NoError(t, err, "GetClient failed")
	t.Logf("Update client: %+v", createdClient)
	assert.Equal(t, *createdClient, *updatedClient)

	// Deleting the client
	err = client.DeleteClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		createdClientID,
	)
	assert.NoError(t, err, "DeleteClient failed")

	// Verifying that the client was deleted
	clients, err = client.GetClients(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetClientsParams{
			ClientID: clientID,
		},
	)
	assert.NoError(t, err, "CreateClients failed")
	assert.Len(t, clients, 0, "GetClients should not return any clients")
}

func TestGocloak_GetGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	_, err := client.GetGroups(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetGroupsParams{})
	assert.NoError(t, err, "GetGroups failed")
}

func TestGocloak_GetGroupsFull(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	groups, err := client.GetGroups(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetGroupsParams{
			Full: BoolP(true),
		})
	assert.NoError(t, err, "GetGroups failed")

	for _, group := range groups {
		if NilOrEmpty(group.ID) {
			continue
		}
		if *(group.ID) == groupID {
			ok := client.UserAttributeContains(group.Attributes, "foo", "alice")
			assert.True(t, ok, "UserAttributeContains")
			return
		}
	}

	assert.Fail(t, "GetGroupsFull failed")
}

func TestGocloak_GetGroupFull(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	createdGroup, err := client.GetGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	assert.NoError(t, err, "GetGroup failed")

	ok := client.UserAttributeContains(createdGroup.Attributes, "foo", "alice")
	assert.True(t, ok, "UserAttributeContains")
}

func TestGocloak_GetGroupMembers(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()

	err := client.AddUserToGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	assert.NoError(t, err, "AddUserToGroup failed")

	users, err := client.GetGroupMembers(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		GetGroupsParams{},
	)
	assert.NoError(t, err, "AddUserToGroup failed")

	assert.Equal(
		t,
		1,
		len(users),
	)
}

func TestGocloak_GetClientRoles(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	testClient := GetClientByClientID(t, client, cfg.GoCloak.ClientID)

	_, err := client.GetClientRoles(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*(testClient.ID))
	assert.NoError(t, err, "GetClientRoles failed")
}

func TestGocloak_GetRoleMappingByGroupID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	_, err := client.GetRoleMappingByGroupID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID)
	assert.NoError(t, err, "GetRoleMappingByGroupID failed")
}

func TestGocloak_GetRoleMappingByUserID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	_, err := client.GetRoleMappingByUserID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	assert.NoError(t, err, "GetRoleMappingByUserID failed")
}

func TestGocloak_ExecuteActionsEmail_UpdatePassword(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	params := ExecuteActionsEmail{
		ClientID: &(cfg.GoCloak.ClientID),
		UserID:   &userID,
		Actions:  []string{"UPDATE_PASSWORD"},
	}

	err := client.ExecuteActionsEmail(
		token.AccessToken,
		cfg.GoCloak.Realm,
		params)

	if err != nil {
		if err.Error() == "500 Internal Server Error: Failed to send execute actions email" {
			return
		}
		assert.NoError(t, err, "ExecuteActionsEmail failed")
	}
}

func TestGocloak_Logout(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetUserToken(t, client)

	err := client.Logout(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		token.RefreshToken)
	assert.NoError(t, err, "Logout failed")
}

func TestGocloak_GetRealm(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	r, err := client.GetRealm(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Logf("%+v", r)
	assert.NoError(t, err, "GetRealm failed")
}

func TestGocloak_GetRealms(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	r, err := client.GetRealms(token.AccessToken)
	t.Logf("%+v", r)
	assert.NoError(t, err, "GetRealms failed")
}

// -----------
// Realm
// -----------

func CreateRealm(t *testing.T, client GoCloak) (func(), string) {
	token := GetAdminToken(t, client)

	realmName := GetRandomName("Realm")
	t.Logf("Creating Realm: %s", realmName)
	realmID, err := client.CreateRealm(
		token.AccessToken,
		RealmRepresentation{
			Realm: &realmName,
		})
	assert.NoError(t, err, "CreateRealm failed")
	assert.Equal(t, realmID, realmName)
	tearDown := func() {
		err := client.DeleteRealm(
			token.AccessToken,
			realmName)
		assert.NoError(t, err, "DeleteRealm failed")
	}
	return tearDown, realmName
}

func TestGocloak_CreateRealm(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	tearDown, _ := CreateRealm(t, client)
	defer tearDown()
}

func TestGocloak_ClearRealmCache(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	ClearRealmCache(t, client)
}

// -----------
// Realm Roles
// -----------

func CreateRealmRole(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	roleName := GetRandomName("Role")
	t.Logf("Creating RoleName: %s", roleName)
	realmRoleID, err := client.CreateRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Role{
			Name:        &roleName,
			ContainerID: StringP("asd"),
		})
	assert.NoError(t, err, "CreateRealmRole failed")
	assert.Equal(t, roleName, realmRoleID)
	tearDown := func() {
		err := client.DeleteRealmRole(
			token.AccessToken,
			cfg.GoCloak.Realm,
			roleName)
		assert.NoError(t, err, "DeleteRealmRole failed")
	}
	return tearDown, roleName
}

func TestGocloak_CreateRealmRole(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	tearDown, _ := CreateRealmRole(t, client)
	defer tearDown()
}

func TestGocloak_GetRealmRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, roleName := CreateRealmRole(t, client)
	defer tearDown()

	role, err := client.GetRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	assert.NoError(t, err, "GetRealmRole failed")
	t.Logf("Role: %+v", *role)
	assert.False(
		t,
		*(role.Name) != roleName,
		"GetRealmRole returns unexpected result. Expected: %s; Actual: %+v",
		roleName, role)
}

func TestGocloak_GetRealmRoles(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, _ := CreateRealmRole(t, client)
	defer tearDown()

	roles, err := client.GetRealmRoles(
		token.AccessToken,
		cfg.GoCloak.Realm)
	assert.NoError(t, err, "GetRealmRoles failed")
	t.Logf("Roles: %+v", roles)
}

func TestGocloak_UpdateRealmRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	newRoleName := GetRandomName("Role")
	_, oldRoleName := CreateRealmRole(t, client)

	err := client.UpdateRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		oldRoleName,
		Role{
			Name: &newRoleName,
		})
	assert.NoError(t, err, "UpdateRealmRole failed")
	err = client.DeleteRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		oldRoleName)
	assert.Error(
		t,
		err,
		"Role with old name was deleted successfully, but it shouldn't. Old role: %s; Updated role: %s",
		oldRoleName, newRoleName)
	err = client.DeleteRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		newRoleName)
	assert.NoError(t, err, "DeleteRealmRole failed")
}

func TestGocloak_DeleteRealmRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	_, roleName := CreateRealmRole(t, client)

	err := client.DeleteRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	assert.NoError(t, err, "DeleteRealmRole failed")
}

func TestGocloak_AddRealmRoleToUser_DeleteRealmRoleFromUser(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()
	tearDownRole, roleName := CreateRealmRole(t, client)
	defer tearDownRole()
	role, err := client.GetRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	assert.NoError(t, err, "GetRealmRole failed")

	roles := []Role{*role}
	err = client.AddRealmRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		roles,
	)
	assert.NoError(t, err, "AddRealmRoleToUser failed")
	err = client.DeleteRealmRoleFromUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		roles,
	)
	assert.NoError(t, err, "DeleteRealmRoleFromUser failed")
}

func TestGocloak_GetRealmRolesByUserID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()
	tearDownRole, roleName := CreateRealmRole(t, client)
	defer tearDownRole()
	role, err := client.GetRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	assert.NoError(t, err)

	err = client.AddRealmRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]Role{
			*role,
		})
	assert.NoError(t, err)

	roles, err := client.GetRealmRolesByUserID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	assert.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *(r.Name) == *(role.Name) {
			return
		}
	}
	assert.Fail(t, "The role has not been found in the assined roles. Role: %+v", *role)
}

func TestGocloak_GetRealmRolesByGroupID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	_, err := client.GetRealmRolesByGroupID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID)
	assert.NoError(t, err, "GetRealmRolesByGroupID failed")
}

func TestGocloak_AddRealmRoleComposite_DeleteRealmRoleComposite(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, compositeRoleName := CreateRealmRole(t, client)
	defer tearDown()

	tearDown, roleName := CreateRealmRole(t, client)
	defer tearDown()

	role, err := client.GetRealmRole(token.AccessToken, cfg.GoCloak.Realm, roleName)
	assert.NoError(t, err, "Can't get just created role with GetRealmRole")

	err = client.AddRealmRoleComposite(token.AccessToken,
		cfg.GoCloak.Realm, compositeRoleName, []Role{*role})
	assert.NoError(t, err)

	err = client.DeleteRealmRoleComposite(token.AccessToken,
		cfg.GoCloak.Realm, compositeRoleName, []Role{*role})
	assert.NoError(t, err)
}

// -----
// Users
// -----

func CreateUser(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	user := User{
		FirstName: GetRandomNameP("FirstName"),
		LastName:  GetRandomNameP("LastName"),
		Email:     StringP(GetRandomName("email") + "@localhost"),
		Enabled:   BoolP(true),
		Attributes: map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
	}
	user.Username = user.Email

	userID, err := client.CreateUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		user)
	assert.NoError(t, err, "CreateUser failed")
	user.ID = &userID
	t.Logf("Created User: %+v", user)
	tearDown := func() {
		err := client.DeleteUser(
			token.AccessToken,
			cfg.GoCloak.Realm,
			*(user.ID))
		assert.NoError(t, err, "DeleteUser")
	}

	return tearDown, *(user.ID)
}

func TestGocloak_CreateUser(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	tearDown, _ := CreateUser(t, client)
	defer tearDown()
}

func TestGocloak_CreateUserCustomAttributes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	fetchedUser, err := client.GetUserByID(token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	assert.NoError(t, err, "GetUserByID failed")
	ok := client.UserAttributeContains(fetchedUser.Attributes, "foo", "alice")
	assert.False(t, !ok, "User doesn't have custom attributes")
	ok = client.UserAttributeContains(fetchedUser.Attributes, "foo2", "alice")
	assert.False(t, ok, "User's custom attributes contains unexpected attribute")
	t.Log(fetchedUser)
}

func TestGocloak_GetUserByID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	fetchedUser, err := client.GetUserByID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	assert.NoError(t, err, "GetUserById failed")
	t.Log(fetchedUser)
}

func TestGocloak_GetUsers(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	users, err := client.GetUsers(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetUsersParams{
			Username: &(cfg.GoCloak.UserName),
		})
	assert.NoError(t, err, "GetUsers failed")
	t.Log(users)
}

func TestGocloak_GetUserCount(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	count, err := client.GetUserCount(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Logf("Users in Realm: %d", count)
	assert.NoError(t, err, "GetUserCount failed")
}

func TestGocloak_AddUserToGroup(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()

	err := client.AddUserToGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	assert.NoError(t, err, "AddUserToGroup failed")
}

func TestGocloak_DeleteUserFromGroup(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()
	err := client.AddUserToGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	assert.NoError(t, err, "AddUserToGroup failed")
	err = client.DeleteUserFromGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	assert.NoError(t, err, "DeleteUserFromGroup failed")
}

func TestGocloak_GetUserGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()

	err := client.AddUserToGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	assert.NoError(t, err)
	groups, err := client.GetUserGroups(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	assert.NoError(t, err)
	assert.NotEqual(
		t,
		len(groups),
		0,
	)
	assert.Equal(
		t,
		groupID,
		*(groups[0].ID))
}

func TestGocloak_DeleteUser(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	tearDown, _ := CreateUser(t, client)
	defer tearDown()
}

func TestGocloak_UpdateUser(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()
	user, err := client.GetUserByID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	assert.NoError(t, err, "GetUserByID failed")
	user.FirstName = GetRandomNameP("UpdateUserFirstName")
	err = client.UpdateUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*user)
	assert.NoError(t, err, "UpdateUser failed")
}

func TestGocloak_UpdateUserSetEmptyEmail(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()
	user, err := client.GetUserByID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	assert.NoError(t, err)
	user.Email = StringP("")
	err = client.UpdateUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*user)
	assert.NoError(t, err)
	user, err = client.GetUserByID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	assert.NoError(t, err)
	assert.Nil(t, user.Email)
}

func TestGocloak_GetUsersByRoleName(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	tearDownRole, roleName := CreateRealmRole(t, client)
	defer tearDownRole()

	role, err := client.GetRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	assert.NoError(t, err)
	err = client.AddRealmRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]Role{
			*role,
		})
	assert.NoError(t, err)

	users, err := client.GetUsersByRoleName(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	assert.NoError(t, err)

	assert.NotEqual(
		t,
		len(users),
		0,
	)
	assert.Equal(
		t,
		userID,
		*(users[0].ID),
	)
}

func TestGocloak_GetUserSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:     &(cfg.GoCloak.ClientID),
			ClientSecret: &(cfg.GoCloak.ClientSecret),
			Username:     &(cfg.GoCloak.UserName),
			Password:     &(cfg.GoCloak.Password),
			GrantType:    StringP("password"),
		},
	)
	assert.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetUserSessions(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
	)
	assert.NoError(t, err, "GetUserSessions failed")
	assert.False(t, len(sessions) == 0, "GetUserSessions returned an empty list")
}

func TestGocloak_GetUserOfflineSessionsForClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:      &(cfg.GoCloak.ClientID),
			ClientSecret:  &(cfg.GoCloak.ClientSecret),
			Username:      &(cfg.GoCloak.UserName),
			Password:      &(cfg.GoCloak.Password),
			GrantType:     StringP("password"),
			ResponseTypes: []string{"token", "id_token"},
			Scopes:        []string{"openid", "offline_access"},
		},
	)
	assert.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetUserOfflineSessionsForClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
		gocloakClientID,
	)
	assert.NoError(t, err, "GetUserOfflineSessionsForClient failed")
	assert.False(t, len(sessions) == 0, "GetUserOfflineSessionsForClient returned an empty list")
}

func TestGocloak_GetClientUserSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:     &(cfg.GoCloak.ClientID),
			ClientSecret: &(cfg.GoCloak.ClientSecret),
			Username:     &(cfg.GoCloak.UserName),
			Password:     &(cfg.GoCloak.Password),
			GrantType:    StringP("password"),
		},
	)
	assert.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetClientUserSessions(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	assert.NoError(t, err, "GetClientUserSessions failed")
	assert.False(t, len(sessions) == 0, "GetClientUserSessions returned an empty list")
}

func TestGocloak_CreateDeleteClientProtocolMapper(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	id := GetRandomName("protocol-mapper-id-")
	testClient := GetClientByClientID(t, client, cfg.GoCloak.ClientID)

	found := false
	for _, protocolMapper := range testClient.ProtocolMappers {
		if protocolMapper == nil || NilOrEmpty(protocolMapper.ID) {
			continue
		}
		if *(protocolMapper.ID) == id {
			found = true
			break
		}
	}
	assert.False(
		t,
		found,
		"default client should not have a protocol mapper with ID: %s", id,
	)

	token := GetAdminToken(t, client)
	createdID, err := client.CreateClientProtocolMapper(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*(testClient.ID),
		ProtocolMapperRepresentation{
			ID:             &id,
			Name:           StringP("test"),
			Protocol:       StringP("openid-connect"),
			ProtocolMapper: StringP("oidc-usermodel-attribute-mapper"),
			Config: map[string]string{
				"access.token.claim":   "true",
				"aggregate.attrs":      "",
				"claim.name":           "test",
				"id.token.claim":       "true",
				"jsonType.label":       "String",
				"multivalued":          "",
				"user.attribute":       "test",
				"userinfo.token.claim": "true",
			},
		},
	)
	assert.NoError(t, err, "CreateClientProtocolMapper failed")
	assert.Equal(t, id, createdID)
	testClientAfter := GetClientByClientID(t, client, cfg.GoCloak.ClientID)

	found = false
	for _, protocolMapper := range testClientAfter.ProtocolMappers {
		if protocolMapper == nil || NilOrEmpty(protocolMapper.ID) {
			continue
		}
		if *(protocolMapper.ID) == id {
			found = true
			break
		}
	}
	assert.True(
		t,
		found,
		"protocol mapper has not been created",
	)
	err = client.DeleteClientProtocolMapper(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*(testClient.ID),
		id,
	)
	assert.NoError(t, err, "DeleteClientProtocolMapper failed")
	testClientAgain := GetClientByClientID(t, client, cfg.GoCloak.ClientID)

	found = false
	for _, protocolMapper := range testClientAgain.ProtocolMappers {
		if protocolMapper == nil || NilOrEmpty(protocolMapper.ID) {
			continue
		}
		if *(protocolMapper.ID) == id {
			found = true
			break
		}
	}
	assert.False(
		t,
		found,
		"default client should not have a protocol mapper with ID: %s", id,
	)
}

func TestGocloak_GetClientOfflineSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:      &(cfg.GoCloak.ClientID),
			ClientSecret:  &(cfg.GoCloak.ClientSecret),
			Username:      &(cfg.GoCloak.UserName),
			Password:      &(cfg.GoCloak.Password),
			GrantType:     StringP("password"),
			ResponseTypes: []string{"token", "id_token"},
			Scopes:        []string{"openid", "offline_access"},
		},
	)
	assert.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetClientOfflineSessions(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	assert.NoError(t, err, "GetClientOfflineSessions failed")
	assert.False(t, len(sessions) == 0, "GetClientOfflineSessions returned an empty list")
}

func TestGoCloak_ClientSecret(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	testClient := Client{
		ID:                      GetRandomNameP("gocloak-client-id-"),
		ClientID:                GetRandomNameP("gocloak-client-secret-client-id-"),
		Secret:                  StringP("initial-secret-key"),
		ServiceAccountsEnabled:  BoolP(true),
		StandardFlowEnabled:     BoolP(true),
		Enabled:                 BoolP(true),
		FullScopeAllowed:        BoolP(true),
		Protocol:                StringP("openid-connect"),
		RedirectURIs:            []string{"localhost"},
		ClientAuthenticatorType: StringP("client-secret"),
	}

	clientID, err := client.CreateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testClient,
	)
	assert.NoError(t, err, "CreateClient failed")
	assert.Equal(t, *(testClient.ID), clientID)

	oldCreds, err := client.GetClientSecret(
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
	)
	assert.NoError(t, err, "GetClientSecret failed")

	regeneratedCreds, err := client.RegenerateClientSecret(
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
	)
	assert.NoError(t, err, "RegenerateClientSecret failed")

	assert.NotEqual(t, *(oldCreds.Value), *(regeneratedCreds.Value))

	err = client.DeleteClient(token.AccessToken, cfg.GoCloak.Realm, clientID)
	assert.NoError(t, err, "DeleteClient failed")
}

func TestGoCloak_ClientServiceAccount(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	serviceAccount, err := client.GetClientServiceAccount(token.AccessToken, cfg.GoCloak.Realm, gocloakClientID)
	assert.NoError(t, err)

	assert.NotNil(t, serviceAccount.ID)
	assert.NotNil(t, serviceAccount.Username)
	assert.NotEqual(t, gocloakClientID, *(serviceAccount.ID))
	assert.Equal(t, "service-account-gocloak", *(serviceAccount.Username))
}

func TestGocloak_AddClientRoleToUser_DeleteClientRoleFromUser(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	tearDown1, roleName1 := CreateClientRole(t, client)
	defer tearDown1()
	token := GetAdminToken(t, client)
	role1, err := client.GetClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName1,
	)
	assert.NoError(t, err, "GetClientRole failed")
	tearDown2, roleName2 := CreateClientRole(t, client)
	defer tearDown2()
	role2, err := client.GetClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName2,
	)
	assert.NoError(t, err, "GetClientRole failed")
	roles := []Role{*role1, *role2}
	err = client.AddClientRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		testUserID,
		roles,
	)
	assert.NoError(t, err, "AddClientRoleToUser failed")

	err = client.DeleteClientRoleFromUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		testUserID,
		roles,
	)
	assert.NoError(t, err, "DeleteClientRoleFromUser failed")
}

func TestGocloak_AddClientRoleToGroup_DeleteClientRoleFromGroup(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	tearDown1, roleName1 := CreateClientRole(t, client)
	defer tearDown1()
	token := GetAdminToken(t, client)
	role1, err := client.GetClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName1,
	)
	assert.NoError(t, err, "GetClientRole failed")
	tearDown2, roleName2 := CreateClientRole(t, client)
	defer tearDown2()
	role2, err := client.GetClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName2,
	)
	assert.NoError(t, err, "GetClientRole failed")

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()

	roles := []Role{*role1, *role2}
	err = client.AddClientRoleToGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID,
		roles,
	)
	assert.NoError(t, err, "AddClientRoleToGroup failed")

	err = client.DeleteClientRoleFromGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID,
		roles,
	)
	assert.NoError(t, err, "DeleteClientRoleFromGroup failed")
}

func TestGocloak_AddDeleteClientRoleComposite(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, compositeRole := CreateClientRole(t, client)
	defer tearDown()

	tearDown, role := CreateClientRole(t, client)
	defer tearDown()

	compositeRoleModel, err := client.GetClientRole(token.AccessToken, cfg.GoCloak.Realm, gocloakClientID, compositeRole)
	assert.NoError(t, err, "Can't get just created role with GetClientRole")

	roleModel, err := client.GetClientRole(token.AccessToken, cfg.GoCloak.Realm, gocloakClientID, role)
	assert.NoError(t, err, "Can't get just created role with GetClientRole")

	err = client.AddClientRoleComposite(token.AccessToken,
		cfg.GoCloak.Realm, *(compositeRoleModel.ID), []Role{*roleModel})
	assert.NoError(t, err, "AddClientRoleComposite failed")

	err = client.DeleteClientRoleComposite(token.AccessToken,
		cfg.GoCloak.Realm, *(compositeRoleModel.ID), []Role{*roleModel})
	assert.NoError(t, err, "DeleteClientRoleComposite failed")
}

func TestGocloak_CreateDeleteClientScopeWithMappers(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	id := GetRandomName("client-scope-id-")
	rolemapperID := GetRandomName("client-rolemapper-id-")
	audiencemapperID := GetRandomName("client-audiencemapper-id-")

	createdID, err := client.CreateClientScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		ClientScope{
			ID:          &id,
			Name:        StringP("test-scope"),
			Description: StringP("testing scope"),
			Protocol:    StringP("openid-connect"),
			ClientScopeAttributes: &ClientScopeAttributes{
				ConsentScreenText:      StringP("false"),
				DisplayOnConsentScreen: StringP("true"),
				IncludeInTokenScope:    StringP("false"),
			},
			ProtocolMappers: []*ProtocolMappers{
				{
					ID:              &rolemapperID,
					Name:            StringP("roles"),
					Protocol:        StringP("openid-connect"),
					ProtocolMapper:  StringP("oidc-usermodel-client-role-mapper"),
					ConsentRequired: BoolP(false),
					ProtocolMappersConfig: &ProtocolMappersConfig{
						UserinfoTokenClaim:                 StringP("false"),
						AccessTokenClaim:                   StringP("true"),
						IDTokenClaim:                       StringP("true"),
						ClaimName:                          StringP("test"),
						Multivalued:                        StringP("true"),
						UsermodelClientRoleMappingClientID: StringP("test"),
					},
				},
				{
					ID:              &audiencemapperID,
					Name:            StringP("audience"),
					Protocol:        StringP("openid-connect"),
					ProtocolMapper:  StringP("oidc-audience-mapper"),
					ConsentRequired: BoolP(false),
					ProtocolMappersConfig: &ProtocolMappersConfig{
						UserinfoTokenClaim:     StringP("false"),
						IDTokenClaim:           StringP("true"),
						AccessTokenClaim:       StringP("true"),
						IncludedClientAudience: StringP("test"),
					},
				},
			},
		},
	)
	assert.NoError(t, err, "CreateClientScope failed")
	assert.Equal(t, id, createdID)
	clientScopeActual, err := client.GetClientScope(token.AccessToken, cfg.GoCloak.Realm, id)
	assert.NoError(t, err)

	assert.NotNil(t, clientScopeActual, "client scope has not been created")
	assert.Len(t, clientScopeActual.ProtocolMappers, 2, "unexpected number of protocol mappers created")
	err = client.DeleteClientScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		id,
	)
	assert.NoError(t, err, "DeleteClientScope failed")
	clientScopeActual, err = client.GetClientScope(token.AccessToken, cfg.GoCloak.Realm, id)
	assert.EqualError(t, err, "404 Not Found: Could not find client scope")
	assert.Nil(t, clientScopeActual, "client scope has not been deleted")
}

// -----------------
// identity provider
// -----------------

func TestGocloak_CreateProvider(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	t.Run("create google provider", func(t *testing.T) {
		repr := IdentityProviderRepresentation{
			Alias:                     StringP("google"),
			DisplayName:               StringP("Google"),
			Enabled:                   BoolP(true),
			ProviderID:                StringP("google"),
			TrustEmail:                BoolP(true),
			FirstBrokerLoginFlowAlias: StringP("first broker login"),
			Config: map[string]string{
				"clientId":     cfg.GoCloak.ClientID,
				"clientSecret": cfg.GoCloak.ClientSecret,
				"hostedDomain": "test.io",
			},
		}
		provider, err := client.CreateIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, repr)
		assert.NoError(t, err)
		assert.Equal(t, "google", provider)
	})

	t.Run("create azure provider", func(t *testing.T) {
		repr := IdentityProviderRepresentation{
			Alias:                     StringP("azure-oidc"),
			DisplayName:               StringP("Azure"),
			Enabled:                   BoolP(true),
			ProviderID:                StringP("oidc"),
			TrustEmail:                BoolP(true),
			FirstBrokerLoginFlowAlias: StringP("first broker login"),
			Config: map[string]string{
				"clientId":         cfg.GoCloak.ClientID,
				"clientSecret":     cfg.GoCloak.ClientSecret,
				"authorizationUrl": "authorization-url",
				"tokenUrl":         "token-url",
				"logoutUrl":        "logout-url",
				"userInfoUrl":      "userinfo-url",
				"issuer":           "test-issuer",
				"jwksUrl":          "jwks-url",
			},
		}
		provider, err := client.CreateIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, repr)
		assert.NoError(t, err)
		assert.Equal(t, "azure-oidc", provider)
	})

	t.Run("create OIDC V1.0 provider", func(t *testing.T) {
		repr := IdentityProviderRepresentation{
			Alias:                     StringP("oidc"),
			DisplayName:               StringP("custom-oidc"),
			Enabled:                   BoolP(true),
			ProviderID:                StringP("oidc"),
			TrustEmail:                BoolP(true),
			FirstBrokerLoginFlowAlias: StringP("first broker login"),
			Config: map[string]string{
				"clientId":                 cfg.GoCloak.ClientID,
				"clientSecret":             cfg.GoCloak.ClientSecret,
				"authorizationUrl":         "authorization-url",
				"tokenUrl":                 "token-url",
				"logoutUrl":                "logout-url",
				"userInfoUrl":              "userinfo-url",
				"issuer":                   "test-issuer",
				"loginHint":                "true",
				"validateSignature":        "true",
				"backchannelLogout":        "false",
				"useJwksUrl":               "true",
				"uiLocales":                "true",
				"disableUserInfo":          "true",
				"defaultScopes":            "default-scope",
				"prompt":                   "false",
				"allowedClockSkew":         "10",
				"forwardedQueryParameters": "forwarded-query-parameters",
			},
		}
		provider, err := client.CreateIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, repr)
		assert.NoError(t, err)
		assert.Equal(t, "oidc", provider)
	})

	t.Run("Update google provider", func(t *testing.T) {
		repr := IdentityProviderRepresentation{
			Alias:                     StringP("google"),
			DisplayName:               StringP("Google"),
			Enabled:                   BoolP(true),
			ProviderID:                StringP("google"),
			TrustEmail:                BoolP(true),
			FirstBrokerLoginFlowAlias: StringP("first broker login"),
			Config: map[string]string{
				"clientId":     cfg.GoCloak.ClientID,
				"clientSecret": cfg.GoCloak.ClientSecret,
				"hostedDomain": "updated-test.io",
			},
		}
		err := client.UpdateIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, "google", repr)
		assert.NoError(t, err)

		// listing identity providers here must now show three
		providers, err := client.GetIdentityProviders(token.AccessToken, cfg.GoCloak.Realm)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(providers))
	})

	t.Run("Delete google provider", func(t *testing.T) {
		err := client.DeleteIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, "google")
		assert.NoError(t, err)
	})

	t.Run("List providers", func(t *testing.T) {
		providers, err := client.GetIdentityProviders(token.AccessToken, cfg.GoCloak.Realm)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(providers))
	})

	t.Run("Get Azure provider", func(t *testing.T) {
		provider, err := client.GetIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, "azure-oidc")
		assert.NoError(t, err)
		assert.Equal(t, "azure-oidc", *(provider.Alias))
	})

	t.Run("Delete Azure provider", func(t *testing.T) {
		err := client.DeleteIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, "azure-oidc")
		assert.NoError(t, err)
	})

	t.Run("Delete OIDC V1.0 provider", func(t *testing.T) {
		err := client.DeleteIdentityProvider(token.AccessToken, cfg.GoCloak.Realm, "oidc")
		assert.NoError(t, err)
	})
}

// -----------------
// Protection API
// -----------------

func TestGocloak_CreateListGetUpdateDeleteResource(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Create
	tearDown, resourceID := CreateResource(t, client, gocloakClientID)
	// Delete
	defer tearDown()

	// List
	createdResource, err := client.GetResource(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		resourceID,
	)

	assert.NoError(t, err, "GetResource failed")
	t.Logf("Created Resource: %+v", *(createdResource.ID))
	assert.Equal(t, resourceID, *(createdResource.ID))

	// Looking for a created resource
	resources, err := client.GetResources(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		GetResourceParams{
			Name: createdResource.Name,
		},
	)
	assert.NoError(t, err, "GetResources failed")
	assert.Len(t, resources, 1, "GetResources should return exact 1 resource")
	assert.Equal(t, *(createdResource.ID), *(resources[0].ID))
	t.Logf("Resources: %+v", resources)

	err = client.UpdateResource(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		ResourceRepresentation{},
	)
	assert.Error(t, err, "Should fail because of missing ID of the resource")

	createdResource.Name = GetRandomNameP("ResourceName")
	err = client.UpdateResource(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdResource,
	)
	assert.NoError(t, err, "UpdateResource failed")

	updatedResource, err := client.GetResource(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		resourceID,
	)
	assert.NoError(t, err, "GetResource failed")
	assert.Equal(t, *(createdResource.Name), *(updatedResource.Name))
}

func TestGocloak_CreateListGetUpdateDeleteScope(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Create
	tearDown, scopeID := CreateScope(t, client, gocloakClientID)
	// Delete
	defer tearDown()

	// List
	createdScope, err := client.GetScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	assert.NoError(t, err, "GetScope failed")
	t.Logf("Created Scope: %+v", *(createdScope.ID))
	assert.Equal(t, scopeID, *(createdScope.ID))

	// Looking for a created scope
	scopes, err := client.GetScopes(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		GetScopeParams{
			Name: createdScope.Name,
		},
	)
	assert.NoError(t, err, "GetScopes failed")
	assert.Len(t, scopes, 1, "GetScopes should return exact 1 scope")
	assert.Equal(t, *(createdScope.ID), *(scopes[0].ID))
	t.Logf("Scopes: %+v", scopes)

	err = client.UpdateScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		ScopeRepresentation{},
	)
	assert.Error(t, err, "Should fail because of missing ID of the scope")

	createdScope.Name = GetRandomNameP("ScopeName")
	err = client.UpdateScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdScope,
	)
	assert.NoError(t, err, "UpdateScope failed")

	updatedScope, err := client.GetScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	assert.NoError(t, err, "GetScope failed")
	assert.Equal(t, *(createdScope.Name), *(updatedScope.Name))
}

func TestGocloak_CreateListGetUpdateDeletePolicy(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Create
	tearDown, policyID := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("Policy Description"),
		Type:        StringP("js"),
		Logic:       NEGATIVE,
		JSPolicyRepresentation: JSPolicyRepresentation{
			Code: StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDown()

	// List
	createdPolicy, err := client.GetPolicy(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		policyID,
	)
	assert.NoError(t, err, "GetPolicy failed")
	t.Logf("Created Policy: %+v", *(createdPolicy.ID))
	assert.Equal(t, policyID, *(createdPolicy.ID))

	// Looking for a created policy
	policies, err := client.GetPolicies(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		GetPolicyParams{
			Name: createdPolicy.Name,
		},
	)
	assert.NoError(t, err, "GetPolicies failed")
	assert.Len(t, policies, 1, "GetPolicies should return exact 1 policy")
	assert.Equal(t, *(createdPolicy.ID), *(policies[0].ID))
	t.Logf("Policies: %+v", policies)

	err = client.UpdatePolicy(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		PolicyRepresentation{},
	)
	assert.Error(t, err, "Should fail because of missing ID of the policy")

	createdPolicy.Name = GetRandomNameP("PolicyName")
	err = client.UpdatePolicy(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdPolicy,
	)
	assert.NoError(t, err, "UpdatePolicy failed")

	updatedPolicy, err := client.GetPolicy(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		policyID,
	)
	assert.NoError(t, err, "GetPolicy failed")
	assert.Equal(t, *(createdPolicy.Name), *(updatedPolicy.Name))
}

func TestGocloak_RolePolicy(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	roles, err := client.GetRealmRoles(token.AccessToken, cfg.GoCloak.Realm)
	assert.NoError(t, err, "GetRealmRoles failed")
	assert.GreaterOrEqual(t, len(roles), 1, "GetRealmRoles failed")

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("Role Policy"),
		Type:        StringP("role"),
		Logic:       NEGATIVE,
		RolePolicyRepresentation: RolePolicyRepresentation{
			Roles: []*RoleDefinition{
				{
					ID: roles[0].ID,
				},
			},
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_JSPolicy(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("JS Policy"),
		Type:        StringP("js"),
		Logic:       POSITIVE,
		JSPolicyRepresentation: JSPolicyRepresentation{
			Code: StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_ClientPolicy(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("Client Policy"),
		Type:        StringP("client"),
		ClientPolicyRepresentation: ClientPolicyRepresentation{
			Clients: []string{
				gocloakClientID,
			},
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_TimePolicy(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("Time Policy"),
		Type:        StringP("time"),
		TimePolicyRepresentation: TimePolicyRepresentation{
			NotBefore:    StringP("2019-12-30 12:00:00"),
			NotOnOrAfter: StringP("2020-12-30 12:00:00"),
			DayMonth:     StringP("1"),
			DayMonthEnd:  StringP("31"),
			Month:        StringP("1"),
			MonthEnd:     StringP("12"),
			Year:         StringP("1900"),
			YearEnd:      StringP("2100"),
			Hour:         StringP("1"),
			HourEnd:      StringP("24"),
			Minute:       StringP("0"),
			MinuteEnd:    StringP("60"),
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_UserPolicy(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("User Policy"),
		Type:        StringP("user"),
		UserPolicyRepresentation: UserPolicyRepresentation{
			Users: []string{
				userID,
			},
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_AggregatedPolicy(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	tearDownClient, clientPolicyID := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("Client Policy"),
		Type:        StringP("client"),
		ClientPolicyRepresentation: ClientPolicyRepresentation{
			Clients: []string{
				gocloakClientID,
			},
		},
	})
	defer tearDownClient()

	tearDownJS, jsPolicyID := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("JS Policy"),
		Type:        StringP("js"),
		Logic:       POSITIVE,
		JSPolicyRepresentation: JSPolicyRepresentation{
			Code: StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDownJS()

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("Aggregated Policy"),
		Type:        StringP("aggregate"),
		AggregatedPolicyRepresentation: AggregatedPolicyRepresentation{
			Policies: []string{
				clientPolicyID,
				jsPolicyID,
			},
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_GroupPolicy(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("Group Policy"),
		Type:        StringP("group"),
		GroupPolicyRepresentation: GroupPolicyRepresentation{
			Groups: []*GroupDefinition{
				{
					ID: StringP(groupID),
				},
			},
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_CreateListGetUpdateDeletePermission(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Create
	tearDownResource, resourceID := CreateResource(t, client, gocloakClientID)
	// Delete
	defer tearDownResource()

	tearDownPolicy, policyID := CreatePolicy(t, client, gocloakClientID, PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: StringP("JS Policy"),
		Type:        StringP("js"),
		Logic:       POSITIVE,
		JSPolicyRepresentation: JSPolicyRepresentation{
			Code: StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDownPolicy()

	// Create
	tearDown, permissionID := CreatePermission(t, client, gocloakClientID, PermissionRepresentation{
		Name:        GetRandomNameP("PermissionName"),
		Description: StringP("Permission Description"),
		Type:        StringP("resource"),
		Policies: []string{
			policyID,
		},
		Resources: []string{
			resourceID,
		},
	})
	// Delete
	defer tearDown()

	// List
	createdPermission, err := client.GetPermission(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		permissionID,
	)
	assert.NoError(t, err, "GetPermission failed")
	t.Logf("Created Permission: %+v", *(createdPermission.ID))
	assert.Equal(t, permissionID, *(createdPermission.ID))

	// Looking for a created permission
	permissions, err := client.GetPermissions(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		GetPermissionParams{
			Name: createdPermission.Name,
		},
	)
	assert.NoError(t, err, "GetPermissions failed")
	assert.Len(t, permissions, 1, "GetPermissions should return exact 1 permission")
	assert.Equal(t, *(createdPermission.ID), *(permissions[0].ID))
	t.Logf("Permissions: %+v", permissions)

	err = client.UpdatePermission(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		PermissionRepresentation{},
	)
	assert.Error(t, err, "Should fail because of missing ID of the permission")

	createdPermission.Name = GetRandomNameP("PermissionName")
	err = client.UpdatePermission(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdPermission,
	)
	assert.NoError(t, err, "UpdatePermission failed")

	updatedPermission, err := client.GetPermission(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		permissionID,
	)
	assert.NoError(t, err, "GetPermission failed")
	assert.Equal(t, *(createdPermission.Name), *(updatedPermission.Name))
}
