package core

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Nerzal/gocloak/models"
)

type Admin interface {
	Login(username string, password string, realm string) (*models.JWT, error)
}

type admin struct {
	basePath string
}

type loginData struct {
	ClientID  string `json:"client_id"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}

const adminClientID string = "admin-cli"

func NewAdminClient(basePath string) Admin {
	return &admin{
		basePath: basePath,
	}
}

func (client *admin) Login(username, password, realm string) (*models.JWT, error) {
	firstPart := "/auth/realms/"
	lastPart := "/protocol/openid-connect/token"
	loginPath := firstPart + realm + lastPart

	loginData := loginData{
		ClientID:  adminClientID,
		UserName:  username,
		Password:  password,
		GrantType: "password",
	}

	payload, err := json.Marshal(loginData)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", client.basePath+loginPath, bytes.NewReader(payload))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	jwt := &models.JWT{}
	err = json.Unmarshal(body, jwt)
	return jwt, err
}
