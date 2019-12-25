package gocloak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkLogin(b *testing.B) {
	cfg := GetConfig(b)
	client := NewClient(cfg.HostName)
	SetUpTestUser(b, client)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Login(
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
	client := NewClient(cfg.HostName)
	SetUpTestUser(b, client)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.Login(
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
