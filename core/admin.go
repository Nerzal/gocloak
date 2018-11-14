package core

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Admin interface {
	Login(string, string) error
}

type admin struct {
}

type loginData struct {
	ClientID  string `json:"client_id"`
	UserName  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
}

func Login(username, password, realm string) error {
	url := "/auth/realms/master/protocol/openid-connect/token"
	
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

	return nil
}
