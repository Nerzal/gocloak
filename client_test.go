package gocloak_test

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
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

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/pkcs12"

	"github.com/Nerzal/gocloak/v8"
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
		require.NoError(t, err, "cannot open config.json")
		defer func() {
			err := configFile.Close()
			require.NoError(t, err, "cannot close config file")
		}()
		data, err := ioutil.ReadAll(configFile)
		require.NoError(t, err, "cannot read config.json")
		config = &Config{}
		err = json.Unmarshal(data, config)
		require.NoError(t, err, "cannot parse config.json")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		if len(config.Proxy) != 0 {
			proxy, err := url.Parse(config.Proxy)
			require.NoError(t, err, "incorrect proxy url: "+config.Proxy)
			http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(proxy)
		}
		if config.GoCloak.UserName == "" {
			config.GoCloak.UserName = "test_user"
		}
	})
	return config
}

func GetClientToken(t *testing.T, client gocloak.GoCloak) *gocloak.JWT {
	cfg := GetConfig(t)
	token, err := client.LoginClient(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	require.NoError(t, err, "Login failed")
	return token
}

func GetUserToken(t *testing.T, client gocloak.GoCloak) *gocloak.JWT {
	SetUpTestUser(t, client)
	cfg := GetConfig(t)
	token, err := client.Login(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password)
	require.NoError(t, err, "Login failed")
	return token
}

func GetAdminToken(t testing.TB, client gocloak.GoCloak) *gocloak.JWT {
	cfg := GetConfig(t)
	token, err := client.LoginAdmin(
		context.Background(),
		cfg.Admin.UserName,
		cfg.Admin.Password,
		cfg.Admin.Realm)
	require.NoError(t, err, "Login Admin failed")
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

func GetClientByClientID(t *testing.T, client gocloak.GoCloak, clientID string) *gocloak.Client {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	clients, err := client.GetClients(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetClientsParams{
			ClientID: &clientID,
		})
	require.NoError(t, err, "GetClients failed")
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

func CreateGroup(t *testing.T, client gocloak.GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	group := gocloak.Group{
		Name: GetRandomNameP("GroupName"),
		Attributes: &map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
	}
	groupID, err := client.CreateGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		group)
	require.NoError(t, err, "CreateGroup failed")
	t.Logf("Created Group ID: %s ", groupID)

	tearDown := func() {
		err := client.DeleteGroup(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			groupID)
		require.NoError(t, err, "DeleteGroup failed")
	}
	return tearDown, groupID
}

func CreateResource(t *testing.T, client gocloak.GoCloak, clientID string) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	resource := gocloak.ResourceRepresentation{
		Name:        GetRandomNameP("ResourceName"),
		DisplayName: gocloak.StringP("Resource Display Name"),
		Type:        gocloak.StringP("urn:gocloak:resources:test"),
		IconURI:     gocloak.StringP("/resource/test/icon"),
		Attributes: &map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
		URIs: &[]string{
			"/resource/1",
			"/resource/2",
		},
		OwnerManagedAccess: gocloak.BoolP(true),
	}
	createdResource, err := client.CreateResource(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		resource)
	require.NoError(t, err, "CreateResource failed")
	t.Logf("Created Resource ID: %s ", *(createdResource.ID))

	tearDown := func() {
		err := client.DeleteResource(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*createdResource.ID)
		require.NoError(t, err, "DeleteResource failed")
	}
	return tearDown, *createdResource.ID
}

func CreateResourceClientWithScopes(t *testing.T, client gocloak.GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetClientToken(t, client)
	resource := gocloak.ResourceRepresentation{
		Name:        GetRandomNameP("ResourceName"),
		DisplayName: gocloak.StringP("Resource Display Name"),
		Type:        gocloak.StringP("urn:gocloak:resources:test"),
		IconURI:     gocloak.StringP("/resource/test/icon"),
		Attributes: &map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
		URIs: &[]string{
			"/resource/1",
			"/resource/2",
		},
		OwnerManagedAccess: gocloak.BoolP(true),
		ResourceScopes: &[]gocloak.ScopeRepresentation{
			{Name: gocloak.StringP("read-public")},
			{Name: gocloak.StringP("read-private")},
			{Name: gocloak.StringP("post-update")},
			{Name: gocloak.StringP("message-view")},
			{Name: gocloak.StringP("message-post")},
		},
	}
	createdResource, err := client.CreateResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		resource)
	require.NoError(t, err, "CreateResource failed")
	t.Logf("Created Resource ID: %s ", *(createdResource.ID))

	tearDown := func() {
		err := client.DeleteResourceClient(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			*createdResource.ID)
		require.NoError(t, err, "DeleteResource failed")
	}
	return tearDown, *createdResource.ID
}

func CreateResourceClient(t *testing.T, client gocloak.GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetClientToken(t, client)
	resource := gocloak.ResourceRepresentation{
		Name:        GetRandomNameP("ResourceName"),
		DisplayName: gocloak.StringP("Resource Display Name"),
		Type:        gocloak.StringP("urn:gocloak:resources:test"),
		IconURI:     gocloak.StringP("/resource/test/icon"),
		Attributes: &map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
		URIs: &[]string{
			"/resource/1",
			"/resource/2",
		},
		OwnerManagedAccess: gocloak.BoolP(true),
	}
	createdResource, err := client.CreateResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		resource)
	require.NoError(t, err, "CreateResource failed")
	t.Logf("Created Resource ID: %s ", *(createdResource.ID))

	tearDown := func() {
		err := client.DeleteResourceClient(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			*createdResource.ID)
		require.NoError(t, err, "DeleteResource failed")
	}
	return tearDown, *createdResource.ID
}

func CreateScope(t *testing.T, client gocloak.GoCloak, clientID string) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	scope := gocloak.ScopeRepresentation{
		Name:        GetRandomNameP("ScopeName"),
		DisplayName: gocloak.StringP("Scope Display Name"),
		IconURI:     gocloak.StringP("/scope/test/icon"),
	}
	createdScope, err := client.CreateScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		scope)
	require.NoError(t, err, "CreateScope failed")
	t.Logf("Created Scope ID: %s ", *(createdScope.ID))

	tearDown := func() {
		err := client.DeleteScope(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*createdScope.ID)
		require.NoError(t, err, "DeleteScope failed")
	}
	return tearDown, *createdScope.ID
}

func CreatePolicy(t *testing.T, client gocloak.GoCloak, clientID string, policy gocloak.PolicyRepresentation) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	createdPolicy, err := client.CreatePolicy(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		policy)
	require.NoError(t, err, "CreatePolicy failed")

	t.Logf("Created Policy ID: %s ", *(createdPolicy.ID))

	tearDown := func() {
		err := client.DeletePolicy(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*createdPolicy.ID)
		require.NoError(t, err, "DeletePolicy failed")
	}
	return tearDown, *createdPolicy.ID
}

func CreatePermission(t *testing.T, client gocloak.GoCloak, clientID string, permission gocloak.PermissionRepresentation) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	createdPermission, err := client.CreatePermission(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		permission)
	require.NoError(t, err, "CreatePermission failed")
	t.Logf("Created RequestingPartyPermission ID: %s ", *(createdPermission.ID))

	tearDown := func() {
		err := client.DeletePermission(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			*createdPermission.ID)
		require.NoError(t, err, "DeletePermission failed")
	}
	return tearDown, *createdPermission.ID
}

func CreateClient(t *testing.T, client gocloak.GoCloak, newClient *gocloak.Client) (func(), string) {
	if newClient == nil {
		newClient = &gocloak.Client{
			ClientID: GetRandomNameP("CliendID"),
			Name:     GetRandomNameP("Name"),
			BaseURL:  gocloak.StringP("http://example.com"),
		}
	}
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	createdID, err := client.CreateClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*newClient,
	)
	require.NoError(t, err, "CreateClient failed")
	tearDown := func() {
		_ = client.DeleteClient(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			createdID,
		)
	}
	return tearDown, createdID
}

func SetUpTestUser(t testing.TB, client gocloak.GoCloak) {
	setupOnce.Do(func() {
		cfg := GetConfig(t)
		token := GetAdminToken(t, client)

		user := gocloak.User{
			Username:      gocloak.StringP(cfg.GoCloak.UserName),
			Email:         gocloak.StringP(cfg.GoCloak.UserName + "@localhost"),
			EmailVerified: gocloak.BoolP(true),
			Enabled:       gocloak.BoolP(true),
		}

		createdUserID, err := client.CreateUser(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			user,
		)

		apiError, ok := err.(*gocloak.APIError)
		if ok && apiError.Code == http.StatusConflict {
			users, err := client.GetUsers(
				context.Background(),
				token.AccessToken,
				cfg.GoCloak.Realm,
				gocloak.GetUsersParams{
					Username: gocloak.StringP(cfg.GoCloak.UserName),
				})
			require.NoError(t, err, "GetUsers failed")
			for _, user := range users {
				if gocloak.PString(user.Username) == cfg.GoCloak.UserName {
					testUserID = gocloak.PString(user.ID)
					break
				}
			}
		} else {
			require.NoError(t, err, "CreateUser failed")
			testUserID = createdUserID
		}

		err = client.SetPassword(
			context.Background(),
			token.AccessToken,
			testUserID,
			cfg.GoCloak.Realm,
			cfg.GoCloak.Password,
			false)
		require.NoError(t, err, "SetPassword failed")
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

func NewClientWithDebug(t testing.TB) gocloak.GoCloak {
	cfg := GetConfig(t)
	client := gocloak.NewClient(cfg.HostName)
	cond := func(resp *resty.Response, err error) bool {
		if resp != nil && resp.IsError() {
			if e, ok := resp.Error().(*gocloak.HTTPErrorResponse); ok {
				msg := e.String()
				return strings.Contains(msg, "Cached clientScope not found") || strings.Contains(msg, "unknown_error")
			}
		}
		return false
	}

	restyClient := client.RestyClient()
	restyClient.
		SetDebug(true).
		SetLogger(&RestyLogWriter{
			t: t,
		}).
		SetRetryCount(10).
		SetRetryWaitTime(2 * time.Second).
		AddRetryCondition(cond)

	return client
}

// FailRequest fails requests and returns an error
//   err - returned error or nil to return the default error
//   failN - number of requests to be failed
//   skipN = number of requests to be executed and not failed by this function
func FailRequest(client gocloak.GoCloak, err error, failN, skipN int) gocloak.GoCloak {
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

func ClearRealmCache(t testing.TB, client gocloak.GoCloak, realm ...string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)
	if len(realm) == 0 {
		realm = append(realm, cfg.Admin.Realm, cfg.GoCloak.Realm)
	}
	ctx := context.Background()
	for _, r := range realm {
		err := client.ClearRealmCache(ctx, token.AccessToken, r)
		require.NoError(t, err, "ClearRealmCache failed for a realm: %s", r)
		err = client.ClearUserCache(ctx, token.AccessToken, r)
		require.NoError(t, err, "ClearUserCache failed for a realm: %s", r)
		err = client.ClearKeysCache(ctx, token.AccessToken, r)
		require.NoError(t, err, "ClearKeysCache failed for a realm: %s", r)
	}
}

// -----
// Tests
// -----

func TestGocloak_RestyClient(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	restyClient := client.RestyClient()
	require.NotEqual(t, restyClient, resty.New())
}

func TestGocloak_SetRestyClient(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	newRestyClient := resty.New()
	client.SetRestyClient(newRestyClient)
	restyClient := client.RestyClient()
	require.Equal(t, newRestyClient, restyClient)
}

func TestGocloak_checkForError(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	FailRequest(client, nil, 1, 0)
	_, err := client.Login(context.Background(), "", "", "", "", "")
	require.Error(t, err, "All requests must fail with NewClientWithError")
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
		context.Background(),
		token.AccessToken,
	)
	require.NoError(t, err, "Failed to fetch server info")
	t.Logf("Server Info: %+v", serverInfo)

	FailRequest(client, nil, 1, 0)
	_, err = client.GetServerInfo(
		context.Background(),
		token.AccessToken,
	)
	require.Error(t, err)
}

func TestGocloak_GetUserInfo(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)
	userInfo, err := client.GetUserInfo(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "Failed to fetch userinfo")
	t.Log(userInfo)
	FailRequest(client, nil, 1, 0)
	_, err = client.GetUserInfo(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm)
	require.Error(t, err, "")
}

func TestGocloak_GetRawUserInfo(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)
	userInfo, err := client.GetUserInfo(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "Failed to fetch userinfo")
	t.Log(userInfo)
	require.NotEmpty(t, userInfo)
}

func TestGocloak_RequestPermission(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	token, err := client.Login(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password)
	require.NoError(t, err, "login failed")

	rpt, err := client.GetRequestingPartyToken(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.RequestingPartyTokenOptions{
			Audience: gocloak.StringP(cfg.GoCloak.ClientID),
			Permissions: &[]string{
				"Fake Resource",
			},
		})
	require.Error(t, err, "GetRequestingPartyToken failed")
	require.Nil(t, rpt)

	rpt, err = client.GetRequestingPartyToken(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.RequestingPartyTokenOptions{
			Audience: gocloak.StringP(cfg.GoCloak.ClientID),
			Permissions: &[]string{
				"Default Resource",
			},
		})
	require.NoError(t, err, "GetRequestingPartyToken failed")
	require.NotNil(t, rpt)

	rptResult, err := client.RetrospectToken(
		context.Background(),
		rpt.AccessToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
	)
	t.Log(rptResult)
	require.NoError(t, err, "inspection failed")
	require.True(t, gocloak.PBool(rptResult.Active), "Inactive Token oO")
	require.NotNil(t, *rptResult.Permissions)
	permissions := *rptResult.Permissions
	require.Len(t, permissions, 1, "GetRequestingPartyToken failed")
	require.Equal(t, "Default Resource", *permissions[0].RSName, "GetRequestingPartyToken failed")
}

