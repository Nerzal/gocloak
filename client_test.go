package gocloak

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/pkg/jwx"
)

func TestRequestPermission(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.RequestPermission(clientid, clientSecret, realm, username, password, "Permission foo # 3")
	if err != nil {
		t.Log("Login failed", err.Error())
		t.FailNow()
	}

	rptResult, err := client.RetrospectToken(token.AccessToken, clientid, clientSecret, realm)
	if err != nil {
		t.Log("Inspection failed:", err.Error())
		t.FailNow()
	}

	if !rptResult.Active {
		t.Log("Inactive Token oO")
		t.FailNow()
	}

	t.Log(rptResult)
}

func TestGetCerts(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	certs, err := client.GetCerts(realm)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log(certs)
}
func Test_LoginClient_UnknownRealm(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	_, err := client.LoginClient(clientid, clientSecret, "ThisRealmDoesNotExist")
	if err == nil {
		t.Log("Login shouldn't be succesful", err.Error())
		t.FailNow()
	}

	errorMessage := err.Error()
	if errorMessage != "404 Not Found" {
		t.Log("Unexpected error message", err.Error())
		t.FailNow()
	}
}

func TestGetIssuer(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	issuer, err := client.GetIssuer(realm)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log(issuer)
}

func TestRetrospectTokenInactiveToken(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	_, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("Login failed", err.Error())
		t.FailNow()
	}

	rptResult, err := client.RetrospectToken("foobar", clientid, clientSecret, realm)
	if err != nil {
		t.Log("Inspection failed:", err.Error())
		t.FailNow()
	}

	if rptResult.Active {
		t.Log("That should never happen. Token is active")
		t.FailNow()
	}

	t.Log(rptResult)
}

func TestGetUserInfo(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	userInfo, err := client.GetUserInfo(token.AccessToken, realm)
	if err != nil {
		t.Log("Failed to fetch userinfo", err.Error())
		t.FailNow()
	}
	t.Log(userInfo)
}

func TestRetrospectToken(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("Login failed", err.Error())
		t.FailNow()
	}

	rptResult, err := client.RetrospectToken(token.AccessToken, clientid, clientSecret, realm)
	if err != nil {
		t.Log("Inspection failed:", err.Error())
		t.FailNow()
	}

	if !rptResult.Active {
		t.Log("Inactive Token oO")
		t.FailNow()
	}

	t.Log(rptResult)
}

func TestDecodeAccessToken(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	resultToken, claims, err := client.DecodeAccessToken(token.AccessToken, realm)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	t.Log(resultToken)
	t.Log(claims)
}

func TestDecodeAccessTokenCustomClaims(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	claims := jwx.Claims{}
	_, err = client.DecodeAccessTokenCustomClaims(token.AccessToken, realm, claims)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func TestRefreshToken(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	token, err = client.RefreshToken(token.RefreshToken, clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	t.Log(token)
}

func TestUserAttributeContains(t *testing.T) {
	t.Parallel()

	attributes := map[string][]string{}
	attributes["foo"] = []string{"bar", "alice", "bob", "roflcopter"}
	attributes["bar"] = []string{"baz"}

	client := NewClient(hostname)
	if !client.UserAttributeContains(attributes, "foo", "alice") {
		t.FailNow()
	}
}

func TestGetUserByID(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	randomNumber := r1.Intn(100000)
	user := User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = "trololo" + strconv.Itoa(randomNumber) + "@mail.com"
	user.Enabled = true
	user.Username = user.Email
	user.Attributes = map[string][]string{}
	user.Attributes["foo"] = []string{"bar", "alice", "bob", "roflcopter"}
	user.Attributes["bar"] = []string{"baz"}

	id, err := client.CreateUser(token.AccessToken, realm, user)
	if err != nil {
		t.Log("CreateUser failed", err.Error())
		t.FailNow()
	}

	fetchedUser, err := client.GetUserByID(token.AccessToken, realm, *id)
	if err != nil {
		t.Log("GetUserById failed", err.Error())
		t.FailNow()
	}

	t.Log(fetchedUser)
}

func TestGetKeys(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	config, err := client.GetKeyStoreConfig(token.AccessToken, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	t.Log(config)
}

func TestLogin(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	_, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}
}

func TestLoginClient(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	_, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}
}

func TestLoginAdmin(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	_, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}
}

