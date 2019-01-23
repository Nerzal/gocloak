package gocloak

import (
	"strconv"
	"testing"
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

func Test_DecodeAccessTokenCustomClaims(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	_, _, err = client.DecodeAccessTokenCustomClaims(token.AccessToken, realm)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
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
	user.Email = "somm@ting.wong"
	user.Enabled = true
	user.Username = user.Email
	err = client.CreateUser(token.AccessToken, realm, user)
	if err != nil {
		t.Log("Create User Failed: ", err.Error())
		t.Fail()
	}
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
