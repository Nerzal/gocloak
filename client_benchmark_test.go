package gocloak_test

import (
	"context"
	"testing"

	"github.com/Nerzal/gocloak/v14"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkLogin(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	SetUpTestUser(b, client)
	b.ResetTimer()
	for i := range b.N {
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

// func BenchmarkLoginParallel(b *testing.B) {
// 	cfg := GetConfig(b)
// 	client := gocloak.NewClient(cfg.HostName)
// 	SetUpTestUser(b, client)
// 	b.ResetTimer()
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			_, err := client.Login(
// 				context.Background(),
// 				cfg.GoCloak.ClientID,
// 				cfg.GoCloak.ClientSecret,
// 				cfg.GoCloak.Realm,
// 				cfg.GoCloak.UserName,
// 				cfg.GoCloak.Password,
// 			)
// 			assert.NoError(b, err)
// 		}
// 	})
// }

func BenchmarkGetGroups(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	token := GetAdminToken(b, client)
	b.ResetTimer()
	for range b.N {
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
		Full: new(true),
	}
	b.ResetTimer()
	for range b.N {
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
		BriefRepresentation: new(true),
	}
	token := GetAdminToken(b, client)
	b.ResetTimer()
	for range b.N {
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
	for range b.N {
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
	defer teardown()
	token := GetAdminToken(b, client)
	group, err := client.GetGroup(context.Background(), token.AccessToken, cfg.GoCloak.Realm, groupID)
	require.NoError(b, err)
	b.ResetTimer()
	for range b.N {
		_, err := client.GetGroupByPath(
			context.Background(),
			token.AccessToken,
			cfg.GoCloak.Realm,
			*group.Path,
		)
		assert.NoError(b, err)
	}
}
