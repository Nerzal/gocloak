package core

import (
	"testing"
)

func TestLogin(t *testing.T) {
	client := NewAdminClient(hostname)
	_, err := client.Login(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}
	t.Log("Successfully logged in")
}

func TestGetallUsers(t *testing.T){
	client := NewAdminClient(hostname)
	token, err := client.Login(username, password, realm)
	if err != nil {
		t.Log("TestLogin failed", err.Error())
		t.Fail()
	}

	err = client.GetAllUsers(token)
	if err != nil {
		t.Log("GetAllUsers failed", err.Error())
		t.Fail()
	}
}
