package gocloak

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/resty.v1"
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

func FailIfErr(t *testing.T, err error, msg string) {
	_, objectAlreadyExists := err.(*ObjectAlreadyExists)
	if objectAlreadyExists {
		return
	}

	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		if len(msg) == 0 {
			msg = "unexpected error"
		}
		t.Fatalf("%s:%d: %s: %s", filepath.Base(file), line, msg, err.Error())
	}
}

func FailIf(t *testing.T, cond bool, msg string, args ...interface{}) {
	if cond {
		if len(args) > 0 {
			t.Fatalf(msg, args...)
		} else {
			t.Fatal(msg)
		}
	}
}

func AssertEquals(t *testing.T, exp interface{}, act interface{}) {
	FailIf(
		t,
		!reflect.DeepEqual(exp, act),
		"The expected and actual results are not equal.\nExpected: %+v.\nActual:   %+v", exp, act)
}

func GetConfig(t *testing.T) *Config {
	configOnce.Do(func() {
		configFileName, ok := os.LookupEnv("GOCLOAK_TEST_CONFIG")
		if !ok {
			configFileName = filepath.Join("testdata", "config.json")
		}
		configFile, err := os.Open(configFileName)
		FailIfErr(t, err, "cannot open config.json")
		defer func() {
			err := configFile.Close()
			FailIfErr(t, err, "cannot close config file")
		}()
		data, err := ioutil.ReadAll(configFile)
		FailIfErr(t, err, "cannot read config.json")
		config = &Config{}
		err = json.Unmarshal(data, config)
		FailIfErr(t, err, "cannot parse config.json")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		if len(config.Proxy) != 0 {
			proxy, err := url.Parse(config.Proxy)
			FailIfErr(t, err, "incorrect proxy url: "+config.Proxy)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(proxy)
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
	FailIfErr(t, err, "Login failed")
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
	FailIfErr(t, err, "Login failed")
	return token
}

func GetAdminToken(t *testing.T, client GoCloak) *JWT {
	cfg := GetConfig(t)
	token, err := client.LoginAdmin(
		cfg.Admin.UserName,
		cfg.Admin.Password,
		cfg.Admin.Realm)
	FailIfErr(t, err, "Login failed")
	return token
}

func GetRandomName(name string) string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	randomNumber := r1.Intn(100000)
	return name + strconv.Itoa(randomNumber)
}

func GetClientByClientID(t *testing.T, client GoCloak, clientID string) *Client {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	clients, err := client.GetClients(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetClientsParams{
			ClientID: clientID,
		})
	FailIfErr(t, err, "GetClients failed")
	for _, fetchedClient := range *clients {
		if fetchedClient.ClientID == clientID {
			return &fetchedClient
		}
	}
	t.Fatal("Client not found")
	return nil
}

func CreateGroup(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	group := Group{
		Name: GetRandomName("GroupName"),
	}
	err := client.CreateGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		group)
	FailIfErr(t, err, "CreateGroup failed")
	groups, err := client.GetGroups(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetGroupsParams{
			Search: group.Name,
		})
	FailIfErr(t, err, "GetGroups failed")
	var groupID string
	for _, fetchedGroup := range *groups {
		if fetchedGroup.Name == group.Name {
			groupID = fetchedGroup.ID
			break
		}
	}
	t.Logf("Created Group with ID: %s. Group: %+v", groupID, group)

	tearDown := func() {
		err := client.DeleteGroup(
			token.AccessToken,
			cfg.GoCloak.Realm,
			groupID)
		FailIfErr(t, err, "DeleteGroup failed")
	}
	return tearDown, groupID
}

func SetUpTestUser(t *testing.T, client GoCloak) {
	setupOnce.Do(func() {
		cfg := GetConfig(t)
		token := GetAdminToken(t, client)

		user := User{
			Username:      cfg.GoCloak.UserName,
			Email:         cfg.GoCloak.UserName + "@localhost",
			EmailVerified: true,
			Enabled:       true,
		}

		createdUserID, err := client.CreateUser(
			token.AccessToken,
			cfg.GoCloak.Realm,
			user)
		if err != nil && err.Error() == "Conflict: Object already exists" {
			err = nil
			users, err := client.GetUsers(
				token.AccessToken,
				cfg.GoCloak.Realm,
				GetUsersParams{
					Username: cfg.GoCloak.UserName,
				})
			FailIfErr(t, err, "GetUsers failed")
			for _, user := range *users {
				if user.Username == cfg.GoCloak.UserName {
					testUserID = user.ID
					break
				}
			}
		} else {
			FailIfErr(t, err, "CreateUser failed")
			testUserID = *createdUserID
		}

		err = client.SetPassword(
			token.AccessToken,
			testUserID,
			cfg.GoCloak.Realm,
			cfg.GoCloak.Password,
			false)
		//client.GetRealmRole(token.AccessToken, cfg.GoCloak.Realm, "")
		FailIfErr(t, err, "SetPassword failed")
	})
}

type RestyLogWriter struct {
	io.Writer
	t *testing.T
}

func (w *RestyLogWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

func NewClientWithDebug(t *testing.T) GoCloak {
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	restyClient := client.RestyClient()
	restyClient.SetDebug(true)
	restyClient.SetLogger(&RestyLogWriter{
		t: t,
	})

	return client
}

func TestGetQueryParams(t *testing.T) {
	t.Parallel()

	type TestParams struct {
		IntField    int    `json:"int_field,string,omitempty"`
		StringField string `json:"string_field,omitempty"`
		BoolField   bool   `json:"bool_field,string,omitempty"`
	}

	params, err := GetQueryParams(TestParams{})
	FailIfErr(t, err, "GetQueryParams failed")
	FailIf(
		t,
		len(params) > 0,
		"Params must be empty, but got: %+v", params)

	params, err = GetQueryParams(TestParams{
		IntField:    1,
		StringField: "fake",
		BoolField:   true,
	})
	FailIfErr(t, err, "GetQueryParams failed")
	AssertEquals(t, map[string]string{
		"int_field":    "1",
		"string_field": "fake",
		"bool_field":   "true",
	}, params)
}

func TestGocloak_RestyClient(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	restyClient := client.RestyClient()
	FailIf(
		t,
		restyClient == resty.DefaultClient,
		"Resty client of the GoCloak client and the Default resty client are equal",
	)
}

// ---------
// API tests
// ---------

func TestGocloak_GetUserInfo(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	userInfo, err := client.GetUserInfo(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Log(userInfo)
	FailIfErr(t, err, "Failed to fetch userinfo")
}

func TestGocloak_RequestPermission(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	token, err := client.RequestPermission(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password,
		"Permission foo # 3")
	FailIfErr(t, err, "login failed")

	rptResult, err := client.RetrospectToken(
		token.AccessToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(rptResult)
	FailIfErr(t, err, "inspection failed")
	FailIf(t, !rptResult.Active, "Inactive Token oO")
}

func TestGocloak_GetCerts(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	certs, err := client.GetCerts(cfg.GoCloak.Realm)
	t.Log(certs)
	FailIfErr(t, err, "get certs")
}

func TestGocloak_LoginClient_UnknownRealm(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginClient(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		"ThisRealmDoesNotExist")
	FailIf(t, err == nil, "Login shouldn't be successful")

	errorMessage := err.Error()
	FailIf(t, errorMessage != "404 Not Found", "Unexpected error message: "+errorMessage)
}

func TestGocloak_GetIssuer(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	issuer, err := client.GetIssuer(cfg.GoCloak.Realm)
	t.Log(issuer)
	FailIfErr(t, err, "get issuer")
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
	FailIfErr(t, err, "inspection failed")
	FailIf(t, rptResult.Active, "That should never happen. Token is active")

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
	FailIfErr(t, err, "Inspection failed")
	FailIf(t, !rptResult.Active, "Inactive Token oO")
}

func TestGocloak_DecodeAccessToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	resultToken, claims, err := client.DecodeAccessToken(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Log(resultToken)
	t.Log(claims)
	FailIfErr(t, err, "DecodeAccessToken")
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
		claims)
	t.Log(resultToken)
	t.Log(claims)
	FailIfErr(t, err, "DecodeAccessTokenCustomClaims")
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
	FailIfErr(t, err, "RefreshToken failed")
}

func TestGocloak_UserAttributeContains(t *testing.T) {
	t.Parallel()

	attributes := map[string][]string{}
	attributes["foo"] = []string{"bar", "alice", "bob", "roflcopter"}
	attributes["bar"] = []string{"baz"}

	client := NewClientWithDebug(t)
	ok := client.UserAttributeContains(attributes, "foo", "alice")
	FailIf(t, !ok, "UserAttributeContains")
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
	FailIfErr(t, err, "GetKeyStoreConfig")
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
	FailIfErr(t, err, "Login failed")
}

func TestGocloak_GetToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	newToken, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:      cfg.GoCloak.ClientID,
			ClientSecret:  cfg.GoCloak.ClientSecret,
			Username:      cfg.GoCloak.UserName,
			Password:      cfg.GoCloak.Password,
			GrantType:     "password",
			ResponseTypes: []string{"token", "id_token"},
			Scopes:        []string{"openid", "offline_access"},
		},
	)
	FailIfErr(t, err, "Login failed")
	t.Logf("New token: %+v", *newToken)
	FailIf(t, newToken.RefreshExpiresIn > 0, "Got a refresh token instead of offline")
	FailIf(t, len(newToken.IDToken) == 0, "Got an empty if token")
}