func TestGocloak_GetCerts(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	certs, err := client.GetCerts(context.Background(), cfg.GoCloak.Realm)
	require.NoError(t, err, "get certs")
	t.Log(certs)
}

func TestGocloak_LoginClient_UnknownRealm(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginClient(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		"ThisRealmDoesNotExist")
	require.Error(t, err, "Login shouldn't be successful")
	require.EqualError(t, err, "404 Not Found: Realm does not exist")
}

func TestGocloak_GetIssuer(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	issuer, err := client.GetIssuer(context.Background(), cfg.GoCloak.Realm)
	t.Log(issuer)
	require.NoError(t, err, "get issuer")
}

func TestGocloak_RetrospectToken_InactiveToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)

	rptResult, err := client.RetrospectToken(
		context.Background(),
		"foobar",
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(rptResult)
	require.NoError(t, err, "inspection failed")
	require.False(t, gocloak.PBool(rptResult.Active), "That should never happen. Token is active")
}

func TestGocloak_RetrospectToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	rptResult, err := client.RetrospectToken(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(rptResult)
	require.NoError(t, err, "Inspection failed")
	require.True(t, gocloak.PBool(rptResult.Active), "Inactive Token oO")
}

func TestGocloak_DecodeAccessToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	resultToken, claims, err := client.DecodeAccessToken(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		"",
	)
	require.NoError(t, err)
	t.Log(resultToken)
	t.Log(claims)
}

func TestGocloak_DecodeAccessTokenCustomClaims(t *testing.T) {
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	claims := jwt.MapClaims{}
	resultToken, err := client.DecodeAccessTokenCustomClaims(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		"",
		claims,
	)
	require.NoError(t, err)
	t.Log(resultToken)
	t.Log(claims)
}

func TestGocloak_RefreshToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	token, err := client.RefreshToken(
		context.Background(),
		token.RefreshToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	t.Log(token)
	require.NoError(t, err, "RefreshToken failed")
}

func TestGocloak_UserAttributeContains(t *testing.T) {
	t.Parallel()

	attributes := map[string][]string{}
	attributes["foo"] = []string{"bar", "alice", "bob", "roflcopter"}
	attributes["bar"] = []string{"baz"}

	ok := gocloak.UserAttributeContains(attributes, "foo", "alice")
	require.False(t, !ok, "UserAttributeContains")
}

func TestGocloak_GetKeyStoreConfig(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	config, err := client.GetKeyStoreConfig(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Log(config)
	require.NoError(t, err, "GetKeyStoreConfig")
}

func TestGocloak_Login(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.Login(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password)
	require.NoError(t, err, "Login failed")
}

func TestGocloak_LoginSignedJWT(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	keystore := filepath.Join("testdata", "keystore.p12")
	f, err := os.Open(keystore)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()
	pfxData, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	pKey, cert, err := pkcs12.Decode(pfxData, "secret")
	require.NoError(t, err)
	rsaKey, ok := pKey.(*rsa.PrivateKey)
	require.True(t, ok)

	client := NewClientWithDebug(t)
	testClient := gocloak.Client{
		ID:                      GetRandomNameP("client-id-"),
		ClientID:                GetRandomNameP("client-signed-jwt-client-id-"),
		ClientAuthenticatorType: gocloak.StringP("client-jwt"),
		RedirectURIs:            &[]string{"localhost"},
		StandardFlowEnabled:     gocloak.BoolP(true),
		ServiceAccountsEnabled:  gocloak.BoolP(true),
		Enabled:                 gocloak.BoolP(true),
		FullScopeAllowed:        gocloak.BoolP(true),
		Protocol:                gocloak.StringP("openid-connect"),
		PublicClient:            gocloak.BoolP(false),
		Attributes: &map[string]string{
			"jwt.credential.certificate": base64.StdEncoding.EncodeToString(cert.Raw),
		},
	}
	tearDown, _ := CreateClient(t, client, &testClient)
	defer tearDown()
	_, err = client.LoginClientSignedJWT(
		context.Background(),
		*testClient.ClientID,
		cfg.GoCloak.Realm,
		rsaKey,
		jwt.SigningMethodRS256,
		&jwt.Time{
			Time: time.Now().Add(time.Hour),
		},
	)
	require.NoError(t, err, "Login failed")
}

func TestGocloak_LoginOtp(t *testing.T) {

	totp := "123456"

	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.LoginOtp(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		cfg.GoCloak.UserName,
		cfg.GoCloak.Password,
		totp,
	)
	require.NoError(t, err, "Login failed")
}

func TestGocloak_GetToken(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	newToken, err := client.GetToken(
		context.Background(),
		cfg.GoCloak.Realm,
		gocloak.TokenOptions{
			ClientID:      &cfg.GoCloak.ClientID,
			ClientSecret:  &cfg.GoCloak.ClientSecret,
			Username:      &cfg.GoCloak.UserName,
			Password:      &cfg.GoCloak.Password,
			GrantType:     gocloak.StringP("password"),
			ResponseTypes: &[]string{"token", "id_token"},
			Scopes:        &[]string{"openid", "offline_access"},
		},
	)
	require.NoError(t, err, "Login failed")
	t.Logf("New token: %+v", *newToken)
	require.Equal(t, newToken.RefreshExpiresIn, 0, "Got a refresh token instead of offline")
	require.NotEmpty(t, newToken.IDToken, "Got an empty if token")
}

func TestGocloak_GetRequestingPartyToken(t *testing.T) {
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	newToken, err := client.GetToken(
		context.Background(),
		cfg.GoCloak.Realm,
		gocloak.TokenOptions{
			ClientID:      &cfg.GoCloak.ClientID,
			ClientSecret:  &cfg.GoCloak.ClientSecret,
			Username:      &cfg.GoCloak.UserName,
			Password:      &cfg.GoCloak.Password,
			GrantType:     gocloak.StringP("password"),
			ResponseTypes: &[]string{"token", "id_token"},
			Scopes:        &[]string{"openid"},
		},
	)
	require.NoError(t, err, "Login failed")
	t.Logf("New token: %+v", *newToken)
	require.NotEmpty(t, newToken.IDToken, "Got an empty id token")

	rpt, err := client.GetRequestingPartyToken(
		context.Background(),
		newToken.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.RequestingPartyTokenOptions{
			Audience: &cfg.GoCloak.ClientID,
		},
	)
	require.NoError(t, err, "Get requesting party token failed")
	t.Logf("New RPT: %+v", *rpt)

	_, err = client.RetrospectToken(
		context.Background(),
		rpt.AccessToken,
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "RetrospectToken failed")
}

func TestGocloak_LoginClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginClient(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm)
	require.NoError(t, err, "LoginClient failed")
}

func TestGocloak_LoginAdmin(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	_, err := client.LoginAdmin(
		context.Background(),
		cfg.Admin.UserName,
		cfg.Admin.Password,
		cfg.Admin.Realm)
	require.NoError(t, err, "LoginAdmin failed")
}

func TestGocloak_SetPassword(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	err := client.SetPassword(
		context.Background(),
		token.AccessToken,
		userID,
		cfg.GoCloak.Realm,
		"passwort1234!",
		false)
	require.NoError(t, err, "Failed to set password")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	require.NoError(t, err, "GetGroup failed")
	t.Logf("Created Group: %+v", createdGroup)
	require.Equal(t, groupID, *(createdGroup.ID))

	err = client.UpdateGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.Group{},
	)
	require.Error(t, err, "Should fail because of missing ID of the group")

	createdGroup.Name = GetRandomNameP("GroupName")
	err = client.UpdateGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*createdGroup,
	)
	require.NoError(t, err, "UpdateGroup failed")

	updatedGroup, err := client.GetGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	require.NoError(t, err, "GetGroup failed")
	require.Equal(t, *(createdGroup.Name), *(updatedGroup.Name))

	childGroupID, err := client.CreateChildGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		gocloak.Group{
			Name: GetRandomNameP("GroupName"),
		},
	)
	require.NoError(t, err, "CreateChildGroup failed")

	_, err = client.GetGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		childGroupID,
	)
	require.NoError(t, err, "GetGroup failed")
}

func CreateClientRole(t *testing.T, client gocloak.GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	roleName := GetRandomName("Role")
	t.Logf("Creating Client Role: %s", roleName)
	clientRoleID, err := client.CreateClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.Role{
			Name: &roleName,
		})
	t.Logf("Created Client Role ID: %s", clientRoleID)
	require.Equal(t, roleName, clientRoleID)

	require.NoError(t, err, "CreateClientRole failed")
	tearDown := func() {
		err := client.DeleteClientRole(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			gocloakClientID,
			roleName)
		require.NoError(t, err, "DeleteClientRole failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName,
	)
	require.NoError(t, err, "GetClientRoleI failed")
	require.NotNil(t, role)

	role, err = client.GetClientRoleByID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*role.ID,
	)
	require.NoError(t, err, "GetClientRoleI failed")
	require.NotNil(t, role)

	token = GetAdminToken(t, client)
	role, err = client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		"Fake-Role-Name",
	)
	require.Error(t, err)
	require.Nil(t, role)
}

func CreateClientScope(t *testing.T, client gocloak.GoCloak, scope *gocloak.ClientScope) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	if scope == nil {
		scope = &gocloak.ClientScope{
			ID:   GetRandomNameP("client-scope-id-"),
			Name: GetRandomNameP("client-scope-name-"),
		}
	}

	t.Logf("Creating Client Scope: %+v", scope)
	clientScopeID, err := client.CreateClientScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*scope,
	)
	if !gocloak.NilOrEmpty(scope.ID) {
		require.Equal(t, clientScopeID, *scope.ID)
	}
	require.NoError(t, err, "CreateClientScope failed")
	tearDown := func() {
		err := client.DeleteClientScope(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientScopeID,
		)
		require.NoError(t, err, "DeleteClientScope failed")
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

	scope := gocloak.ClientScope{
		ID:       GetRandomNameP("client-scope-id-"),
		Name:     GetRandomNameP("client-scope-name-"),
		Protocol: gocloak.StringP("openid-connect"),
		ClientScopeAttributes: &gocloak.ClientScopeAttributes{
			IncludeInTokenScope: gocloak.StringP("true"),
		},
	}

	tearDown, scopeID := CreateClientScope(t, client, &scope)
	defer tearDown()

	scopesBeforeAdding, err := client.GetClientsDefaultScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientsDefaultScopes failed")

	err = client.AddDefaultScopeToClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	require.NoError(t, err, "AddDefaultScopeToClient failed")

	scopesAfterAdding, err := client.GetClientsDefaultScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientsDefaultScopes failed")

	require.NotEqual(t, len(scopesBeforeAdding), len(scopesAfterAdding), "scope should have been added")

	err = client.RemoveDefaultScopeFromClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	require.NoError(t, err, "RemoveDefaultScopeFromClient failed")

	scopesAfterRemoving, err := client.GetClientsDefaultScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientsDefaultScopes failed")

	require.Equal(t, len(scopesAfterRemoving), len(scopesBeforeAdding), "scope should have been removed")
}

func TestGocloak_ListAddRemoveOptionalClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	scope := gocloak.ClientScope{
		ID:       GetRandomNameP("client-scope-id-"),
		Name:     GetRandomNameP("client-scope-name-"),
		Protocol: gocloak.StringP("openid-connect"),
		ClientScopeAttributes: &gocloak.ClientScopeAttributes{
			IncludeInTokenScope: gocloak.StringP("true"),
		},
	}
	tearDown, scopeID := CreateClientScope(t, client, &scope)
	defer tearDown()

	scopesBeforeAdding, err := client.GetClientsOptionalScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID)
	require.NoError(t, err, "GetClientsOptionalScopes failed")

	err = client.AddOptionalScopeToClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID)
	require.NoError(t, err, "AddOptionalScopeToClient failed")

	scopesAfterAdding, err := client.GetClientsOptionalScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID)
	require.NoError(t, err, "GetClientsOptionalScopes failed")

	require.NotEqual(t, len(scopesAfterAdding), len(scopesBeforeAdding), "scope should have been added")

	err = client.RemoveOptionalScopeFromClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID)
	require.NoError(t, err, "RemoveOptionalScopeFromClient failed")

	scopesAfterRemoving, err := client.GetClientsOptionalScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID)
	require.NoError(t, err, "GetClientsOptionalScopes failed")

	require.Equal(t, len(scopesBeforeAdding), len(scopesAfterRemoving), "scope should have been removed")
}

