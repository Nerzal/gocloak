package gocloak

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/pkg/jwx"
)

func Test_RetrospectToken(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
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
		t.Log("Inactive Token o_O")
		t.FailNow()
	}	

	t.Log(rptResult)
}

func Test_DecodeAccessToken(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	_, _, err = client.DecodeAccessToken(token.AccessToken, token.AccessToken, realm)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func Test_DecodeAccessTokenCustomClaims(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}

	claims := jwx.Claims{}
	_, err = client.DecodeAccessTokenCustomClaims(token.AccessToken, token.AccessToken, realm, claims)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func Test_RefreshToken(t *testing.T) {
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

func Test_UserAttributeContains(t *testing.T) {
	t.Parallel()

	attributes := map[string][]string{}
	attributes["foo"] = []string{"bar", "alice", "bob", "roflcopter"}
	attributes["bar"] = []string{"baz"}

	client := NewClient(hostname)
	if !client.UserAttributeContains(attributes, "foo", "alice") {
		t.FailNow()
	}
}

func Test_GetUserByID(t *testing.T) {
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

func Test_GetKeys(t *testing.T) {
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

func Test_Login(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	_, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}
}

func Test_LoginAdmin(t *testing.T) {
	t.Parallel()
	client := NewClient(hostname)
	_, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.FailNow()
	}
}

func Test_SetPassword(t *testing.T) {
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
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.FailNow()
	}
}

func Test_CreateUser_CustomAttributes(t *testing.T) {
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
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
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

	users, err := client.GetUsers(token.AccessToken, realm)
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

	users, err := client.GetUsers(token.AccessToken, realm)
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.FailNow()
	}

	dereferencedUsers := *users
	user, err := client.GetUser(token.AccessToken, realm, dereferencedUsers[0].ID)
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

	users, err := client.GetUsers(token.AccessToken, realm)
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