func TestGocloak_LoginClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginClient(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	FailIfErr(t, err, "LoginClient failed")
}

func TestGocloak_LoginAdmin(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginAdmin(
		cfg.Admin.UserName,
		cfg.Admin.Password,
		cfg.Admin.Realm)
	FailIfErr(t, err, "LoginAdmin failed")
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
	FailIfErr(t, err, "Failed to set password")
}

func TestGocloak_CreateListGetUpdateDeleteGroup(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Create, List
	tearDown, groupID := CreateGroup(t, client)

	createdGroup, err := client.GetGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	FailIfErr(t, err, "GetGroup failed")
	t.Logf("Created Group: %+v", createdGroup)
	AssertEquals(t, groupID, createdGroup.ID)

	err = client.UpdateGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Group{},
	)
	FailIf(t, err == nil, "Should fail because of missing ID of the group")

	createdGroup.Name = GetRandomName("GroupName")
	err = client.UpdateGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*createdGroup,
	)
	FailIfErr(t, err, "UpdateGroup failed")

	updatedGroup, err := client.GetGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	FailIfErr(t, err, "GetGroup failed")
	AssertEquals(t, createdGroup.Name, updatedGroup.Name)

	// Delete
	defer tearDown()
}

func TestGocloak_CreateClientRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	role := Role{
		Name: GetRandomName("mySuperCoolRole"),
	}
	testClient := GetClientByClientID(t, client, cfg.GoCloak.ClientID)

	err := client.CreateClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testClient.ID,
		role)
	FailIfErr(t, err, "CreateClientRole failed")
	err = client.DeleteClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testClient.ID,
		role.Name,
	)
	FailIfErr(t, err, "DeleteClientRole failed")
}

