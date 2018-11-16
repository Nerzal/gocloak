package core

import (
	"testing"
)

func TestLogin(t *testing.T) {
	client := NewClient(hostname)
	_, err := client.Login(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}
}

func TestGetUsers(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.Login(username, password, realm)
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

func TestGetGroups(t *testing.T) {
	client := NewClient(hostname)
	token, err := client.Login(username, password, realm)
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
	token, err := client.Login(username, password, realm)
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
	token, err := client.Login(username, password, realm)
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
	token, err := client.Login(username, password, realm)
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
	token, err := client.Login(username, password, realm)
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
	token, err := client.Login(username, password, realm)
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
