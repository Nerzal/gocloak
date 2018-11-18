package core

import (
	"strconv"
	"testing"

	"github.com/Nerzal/gocloak/models"
)

func TestLogin(t *testing.T) {
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

	user := models.User{}
	user.FirstName = "Klaus"
	user.LastName = "Peter"
	user.Email = "somm@ting.wong"
	user.Enabled = true
	user.Username = user.Email
	err = client.CreateUser(token, realm, user)
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

	group := models.Group{}
	group.Name = "MySuperCoolNewGroup"
	err = client.CreateGroup(token, realm, group)
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

	role := models.Role{}
	role.Name = "mySuperCoolRole"
	err = client.CreateRole(token, realm, "9204c840-f857-4507-8b00-784c9c845e6e", role)
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

	newClient := models.Client{}
	newClient.ClientID = "KYCnow"
	err = client.CreateClient(token, realm, newClient)
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

	_, err = client.GetUsers(token, realm)
	if err != nil {
		t.Log("GetUsers failed", err.Error())
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

	users, err := client.GetUsers(token, realm)
	if err != nil {
		t.Log("GetUsers failed", err.Error())
		t.Fail()
	}

	dereferencedUsers := *users
	_, err = client.GetUser(token, realm, dereferencedUsers[0].ID)
}

func TestGetUserCount(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.LoginAdmin(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	count, err := client.GetUserCount(token, realm)
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

	_, err = client.GetGroups(token, realm)
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

	users, err := client.GetUsers(token, realm)
	if err != nil {
		t.Log("GetAllUsers failed", err.Error())
		t.Fail()
	}

	realUsers := *users

	_, err = client.GetUserGroups(token, realm, realUsers[0].ID)
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

	_, err = client.GetRoles(token, realm)
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

	_, err = client.GetClients(token, realm)
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

	clients, err := client.GetClients(token, realm)
	if err != nil {
		t.Log("GetClients failed", err.Error())
		t.Fail()
	}

	clientsDeferenced := *clients
	_, err = client.GetRolesByClientID(token, realm, clientsDeferenced[4].ClientID)
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

	groups, err := client.GetGroups(token, realm)
	if err != nil {
		t.Log("GetGroups failed", err.Error())
		t.Fail()
	}

	if len(*groups) == 0 {
		return
	}

	groupsDeferenced := *groups
	_, err = client.GetRoleMappingByGroupID(token, realm, groupsDeferenced[0].ID)
	if err != nil {
		t.Log("GetRoleMappingByGroupID failed")
		t.Fail()
	}
}