func TestGocloak_CreateClientScope(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	newScope := ClientScope{
		ID:   "foo",
		Name: "bar",
	}
	err := client.CreateClientScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		newScope)
	FailIfErr(t, err, "CreateClientScope failed")

	err = client.DeleteClientScope(
		token.AccessToken,
		cfg.GoCloak.Realm,
		newScope.ID)

	FailIfErr(t, err, "DeleteClientScope failed")
}

func TestGocloak_CreateListGetUpdateDeleteClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	clientID := GetRandomName("ClientID")
	t.Logf("Client ID: %s", clientID)

	// Creating a client
	err := client.CreateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Client{
			ClientID: clientID,
			Name:     GetRandomName("Name"),
		},
	)
	FailIfErr(t, err, "CreateClient failed")

	// Looking for a created client
	clients, err := client.GetClients(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetClientsParams{
			ClientID: clientID,
		},
	)
	FailIfErr(t, err, "CreateClients failed")
	FailIf(t, len(*clients) != 1, "GetClients should return exact 1 client")
	t.Logf("Clients: %+v", *clients)

	// Getting exact client
	createdClient, err := client.GetClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		(*clients)[0].ID,
	)
	FailIfErr(t, err, "GetClient failed")
	t.Logf("Created client: %+v", createdClient)
	// Checking that GetClient returns same client
	AssertEquals(t, (*clients)[0], *createdClient)

	// Updating the client

	// Should fail
	err = client.UpdateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Client{},
	)
	FailIf(t, err == nil, "Should fail because of missing ID of the client")

	// Update existing client
	createdClient.Name = GetRandomName("Name")
	err = client.UpdateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*createdClient,
	)
	FailIfErr(t, err, "GetClient failed")

	// Getting updated client
	updatedClient, err := client.GetClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		(*clients)[0].ID,
	)
	FailIfErr(t, err, "GetClient failed")
	t.Logf("Update client: %+v", createdClient)
	AssertEquals(t, *createdClient, *updatedClient)

	// Deleting the client
	err = client.DeleteClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		createdClient.ID,
	)
	FailIfErr(t, err, "DeleteClient failed")

	// Verifying that the client was deleted
	clients, err = client.GetClients(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetClientsParams{
			ClientID: clientID,
		},
	)
	FailIfErr(t, err, "CreateClients failed")
	FailIf(t, len(*clients) != 0, "GetClients should not return any clients")

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
	FailIfErr(t, err, "GetGroups failed")
}

