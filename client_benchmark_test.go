package gocloak_test

import (
	"context"
	"testing"

	"github.com/Nerzal/gocloak/v13"

	"github.com/stretchr/testify/assert"
)

func BenchmarkLogin(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	SetUpTestUser(b, client)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Login(
			context.Background(),
			cfg.GoCloak.ClientID,
			cfg.GoCloak.ClientSecret,
			cfg.GoCloak.Realm,
			cfg.GoCloak.UserName,
			cfg.GoCloak.Password,
		)
		assert.NoError(b, err, "Failed %d", i)
	}
}

func BenchmarkLoginParallel(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	SetUpTestUser(b, client)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.Login(
				context.Background(),
				cfg.GoCloak.ClientID,
				cfg.GoCloak.ClientSecret,
				cfg.GoCloak.Realm,
				cfg.GoCloak.UserName,
				cfg.GoCloak.Password,
			)
			assert.NoError(b, err)
		}
	})
}

func BenchmarkGetGroups(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	token := GetAdminToken(b, client)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetGroups(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			gocloak.GetGroupsParams{},
		)
		assert.NoError(b, err)
	}
}

func BenchmarkGetGroupsFull(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	token := GetAdminToken(b, client)
	params := gocloak.GetGroupsParams{
		Full: gocloak.BoolP(true),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetGroups(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			params,
		)
		assert.NoError(b, err)
	}
}

func BenchmarkGetGroupsBrief(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	params := gocloak.GetGroupsParams{
		BriefRepresentation: gocloak.BoolP(true),
	}
	token := GetAdminToken(b, client)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetGroups(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			params,
		)
		assert.NoError(b, err)
	}
}

func BenchmarkGetGroup(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	teardown, groupID := CreateGroup(b, client)
	defer teardown()
	token := GetAdminToken(b, client)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetGroup(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			groupID,
		)
		assert.NoError(b, err)
	}
}

func BenchmarkGetGroupByPath(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	teardown, groupID := CreateGroup(b, client)
	token := GetAdminToken(b, client)
	group, err := client.GetGroup(context.Background(), token.AccessToken, cfg.GoCloak.Realm, groupID)
	assert.NoError(b, err)
	defer teardown()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetGroupByPath(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			*group.Path,
		)
		assert.NoError(b, err)
	}
}