func TestGocloak_GetDefaultOptionalClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	scopes, err := client.GetDefaultOptionalClientScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm)

	require.NoError(t, err, "GetDefaultOptionalClientScopes failed")

	require.NotEqual(t, 0, len(scopes), "there should be default optional client scopes")
}

func TestGocloak_GetDefaultDefaultClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	scopes, err := client.GetDefaultDefaultClientScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm)

	require.NoError(t, err, "GetDefaultDefaultClientScopes failed")
	require.NotEmpty(t, scopes, "there should be default default client scopes")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		scopeID,
	)
	require.NoError(t, err, "GetClientScope failed")
	// Checking that GetClientScope returns same client scope
	require.NotNil(t, createdClientScope.ID)
	require.Equal(t, scopeID, *(createdClientScope.ID))
}

func TestGocloak_GetClientScopes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Getting client scopes
	scopes, err := client.GetClientScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm)
	require.NoError(t, err, "GetClientScopes failed")
	// Checking that GetClientScopes returns scopes
	require.NotZero(t, len(scopes), "there should be client scopes")
}

func CreateClientScopeMappingsRealmRoles(t *testing.T, client gocloak.GoCloak, clientID string, roles []gocloak.Role) func() {
	token := GetAdminToken(t, client)
	cfg := GetConfig(t)

	// Creating client scope mappings
	err := client.CreateClientScopeMappingsRealmRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		roles,
	)
	require.NoError(t, err, "CreateClientScopeMappingsRealmRoles failed")

	tearDown := func() {
		err = client.DeleteClientScopeMappingsRealmRoles(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			roles,
		)
		require.NoError(t, err, "DeleteClientScopeMappingsRealmRoles failed")
	}
	return tearDown
}

func CreateClientScopeMappingsClientRoles(t *testing.T, client gocloak.GoCloak, clientID, clients string, roles []gocloak.Role) func() {
	token := GetAdminToken(t, client)
	cfg := GetConfig(t)

	// Creating client scope mappings
	err := client.CreateClientScopeMappingsClientRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		clients,
		roles,
	)
	require.NoError(t, err, "CreateClientScopeMappingsClientRoles failed")

	tearDown := func() {
		err = client.DeleteClientScopeMappingsClientRoles(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			clientID,
			clients,
			roles,
		)
		require.NoError(t, err, "DeleteClientScopeMappingsClientRoles failed")
	}
	return tearDown
}

func TestGocloak_ClientScopeMappingsClientRoles(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	testClient := gocloak.Client{
		ClientID:         GetRandomNameP("ClientID"),
		BaseURL:          gocloak.StringP("http://example.com"),
		FullScopeAllowed: gocloak.BoolP(false),
	}
	// Creating client
	tearDownClient, clientID := CreateClient(t, client, &testClient)
	defer tearDownClient()

	// Creating client roles
	var roles []gocloak.Role
	tearDownRole1, roleName := CreateClientRole(t, client)
	defer tearDownRole1()
	role, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName)
	require.NoError(t, err, "CreateClientRole failed")
	roles = append(roles, *role)
	tearDownRole2, roleName := CreateClientRole(t, client)
	defer tearDownRole2()
	role, err = client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName)
	require.NoError(t, err, "CreateClientRole failed")
	roles = append(roles, *role)

	// Creating client client roles for client scope mappings
	tearDownScopeMappingsClientRoles := CreateClientScopeMappingsClientRoles(t, client, clientID, gocloakClientID, roles)
	defer tearDownScopeMappingsClientRoles()

	// Check client roles
	clientRoles, err := client.GetClientScopeMappingsClientRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientScopeMappingsClientRoles failed")
	require.Len(
		t, clientRoles, len(roles),
		"GetClientScopeMappingsClientRoles should return exact %s roles", len(roles),
	)

	clientRoles, err = client.GetClientRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientRoles failed")

	clientRolesAvailable, err := client.GetClientScopeMappingsClientRolesAvailable(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientScopeMappingsClientRolesAvailable failed")
	require.Len(
		t, clientRolesAvailable, len(clientRoles)-len(roles),
		"GetClientScopeMappingsClientRolesAvailable should return exact %s roles", len(clientRoles)-len(roles),
	)
}

func TestGocloak_ClientScopeMappingsRealmRoles(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	testClient := gocloak.Client{
		ClientID:         GetRandomNameP("ClientID"),
		BaseURL:          gocloak.StringP("http://example.com"),
		FullScopeAllowed: gocloak.BoolP(false),
	}
	// Creating client
	tearDownClient, clientID := CreateClient(t, client, &testClient)
	defer tearDownClient()

	// Creating realm role
	var roles []gocloak.Role
	tearDownRealmRole1, roleName := CreateRealmRole(t, client)
	defer tearDownRealmRole1()
	role, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName,
	)
	require.NoError(t, err, "CreateRealmRole failed")
	roles = append(roles, *role)
	tearDownRealmRole2, roleName := CreateRealmRole(t, client)
	defer tearDownRealmRole2()
	role, err = client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName,
	)
	require.NoError(t, err, "CreateRealmRole failed")
	roles = append(roles, *role)

	// Creating client realm roles for client scope mappings
	tearDownScopeMappingsRealmRoles := CreateClientScopeMappingsRealmRoles(t, client, clientID, roles)
	defer tearDownScopeMappingsRealmRoles()

	// Check realm roles
	realmRoles, err := client.GetClientScopeMappingsRealmRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
	)
	require.NoError(t, err, "GetClientScopeMappingsRealmRoles failed")
	require.Len(
		t, realmRoles, len(roles),
		"GetClientScopeMappingsRealmRoles should return exact %s realm", len(roles),
	)

	realmRoles, err = client.GetRealmRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "GetRealmRoles failed")

	realmRolesAvailable, err := client.GetClientScopeMappingsRealmRolesAvailable(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
	)
	require.NoError(t, err, "GetClientScopeMappingsRealmRolesAvailable failed")
	require.Len(
		t, realmRolesAvailable, len(realmRoles)-len(roles),
		"GetClientScopeMappingsRealmRolesAvailable should return exact %s realm", len(realmRoles)-len(roles),
	)
}

func TestGocloak_CreateListGetUpdateDeleteClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	clientID := GetRandomNameP("ClientID")
	testClient := gocloak.Client{
		ClientID: clientID,
		BaseURL:  gocloak.StringP("http://example.com"),
	}
	t.Logf("Client ID: %s", *clientID)

	// Creating a client
	tearDown, createdClientID := CreateClient(t, client, &testClient)
	defer tearDown()

	// Looking for a created client
	clients, err := client.GetClients(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetClientsParams{
			ClientID: clientID,
		},
	)
	require.NoError(t, err, "CreateClients failed")
	require.Len(t, clients, 1, "GetClients should return exact 1 client")
	require.Equal(t, createdClientID, *(clients[0].ID))
	t.Logf("Clients: %+v", clients)

	// Getting exact client
	createdClient, err := client.GetClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		createdClientID,
	)
	require.NoError(t, err, "GetClient failed")
	t.Logf("Created client: %+v", createdClient)
	// Checking that GetClient returns same client
	require.Equal(t, clients[0], createdClient)

	// Updating the client

	// Should fail
	err = client.UpdateClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.Client{},
	)
	require.Error(t, err, "Should fail because of missing ID of the client")

	// Update existing client
	createdClient.Name = GetRandomNameP("Name")
	err = client.UpdateClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*createdClient,
	)
	require.NoError(t, err, "GetClient failed")

	// Getting updated client
	updatedClient, err := client.GetClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		createdClientID,
	)
	require.NoError(t, err, "GetClient failed")
	t.Logf("Update client: %+v", createdClient)
	require.Equal(t, *createdClient, *updatedClient)

	// Deleting the client
	err = client.DeleteClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		createdClientID,
	)
	require.NoError(t, err, "DeleteClient failed")

	// Verifying that the client was deleted
	clients, err = client.GetClients(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetClientsParams{
			ClientID: clientID,
		},
	)
	require.NoError(t, err, "CreateClients failed")
	require.Len(t, clients, 0, "GetClients should not return any clients")
}

func TestGocloak_GetGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	_, err := client.GetGroups(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetGroupsParams{})
	require.NoError(t, err, "GetGroups failed")
}

func TestGocloak_GetGroupsFull(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	groups, err := client.GetGroups(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetGroupsParams{
			Full: gocloak.BoolP(true),
		})
	require.NoError(t, err, "GetGroups failed")

	for _, group := range groups {
		if gocloak.NilOrEmpty(group.ID) {
			continue
		}
		require.NotNil(t, group.Attributes)
		if *group.ID == groupID {
			ok := gocloak.UserAttributeContains(*group.Attributes, "foo", "alice")
			require.True(t, ok, "UserAttributeContains")
			return
		}
	}

	require.Fail(t, "GetGroupsFull failed")
}

func TestGocloak_GetGroupsBriefRepresentation(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	groups, err := client.GetGroups(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetGroupsParams{
			BriefRepresentation: gocloak.BoolP(false),
		})
	require.NoError(t, err, "GetGroups failed")

	for _, group := range groups {
		if gocloak.NilOrEmpty(group.ID) {
			continue
		}
		if *group.ID == groupID {
			require.NotNil(t, group.Attributes)
			ok := gocloak.UserAttributeContains(*group.Attributes, "foo", "alice")
			require.True(t, ok, "UserAttributeContains")
			return
		}
	}

	require.Fail(t, "GetGroupsBriefRepresentation failed")
}

func TestGocloak_GetGroupFull(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	createdGroup, err := client.GetGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	require.NoError(t, err, "GetGroup failed")

	require.NotNil(t, createdGroup.Attributes)
	ok := gocloak.UserAttributeContains(*createdGroup.Attributes, "foo", "alice")
	require.True(t, ok, "UserAttributeContains")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	require.NoError(t, err, "AddUserToGroup failed")

	users, err := client.GetGroupMembers(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		gocloak.GetGroupsParams{},
	)
	require.NoError(t, err, "AddUserToGroup failed")
	require.Len(t, users, 1)
}

func TestGocloak_ListAddRemoveDefaultGroups(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	groupsBeforeAdding, err := client.GetDefaultGroups(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "GetDefaultGroups failed")

	err = client.AddDefaultGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	require.NoError(t, err, "AddDefaultGroup failed")

	groupsAfterAdding, err := client.GetDefaultGroups(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "GetDefaultGroups failed")
	require.NotEqual(t, len(groupsBeforeAdding), len(groupsAfterAdding), "group should have been added")

	err = client.RemoveDefaultGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
	)
	require.NoError(t, err, "RemoveDefaultGroup failed")

	groupsAfterRemoving, err := client.GetDefaultGroups(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "GetDefaultGroups failed")
	require.Equal(t, len(groupsAfterRemoving), len(groupsBeforeAdding), "group should have been removed")
}

func TestGocloak_GetClientRoles(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	testClient := GetClientByClientID(t, client, cfg.GoCloak.ClientID)

	_, err := client.GetClientRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*testClient.ID)
	require.NoError(t, err, "GetClientRoles failed")
}

func TestGocloak_GetRoleMappingByGroupID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	_, err := client.GetRoleMappingByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID)
	require.NoError(t, err, "GetRoleMappingByGroupID failed")
}

func TestGocloak_GetRoleMappingByUserID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	_, err := client.GetRoleMappingByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err, "GetRoleMappingByUserID failed")
}

func TestGocloak_ExecuteActionsEmail_UpdatePassword(t *testing.T) {
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	params := gocloak.ExecuteActionsEmail{
		ClientID: &(cfg.GoCloak.ClientID),
		UserID:   &userID,
		Actions:  &[]string{"UPDATE_PASSWORD"},
	}

	err := client.ExecuteActionsEmail(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		params)
	if err != nil {
		if err.Error() == "500 Internal Server Error: Failed to send execute actions email" {
			return
		}
		require.NoError(t, err, "ExecuteActionsEmail failed")
	}
}

func TestGocloak_Logout(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetUserToken(t, client)

	err := client.Logout(
		context.Background(),
		cfg.GoCloak.ClientID,
		cfg.GoCloak.ClientSecret,
		cfg.GoCloak.Realm,
		token.RefreshToken)
	require.NoError(t, err, "Logout failed")
}

func TestGocloak_LogoutAllSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()

	err := client.LogoutAllSessions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	require.NoError(t, err, "Logout failed")
}

func TestGocloak_LogoutUserSession(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	uToken := GetUserToken(t, client)
	aToken := GetAdminToken(t, client)

	err := client.LogoutUserSession(
		context.Background(),
		aToken.AccessToken,
		cfg.GoCloak.Realm,
		uToken.SessionState,
	)
	require.NoError(t, err, "Logout failed")
}

func TestGocloak_GetRealm(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	r, err := client.GetRealm(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm)
	t.Logf("%+v", r)
	require.NoError(t, err, "GetRealm failed")
}

func TestGocloak_GetRealms(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	r, err := client.GetRealms(
		context.Background(),
		token.AccessToken,
	)
	t.Logf("%+v", r)
	require.NoError(t, err, "GetRealms failed")
}

// -----------
// Realm
// -----------