func TestGocloak_GetClients(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	_, err := client.GetClients(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetClientsParams{
			ClientID: cfg.GoCloak.ClientID,
		})
	FailIfErr(t, err, "GetClients failed")
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
		testClient.ID)
	FailIfErr(t, err, "GetClientRoles failed")
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
	FailIfErr(t, err, "GetRoleMappingByGroupID failed")
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
	FailIfErr(t, err, "GetRoleMappingByUserID failed")
}

func TestGocloak_ExecuteActionsEmail_UpdatePassword(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	params := ExecuteActionsEmail{
		ClientID: cfg.GoCloak.ClientID,
		UserID:   userID,
		Actions:  []string{"UPDATE_PASSWORD"},
	}

	err := client.ExecuteActionsEmail(
		token.AccessToken,
		cfg.GoCloak.Realm,
		params)

	if err != nil {
		if err.Error() == "500 Internal Server Error" {
			return
		}
		FailIfErr(t, err, "ExecuteActionsEmail failed")
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
	FailIfErr(t, err, "Logout failed")
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
	FailIfErr(t, err, "GetRealm failed")
}

// -----------
// Realm
// -----------

func CreateRealm(t *testing.T, client GoCloak) (func(), string) {
	token := GetAdminToken(t, client)

	realmName := GetRandomName("Realm")
	t.Logf("Creating Realm: %s", realmName)
	err := client.CreateRealm(
		token.AccessToken,
		RealmRepresentation{
			Realm: realmName,
		})
	FailIfErr(t, err, "CreateRealm failed")
	tearDown := func() {
		err := client.DeleteRealm(
			token.AccessToken,
			realmName)
		FailIfErr(t, err, "DeleteRealm failed")
	}
	return tearDown, realmName
}

func TestGocloak_CreateRealm(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	tearDown, _ := CreateRealm(t, client)
	defer tearDown()
}

// -----------
// Realm Roles
// -----------

func CreateRealmRole(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	roleName := GetRandomName("Role")
	t.Logf("Creating RoleName: %s", roleName)
	err := client.CreateRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		Role{
			Name:        roleName,
			ContainerID: "asd",
		})
	FailIfErr(t, err, "CreateRealmRole failed")
	tearDown := func() {
		err := client.DeleteRealmRole(
			token.AccessToken,
			cfg.GoCloak.Realm,
			roleName)
		FailIfErr(t, err, "DeleteRealmRole failed")
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
	FailIfErr(t, err, "GetRealmRole failed")
	t.Logf("Role: %+v", *role)
	FailIf(
		t,
		role.Name != roleName,
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
	FailIfErr(t, err, "GetRealmRoles failed")
	t.Logf("Roles: %+v", *roles)
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
			Name: newRoleName,
		})
	FailIfErr(t, err, "UpdateRealmRole failed")
	err = client.DeleteRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		oldRoleName)
	FailIf(
		t,
		err == nil,
		"Role with old name was deleted successfully, but it shouldn't. Old role: %s; Updated role: %s",
		oldRoleName, newRoleName)
	err = client.DeleteRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		newRoleName)
	FailIfErr(t, err, "DeleteRealmRole failed")
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
	FailIfErr(t, err, "DeleteRealmRole failed")
}