func TestSetPassword(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("Failed to login: ", err.Error())
		t.FailNow()
	}
	user := User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = "olaf5@mail.com"
	user.Enabled = true
	user.Username = user.Email

	userID, err := client.CreateUser(token.AccessToken, realm, user)
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.FailNow()
	}

	err = client.SetPassword(token.AccessToken, *userID, realm, "passwort1234!", false)
	if err != nil {
		t.Log("Failed to set password: ", err.Error())
		t.FailNow()
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	user := User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = "trololo@mail.com"
	user.Enabled = true
	user.Username = user.Email
	_, err = client.CreateUser(token.AccessToken, realm, user)
	_, ok := err.(*ObjectAllreadyExists)
	if ok {
		return
	}

	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.FailNow()
	}
}

func TestCreateUserCustomAttributes(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	user := User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = "trololo4234@mail.com"
	user.Enabled = true
	user.Username = user.Email
	user.Attributes = map[string][]string{}
	user.Attributes["foo"] = []string{"bar"}
	user.Attributes["bar"] = []string{"baz"}

	id, err := client.CreateUser(token.AccessToken, realm, user)
	_, ok := err.(*ObjectAllreadyExists)
	if ok {
		return
	}

	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.FailNow()
	}

	t.Log(id)
}

func TestCreateGroup(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	group := Group{}
	group.Name = "MySuperCoolNewGroup"
	err = client.CreateGroup(token.AccessToken, realm, group)
	_, ok := err.(*ObjectAllreadyExists)
	if ok {
		return
	}

	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.FailNow()
	}
}

func TestCreateRole(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	role := Role{}
	role.Name = "mySuperCoolRole"
	err = client.CreateRole(token.AccessToken, realm, "9204c840-f857-4507-8b00-784c9c845e6e", role)
	_, ok := err.(*ObjectAllreadyExists)
	if ok {
		return
	}

	if err != nil {
		t.Log("Create Role Failed: ", err.Error())
		t.FailNow()
	}
}

func TestCreateClient(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	newClient := Client{}
	newClient.ClientID = "KYCnow"
	err = client.CreateClient(token.AccessToken, realm, newClient)
	_, ok := err.(*ObjectAllreadyExists)
	if ok {
		return
	}

	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.FailNow()
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	users, err := client.GetUsers(token.AccessToken, realm, GetUsersParams{})
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.FailNow()
	}

	t.Log(users)
}

func TestGetKeyStoreConfig(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	_, err = client.GetKeyStoreConfig(token.AccessToken, realm)
	if err != nil {
		t.Log("GetKeyStoreConfig failed", err.Error())
		t.FailNow()
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	users, err := client.GetUsers(token.AccessToken, realm, GetUsersParams{})
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.FailNow()
	}

	dereferencedUsers := *users
	user, err := client.GetUserByID(token.AccessToken, realm, dereferencedUsers[0].ID)
	if err != nil {
		t.Log("GetUser failed", err.Error())
		t.FailNow()
	}

	t.Log(user)
}

func TestGetUserCount(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	count, err := client.GetUserCount(token.AccessToken, realm)
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.FailNow()
	}

	t.Log("Users in Realm: " + strconv.Itoa(count))
}

func TestGetGroups(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	_, err = client.GetGroups(token.AccessToken, realm)
	if err != nil {
		t.Log("GetGroups failed", err.Error())
		t.FailNow()
	}
}