func CreateRealm(t *testing.T, client gocloak.GoCloak) (func(), string) {
	token := GetAdminToken(t, client)

	realmName := GetRandomName("Realm")
	t.Logf("Creating Realm: %s", realmName)
	realmID, err := client.CreateRealm(
		context.Background(),
		token.AccessToken,
		gocloak.RealmRepresentation{
			Realm: &realmName,
			Roles: &gocloak.RolesRepresentation{
				Realm: &[]gocloak.Role{
					{
						Name: GetRandomNameP("Role"),
					},
				},
				Client: &map[string][]gocloak.Role{
					"account": {
						{
							Name: GetRandomNameP("Role"),
						},
					},
				},
			},
		})
	require.NoError(t, err, "CreateRealm failed")
	require.Equal(t, realmID, realmName)
	tearDown := func() {
		token := GetAdminToken(t, client)
		err := client.DeleteRealm(
			context.Background(),
			token.AccessToken,
			realmName)
		require.NoError(t, err, "DeleteRealm failed")
	}
	return tearDown, realmName
}

func TestGocloak_CreateRealm(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	tearDown, _ := CreateRealm(t, client)
	defer tearDown()
}

func TestGocloak_UpdateRealm(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, realmID := CreateRealm(t, client)
	defer tearDown()

	realm, err := client.GetRealm(
		context.Background(),
		token.AccessToken,
		realmID)
	require.NoError(t, err, "GetRealm failed")

	realm.Enabled = gocloak.BoolP(false)
	err = client.UpdateRealm(
		context.Background(),
		token.AccessToken,
		*realm)
	require.NoError(t, err, "UpdateRealm failed")
}

func TestGocloak_ClearRealmCache(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)
	ClearRealmCache(t, client)
}

// -----------
// Realm Roles
// -----------

func CreateRealmRole(t *testing.T, client gocloak.GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	roleName := GetRandomName("Role")
	t.Logf("Creating RoleName: %s", roleName)
	realmRoleID, err := client.CreateRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.Role{
			Name:        &roleName,
			ContainerID: gocloak.StringP("asd"),
		})
	require.NoError(t, err, "CreateRealmRole failed")
	require.Equal(t, roleName, realmRoleID)
	tearDown := func() {
		err := client.DeleteRealmRole(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			roleName)
		require.NoError(t, err, "DeleteRealmRole failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	require.NoError(t, err, "GetRealmRole failed")
	t.Logf("Role: %+v", *role)
	require.False(
		t,
		*role.Name != roleName,
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm)
	require.NoError(t, err, "GetRealmRoles failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		oldRoleName,
		gocloak.Role{
			Name: &newRoleName,
		})
	require.NoError(t, err, "UpdateRealmRole failed")
	err = client.DeleteRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		oldRoleName)
	require.Error(
		t,
		err,
		"Role with old name was deleted successfully, but it shouldn't. Old role: %s; Updated role: %s",
		oldRoleName, newRoleName)
	err = client.DeleteRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		newRoleName)
	require.NoError(t, err, "DeleteRealmRole failed")
}

func TestGocloak_DeleteRealmRole(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	_, roleName := CreateRealmRole(t, client)

	err := client.DeleteRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	require.NoError(t, err, "DeleteRealmRole failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	require.NoError(t, err, "GetRealmRole failed")

	roles := []gocloak.Role{*role}
	err = client.AddRealmRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		roles,
	)
	require.NoError(t, err, "AddRealmRoleToUser failed")
	err = client.DeleteRealmRoleFromUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		roles,
	)
	require.NoError(t, err, "DeleteRealmRoleFromUser failed")
}

func TestGocloak_AddRealmRoleToGroup_DeleteRealmRoleFromGroup(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()
	tearDownRole, roleName := CreateRealmRole(t, client)
	defer tearDownRole()
	role, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	require.NoError(t, err, "GetRealmRole failed")

	roles := []gocloak.Role{*role}
	err = client.AddRealmRoleToGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		roles,
	)
	require.NoError(t, err, "AddRealmRoleToGroup failed")
	err = client.DeleteRealmRoleFromGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		roles,
	)
	require.NoError(t, err, "DeleteRealmRoleFromGroup failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	require.NoError(t, err)

	err = client.AddRealmRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]gocloak.Role{
			*role,
		})
	require.NoError(t, err)

	roles, err := client.GetRealmRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	var found bool
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role.Name {
			found = true
			break
		}
	}
	require.True(t, found, "The role has not been found in the assigned roles. Role: %+v", *role)

	roles, err = client.GetCompositeRealmRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role.Name {
			return
		}
	}
	require.Fail(t, "The role has not been found in the assigned composite roles. Role: %+v", *role)
}

func TestGocloak_GetRealmRolesByGroupID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	tearDown, roleName := CreateRealmRole(t, client)
	defer tearDown()

	role, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName,
	)
	require.NoError(t, err, "Can't get just created role with GetRealmRole")

	err = client.AddRealmRoleToGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		[]gocloak.Role{
			*role,
		})
	require.NoError(t, err)

	roles, err := client.GetRealmRolesByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID)
	require.NoError(t, err, "GetRealmRolesByGroupID failed")

	require.Len(t, roles, 1, "GetRealmRolesByGroupID failed")
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

	role, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName,
	)
	require.NoError(t, err, "Can't get just created role with GetRealmRole")

	err = client.AddRealmRoleComposite(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		compositeRoleName,
		[]gocloak.Role{*role},
	)
	require.NoError(t, err)

	err = client.DeleteRealmRoleComposite(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		compositeRoleName,
		[]gocloak.Role{*role},
	)
	require.NoError(t, err)
}

// -----
// Users
// -----

func CreateUser(t *testing.T, client gocloak.GoCloak) (func(), string) {
	cfg := GetConfig(t)
	token := GetAdminToken(t, client)

	user := gocloak.User{
		FirstName: GetRandomNameP("FirstName"),
		LastName:  GetRandomNameP("LastName"),
		Email:     gocloak.StringP(GetRandomName("email") + "@localhost"),
		Enabled:   gocloak.BoolP(true),
		Attributes: &map[string][]string{
			"foo": {"bar", "alice", "bob", "roflcopter"},
			"bar": {"baz"},
		},
	}
	user.Username = user.Email

	userID, err := client.CreateUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		user)
	require.NoError(t, err, "CreateUser failed")
	user.ID = &userID
	t.Logf("Created User: %+v", user)
	tearDown := func() {
		err := client.DeleteUser(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			userID)
		require.NoError(t, err, "DeleteUser")
	}

	return tearDown, userID
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

	fetchedUser, err := client.GetUserByID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err, "GetUserByID failed")
	require.NotNil(t, fetchedUser.Attributes)
	ok := gocloak.UserAttributeContains(*fetchedUser.Attributes, "foo", "alice")
	require.False(t, !ok, "User doesn't have custom attributes")
	ok = gocloak.UserAttributeContains(*fetchedUser.Attributes, "foo2", "alice")
	require.False(t, ok, "User's custom attributes contains unexpected attribute")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err, "GetUserById failed")
	t.Log(fetchedUser)
}

func TestGocloak_GetUsers(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	users, err := client.GetUsers(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetUsersParams{
			Username: &cfg.GoCloak.UserName,
		})
	require.NoError(t, err, "GetUsers failed")
	t.Log(users)
}

func TestGocloak_GetUserCount(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	count, err := client.GetUserCount(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetUsersParams{})

	t.Logf("Users in Realm: %d", count)
	require.NoError(t, err, "GetUserCount failed")
}

func TestGocloak_GetGroupsCount(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	count, err := client.GetGroupsCount(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetGroupsParams{})
	t.Logf("Groups in Realm: %d", count)
	require.NoError(t, err, "GetGroupsCount failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	require.NoError(t, err, "AddUserToGroup failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	require.NoError(t, err, "AddUserToGroup failed")
	err = client.DeleteUserFromGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	require.NoError(t, err, "DeleteUserFromGroup failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		groupID,
	)
	require.NoError(t, err)
	groups, err := client.GetUserGroups(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		gocloak.GetGroupsParams{})
	require.NoError(t, err)
	require.NotEmpty(t, groups)
	require.Equal(t, groupID, *groups[0].ID)
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err, "GetUserByID failed")
	user.FirstName = GetRandomNameP("UpdateUserFirstName")
	err = client.UpdateUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*user)
	require.NoError(t, err, "UpdateUser failed")
}

func TestGocloak_UpdateUserSetEmptyRequiredActions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	_, userID := CreateUser(t, client)
	// tearDown, userID := CreateUser(t, client)
	// defer tearDown()

	user, err := client.GetUserByID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err, "GetUserByID failed")
	user.RequiredActions = &[]string{"VERIFY_EMAIL"}
	err = client.UpdateUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*user)
	require.NoError(t, err, "UpdateUser failed")

	user, err = client.GetUserByID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err, "GetUserByID failed")
	require.False(t, gocloak.NilOrEmptySlice(user.RequiredActions))
	require.Contains(t, *user.RequiredActions, "VERIFY_EMAIL")

	user.RequiredActions = &[]string{""}
	err = client.UpdateUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*user)
	require.NoError(t, err, "UpdateUser failed")

	user, err = client.GetUserByID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err, "GetUserByID failed")
	require.True(t, gocloak.NilOrEmptySlice(user.RequiredActions))
}

func TestGocloak_UpdateUserSetEmptyEmail(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, userID := CreateUser(t, client)
	defer tearDown()
	user, err := client.GetUserByID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	require.NoError(t, err)
	user.Email = gocloak.StringP("")
	err = client.UpdateUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*user)
	require.NoError(t, err)
	user, err = client.GetUserByID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	require.NoError(t, err)
	require.Nil(t, user.Email)
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	require.NoError(t, err)
	err = client.AddRealmRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]gocloak.Role{
			*role,
		})
	require.NoError(t, err)

	users, err := client.GetUsersByRoleName(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName)
	require.NoError(t, err)
	require.NotEmpty(t, users)
	require.Equal(t, userID, *users[0].ID)
}

func TestGocloak_GetUsersByClientRoleName(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	tearDownRole, roleName := CreateClientRole(t, client)
	defer tearDownRole()

	role, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName)
	require.NoError(t, err)
	err = client.AddClientRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		userID,
		[]gocloak.Role{*role},
	)
	require.NoError(t, err)

	users, err := client.GetUsersByClientRoleName(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName,
		gocloak.GetUsersByRoleParams{})
	require.NoError(t, err)
	require.NotEmpty(t, users)
	require.Equal(t, userID, *users[0].ID)
}

func TestGocloak_GetUserSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		context.Background(),
		cfg.GoCloak.Realm,
		gocloak.TokenOptions{
			ClientID:     &cfg.GoCloak.ClientID,
			ClientSecret: &cfg.GoCloak.ClientSecret,
			Username:     &cfg.GoCloak.UserName,
			Password:     &cfg.GoCloak.Password,
			GrantType:    gocloak.StringP("password"),
		},
	)
	require.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetUserSessions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
	)
	require.NoError(t, err, "GetUserSessions failed")
	require.NotEmpty(t, sessions, "GetUserSessions returned an empty list")
}

func TestGocloak_GetUserOfflineSessionsForClient(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		context.Background(),
		cfg.GoCloak.Realm,
		gocloak.TokenOptions{
			ClientID:      &cfg.GoCloak.ClientID,
			ClientSecret:  &cfg.GoCloak.ClientSecret,
			Username:      &cfg.GoCloak.UserName,
			Password:      &cfg.GoCloak.Password,
			GrantType:     gocloak.StringP("password"),
			ResponseTypes: &[]string{"token", "id_token"},
			Scopes:        &[]string{"openid", "offline_access"},
		},
	)
	require.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetUserOfflineSessionsForClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
		gocloakClientID,
	)
	require.NoError(t, err, "GetUserOfflineSessionsForClient failed")
	require.NotEmpty(t, sessions, "GetUserOfflineSessionsForClient returned an empty list")
}

func TestGocloak_GetClientUserSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		context.Background(),
		cfg.GoCloak.Realm,
		gocloak.TokenOptions{
			ClientID:     &cfg.GoCloak.ClientID,
			ClientSecret: &cfg.GoCloak.ClientSecret,
			Username:     &cfg.GoCloak.UserName,
			Password:     &cfg.GoCloak.Password,
			GrantType:    gocloak.StringP("password"),
		},
	)
	require.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetClientUserSessions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientUserSessions failed")
	require.NotEmpty(t, sessions, "GetClientUserSessions returned an empty list")
}

func findProtocolMapperByID(t *testing.T, client *gocloak.Client, id string) *gocloak.ProtocolMapperRepresentation {
	require.NotNil(t, client.ProtocolMappers)
	for _, protocolMapper := range *client.ProtocolMappers {
		if gocloak.NilOrEmpty(protocolMapper.ID) {
			continue
		}
		if *protocolMapper.ID == id {
			return &protocolMapper
		}
	}
	return nil
}