func TestGocloak_AddRealmRoleToUser(t *testing.T) {
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
	FailIfErr(t, err, "GetRealmRole failed")

	err = client.AddRealmRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]Role{
			*role,
		})
	FailIfErr(t, err, "AddRealmRoleToUser failed")
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
	FailIfErr(t, err, "GetRealmRole failed")

	err = client.AddRealmRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]Role{
			*role,
		})
	FailIfErr(t, err, "AddRealmRoleToUser failed")

	roles, err := client.GetRealmRolesByUserID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	FailIfErr(t, err, "GetRealmRolesByUserID failed")
	t.Logf("User roles: %+v", *roles)
	for _, r := range *roles {
		if r.Name == role.Name {
			return
		}
	}
	t.Fatalf("The role has not been found in the assined roles. Role: %+v", *role)
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
	FailIfErr(t, err, "GetRealmRolesByGroupID failed")
}

func TestGocloak_DeleteRealmRoleFromUser(t *testing.T) {
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
	FailIfErr(t, err, "GetRealmRole failed")

	err = client.AddRealmRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]Role{
			*role,
		})
	FailIfErr(t, err, "AddRealmRoleToUser failed")
	err = client.DeleteRealmRoleFromUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]Role{
			*role,
		})
	FailIfErr(t, err, "DeleteRealmRoleFromUser failed")

	roles, err := client.GetRealmRolesByUserID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	FailIfErr(t, err, "GetRealmRolesByUserID failed")
	for _, r := range *roles {
		FailIf(
			t,
			r.Name == role.Name,
			"The role has been found in asigned roles. Role: %+v", role)
	}
}

func TestGocloak_AddRealmRoleComposite(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, compositeRole := CreateRealmRole(t, client)
	defer tearDown()

	tearDown, role := CreateRealmRole(t, client)
	defer tearDown()

	roleModel, err := client.GetRealmRole(token.AccessToken, cfg.GoCloak.Realm, role)
	FailIfErr(t, err, "Can't get just created role with GetRealmRole")

	err = client.AddRealmRoleComposite(token.AccessToken,
		cfg.GoCloak.Realm, compositeRole, []Role{*roleModel})
	FailIfErr(t, err, "AddRealmRoleComposite failed")
}

func TestGocloak_DeleteRealmRoleComposite(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, compositeRole := CreateRealmRole(t, client)
	defer tearDown()

	tearDown, role := CreateRealmRole(t, client)
	defer tearDown()

	roleModel, err := client.GetRealmRole(token.AccessToken, cfg.GoCloak.Realm, role)
	FailIfErr(t, err, "Can't get just created role with GetRealmRole")

	err = client.AddRealmRoleComposite(token.AccessToken,
		cfg.GoCloak.Realm, compositeRole, []Role{*roleModel})
	FailIfErr(t, err, "AddRealmRoleComposite failed")

	err = client.DeleteRealmRoleComposite(token.AccessToken,
		cfg.GoCloak.Realm, compositeRole, []Role{*roleModel})
	FailIfErr(t, err, "DeleteRealmRoleComposite failed")
}

// -----
// Users
// -----

