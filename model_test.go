package gocloak

import (
	"reflect"
	"testing"
)

func TestRequestingPartyTokenOptionsFormData(t *testing.T) {
	tests := []struct {
		name     string
		input    *RequestingPartyTokenOptions
		expected map[string]string
	}{
		{
			name:  "Empty input",
			input: &RequestingPartyTokenOptions{},
			expected: map[string]string{
				"grant_type":                     "urn:ietf:params:oauth:grant-type:uma-ticket",
				"response_include_resource_name": "true",
			},
		},
		{
			name: "With grant type and response include resource name",
			input: &RequestingPartyTokenOptions{
				GrantType:                   ptr("custom_grant_type"),
				ResponseIncludeResourceName: ptr(false),
			},
			expected: map[string]string{
				"grant_type":                     "custom_grant_type",
				"response_include_resource_name": "false",
			},
		},
		{
			name: "With various field types",
			input: &RequestingPartyTokenOptions{
				Ticket:                        ptr("ticket123"),
				PermissionResourceMatchingURI: ptr(true),
				ResponsePermissionsLimit:      ptr(uint32(10)),
			},
			expected: map[string]string{
				"grant_type":                       "urn:ietf:params:oauth:grant-type:uma-ticket",
				"response_include_resource_name":   "true",
				"ticket":                           "ticket123",
				"permission_resource_matching_uri": "true",
				"response_permissions_limit":       "10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.FormData()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FormData() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function for creating pointers to values
func ptr[T any](v T) *T {
	return &v
}