func TestGocloak_CreateUpdateDeleteClientProtocolMapper(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	id := GetRandomName("protocol-mapper-id-")

	testClient := GetClientByClientID(t, client, cfg.GoCloak.ClientID)
	require.Nil(
		t,
		findProtocolMapperByID(t, testClient, id),
		"default client should not have a protocol mapper with ID: %s", id,
	)

	token := GetAdminToken(t, client)
	createdID, err := client.CreateClientProtocolMapper(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*testClient.ID,
		gocloak.ProtocolMapperRepresentation{
			ID:             &id,
			Name:           gocloak.StringP("test"),
			Protocol:       gocloak.StringP("openid-connect"),
			ProtocolMapper: gocloak.StringP("oidc-usermodel-attribute-mapper"),
			Config: &map[string]string{
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
	require.NoError(t, err, "CreateClientProtocolMapper failed")
	require.Equal(t, id, createdID)

	testClientAfter := GetClientByClientID(t, client, cfg.GoCloak.ClientID)
	require.NotNil(
		t,
		findProtocolMapperByID(t, testClientAfter, id),
		"protocol mapper has not been created",
	)

	err = client.UpdateClientProtocolMapper(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*testClient.ID,
		createdID,
		gocloak.ProtocolMapperRepresentation{
			ID:             &id,
			Name:           gocloak.StringP("test"),
			Protocol:       gocloak.StringP("openid-connect"),
			ProtocolMapper: gocloak.StringP("oidc-usermodel-attribute-mapper"),
			Config: &map[string]string{
				"access.token.claim":   "true",
				"aggregate.attrs":      "",
				"claim.name":           "testUpdated",
				"id.token.claim":       "true",
				"jsonType.label":       "String",
				"multivalued":          "",
				"user.attribute":       "test",
				"userinfo.token.claim": "true",
			},
		},
	)
	require.NoError(t, err, "UpdateClientProtocolMapper failed")

	testClientAfterUpdate := GetClientByClientID(t, client, cfg.GoCloak.ClientID)
	mapper := findProtocolMapperByID(t, testClientAfterUpdate, id)
	require.NotNil(t, mapper)
	mapperConfig := *mapper.Config
	require.Equal(
		t,
		mapperConfig["claim.name"],
		"testUpdated",
	)

	err = client.DeleteClientProtocolMapper(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*testClient.ID,
		id,
	)
	require.NoError(t, err, "DeleteClientProtocolMapper failed")

	testClientAgain := GetClientByClientID(t, client, cfg.GoCloak.ClientID)
	require.Nil(
		t,
		findProtocolMapperByID(t, testClientAgain, id),
		"default client should not have a protocol mapper with ID: %s", id,
	)
}

func TestGocloak_GetClientOfflineSessions(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	SetUpTestUser(t, client)
	_, err := client.GetToken(
		context.Background(),
		cfg.GoCloak.Realm,
		gocloak.TokenOptions{
			ClientID:      &cfg.GoCloak.ClientID,
			ClientSecret:  &cfg.GoCloak.ClientSecret,
			Username:      &cfg.GoCloak.UserName,
			Password:      &cfg.GoCloak.Password,
			GrantType:     gocloak.StringP("password"),
			ResponseTypes: &[]string{"token", "id_token"},
			Scopes:        &[]string{"openid", "offline_access"},
		},
	)
	require.NoError(t, err, "Login failed")
	token := GetAdminToken(t, client)
	sessions, err := client.GetClientOfflineSessions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	require.NoError(t, err, "GetClientOfflineSessions failed")
	require.NotEmpty(t, sessions, "GetClientOfflineSessions returned an empty list")
}

func TestGoCloak_ClientSecret(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	testClient := gocloak.Client{
		ID:                      GetRandomNameP("gocloak-client-id-"),
		ClientID:                GetRandomNameP("gocloak-client-secret-client-id-"),
		Secret:                  gocloak.StringP("initial-secret-key"),
		ServiceAccountsEnabled:  gocloak.BoolP(true),
		StandardFlowEnabled:     gocloak.BoolP(true),
		Enabled:                 gocloak.BoolP(true),
		FullScopeAllowed:        gocloak.BoolP(true),
		Protocol:                gocloak.StringP("openid-connect"),
		RedirectURIs:            &[]string{"localhost"},
		ClientAuthenticatorType: gocloak.StringP("client-secret"),
	}

	tearDown, clientID := CreateClient(t, client, &testClient)
	defer tearDown()
	require.Equal(t, *testClient.ID, clientID)

	oldCreds, err := client.GetClientSecret(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
	)
	require.NoError(t, err, "GetClientSecret failed")

	regeneratedCreds, err := client.RegenerateClientSecret(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
	)
	require.NoError(t, err, "RegenerateClientSecret failed")
	require.NotEqual(t, *oldCreds.Value, *regeneratedCreds.Value)

	err = client.DeleteClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		clientID,
	)
	require.NoError(t, err, "DeleteClient failed")
}

func TestGoCloak_ClientServiceAccount(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	serviceAccount, err := client.GetClientServiceAccount(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
	)
	require.NoError(t, err)

	require.NotNil(t, serviceAccount.ID)
	require.NotNil(t, serviceAccount.Username)
	require.NotEqual(t, gocloakClientID, *(serviceAccount.ID))
	require.Equal(t, "service-account-gocloak", *(serviceAccount.Username))
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName1,
	)
	require.NoError(t, err, "GetClientRole failed")
	tearDown2, roleName2 := CreateClientRole(t, client)
	defer tearDown2()
	role2, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName2,
	)
	require.NoError(t, err, "GetClientRole failed")
	roles := []gocloak.Role{*role1, *role2}
	err = client.AddClientRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		testUserID,
		roles,
	)
	require.NoError(t, err, "AddClientRoleToUser failed")

	err = client.DeleteClientRoleFromUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		testUserID,
		roles,
	)
	require.NoError(t, err, "DeleteClientRoleFromUser failed")
}

func TestGocloak_GetClientRolesByUserID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()
	tearDownRole, roleName := CreateClientRole(t, client)
	defer tearDownRole()
	role, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName)
	require.NoError(t, err)

	err = client.AddClientRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		userID,
		[]gocloak.Role{*role},
	)
	require.NoError(t, err)

	roles, err := client.GetClientRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	var found bool
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role.Name {
			found = true
			break
		}
	}
	require.True(t, found, "The role has not been found in the assigned roles. Role: %+v", *role)

	roles, err = client.GetCompositeClientRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role.Name {
			return
		}
	}
	require.Fail(t, "The role has not been found in the assigned composite roles. Role: %+v", *role)
}

func TestGoCloak_GetAvailableClientRolesByUserID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()
	tearDownRole, roleName1 := CreateClientRole(t, client)
	defer tearDownRole()
	tearDownRole2, roleName2 := CreateClientRole(t, client)
	defer tearDownRole2()

	role1, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName1)
	require.NoError(t, err)

	role2, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName2)
	require.NoError(t, err)

	err = client.AddClientRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		userID,
		[]gocloak.Role{*role1},
	)
	require.NoError(t, err)

	roles, err := client.GetClientRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	var found bool
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role1.Name {
			found = true
			break
		}
	}
	require.True(t, found, "The role1 has not been found in the assigned roles. Role: %+v", *role1)

	roles, err = client.GetAvailableClientRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role2.Name {
			return
		}
	}
	require.Fail(t, "The role2 has not been found in the assigned composite roles. Role: %+v", *role2)
}

func TestGoCloak_GetAvailableRealmRolesByUserID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()
	tearDownRole, roleName1 := CreateRealmRole(t, client)
	defer tearDownRole()
	tearDownRole2, roleName2 := CreateRealmRole(t, client)
	defer tearDownRole2()

	role1, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName1)
	require.NoError(t, err)

	role2, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName2)
	require.NoError(t, err)

	err = client.AddRealmRoleToUser(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		[]gocloak.Role{*role1},
	)
	require.NoError(t, err)

	roles, err := client.GetRealmRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	var found bool
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role1.Name {
			found = true
			break
		}
	}
	require.True(t, found, "The role1 has not been found in the assigned roles. Role: %+v", *role1)

	roles, err = client.GetAvailableRealmRolesByUserID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID)
	require.NoError(t, err)
	t.Logf("User roles: %+v", roles)
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role2.Name {
			return
		}
	}
	require.Fail(t, "The role2 has not been found in the assigned composite roles. Role: %+v", *role2)
}

func TestGoCloak_GetAvailableClientRolesByGroupID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()
	tearDownRole, roleName1 := CreateClientRole(t, client)
	defer tearDownRole()
	tearDownRole2, roleName2 := CreateClientRole(t, client)
	defer tearDownRole2()

	role1, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName1)
	require.NoError(t, err)

	role2, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName2)
	require.NoError(t, err)

	err = client.AddClientRoleToGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID,
		[]gocloak.Role{*role1},
	)
	require.NoError(t, err)

	roles, err := client.GetClientRolesByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID)
	require.NoError(t, err)
	t.Logf("Group roles: %+v", roles)
	var found bool
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role1.Name {
			found = true
			break
		}
	}
	require.True(t, found, "The role1 has not been found in the assigned roles. Role: %+v", *role1)

	roles, err = client.GetAvailableClientRolesByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID)
	require.NoError(t, err)
	t.Logf("Group roles: %+v", roles)
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role2.Name {
			return
		}
	}
	require.Fail(t, "The role2 has not been found in the assigned composite roles. Role: %+v", *role2)
}

func TestGoCloak_GetAvailableRealmRolesByGroupID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()
	tearDownRole, roleName1 := CreateRealmRole(t, client)
	defer tearDownRole()
	tearDownRole2, roleName2 := CreateRealmRole(t, client)
	defer tearDownRole2()

	role1, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName1)
	require.NoError(t, err)

	role2, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		roleName2)
	require.NoError(t, err)

	err = client.AddRealmRoleToGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID,
		[]gocloak.Role{*role1},
	)
	require.NoError(t, err)

	roles, err := client.GetRealmRolesByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID)
	require.NoError(t, err)

	t.Logf("Group roles: %+v", roles)
	var found bool
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role1.Name {
			found = true
			break
		}
	}
	require.True(t, found, "The role1 has not been found in the assigned roles. Role: %+v", *role1)

	roles, err = client.GetAvailableRealmRolesByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		groupID)
	require.NoError(t, err)
	t.Logf("Group roles: %+v", roles)
	for _, r := range roles {
		if r.Name == nil {
			continue
		}
		if *r.Name == *role2.Name {
			return
		}
	}
	require.Fail(t, "The role2 has not been found in the assigned composite roles. Role: %+v", *role2)
}

func TestGocloak_GetClientRolesByGroupID(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, groupID := CreateGroup(t, client)
	defer tearDown()

	_, err := client.GetClientRolesByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID)
	require.NoError(t, err, "GetClientRolesByGroupID failed")

	_, err = client.GetCompositeClientRolesByGroupID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID)
	require.NoError(t, err, "GetCompositeClientRolesByGroupID failed")
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName1,
	)
	require.NoError(t, err, "GetClientRole failed")
	tearDown2, roleName2 := CreateClientRole(t, client)
	defer tearDown2()
	role2, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		roleName2,
	)
	require.NoError(t, err, "GetClientRole failed")

	tearDownGroup, groupID := CreateGroup(t, client)
	defer tearDownGroup()

	roles := []gocloak.Role{*role1, *role2}
	err = client.AddClientRoleToGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID,
		roles,
	)
	require.NoError(t, err, "AddClientRoleToGroup failed")

	err = client.DeleteClientRoleFromGroup(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		groupID,
		roles,
	)
	require.NoError(t, err, "DeleteClientRoleFromGroup failed")
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

	compositeRoleModel, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		compositeRole,
	)
	require.NoError(t, err, "Can't get just created role with GetClientRole")

	roleModel, err := client.GetClientRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		role,
	)
	require.NoError(t, err, "Can't get just created role with GetClientRole")

	err = client.AddClientRoleComposite(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*compositeRoleModel.ID,
		[]gocloak.Role{*roleModel},
	)
	require.NoError(t, err, "AddClientRoleComposite failed")

	compositeRoles, err := client.GetCompositeClientRolesByRoleID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*compositeRoleModel.ID,
	)
	require.NoError(t, err, "GetCompositeClientRolesByRoleID failed")
	require.GreaterOrEqual(t, len(compositeRoles), 1, "GetCompositeClientRolesByRoleID didn't return any composite roles")
	require.Equal(t, *(roleModel.ID), *(compositeRoles[0].ID), "GetCompositeClientRolesByRoleID returned wrong composite role")

	err = client.DeleteClientRoleComposite(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*compositeRoleModel.ID,
		[]gocloak.Role{*roleModel},
	)
	require.NoError(t, err, "DeleteClientRoleComposite failed")
}

func TestGocloak_AddDeleteRealmRoleComposite(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDown, compositeRole := CreateRealmRole(t, client)
	defer tearDown()

	tearDown, role := CreateRealmRole(t, client)
	defer tearDown()

	compositeRoleModel, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		compositeRole,
	)
	require.NoError(t, err, "Can't get just created role with GetRealmRole")

	roleModel, err := client.GetRealmRole(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		role,
	)
	require.NoError(t, err, "Can't get just created role with GetRealmRole")

	err = client.AddRealmRoleComposite(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*compositeRoleModel.Name,
		[]gocloak.Role{*roleModel},
	)
	require.NoError(t, err, "AddRealmRoleComposite failed")

	compositeRoles, err := client.GetCompositeRealmRolesByRoleID(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*compositeRoleModel.ID,
	)
	require.NoError(t, err, "GetCompositeRealmRolesByRoleID failed")
	require.GreaterOrEqual(t, len(compositeRoles), 1, "GetCompositeRealmRolesByRoleID didn't return any composite roles")
	require.Equal(t, *(roleModel.ID), *(compositeRoles[0].ID), "GetCompositeRealmRolesByRoleID returned wrong composite role")

	err = client.DeleteRealmRoleComposite(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*compositeRoleModel.Name,
		[]gocloak.Role{*roleModel},
	)
	require.NoError(t, err, "DeleteRealmRoleComposite failed")
}

