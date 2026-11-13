package gocloak_test

import (
	"context"
	"testing"

	"github.com/Nerzal/gocloak/v14"

	"github.com/stretchr/testify/require"
)

func BenchmarkLogin(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	SetUpTestUser(b, client)
	for b.Loop() {
		token, err := client.Login(
			context.Background(),
			cfg.GoCloak.ClientID,
			cfg.GoCloak.ClientSecret,
			cfg.GoCloak.Realm,
			cfg.GoCloak.UserName,
			cfg.GoCloak.Password,
		)
		if err != nil {
			b.Fatal(err)
		}
		// end the session outside the measurement, so sessions do not pile up in Keycloak
		b.StopTimer()
		err = client.Logout(context.Background(), cfg.GoCloak.ClientID, cfg.GoCloak.ClientSecret, cfg.GoCloak.Realm, token.RefreshToken)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
	}
}

// benchmarkGetGroups measures GetGroups with the given params.
// Input: b [*testing.B], params [gocloak.GetGroupsParams]. Output: none.
func benchmarkGetGroups(b *testing.B, params gocloak.GetGroupsParams) {
	b.Helper()
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	token := GetAdminToken(b, client)
	for b.Loop() {
		if _, err := client.GetGroups(context.Background(), token.AccessToken, cfg.GoCloak.Realm, params); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetGroups(b *testing.B) {
	benchmarkGetGroups(b, gocloak.GetGroupsParams{})
}

func BenchmarkGetGroupsFull(b *testing.B) {
	benchmarkGetGroups(b, gocloak.GetGroupsParams{Full: new(true)})
}

func BenchmarkGetGroupsBrief(b *testing.B) {
	benchmarkGetGroups(b, gocloak.GetGroupsParams{BriefRepresentation: new(true)})
}

// createBenchmarkGroup creates a group and deletes it after the benchmark with a fresh admin token,
// because the token used for creation may expire during a long run.
// Input: b [*testing.B], client [*gocloak.GoCloak]. Output: [string] group ID.
func createBenchmarkGroup(b *testing.B, client *gocloak.GoCloak) string {
	b.Helper()
	_, groupID := CreateGroup(b, client)
	b.Cleanup(func() {
		token := GetAdminToken(b, client)
		err := client.DeleteGroup(context.Background(), token.AccessToken, GetConfig(b).GoCloak.Realm, groupID)
		require.NoError(b, err, "DeleteGroup failed")
	})
	return groupID
}

func BenchmarkGetGroup(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	groupID := createBenchmarkGroup(b, client)
	token := GetAdminToken(b, client)
	for b.Loop() {
		if _, err := client.GetGroup(context.Background(), token.AccessToken, cfg.GoCloak.Realm, groupID); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetGroupByPath(b *testing.B) {
	cfg := GetConfig(b)
	client := gocloak.NewClient(cfg.HostName)
	groupID := createBenchmarkGroup(b, client)
	token := GetAdminToken(b, client)
	group, err := client.GetGroup(context.Background(), token.AccessToken, cfg.GoCloak.Realm, groupID)
	require.NoError(b, err)
	for b.Loop() {
		if _, err := client.GetGroupByPath(context.Background(), token.AccessToken, cfg.GoCloak.Realm, *group.Path); err != nil {
			b.Fatal(err)
		}
	}
}
