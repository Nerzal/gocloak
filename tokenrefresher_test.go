package gocloak_test

import (
	"context"
	"testing"

	"github.com/Nerzal/gocloak/v7"
)

func TestTokenRefresher(t *testing.T) {
	cfg := GetConfig(t)
	refresher, err := gocloak.NewTokenRefresher(context.Background(), &gocloak.RefreshConfig{
		Domain:           "test.io",
		ClientId:         cfg.GoCloak.ClientID,
		ClientSecret:     cfg.GoCloak.ClientSecret,
		Realm:            cfg.GoCloak.Realm,
		EarlyRefreshSecs: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	token := refresher.AccessToken()
	if token == "" {
		t.Fatalf("expected token but received %s", token)
	}
}