func TestGocloak_CreateGetDeleteUserFederatedIdentity(t *testing.T) {
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	idp := "google"
	idprep := gocloak.IdentityProviderRepresentation{
		ProviderID:                &idp,
		Alias:                     gocloak.StringP("google"),
		DisplayName:               gocloak.StringP("Google"),
		Enabled:                   gocloak.BoolP(true),
		TrustEmail:                gocloak.BoolP(true),
		FirstBrokerLoginFlowAlias: gocloak.StringP("first broker login"),
		Config: &map[string]string{
			"clientId":     cfg.GoCloak.ClientID,
			"clientSecret": cfg.GoCloak.ClientSecret,
			"hostedDomain": "test.io",
		},
	}
	res, err := client.CreateIdentityProvider(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		idprep,
	)
	require.NoError(t, err)
	require.Equal(t, idp, res)

	defer func() {
		err = client.DeleteIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			"google",
		)
		require.NoError(t, err)
	}()

	firep := gocloak.FederatedIdentityRepresentation{
		IdentityProvider: &idp,
		UserID:           gocloak.StringP("my-external-userid"),
		UserName:         gocloak.StringP("my-external-username"),
	}
	err = client.CreateUserFederatedIdentity(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		idp,
		firep,
	)
	require.NoError(t, err)
	require.Equal(t, idp, res)

	defer func() {
		err = client.DeleteUserFederatedIdentity(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			userID,
			idp,
		)
		require.NoError(t, err)
	}()

	arr, err := client.GetUserFederatedIdentities(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(arr))
	require.Equal(t, "my-external-userid", *arr[0].UserID)
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.ClientScope{
			ID:          &id,
			Name:        gocloak.StringP("test-scope"),
			Description: gocloak.StringP("testing scope"),
			Protocol:    gocloak.StringP("openid-connect"),
			ClientScopeAttributes: &gocloak.ClientScopeAttributes{
				ConsentScreenText:      gocloak.StringP("false"),
				DisplayOnConsentScreen: gocloak.StringP("true"),
				IncludeInTokenScope:    gocloak.StringP("false"),
			},
			ProtocolMappers: &[]gocloak.ProtocolMappers{
				{
					ID:              &rolemapperID,
					Name:            gocloak.StringP("roles"),
					Protocol:        gocloak.StringP("openid-connect"),
					ProtocolMapper:  gocloak.StringP("oidc-usermodel-client-role-mapper"),
					ConsentRequired: gocloak.BoolP(false),
					ProtocolMappersConfig: &gocloak.ProtocolMappersConfig{
						UserinfoTokenClaim:                 gocloak.StringP("false"),
						AccessTokenClaim:                   gocloak.StringP("true"),
						IDTokenClaim:                       gocloak.StringP("true"),
						ClaimName:                          gocloak.StringP("test"),
						Multivalued:                        gocloak.StringP("true"),
						UsermodelClientRoleMappingClientID: gocloak.StringP("test"),
					},
				},
				{
					ID:              &audiencemapperID,
					Name:            gocloak.StringP("audience"),
					Protocol:        gocloak.StringP("openid-connect"),
					ProtocolMapper:  gocloak.StringP("oidc-audience-mapper"),
					ConsentRequired: gocloak.BoolP(false),
					ProtocolMappersConfig: &gocloak.ProtocolMappersConfig{
						UserinfoTokenClaim:     gocloak.StringP("false"),
						IDTokenClaim:           gocloak.StringP("true"),
						AccessTokenClaim:       gocloak.StringP("true"),
						IncludedClientAudience: gocloak.StringP("test"),
					},
				},
			},
		},
	)
	require.NoError(t, err, "CreateClientScope failed")
	require.Equal(t, id, createdID)
	clientScopeActual, err := client.GetClientScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		id,
	)
	require.NoError(t, err)

	require.NotNil(t, clientScopeActual, "client scope has not been created")
	require.Len(t, *clientScopeActual.ProtocolMappers, 2, "unexpected number of protocol mappers created")
	err = client.DeleteClientScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		id,
	)
	require.NoError(t, err, "DeleteClientScope failed")
	clientScopeActual, err = client.GetClientScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		id,
	)
	require.EqualError(t, err, "404 Not Found: Could not find client scope")
	require.Nil(t, clientScopeActual, "client scope has not been deleted")
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
		repr := gocloak.IdentityProviderRepresentation{
			Alias:                     gocloak.StringP("google"),
			DisplayName:               gocloak.StringP("Google"),
			Enabled:                   gocloak.BoolP(true),
			ProviderID:                gocloak.StringP("google"),
			TrustEmail:                gocloak.BoolP(true),
			FirstBrokerLoginFlowAlias: gocloak.StringP("first broker login"),
			Config: &map[string]string{
				"clientId":     cfg.GoCloak.ClientID,
				"clientSecret": cfg.GoCloak.ClientSecret,
				"hostedDomain": "test.io",
			},
		}
		provider, err := client.CreateIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			repr,
		)
		require.NoError(t, err)
		require.Equal(t, "google", provider)
	})

	t.Run("create github provider", func(t *testing.T) {
		repr := gocloak.IdentityProviderRepresentation{
			Alias:                     gocloak.StringP("github"),
			DisplayName:               gocloak.StringP("GitHub"),
			Enabled:                   gocloak.BoolP(true),
			ProviderID:                gocloak.StringP("github"),
			TrustEmail:                gocloak.BoolP(true),
			FirstBrokerLoginFlowAlias: gocloak.StringP("first broker login"),
			Config: &map[string]string{
				"clientId":     cfg.GoCloak.ClientID,
				"clientSecret": cfg.GoCloak.ClientSecret,
			},
		}
		provider, err := client.CreateIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			repr,
		)
		require.NoError(t, err)
		require.Equal(t, "github", provider)
	})

	t.Run("create microsoft provider", func(t *testing.T) {
		repr := gocloak.IdentityProviderRepresentation{
			Alias:                     gocloak.StringP("microsoft"),
			DisplayName:               gocloak.StringP("Microsoft"),
			Enabled:                   gocloak.BoolP(true),
			ProviderID:                gocloak.StringP("microsoft"),
			TrustEmail:                gocloak.BoolP(true),
			FirstBrokerLoginFlowAlias: gocloak.StringP("first broker login"),
			Config: &map[string]string{
				"clientId":     cfg.GoCloak.ClientID,
				"clientSecret": cfg.GoCloak.ClientSecret,
			},
		}
		provider, err := client.CreateIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			repr,
		)
		require.NoError(t, err)
		require.Equal(t, "microsoft", provider)
	})

	t.Run("Update google provider", func(t *testing.T) {
		repr := gocloak.IdentityProviderRepresentation{
			Alias:                     gocloak.StringP("google"),
			DisplayName:               gocloak.StringP("Google"),
			Enabled:                   gocloak.BoolP(true),
			ProviderID:                gocloak.StringP("google"),
			TrustEmail:                gocloak.BoolP(true),
			FirstBrokerLoginFlowAlias: gocloak.StringP("first broker login"),
			Config: &map[string]string{
				"clientId":     cfg.GoCloak.ClientID,
				"clientSecret": cfg.GoCloak.ClientSecret,
				"hostedDomain": "updated-test.io",
			},
		}
		err := client.UpdateIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			"google",
			repr,
		)
		require.NoError(t, err)

		// listing identity providers here must now show three
		providers, err := client.GetIdentityProviders(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
		)
		require.NoError(t, err)
		require.Equal(t, 3, len(providers))
	})

	t.Run("Delete google provider", func(t *testing.T) {
		err := client.DeleteIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			"google",
		)
		require.NoError(t, err)
	})

	t.Run("List providers", func(t *testing.T) {
		providers, err := client.GetIdentityProviders(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
		)
		require.NoError(t, err)
		require.Equal(t, 2, len(providers))
	})

	t.Run("Get microsoft provider", func(t *testing.T) {
		provider, err := client.GetIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			"microsoft",
		)
		require.NoError(t, err)
		require.Equal(t, "microsoft", *(provider.Alias))
	})

	t.Run("Delete microsoft provider", func(t *testing.T) {
		err := client.DeleteIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			"microsoft",
		)
		require.NoError(t, err)
	})

	t.Run("Delete github provider", func(t *testing.T) {
		err := client.DeleteIdentityProvider(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			"github",
		)
		require.NoError(t, err)
	})
}

// -----------------
// Protection API
// -----------------

func TestGocloak_ErrorsCreateListGetUpdateDeleteResourceClient(t *testing.T) {

	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)
	token.AccessToken = "" // force unauthorized access attempts

	// Create
	tearDown, resourceID := CreateResourceClient(t, client)
	// Delete
	defer tearDown()

	// List
	_, err := client.GetResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		resourceID,
	)

	require.Error(t, err, "GetResource no error on unauthorized request")

	// Looking for a created resource
	_, err = client.GetResourcesClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetResourceParams{
			Name: gocloak.StringP("nothing"),
		},
	)
	require.Error(t, err, "GetResources no error on unauthorized request")

	err = client.UpdateResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.ResourceRepresentation{},
	)
	require.Error(t, err, "UpdateResourceClient no error on missing ID of the resource")
	emptyResource := gocloak.ResourceRepresentation{}
	err = client.UpdateResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		emptyResource,
	)
	require.Error(t, err, "UpdateResourceClient no error on unauthorized request")

}

func TestGocloak_CreateListGetUpdateDeleteResourceClient(t *testing.T) {

	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	// Create
	tearDown, resourceID := CreateResourceClient(t, client)
	// Delete
	defer tearDown()

	// List
	createdResource, err := client.GetResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		resourceID,
	)

	require.NoError(t, err, "GetResource failed")
	t.Logf("Created Resource: %+v", *(createdResource.ID))
	require.Equal(t, resourceID, *(createdResource.ID))

	// Looking for a created resource
	resources, err := client.GetResourcesClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetResourceParams{
			Name: createdResource.Name,
		},
	)
	require.NoError(t, err, "GetResources failed")
	require.Len(t, resources, 1, "GetResources should return exact 1 resource")
	require.Equal(t, *(createdResource.ID), *(resources[0].ID))
	t.Logf("Resources: %+v", resources)

	err = client.UpdateResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.ResourceRepresentation{},
	)
	require.Error(t, err, "Should fail because of missing ID of the resource")

	createdResource.Name = GetRandomNameP("ResourceName")

	err = client.UpdateResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*createdResource,
	)
	require.NoError(t, err, "UpdateResource failed")

	updatedResource, err := client.GetResourceClient(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		resourceID,
	)
	require.NoError(t, err, "GetResource failed")
	require.Equal(t, *(createdResource.Name), *(updatedResource.Name))

}

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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		resourceID,
	)

	require.NoError(t, err, "GetResource failed")
	t.Logf("Created Resource: %+v", *(createdResource.ID))
	require.Equal(t, resourceID, *(createdResource.ID))

	// Looking for a created resource
	resources, err := client.GetResources(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.GetResourceParams{
			Name: createdResource.Name,
		},
	)
	require.NoError(t, err, "GetResources failed")
	require.Len(t, resources, 1, "GetResources should return exact 1 resource")
	require.Equal(t, *(createdResource.ID), *(resources[0].ID))
	t.Logf("Resources: %+v", resources)

	err = client.UpdateResource(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.ResourceRepresentation{},
	)
	require.Error(t, err, "Should fail because of missing ID of the resource")

	createdResource.Name = GetRandomNameP("ResourceName")
	err = client.UpdateResource(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdResource,
	)
	require.NoError(t, err, "UpdateResource failed")

	updatedResource, err := client.GetResource(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		resourceID,
	)
	require.NoError(t, err, "GetResource failed")
	require.Equal(t, *(createdResource.Name), *(updatedResource.Name))
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
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	require.NoError(t, err, "GetScope failed")
	t.Logf("Created Scope: %+v", *(createdScope.ID))
	require.Equal(t, scopeID, *(createdScope.ID))

	// Looking for a created scope
	scopes, err := client.GetScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.GetScopeParams{
			Name: createdScope.Name,
		},
	)
	require.NoError(t, err, "GetScopes failed")
	require.Len(t, scopes, 1, "GetScopes should return exact 1 scope")
	require.Equal(t, *(createdScope.ID), *(scopes[0].ID))
	t.Logf("Scopes: %+v", scopes)

	err = client.UpdateScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.ScopeRepresentation{},
	)
	require.Error(t, err, "Should fail because of missing ID of the scope")

	createdScope.Name = GetRandomNameP("ScopeName")
	err = client.UpdateScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdScope,
	)
	require.NoError(t, err, "UpdateScope failed")

	updatedScope, err := client.GetScope(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		scopeID,
	)
	require.NoError(t, err, "GetScope failed")
	require.Equal(t, *(createdScope.Name), *(updatedScope.Name))
}

