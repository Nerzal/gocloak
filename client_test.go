package gocloak

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func Test_DecodeAccessToken(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	_, _, err = client.DecodeAccessToken(token.AccessToken, realm)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
}

func Test_RefreshToken(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	token, err = client.RefreshToken(token.RefreshToken, clientid, clientSecret, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
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
	token, err := client.Login(clientid, clientSecret, realm, username, password)
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

func Test_DecodeAccessTokenCustomClaims(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	_, claims, err := client.DecodeAccessTokenCustomClaims(token.AccessToken, realm)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	t.Log(claims)
}

func Test_GetKeys(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	config, err := client.GetKeyStoreConfig(token.AccessToken, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	t.Log(config)
}

func Test_Login(t *testing.T) {
	client := NewClient(hostname)
	_, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}
}

func Test_LoginAdmin(t *testing.T) {
	client := NewClient(hostname)
	_, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}
}

func Test_SetPassword(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginClient(clientid, clientSecret, realm)
	if err != nil {
		t.Log("Failed to login: ", err.Error())
		t.FailNow()
	}
	user := User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = "olaf4@mail.com"
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
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
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
		t.Fail()
	}
}

func Test_CreateUser_CustomAttributes(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	user := User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = "trololo2@mail.com"
	user.Enabled = true
	user.Username = user.Email
	user.Attributes = map[string][]string{}
	user.Attributes["foo"] = []string{"bar"}
	user.Attributes["bar"] = []string{"baz"}

	id, err := client.CreateUser(token.AccessToken, realm, user)
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.Fail()
	}

	t.Log(id)
}

func TestCreateGroup(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	group := Group{}
	group.Name = "MySuperCoolNewGroup"
	err = client.CreateGroup(token.AccessToken, realm, group)
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.Fail()
	}
}

func TestCreateRole(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	role := Role{}
	role.Name = "mySuperCoolRole"
	err = client.CreateRole(token.AccessToken, realm, "9204c840-f857-4507-8b00-784c9c845e6e", role)
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.Fail()
	}
}

func TestCreateClient(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	newClient := Client{}
	newClient.ClientID = "KYCnow"
	err = client.CreateClient(token.AccessToken, realm, newClient)
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.Fail()
	}
}

func TestGetUsers(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	users, err := client.GetUsers(token.AccessToken, realm)
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.Fail()
	}

	t.Log(users)
}

func TestGetKeyStoreConfig(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	_, err = client.GetKeyStoreConfig(token.AccessToken, realm)
	if err != nil {
		t.Log("GetKeyStoreConfig failed", err.Error())
		t.Fail()
	}
}

func TestGetUser(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	users, err := client.GetUsers(token.AccessToken, realm)
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.Fail()
	}

	dereferencedUsers := *users
	user, err := client.GetUser(token.AccessToken, realm, dereferencedUsers[0].ID)
	if err != nil {
		t.Log("GetUser failed", err.Error())
		t.Fail()
	}

	t.Log(user)
}

func TestGetUserCount(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	count, err := client.GetUserCount(token.AccessToken, realm)
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.Fail()
	}

	t.Log("Users in Realm: " + strconv.Itoa(count))
}

func TestGetGroups(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	_, err = client.GetGroups(token.AccessToken, realm)
	if err != nil {
		t.Log("GetGroups failed", err.Error())
		t.Fail()
	}
}

func TestGetUserGroups(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	users, err := client.GetUsers(token.AccessToken, realm)
	if err != nil {
		t.Log("GetAllUsers failed", err.Error())
		t.Fail()
	}

	realUsers := *users

	_, err = client.GetUserGroups(token.AccessToken, realm, realUsers[0].ID)
	if err != nil {
		t.Log("GetUserGroups failed", err.Error())
		t.Fail()
	}
}

func TestGetRoles(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	_, err = client.GetRoles(token.AccessToken, realm)
	if err != nil {
		t.Log("GetRoles failed", err.Error())
		t.Fail()
	}
}

func TestGetClients(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	_, err = client.GetClients(token.AccessToken, realm)
	if err != nil {
		t.Log("GetClients failed", err.Error())
		t.Fail()
	}
}

func TestGetRolesByClientId(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	clients, err := client.GetClients(token.AccessToken, realm)
	if err != nil {
		t.Log("GetClients failed", err.Error())
		t.Fail()
	}

	clientsDeferenced := *clients
	_, err = client.GetRolesByClientID(token.AccessToken, realm, clientsDeferenced[4].ClientID)
	if err != nil {
		t.Log("GetRolesByClientID failed", err.Error())
		t.Fail()
	}
}

func TestGetRoleMappingByGroupID(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	groups, err := client.GetGroups(token.AccessToken, realm)
	if err != nil {
		t.Log("GetGroups failed", err.Error())
		t.Fail()
	}

	if len(*groups) == 0 {
		return
	}

	groupsDeferenced := *groups
	_, err = client.GetRoleMappingByGroupID(token.AccessToken, realm, groupsDeferenced[0].ID)
	if err != nil {
		t.Log("GetRoleMappingByGroupID failed")
		t.Fail()
	}
}
