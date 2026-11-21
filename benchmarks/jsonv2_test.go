//go:build goexperiment.jsonv2

package benchmarks

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"reflect"
	"testing"
)

func init() {
	jsonAPIs = append(jsonAPIs, jsonAPI{
		name:      "v2",
		marshal:   func(v any) ([]byte, error) { return jsonv2.Marshal(v) },
		unmarshal: func(data []byte, v any) error { return jsonv2.Unmarshal(data, v) },
	})
}

// TestJSONv2Equivalence checks that the direct v2 API decodes every payload exactly like v1, so a switch would
// not change results. v2 differs in defaults (case-sensitive names, duplicate names rejected, invalid UTF-8 rejected).
func TestJSONv2Equivalence(t *testing.T) {
	p := payloads()
	for _, c := range decodeCases {
		t.Run(c.name, func(t *testing.T) {
			want, got := c.target(), c.target()
			if err := jsonv1.Unmarshal(c.body(p), want); err != nil {
				t.Fatal(err)
			}
			if err := jsonv2.Unmarshal(c.body(p), got); err != nil {
				t.Fatalf("v2 unmarshal: %v", err)
			}
			if !reflect.DeepEqual(want, got) {
				t.Errorf("v2 result differs from v1")
			}
		})
	}
}