func TestGocloak_CreateListGetUpdateDeletePolicy(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	// Create
	tearDown, policyID := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("Policy Description"),
		Type:        gocloak.StringP("js"),
		Logic:       gocloak.NEGATIVE,
		JSPolicyRepresentation: gocloak.JSPolicyRepresentation{
			Code: gocloak.StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDown()

	// List
	createdPolicy, err := client.GetPolicy(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		policyID,
	)
	require.NoError(t, err, "GetPolicy failed")
	t.Logf("Created Policy: %+v", *(createdPolicy.ID))
	require.Equal(t, policyID, *(createdPolicy.ID))

	// Looking for a created policy
	policies, err := client.GetPolicies(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.GetPolicyParams{
			Name: createdPolicy.Name,
		},
	)
	require.NoError(t, err, "GetPolicies failed")
	require.Len(t, policies, 1, "GetPolicies should return exact 1 policy")
	require.Equal(t, *(createdPolicy.ID), *(policies[0].ID))
	t.Logf("Policies: %+v", policies)

	// Looking for a created policy using type
	policies, err = client.GetPolicies(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.GetPolicyParams{
			Name: createdPolicy.Name,
			Type: gocloak.StringP("js"),
		},
	)
	require.NoError(t, err, "GetPolicies failed")
	require.Len(t, policies, 1, "GetPolicies should return exact 1 policy")
	require.Equal(t, *(createdPolicy.ID), *(policies[0].ID))
	t.Logf("Policies: %+v", policies)

	err = client.UpdatePolicy(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.PolicyRepresentation{},
	)
	require.Error(t, err, "Should fail because of missing ID of the policy")

	createdPolicy.Name = GetRandomNameP("PolicyName")
	err = client.UpdatePolicy(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdPolicy,
	)
	require.NoError(t, err, "UpdatePolicy failed")

	updatedPolicy, err := client.GetPolicy(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		policyID,
	)
	require.NoError(t, err, "GetPolicy failed")
	require.Equal(t, *(createdPolicy.Name), *(updatedPolicy.Name))
}

func TestGocloak_CreateGetUpdateDeleteResourcePolicy(t *testing.T) {

	// parallel is causing intermittent conflict with role-based test GetClientScopeMappingsClientRolesAvailable
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)
	adminToken := GetAdminToken(t, client)

	tearDownResource, resourceID := CreateResourceClientWithScopes(t, client)
	defer tearDownResource()

	roleName := GetRandomName("editor")
	role := gocloak.Role{
		Name: &roleName,
	}

	roleID, err := client.CreateClientRole(context.Background(), adminToken.AccessToken, cfg.GoCloak.Realm, gocloakClientID, role)

	defer func() {
		err := client.DeleteClientRole(context.Background(), adminToken.AccessToken, cfg.GoCloak.Realm, gocloakClientID, roleName)
		require.NoError(t, err, "could not delete client role")
	}()

	require.NoError(t, err, "could not create client role")
	t.Logf("Created ClientRole: %+v", roleID)

	scopes := []string{"message-post"}
	policyNameP := GetRandomNameP("PolicyName")
	policy := gocloak.ResourcePolicyRepresentation{
		Name:        policyNameP,
		Description: gocloak.StringP("Role Policy"),
		Scopes:      &scopes,
		Roles:       &[]string{roleName},
	}

	result, err := client.CreateResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, resourceID, policy)
	require.NoError(t, err, "could not create resource policy")
	require.Equal(t, "Role Policy", *(result.Description))

	result, err = client.GetResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, *(result.ID))
	require.NoError(t, err, "could not get resource policy")
	require.Equal(t, scopes, *(result.Scopes))

	newScopes := []string{"message-view"}
	result.Scopes = &newScopes

	err = client.UpdateResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, *(result.ID), *result)
	require.NoError(t, err, "could not get resource policy")

	result, err = client.GetResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, *(result.ID))
	require.NoError(t, err, "could not get resource policy")
	require.Equal(t, newScopes, *(result.Scopes))

	params := gocloak.GetResourcePoliciesParams{
		Scope: gocloak.StringP("message-view"),
	}
	policies, err := client.GetResourcePolicies(context.Background(), token.AccessToken, cfg.GoCloak.Realm, params)
	require.NoError(t, err, "could not get resource policies")
	require.Equal(t, 1, len(policies))
	require.False(t, policies[0] == nil)

	if len(policies) == 1 && policies[0] != nil {
		require.Equal(t, *policyNameP, *(policies[0].Name))
	}
	err = client.DeleteResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, *(result.ID))
	require.NoError(t, err, "could not delete resource policies")

	policies, err = client.GetResourcePolicies(context.Background(), token.AccessToken, cfg.GoCloak.Realm, params)
	require.NoError(t, err, "could not get resource policies")
	require.Equal(t, 0, len(policies))

	// Test error handling
	_, err = client.CreateResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, "", policy)
	require.Error(t, err, "should not create resource policy without resourceID")

	_, err = client.GetResourcePolicy(context.Background(), "", cfg.GoCloak.Realm, "asdfasdfasdfasdf")
	require.Error(t, err, "should not get resource policy without token")

	err = client.UpdateResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, "", policy)
	require.Error(t, err, "should not update resource policy without token")

	_, err = client.GetResourcePolicies(context.Background(), "", cfg.GoCloak.Realm, params)
	require.Error(t, err, "should not get resource policies without token")

	err = client.DeleteResourcePolicy(context.Background(), token.AccessToken, cfg.GoCloak.Realm, "")
	require.Error(t, err, "should not delete resource policy without permission ID")

}

