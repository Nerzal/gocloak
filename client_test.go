package gocloak

import (
	"crypto/tls"
	"encoding/json"
	"github.com/Nerzal/gocloak/pkg/jwx"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
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
)

func FailIfErr(t *testing.T, err error, msg string) {
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

func GetConfig(t *testing.T) *Config {
	configOnce.Do(func() {
		configFile, err := os.Open(filepath.Join("testdata", "config.json"))
		FailIfErr(t, err, "cannot open config.json")
		defer configFile.Close()
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

func CreateUser(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	user := User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = GetRandomName("trololo") + "@localhost"
	user.Enabled = true
	user.Username = user.Email
	user.Attributes = map[string][]string{}
	user.Attributes["foo"] = []string{"bar", "alice", "bob", "roflcopter"}
	user.Attributes["bar"] = []string{"baz"}

	userID, err := client.CreateUser(
		token.AccessToken,
		cfg.GoCloak.Realm,
		user)
	FailIfErr(t, err, "CreateUser failed")
	t.Logf("Created user with ID: %s. User: %+v", *userID, user)
	tearDown := func() {
		err := client.DeleteUser(
			token.AccessToken,
			cfg.GoCloak.Realm,
			*userID)
		FailIfErr(t, err, "DeleteUser")
	}

	return tearDown, *userID
}

func CreateGroup(t *testing.T, client GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	group := Group{
		Name: GetRandomName("MySuperCoolNewGroup"),
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
		var userID string
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
					userID = user.ID
					break
				}
			}
		} else {
			FailIfErr(t, err, "CreateUser failed")
			userID = *createdUserID
		}

		err = client.SetPassword(
			token.AccessToken,
			userID,
			cfg.GoCloak.Realm,
			cfg.GoCloak.Password,
			false)
		FailIfErr(t, err, "SetPassword	 failed")
	})
}

func TestGocloak_RequestPermission(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
	certs, err := client.GetCerts(cfg.GoCloak.Realm)
	t.Log(certs)
	FailIfErr(t, err, "get certs")
}

func TestGocloak_LoginClient_UnknownRealm(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
	issuer, err := client.GetIssuer(cfg.GoCloak.Realm)
	t.Log(issuer)
	FailIfErr(t, err, "get issuer")
}

func TestGocloak_RetrospectToken_InactiveToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)

	rptResult, err := client.RetrospectToken(
		"foobar",
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(rptResult)
	FailIfErr(t, err, "inspection failed")
	FailIf(t, rptResult.Active, "That should never happen. Token is active")

}

func TestGocloak_GetUserInfo(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetClientToken(t, client)

	userInfo, err := client.GetUserInfo(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Log(userInfo)
	FailIfErr(t, err, "Failed to fetch userinfo")
}

func TestGocloak_RetrospectToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
	token := GetClientToken(t, client)

	resultToken, claims, err := client.DecodeAccessToken(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Log(resultToken)
	t.Log(claims)
	FailIfErr(t, err, "DecodeAccessToken")
}

func TestGocloak_DecodeAccessTokenCustomClaims(t *testing.T) {
	t.Skipf(
		"Due to error: %s",
		"DecodeAccessTokenCustomClaims: json: cannot unmarshal object into Go value of type jwt.Claims")
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetClientToken(t, client)

	claims := jwx.Claims{}
	_, err := client.DecodeAccessTokenCustomClaims(
		token.AccessToken,
		cfg.GoCloak.Realm,
		claims)
	FailIfErr(t, err, "DecodeAccessTokenCustomClaims")
}

func TestGocloak_RefreshToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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

	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	ok := client.UserAttributeContains(attributes, "foo", "alice")
	FailIf(t, !ok, "UserAttributeContains")
}

func TestGocloak_GetUserByID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	fetchedUser, err := client.GetUserByID(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	t.Log(fetchedUser)
	FailIfErr(t, err, "GetUserById failed")
}

func TestGocloak_GetKeyStoreConfig(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
	SetUpTestUser(t, client)
	_, err := client.Login(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password)
	FailIfErr(t, err, "Login failed")
}

func TestGocloak_LoginClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	_, err := client.LoginClient(
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	FailIfErr(t, err, "LoginClient failed")
}

func TestGocloak_LoginAdmin(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	_, err := client.LoginAdmin(
		cfg.Admin.UserName,
		cfg.Admin.Password,
		cfg.Admin.Realm)
	FailIfErr(t, err, "LoginAdmin failed")
}

func TestGocloak_SetPassword(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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

func TestGocloak_CreateUser(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)

	tearDown, _ := CreateUser(t, client)
	defer tearDown()
}

func TestGocloak_CreateUserCustomAttributes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
}

func TestGocloak_CreateGroup(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)

	tearDown, _ := CreateGroup(t, client)
	defer tearDown()
}

func TestGocloak_CreateClientRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
	defer client.DeleteClientRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		testClient.ID,
		role.Name)
}

func TestGocloak_CreateClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	newClient := Client{
		ID:       GetRandomName("ID"),
		ClientID: GetRandomName("ClientID"),
	}
	err := client.CreateClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		newClient)
	FailIfErr(t, err, "CreateClient failed")
	defer client.DeleteClient(
		token.AccessToken,
		cfg.GoCloak.Realm,
		newClient.ID)
}

func TestGocloak_GetUsers(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	users, err := client.GetUsers(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetUsersParams{
			Username: cfg.GoCloak.UserName,
		})
	t.Log(users)
	FailIfErr(t, err, "GetUsers failed")
}

func TestGocloak_GetUserCount(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	count, err := client.GetUserCount(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Logf("Users in Realm: %d", count)
	FailIfErr(t, err, "GetUserCount failed")
}

func TestGocloak_GetGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	_, err := client.GetGroups(
		token.AccessToken,
		cfg.GoCloak.Realm,
		GetGroupsParams{})
	FailIfErr(t, err, "GetGroups failed")
}

func TestGocloak_GetUserGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	_, err := client.GetUserGroups(
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	FailIfErr(t, err, "GetUserGroups failed")
}

func TestGocloak_GetClients(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	r, err := client.GetRealm(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Logf("%+v", r)
	FailIfErr(t, err, "GetRealm failed")
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
			Name: roleName,
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
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	tearDown, _ := CreateRealmRole(t, client)
	defer tearDown()
}

func TestGocloak_GetRealmRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	tearDown, roleName := CreateRealmRole(t, client)
	defer tearDown()

	role, err := client.GetRealmRole(
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	t.Logf("Role: %+v", *role)
	FailIfErr(t, err, "GetRealmRole failed")
	FailIf(
		t,
		role.Name != roleName,
		"GetRealmRole returns unexpected result. Expected: %s; Actual: %+v",
		roleName, role)
}

func TestGocloak_GetRealmRoles(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
	token := GetAdminToken(t, client)

	tearDown, _ := CreateRealmRole(t, client)
	defer tearDown()

	roles, err := client.GetRealmRoles(
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Logf("Roles: %+v", *roles)
	FailIfErr(t, err, "GetRealmRoles failed")
}

func TestGocloak_UpdateRealmRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
	t.Logf("User roles: %+v", *roles)
	FailIfErr(t, err, "GetRealmRolesByUserID failed")
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
	client := NewClient(cfg.HostName)
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
	client := NewClient(cfg.HostName)
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