func CreateUser(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	user := User{
		FirstName: GetRandomName("FirstName"),
		LastName:  GetRandomName("LastName"),
		Email:     GetRandomName("email") + "@localhost",
		Enabled:   true,
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
	FailIfErr(t, err, "CreateUser failed")
	user.ID = *userID
	t.Logf("Created User: %+v", user)
	tearDown := func() {
		err := client.DeleteUser(
			token.AccessToken,
			cfg.GoCloak.Realm,
			user.ID)
		FailIfErr(t, err, "DeleteUser")
	}

	return tearDown, user.ID
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
	FailIfErr(t, err, "GetUserByID failed")
	ok := client.UserAttributeContains(fetchedUser.Attributes, "foo", "alice")
	FailIf(t, !ok, "User doesn't have custom attributes")
	ok = client.UserAttributeContains(fetchedUser.Attributes, "foo2", "alice")
	FailIf(t, ok, "User's custom attributes contains unexpected attribute")
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
	FailIfErr(t, err, "GetUserById failed")
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
			Username: cfg.GoCloak.UserName,
		})
	FailIfErr(t, err, "GetUsers failed")
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
	FailIfErr(t, err, "GetUserCount failed")
	t.Logf("Users in Realm: %d", count)
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
	FailIfErr(t, err, "AddUserToGroup failed")
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
	FailIfErr(t, err, "AddUserToGroup failed")
	err = client.DeleteUserFromGroup(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	FailIfErr(t, err, "DeleteUserFromGroup failed")
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
	FailIfErr(t, err, "AddUserToGroup failed")
	groups, err := client.GetUserGroups(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	FailIfErr(t, err, "GetUserGroups failed")
	FailIf(
		t,
		len(*groups) == 0,
		"User is not in the Group")
	AssertEquals(
		t,
		groupID,
		(*groups)[0].ID)
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
	FailIfErr(t, err, "GetUserByID failed")
	user.FirstName = GetRandomName("UpdateUserFirstName")
	err = client.UpdateUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		*user)
	FailIfErr(t, err, "UpdateUser failed")
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
	FailIfErr(t, err, "GetRealmRole failed")
	err = client.AddRealmRoleToUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]Role{
			*role,
		})
	FailIfErr(t, err, "AddRealmRoleToUser failed")

	users, err := client.GetUsersByRoleName(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	FailIfErr(t, err, "GetUsersByRoleName failed")

	FailIf(
		t,
		len(*users) == 0,
		"User is not in the Group")
	AssertEquals(
		t,
		userID,
		(*users)[0].ID)
}

func TestGocloak_GetUserSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:     cfg.GoCloak.ClientID,
			ClientSecret: cfg.GoCloak.ClientSecret,
			Username:     cfg.GoCloak.UserName,
			Password:     cfg.GoCloak.Password,
			GrantType:    "password",
		},
	)
	FailIfErr(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetUserSessions(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
	)
	FailIfErr(t, err, "GetUserSessions failed")
	FailIf(t, len(*sessions) == 0, "GetUserSessions returned an empty list")
}

func TestGocloak_GetUserOfflineSessionsForClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:      cfg.GoCloak.ClientID,
			ClientSecret:  cfg.GoCloak.ClientSecret,
			Username:      cfg.GoCloak.UserName,
			Password:      cfg.GoCloak.Password,
			GrantType:     "password",
			ResponseTypes: []string{"token", "id_token"},
			Scopes:        []string{"openid", "offline_access"},
		},
	)
	FailIfErr(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetUserOfflineSessionsForClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
		gocloakClientID,
	)
	FailIfErr(t, err, "GetUserOfflineSessionsForClient failed")
	FailIf(t, len(*sessions) == 0, "GetUserOfflineSessionsForClient returned an empty list")
}

func TestGocloak_GetClientUserSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:     cfg.GoCloak.ClientID,
			ClientSecret: cfg.GoCloak.ClientSecret,
			Username:     cfg.GoCloak.UserName,
			Password:     cfg.GoCloak.Password,
			GrantType:    "password",
		},
	)
	FailIfErr(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetClientUserSessions(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	FailIfErr(t, err, "GetClientUserSessions failed")
	FailIf(t, len(*sessions) == 0, "GetClientUserSessions returned an empty list")
}

func TestGocloak_GetClientOfflineSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		cfg.GoCloak.Realm,
		TokenOptions{
			ClientID:      cfg.GoCloak.ClientID,
			ClientSecret:  cfg.GoCloak.ClientSecret,
			Username:      cfg.GoCloak.UserName,
			Password:      cfg.GoCloak.Password,
			GrantType:     "password",
			ResponseTypes: []string{"token", "id_token"},
			Scopes:        []string{"openid", "offline_access"},
		},
	)
	FailIfErr(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetClientOfflineSessions(
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	FailIfErr(t, err, "GetClientOfflineSessions failed")
	FailIf(t, len(*sessions) == 0, "GetClientOfflineSessions returned an empty list")
}
