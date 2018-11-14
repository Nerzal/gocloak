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

}
