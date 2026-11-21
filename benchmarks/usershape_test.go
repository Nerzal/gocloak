package benchmarks

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Nerzal/gocloak/v15"
)

// flatUser has the scalar fields of gocloak.User as values instead of pointers; maps and slices stay.
type flatUser struct {
	ID                         string                             `json:"id"`
	CreatedTimestamp           int64                              `json:"createdTimestamp"`
	Username                   string                             `json:"username"`
	Enabled                    bool                               `json:"enabled"`
	Totp                       bool                               `json:"totp"`
	EmailVerified              bool                               `json:"emailVerified"`
	FirstName                  string                             `json:"firstName"`
	LastName                   string                             `json:"lastName"`
	Email                      string                             `json:"email"`
	FederationLink             string                             `json:"federationLink"`
	Attributes                 map[string][]string                `json:"attributes"`
	DisableableCredentialTypes []any                              `json:"disableableCredentialTypes"`
	RequiredActions            []string                           `json:"requiredActions"`
	Access                     map[string]bool                    `json:"access"`
	ClientRoles                map[string][]string                `json:"clientRoles"`
	RealmRoles                 []string                           `json:"realmRoles"`
	Groups                     []string                           `json:"groups"`
	ServiceAccountClientID     string                             `json:"serviceAccountClientId"`
	Credentials                []gocloak.CredentialRepresentation `json:"credentials"`
}

// allowlistUser keeps only the fields a typical consumer maps, unknown fields are skipped by the decoder.
type allowlistUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Enabled       bool   `json:"enabled"`
	EmailVerified bool   `json:"emailVerified"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Email         string `json:"email"`
	Created       int64  `json:"createdTimestamp"`
}

// BenchmarkUserShape compares decoding 1000 users into gocloak.User (pointer scalars), a value-scalar copy of it and
// an 8-field allowlist struct, each from a byte slice and streamed from a reader.
func BenchmarkUserShape(b *testing.B) {
	body := payloads().users1000
	shapes := []struct {
		name   string
		target func() any
	}{
		{"gocloak", func() any { return new([]*gocloak.User) }},
		{"gocloak-values", func() any { return new([]gocloak.User) }},
		{"flat", func() any { return new([]flatUser) }},
		{"allowlist", func() any { return new([]allowlistUser) }},
	}
	for _, s := range shapes {
		b.Run("shape="+s.name+"/read=unmarshal", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := json.Unmarshal(body, s.target()); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("shape="+s.name+"/read=stream", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := json.NewDecoder(bytes.NewReader(body)).Decode(s.target()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
