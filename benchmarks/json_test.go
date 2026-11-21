package benchmarks

import (
	"encoding/json"
	"testing"

	"github.com/Nerzal/gocloak/v15"
)

// jsonAPI is one JSON implementation under test.
type jsonAPI struct {
	name      string
	marshal   func(v any) ([]byte, error)
	unmarshal func(data []byte, v any) error
}

// jsonAPIs lists the implementations; jsonv2_test.go appends the direct v2 API under GOEXPERIMENT=jsonv2.
// Note: under GOEXPERIMENT=jsonv2 "v1" is the v1 API backed by the v2 engine, without any code change.
var jsonAPIs = []jsonAPI{{"v1", json.Marshal, json.Unmarshal}}

// decodeCase is one response body and a constructor for its target type.
type decodeCase struct {
	name   string
	body   func(*payloadSet) []byte
	target func() any
}

var decodeCases = []decodeCase{
	{"token", func(p *payloadSet) []byte { return p.token }, func() any { return new(gocloak.JWT) }},
	{"certs", func(p *payloadSet) []byte { return p.certs }, func() any { return new(gocloak.CertResponse) }},
	{"user", func(p *payloadSet) []byte { return p.user }, func() any { return new(gocloak.User) }},
	{"users100", func(p *payloadSet) []byte { return p.users100 }, func() any { return new([]*gocloak.User) }},
	{"users1000", func(p *payloadSet) []byte { return p.users1000 }, func() any { return new([]*gocloak.User) }},
	{"groups", func(p *payloadSet) []byte { return p.groups }, func() any { return new([]*gocloak.Group) }},
	{"clients", func(p *payloadSet) []byte { return p.clients }, func() any { return new([]*gocloak.Client) }},
	{"realm", func(p *payloadSet) []byte { return p.realm }, func() any { return new(gocloak.RealmRepresentation) }},
}

// BenchmarkJSONUnmarshal decodes typical Keycloak responses into gocloak models.
func BenchmarkJSONUnmarshal(b *testing.B) {
	p := payloads()
	for _, c := range decodeCases {
		body := c.body(p)
		for _, api := range jsonAPIs {
			b.Run("payload="+c.name+"/api="+api.name, func(b *testing.B) {
				b.SetBytes(int64(len(body)))
				b.ReportAllocs()
				for b.Loop() {
					if err := api.unmarshal(body, c.target()); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// BenchmarkJSONMarshal encodes typical request bodies (create user, update client, realm).
func BenchmarkJSONMarshal(b *testing.B) {
	p := payloads()
	var clients []*gocloak.Client
	must(json.Unmarshal(p.clients, &clients))
	var realm gocloak.RealmRepresentation
	must(json.Unmarshal(p.realm, &realm))
	cases := []struct {
		name string
		v    any
	}{
		{"createUser", p.createUser},
		{"client", clients[0]},
		{"realm", realm},
	}
	for _, c := range cases {
		for _, api := range jsonAPIs {
			b.Run("payload="+c.name+"/api="+api.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := api.marshal(c.v); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