func TestGetUserGroups(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	users, err := client.GetUsers(token.AccessToken, realm, GetUsersParams{})
	if err != nil {
		t.Log("GetAllUsers failed", err.Error())
		t.FailNow()
	}

	realUsers := *users

	_, err = client.GetUserGroups(token.AccessToken, realm, realUsers[0].ID)
	if err != nil {
		t.Log("GetUserGroups failed", err.Error())
		t.FailNow()
	}
}

func TestGetRoles(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	_, err = client.GetRoles(token.AccessToken, realm)
	if err != nil {
		t.Log("GetRoles failed", err.Error())
		t.FailNow()
	}
}

func TestGetClients(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	_, err = client.GetClients(token.AccessToken, realm)
	if err != nil {
		t.Log("GetClients failed", err.Error())
		t.FailNow()
	}
}

func TestGetRolesByClientId(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	clients, err := client.GetClients(token.AccessToken, realm)
	if err != nil {
		t.Log("GetClients failed", err.Error())
		t.FailNow()
	}

	clientsDeferenced := *clients
	_, err = client.GetRolesByClientID(token.AccessToken, realm, clientsDeferenced[4].ClientID)
	if err != nil {
		t.Log("GetRolesByClientID failed", err.Error())
		t.FailNow()
	}
}

func TestGetRoleMappingByGroupID(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	groups, err := client.GetGroups(token.AccessToken, realm)
	if err != nil {
		t.Log("GetGroups failed", err.Error())
		t.FailNow()
	}

	if len(*groups) == 0 {
		return
	}

	groupsDeferenced := *groups
	_, err = client.GetRoleMappingByGroupID(token.AccessToken, realm, groupsDeferenced[0].ID)
	if err != nil {
		t.Log("GetRoleMappingByGroupID failed")
		t.FailNow()
	}
}

func TestGetRoleMappingByUserID(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	users, err := client.GetUsers(token.AccessToken, realm, GetUsersParams{})
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.FailNow()
	}

	if len(*users) == 0 {
		return
	}

	usersDeferenced := *users
	_, err = client.GetRoleMappingByGroupID(token.AccessToken, realm, usersDeferenced[0].ID)
	if err != nil {
		t.Log("GetRoleMappingByUserID failed")
		t.FailNow()
	}
}

func TestGetRealmRolesByUserID(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	users, err := client.GetUsers(token.AccessToken, realm, GetUsersParams{})
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.FailNow()
	}

	if len(*users) == 0 {
		return
	}

	usersDeferenced := *users
	_, err = client.GetRealmRolesByUserID(token.AccessToken, realm, usersDeferenced[0].ID)
	if err != nil {
		t.Log("GetRealmRolesByUserID failed")
		t.FailNow()
	}
}

func TestGetRealmRolesByGroupID(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	groups, err := client.GetGroups(token.AccessToken, realm)
	if err != nil {
		t.Log("GetGroups failed", err.Error())
		t.FailNow()
	}

	if len(*groups) == 0 {
		return
	}

	groupsDeferenced := *groups
	_, err = client.GetRealmRolesByGroupID(token.AccessToken, realm, groupsDeferenced[0].ID)
	if err != nil {
		t.Log("GetRealmRolesByGroupID failed")
		t.FailNow()
	}
}

func TestExecuteActionsEmailUpdatePassword(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	params := ExecuteActionsEmail{
		ClientID: clientid,
		UserID:   "7ce47297-f884-43ac-92e2-71820c63969a",
		Actions:  []string{"UPDATE_PASSWORD"},
	}

	err = client.ExecuteActionsEmail(token.AccessToken, realm, params)
	if err != nil {
		t.Log("ExecuteActionsEmail failed", err.Error())
		t.FailNow()
	}
}

func TestGetUsersByEmail(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	users, err := client.GetUsers(token.AccessToken, realm, GetUsersParams{Email: "trololo@mail.com"})
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.FailNow()
	}

	t.Logf("%+v", users)
}

func TestLogout(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	err = client.Logout(clientid, realm, token.RefreshToken)
	if err != nil {
		t.Log("TestLogout failed", err.Error())
		t.FailNow()
	}
}