func TestGocloak_RolePolicy(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	roles, err := client.GetRealmRoles(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "GetRealmRoles failed")
	require.GreaterOrEqual(t, len(roles), 1, "GetRealmRoles failed")

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("Role Policy"),
		Type:        gocloak.StringP("role"),
		Logic:       gocloak.NEGATIVE,
		RolePolicyRepresentation: gocloak.RolePolicyRepresentation{
			Roles: &[]gocloak.RoleDefinition{
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
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("JS Policy"),
		Type:        gocloak.StringP("js"),
		Logic:       gocloak.POSITIVE,
		JSPolicyRepresentation: gocloak.JSPolicyRepresentation{
			Code: gocloak.StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_ClientPolicy(t *testing.T) {
	t.Parallel()
	client := NewClientWithDebug(t)

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("Client Policy"),
		Type:        gocloak.StringP("client"),
		ClientPolicyRepresentation: gocloak.ClientPolicyRepresentation{
			Clients: &[]string{
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
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("Time Policy"),
		Type:        gocloak.StringP("time"),
		TimePolicyRepresentation: gocloak.TimePolicyRepresentation{
			NotBefore:    gocloak.StringP("2019-12-30 12:00:00"),
			NotOnOrAfter: gocloak.StringP("2020-12-30 12:00:00"),
			DayMonth:     gocloak.StringP("1"),
			DayMonthEnd:  gocloak.StringP("31"),
			Month:        gocloak.StringP("1"),
			MonthEnd:     gocloak.StringP("12"),
			Year:         gocloak.StringP("1900"),
			YearEnd:      gocloak.StringP("2100"),
			Hour:         gocloak.StringP("1"),
			HourEnd:      gocloak.StringP("24"),
			Minute:       gocloak.StringP("0"),
			MinuteEnd:    gocloak.StringP("60"),
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
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("User Policy"),
		Type:        gocloak.StringP("user"),
		UserPolicyRepresentation: gocloak.UserPolicyRepresentation{
			Users: &[]string{
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

	tearDownClient, clientPolicyID := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("Client Policy"),
		Type:        gocloak.StringP("client"),
		ClientPolicyRepresentation: gocloak.ClientPolicyRepresentation{
			Clients: &[]string{
				gocloakClientID,
			},
		},
	})
	defer tearDownClient()

	tearDownJS, jsPolicyID := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("JS Policy"),
		Type:        gocloak.StringP("js"),
		Logic:       gocloak.POSITIVE,
		JSPolicyRepresentation: gocloak.JSPolicyRepresentation{
			Code: gocloak.StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDownJS()

	// Create
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("Aggregated Policy"),
		Type:        gocloak.StringP("aggregate"),
		AggregatedPolicyRepresentation: gocloak.AggregatedPolicyRepresentation{
			Policies: &[]string{
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
	tearDown, _ := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("Group Policy"),
		Type:        gocloak.StringP("group"),
		GroupPolicyRepresentation: gocloak.GroupPolicyRepresentation{
			Groups: &[]gocloak.GroupDefinition{
				{
					ID: gocloak.StringP(groupID),
				},
			},
		},
	})
	// Delete
	defer tearDown()
}

func TestGocloak_ErrorsGrantGetUpdateDeleteUserPermission(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	tearDownResource, resourceID := CreateResourceClientWithScopes(t, client)
	defer tearDownResource()

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	// Grant
	scope := "read-private"

	permission := gocloak.PermissionGrantParams{
		RequesterID: &userID,
		ScopeName:   &scope,
	}
	_, err := client.GrantUserPermission(context.Background(), token.AccessToken, cfg.GoCloak.Realm, permission)
	require.Error(t, err, "GrantUserPermission no error on missing ResourceID")

	permission = gocloak.PermissionGrantParams{
		ResourceID: &resourceID,
		ScopeName:  &scope,
	}
	_, err = client.GrantUserPermission(context.Background(), token.AccessToken, cfg.GoCloak.Realm, permission)
	require.Error(t, err, "GrantUserPermission no error on missing RequesterID")

	permission = gocloak.PermissionGrantParams{
		ScopeName: &scope,
	}
	_, err = client.GrantUserPermission(context.Background(), token.AccessToken, cfg.GoCloak.Realm, permission)
	require.Error(t, err, "GrantUserPermission no error on missing Scope")

	permission = gocloak.PermissionGrantParams{
		ResourceID:  &resourceID,
		RequesterID: &userID,
		ScopeName:   &scope,
	}
	_, err = client.GrantUserPermission(context.Background(), "", cfg.GoCloak.Realm, permission)
	require.Error(t, err, "GrantUserPermission no error on unauthorized request")

	// Get
	params := gocloak.GetUserPermissionParams{
		ResourceID: &resourceID,
	}
	_, err = client.GetUserPermissions(context.Background(), "", cfg.GoCloak.Realm, params)
	require.Error(t, err, "GetUserPermission no error on unauthorized request")

	_, err = client.UpdateUserPermission(context.Background(), "", cfg.GoCloak.Realm, permission)
	require.Error(t, err, "UpdateUserPermission no error on unauthorized request")

	// Get (no permission expected to be returned)
	params = gocloak.GetUserPermissionParams{
		ResourceID: &resourceID,
	}
	_, err = client.GetUserPermissions(context.Background(), "", cfg.GoCloak.Realm, params)
	require.Error(t, err, "UpdateUserPermission no error on unauthorized request")

	// Delete
	err = client.DeleteUserPermission(context.Background(), "", cfg.GoCloak.Realm, "someID")
	require.Error(t, err, "DeleteUserPermission no error on unauthorized request")

}

func TestGocloak_GrantGetUpdateDeleteUserPermission(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	tearDownResource, resourceID := CreateResourceClientWithScopes(t, client)
	defer tearDownResource()

	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()

	// Grant
	scope := "read-private"

	permission := gocloak.PermissionGrantParams{
		ResourceID:  &resourceID,
		RequesterID: &userID,
		ScopeName:   &scope,
	}
	result, err := client.GrantUserPermission(context.Background(), token.AccessToken, cfg.GoCloak.Realm, permission)

	require.NoError(t, err, "GrantUserPermission failed")
	require.True(t, nil != result)
	if result != nil {
		require.False(t, result.ResourceID == nil)
		require.False(t, result.RequesterID == nil)
		require.False(t, result.Granted == nil)
		if result.ResourceID != nil {
			require.Equal(t, resourceID, *(result.ResourceID))
		}
		if result.RequesterID != nil {
			require.Equal(t, userID, *(result.RequesterID))
		}
		if result.Granted != nil {
			require.Equal(t, true, *(result.Granted))
		}
	}

	// Get
	params := gocloak.GetUserPermissionParams{
		ResourceID: &resourceID,
	}
	queried, err := client.GetUserPermissions(context.Background(), token.AccessToken, cfg.GoCloak.Realm, params)
	require.NoError(t, err, "GetUserPermissions failed")
	require.Equal(t, 1, len(queried))
	require.Equal(t, userID, *(queried[0].RequesterID))

	// Update
	permission.TicketID = gocloak.StringP(*(result.ID))
	permission.Granted = gocloak.BoolP(false)

	result, err = client.UpdateUserPermission(context.Background(), token.AccessToken, cfg.GoCloak.Realm, permission)

	require.NoError(t, err, "UpdateUserPermission failed")
	require.True(t, nil == result)

	// Get (no permission expected to be returned)
	params = gocloak.GetUserPermissionParams{
		ResourceID: &resourceID,
	}
	queried, err = client.GetUserPermissions(context.Background(), token.AccessToken, cfg.GoCloak.Realm, params)
	require.NoError(t, err, "GetUserPermissions failed")
	require.Equal(t, 0, len(queried))

	// Grant again
	permission = gocloak.PermissionGrantParams{
		ResourceID:  &resourceID,
		RequesterID: &userID,
		ScopeName:   &scope,
	}
	result, err = client.GrantUserPermission(context.Background(), token.AccessToken, cfg.GoCloak.Realm, permission)
	require.NoError(t, err, "GrantUserPermissions failed")

	// Get
	params = gocloak.GetUserPermissionParams{
		ResourceID: &resourceID,
	}
	queried, err = client.GetUserPermissions(context.Background(), token.AccessToken, cfg.GoCloak.Realm, params)
	require.NoError(t, err, "GetUserPermissions failed")
	require.Equal(t, 1, len(queried))
	require.Equal(t, userID, *(queried[0].RequesterID))

	// Delete
	err = client.DeleteUserPermission(context.Background(), token.AccessToken, cfg.GoCloak.Realm, *(result.ID))
	require.NoError(t, err, "DeleteUserPermissions failed")

	// Get (no permission expected to be returned)

	params = gocloak.GetUserPermissionParams{
		ResourceID: &resourceID,
	}
	queried, err = client.GetUserPermissions(context.Background(), token.AccessToken, cfg.GoCloak.Realm, params)
	require.NoError(t, err, "GetUserPermissions failed")
	require.Equal(t, 0, len(queried))

}

func TestGocloak_BadCreatePermissionTicket(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	// Create
	tearDownResource, resourceID := CreateResourceClientWithScopes(t, client)
	// Delete
	defer tearDownResource()

	_, err := client.CreatePermissionTicket(context.Background(), token.AccessToken, cfg.GoCloak.Realm, []gocloak.CreatePermissionTicketParams{})
	require.Error(t, err, "CreatePermissionTicket no error on empty params")

	permissions := gocloak.CreatePermissionTicketParams{
		ResourceID: &resourceID,
	}

	_, err = client.CreatePermissionTicket(context.Background(), token.AccessToken, cfg.GoCloak.Realm, []gocloak.CreatePermissionTicketParams{permissions})
	require.Error(t, err, "CreatePermissionTicket no error on missing ResourceScopes in permission")

	permissions = gocloak.CreatePermissionTicketParams{
		ResourceScopes: &[]string{"read-private"},
	}
	_, err = client.CreatePermissionTicket(context.Background(), token.AccessToken, cfg.GoCloak.Realm, []gocloak.CreatePermissionTicketParams{permissions})
	require.Error(t, err, "CreatePermissionTicket no error on missing ResourceID in permission")

	permissions = gocloak.CreatePermissionTicketParams{
		ResourceID:     &resourceID,
		ResourceScopes: &[]string{"read-private"},
	}

	_, err = client.CreatePermissionTicket(context.Background(), "", cfg.GoCloak.Realm, []gocloak.CreatePermissionTicketParams{permissions})
	require.Error(t, err, "CreatePermissionTicket no error on unauthorized access attempt")

}

func TestGocloak_CreatePermissionTicket(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetClientToken(t, client)

	// Create
	tearDownResource, resourceID := CreateResourceClientWithScopes(t, client)
	// Delete
	defer tearDownResource()

	// Add additional claims
	pushClaims := make(map[string][]string)

	pushClaims["organization"] = []string{"acme", "somecorp"}

	permissions := gocloak.CreatePermissionTicketParams{
		ResourceID:     &resourceID,
		ResourceScopes: &[]string{"read-private"},
		Claims:         &pushClaims,
	}

	ticket, err := client.CreatePermissionTicket(context.Background(), token.AccessToken, cfg.GoCloak.Realm, []gocloak.CreatePermissionTicketParams{permissions})

	require.NoError(t, err, "CreatePermissionTicket failed")
	t.Logf("Created PermissionTicket: %+v", *(ticket.Ticket))

	pt, err := jwt.ParseWithClaims(*(ticket.Ticket), &gocloak.PermissionTicketRepresentation{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil
	})

	//we're expecting validity error because we didn't supply secret
	require.Equal(t, "token signature is invalid", err.Error())

	claims, ok := pt.Claims.(*gocloak.PermissionTicketRepresentation) //ticketClaims)
	require.Equal(t, true, ok)
	require.Equal(t, cfg.GoCloak.Realm, *(claims.AZP))
	require.Equal(t, 1, len(*(claims.Permissions)))
	require.Equal(t, 1, len(*(claims.Permissions)))
	require.Equal(t, 1, len(*(claims.Claims)))
	require.Equal(t, pushClaims["organization"], (*(claims.Claims))["organization"])
	require.Equal(t, *permissions.ResourceID, *((*(claims.Permissions))[0].RSID))

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

	tearDownPolicy, policyID := CreatePolicy(t, client, gocloakClientID, gocloak.PolicyRepresentation{
		Name:        GetRandomNameP("PolicyName"),
		Description: gocloak.StringP("JS Policy"),
		Type:        gocloak.StringP("js"),
		Logic:       gocloak.POSITIVE,
		JSPolicyRepresentation: gocloak.JSPolicyRepresentation{
			Code: gocloak.StringP("$evaluation.grant();"),
		},
	})
	// Delete
	defer tearDownPolicy()

	// Create
	tearDown, permissionID := CreatePermission(t, client, gocloakClientID, gocloak.PermissionRepresentation{
		Name:        GetRandomNameP("PermissionName"),
		Description: gocloak.StringP("RequestingPartyPermission Description"),
		Type:        gocloak.StringP("resource"),
		Policies: &[]string{
			policyID,
		},
		Resources: &[]string{
			resourceID,
		},
	})
	// Delete
	defer tearDown()

	// List
	createdPermission, err := client.GetPermission(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		permissionID,
	)
	require.NoError(t, err, "GetPermission failed")
	t.Logf("Created RequestingPartyPermission: %+v", *(createdPermission.ID))
	require.Equal(t, permissionID, *(createdPermission.ID))

	// Looking for a created permission
	permissions, err := client.GetPermissions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.GetPermissionParams{
			Name: createdPermission.Name,
		},
	)
	require.NoError(t, err, "GetPermissions failed")
	require.Len(t, permissions, 1, "GetPermissions should return exact 1 permission")
	require.Equal(t, *(createdPermission.ID), *(permissions[0].ID))
	t.Logf("Permissions: %+v", permissions)

	err = client.UpdatePermission(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		gocloak.PermissionRepresentation{},
	)
	require.Error(t, err, "Should fail because of missing ID of the permission")

	createdPermission.Name = GetRandomNameP("PermissionName")
	err = client.UpdatePermission(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		*createdPermission,
	)
	require.NoError(t, err, "UpdatePermission failed")

	updatedPermission, err := client.GetPermission(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		permissionID,
	)
	require.NoError(t, err, "GetPermission failed")
	require.Equal(t, *(createdPermission.Name), *(updatedPermission.Name))

	dependentPermissions, err := client.GetDependentPermissions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		policyID,
	)

	require.NoError(t, err, "GetDependentPermissions failed")
	require.Len(t, dependentPermissions, 1, "GetDependentPermissions should return exact 1 permission")
	require.Equal(t, *(createdPermission.Name), *(dependentPermissions[0].Name))

	permissionResources, err := client.GetPermissionResources(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		permissionID,
	)

	require.NoError(t, err, "GetPermissionResource failed")
	require.Len(t, permissionResources, 1, "GetPermissionResource should return exact 1 resource")
	require.Equal(t, resourceID, *permissionResources[0].ResourceID)

	permissionScopes, err := client.GetPermissionScopes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloakClientID,
		permissionID,
	)

	require.NoError(t, err, "GetPermissionScopes failed")
	require.Len(t, permissionScopes, 0, "GetPermissionResource should return exact 0 scopes")

}

func TestGoCloak_CheckError(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	_, err := client.GetClient(
		context.Background(),
		token.AccessToken,
		cfg.Admin.Realm,
		"random_client",
	)
	require.Error(t, err)

	t.Log(err)

	expectedError := &gocloak.APIError{
		Code:    http.StatusNotFound,
		Message: "404 Not Found: Could not find client",
		Type:    gocloak.APIErrTypeUnknown,
	}

	apiError := err.(*gocloak.APIError)
	require.Equal(t, expectedError, apiError)
}

// ---------------
// Credentials API
// ---------------

func TestGoCloak_GetCredentialRegistrators(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)

	res, err := client.GetCredentialRegistrators(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	t.Log(res)
	require.NoError(t, err)
}

func TestGoCloak_GetConfiguredUserStorageCredentialTypes(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	SetUpTestUser(t, client)

	res, err := client.GetConfiguredUserStorageCredentialTypes(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
	)
	t.Log(res)
	require.NoError(t, err)
}

func TestGoCloak_GetUpdateLableDeleteCredentials(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	tearDownUser, userID := CreateUser(t, client)
	defer tearDownUser()
	err := client.SetPassword(
		context.Background(),
		token.AccessToken,
		userID,
		cfg.GoCloak.Realm,
		"fake-password",
		false,
	)
	require.NoError(t, err)

	res, err := client.GetCredentials(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	t.Log(res)
	require.NoError(t, err)
	require.Len(t, res, 1)
	credentialID := *res[0].ID

	err = client.UpdateCredentialUserLabel(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		credentialID,
		"test-label",
	)
	require.NoError(t, err)
	res, err = client.GetCredentials(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	t.Log(res)
	require.NoError(t, err)
	require.Equal(t, "test-label", *res[0].UserLabel)

	err = client.DeleteCredentials(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
		credentialID,
	)
	require.NoError(t, err)

	res, err = client.GetCredentials(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		userID,
	)
	t.Log(res)
	require.NoError(t, err)
	require.Empty(t, res)
}

func TestGoCloak_DisableAllCredentialsByType(t *testing.T) {
	// NOTE(svilgelm): I didn't find a way how to properly test this function,
	// so the test validates that the API call doesn't return an error.
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	SetUpTestUser(t, client)

	err := client.DisableAllCredentialsByType(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		testUserID,
		[]string{"password"},
	)
	require.NoError(t, err)
}

func TestGocloak_TestSetFunctionalOptions(t *testing.T) {
	t.Parallel()

	cfg := GetConfig(t)
	gocloak.NewClient(cfg.HostName, gocloak.SetAuthRealms("foo"), gocloak.SetAuthAdminRealms("bar"))
}

func TestGocloak_GetClientsWithPagination(t *testing.T) {
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	clientID := GetRandomNameP("ClientID")

	testClient := gocloak.Client{
		ClientID: clientID,
		BaseURL:  gocloak.StringP("http://example.com"),
	}
	t.Logf("Client ID: %s", *clientID)

	// Creating a client
	tearDown, createdClientID := CreateClient(t, client, &testClient)
	defer tearDown()
	t.Log(createdClientID)
	first := 0
	max := 1
	// Looking for a created client
	clients, err := client.GetClients(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		gocloak.GetClientsParams{
			First: &first,
			Max:   &max,
		},
	)
	require.NoError(t, err)
	require.Equal(t, max, len(clients))
}

func TestGocloak_GetAuthenticationFlows(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	authFlows, err := client.GetAuthenticationFlows(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.NoError(t, err, "Failed to fetch authentication flows")
	t.Logf("authentication flows: %+v", authFlows)

	FailRequest(client, nil, 1, 0)
	_, err = client.GetAuthenticationFlows(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
	)
	require.Error(t, err)
}

func TestGocloak_CreateAuthenticationFlowsAndCreateAuthenticationExecution(t *testing.T) {
	t.Parallel()
	cfg := GetConfig(t)
	client := NewClientWithDebug(t)
	token := GetAdminToken(t, client)
	authExec :=
		gocloak.CreateAuthenticationExecutionRepresentation{
			Provider: gocloak.StringP("idp-auto-link"),
		}
	authFlow := gocloak.AuthenticationFlowRepresentation{
		Alias:       gocloak.StringP("testauthflow2"),
		BuiltIn:     gocloak.BoolP(false),
		Description: gocloak.StringP("my test description"),
		TopLevel:    gocloak.BoolP(true),
		ProviderID:  gocloak.StringP("basic-flow"),
	}

	err := client.CreateAuthenticationFlow(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		authFlow,
	)
	require.NoError(t, err, "Failed to create authentication flow")
	t.Logf("authentication flows: %+v", authFlow)

	err = client.CreateAuthenticationExecution(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*authFlow.Alias,
		authExec,
	)
	require.NoError(t, err, "Failed to create authentication execution")

	authExecs, err := client.GetAuthenticationExecutions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*authFlow.Alias,
	)

	t.Logf("authentication executions: %+v", authExecs)
	require.NoError(t, err, "Failed to get authentication executions")

	authExecs[0].Requirement = gocloak.StringP("ALTERNATIVE")
	err = client.UpdateAuthenticationExecution(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*authFlow.Alias,
		*authExecs[0],
	)
	require.NoError(t, err, "Failed to update authentication executions")
	authExecs, err = client.GetAuthenticationExecutions(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*authFlow.Alias,
	)
	require.NoError(t, err, "Failed to get authentication executions second time")
	t.Logf("authentication executions after update: %+v", authExecs)

	err = client.DeleteAuthenticationFlow(
		context.Background(),
		token.AccessToken,
		cfg.GoCloak.Realm,
		*authFlow.Alias,
	)
	require.NoError(t, err, "Failed to delete authentication flow")
}
