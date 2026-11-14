package gocloak_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/Nerzal/gocloak/v14"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringOrArray_Unmarshal(t *testing.T) {
	t.Parallel()
	jsonString := []byte("\"123\"")
	var dataString gocloak.StringOrArray
	err := json.Unmarshal(jsonString, &dataString)
	require.NoErrorf(t, err, "Unmarshalling failed for json string: %s", jsonString)
	require.Len(t, dataString, 1)
	assert.Equal(t, "123", dataString[0])

	jsonArray := []byte("[\"1\",\"2\",\"3\"]")
	var dataArray gocloak.StringOrArray
	err = json.Unmarshal(jsonArray, &dataArray)
	require.NoError(t, err, "Unmarshalling failed for json array of strings: %s", jsonArray)
	assert.Equal(t, gocloak.StringOrArray{"1", "2", "3"}, dataArray)
}

func TestStringOrArray_Marshal(t *testing.T) {
	t.Parallel()
	one := gocloak.StringOrArray{"123"}
	many := gocloak.StringOrArray{"1", "2", "3"}
	tests := []struct {
		name     string
		in       any
		expected string
	}{
		{"one string by pointer", &one, `"123"`},
		{"one string by value", one, `"123"`},
		{"array by pointer", &many, `["1","2","3"]`},
		{"array by value", many, `["1","2","3"]`},
		{"single value in struct", struct{ Aud gocloak.StringOrArray }{one}, `{"Aud":"123"}`},
		{"nil", gocloak.StringOrArray(nil), `null`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := json.Marshal(tc.in)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, string(out))
		})
	}
}

func TestEnforcedString_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	type testData struct {
		In  []byte
		Out gocloak.EnforcedString
	}

	data := []testData{{
		In:  []byte(`null`),
		Out: "",
	}, {
		In:  []byte(`{"foo": "b\"a\\r\n"}`),
		Out: `{"foo": "b\"a\\r\n"}`,
	}, {
		In:  []byte(`"string value"`),
		Out: "string value",
	}, {
		In:  []byte(`"\"quoted string value\""`),
		Out: `"quoted string value"`,
	}, {
		In:  []byte(`true`),
		Out: "true",
	}, {
		In:  []byte(`42`),
		Out: "42",
	}, {
		In:  []byte(`{"foo": "bar"}`),
		Out: `{"foo": "bar"}`,
	}, {
		In:  []byte(`["foo"]`),
		Out: `["foo"]`,
	}}

	for _, d := range data {
		t.Run(string(d.In), func(t *testing.T) {
			t.Parallel()
			var val gocloak.EnforcedString
			err := json.Unmarshal(d.In, &val)
			require.NoErrorf(t, err, "Unmarshalling failed with data: %s", d.In)
			assert.Equal(t, d.Out, val)
		})
	}
}

func TestEnforcedString_MarshalJSON(t *testing.T) {
	t.Parallel()

	data := gocloak.EnforcedString("foo")
	jsonString, err := json.Marshal(&data)
	require.NoErrorf(t, err, "Marshaling failed with data: %v", data)
	assert.JSONEq(t, `"foo"`, string(jsonString))
}

func TestGetQueryParams(t *testing.T) {
	t.Parallel()

	type TestParams struct {
		IntField    *int    `json:"int_field,string,omitempty"`
		StringField *string `json:"string_field,omitempty"`
		BoolField   *bool   `json:"bool_field,string,omitempty"`
	}

	params, err := gocloak.GetQueryParams(TestParams{})
	require.NoError(t, err)
	assert.Empty(
		t,
		params,
		"Params must be empty, but got: %+v",
		params,
	)

	params, err = gocloak.GetQueryParams(TestParams{
		IntField:    new(1),
		StringField: new("fake"),
		BoolField:   new(true),
	})
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			"int_field":    "1",
			"string_field": "fake",
			"bool_field":   "true",
		},
		params,
	)

	params, err = gocloak.GetQueryParams(TestParams{
		StringField: new("fake"),
		BoolField:   new(false),
	})
	require.NoError(t, err)
	assert.Equal(
		t,
		map[string]string{
			"string_field": "fake",
			"bool_field":   "false",
		},
		params,
	)
}

func TestParseAPIErrType(t *testing.T) {
	testCases := []struct {
		Name     string
		Error    error
		Expected gocloak.APIErrType
	}{
		{
			Name:     "nil error",
			Error:    nil,
			Expected: gocloak.APIErrTypeUnknown,
		},
		{
			Name:     "invalid grant",
			Error:    errors.New("something invalid_grant something"),
			Expected: gocloak.APIErrTypeInvalidGrant,
		},
		{
			Name:     "other error",
			Error:    errors.New("something unsupported_grant_type something"),
			Expected: gocloak.APIErrTypeUnknown,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			assert.Equal(t, testCase.Expected, gocloak.ParseAPIErrType(testCase.Error))
		})
	}
}

func TestRequestingPartyTokenOptionsFormData(t *testing.T) {
	tests := []struct {
		name     string
		input    *gocloak.RequestingPartyTokenOptions
		expected map[string]string
	}{
		{
			name:  "Empty input",
			input: &gocloak.RequestingPartyTokenOptions{},
			expected: map[string]string{
				"grant_type":                     "urn:ietf:params:oauth:grant-type:uma-ticket",
				"response_include_resource_name": "true",
			},
		},
		{
			name: "With grant type and response include resource name",
			input: &gocloak.RequestingPartyTokenOptions{
				GrantType:                   new("custom_grant_type"),
				ResponseIncludeResourceName: new(false),
			},
			expected: map[string]string{
				"grant_type":                     "custom_grant_type",
				"response_include_resource_name": "false",
			},
		},
		{
			name: "With various field types",
			input: &gocloak.RequestingPartyTokenOptions{
				Ticket:                        new("ticket123"),
				PermissionResourceMatchingURI: new(true),
				ResponsePermissionsLimit:      new(uint32(10)),
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
			assert.Equal(t, tt.expected, tt.input.FormData())
		})
	}
}

// prettyJSON marshals v as tab-indented JSON. Input: t [*testing.T], v [any]. Output: [string].
func prettyJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "	")
	require.NoError(t, err)
	return string(b)
}

// TestPrettyJSON checks that nested optional fields are omitted when marshaling models.
func TestPrettyJSON(t *testing.T) {
	// nested structs
	actions := []string{"someAction", "anotherAction"}
	access := gocloak.AccessRepresentation{
		Manage:      new(true),
		Impersonate: new(false),
	}
	v := gocloak.PermissionTicketDescriptionRepresentation{
		ID:               new("someID"),
		CreatedTimeStamp: new(int64(1607702613)),
		Enabled:          new(true),
		RequiredActions:  actions,
		Access:           &access,
	}

	str := prettyJSON(t, &v)

	expectedStr := `{
	"id": "someID",
	"createdTimestamp": 1607702613,
	"enabled": true,
	"requiredActions": [
		"someAction",
		"anotherAction"
	],
	"access": {
		"impersonate": false,
		"manage": true
	}
}`

	assert.Equal(t, expectedStr, str)

	// nested arrays
	config := make(map[string]string)
	config["bar"] = "foo"
	config["ping"] = "pong"

	pmappers := []gocloak.ProtocolMapperRepresentation{
		{
			Name:   new("someMapper"),
			Config: config,
		},
	}
	clients := []gocloak.Client{
		{
			Name:            new("someClient"),
			ProtocolMappers: pmappers,
		},
		{
			Name: new("AnotherClient"),
		},
	}

	realmRep := gocloak.RealmRepresentation{
		DisplayName: new("someRealm"),
		Clients:     clients,
	}

	str = prettyJSON(t, &realmRep)
	expectedStr = `{
	"clients": [
		{
			"name": "someClient",
			"protocolMappers": [
				{
					"config": {
						"bar": "foo",
						"ping": "pong"
					},
					"name": "someMapper"
				}
			]
		},
		{
			"name": "AnotherClient"
		}
	],
	"displayName": "someRealm"
}`
	assert.Equal(t, expectedStr, str)
}

// TestModelsOmitEmpty checks that zero-value models marshal to an empty JSON object.
func TestModelsOmitEmpty(t *testing.T) {
	customs := []any{
		&gocloak.CertResponseKey{},
		&gocloak.CertResponse{},
		&gocloak.IssuerResponse{},
		&gocloak.ResourcePermission{},
		&gocloak.PermissionResource{},
		&gocloak.PermissionScope{},
		&gocloak.IntroSpectTokenResult{},
		&gocloak.User{},
		&gocloak.SetPasswordRequest{},
		&gocloak.Component{},
		&gocloak.KeyStoreConfig{},
		&gocloak.ActiveKeys{},
		&gocloak.Key{},
		&gocloak.ExecuteActionsEmail{},
		&gocloak.Group{},
		&gocloak.GroupsCount{},
		&gocloak.GetGroupsParams{},
		&gocloak.CompositesRepresentation{},
		&gocloak.Role{},
		&gocloak.GetRoleParams{},
		&gocloak.ClientMappingsRepresentation{},
		&gocloak.MappingsRepresentation{},
		&gocloak.ClientScope{},
		&gocloak.ClientScopeAttributes{},
		&gocloak.ProtocolMappers{},
		&gocloak.ProtocolMappersConfig{},
		&gocloak.Client{},
		&gocloak.ResourceServerRepresentation{},
		&gocloak.RoleDefinition{},
		&gocloak.PolicyRepresentation{},
		&gocloak.RolePolicyRepresentation{},
		&gocloak.JSPolicyRepresentation{},
		&gocloak.ClientPolicyRepresentation{},
		&gocloak.TimePolicyRepresentation{},
		&gocloak.UserPolicyRepresentation{},
		&gocloak.AggregatedPolicyRepresentation{},
		&gocloak.GroupPolicyRepresentation{},
		&gocloak.GroupDefinition{},
		&gocloak.ResourceRepresentation{},
		&gocloak.ResourceOwnerRepresentation{},
		&gocloak.ScopeRepresentation{},
		&gocloak.ProtocolMapperRepresentation{},
		&gocloak.UserInfoAddress{},
		&gocloak.UserInfo{},
		&gocloak.RolesRepresentation{},
		&gocloak.RealmRepresentation{},
		&gocloak.MultiValuedHashMap{},
		&gocloak.TokenOptions{},
		&gocloak.UserSessionRepresentation{},
		&gocloak.SystemInfoRepresentation{},
		&gocloak.MemoryInfoRepresentation{},
		&gocloak.ServerInfoRepresentation{},
		&gocloak.FederatedIdentityRepresentation{},
		&gocloak.IdentityProviderRepresentation{},
		&gocloak.GetResourceParams{},
		&gocloak.GetScopeParams{},
		&gocloak.GetPolicyParams{},
		&gocloak.GetPermissionParams{},
		&gocloak.GetUsersByRoleParams{},
		&gocloak.PermissionRepresentation{},
		&gocloak.CreatePermissionTicketParams{},
		&gocloak.PermissionTicketDescriptionRepresentation{},
		&gocloak.AccessRepresentation{},
		&gocloak.PermissionTicketResponseRepresentation{},
		&gocloak.PermissionTicketRepresentation{},
		&gocloak.PermissionTicketPermissionRepresentation{},
		&gocloak.PermissionGrantParams{},
		&gocloak.PermissionGrantResponseRepresentation{},
		&gocloak.GetUserPermissionParams{},
		&gocloak.ResourcePolicyRepresentation{},
		&gocloak.GetResourcePoliciesParams{},
		&gocloak.CredentialRepresentation{},
		&gocloak.GetUsersParams{},
		&gocloak.GetComponentsParams{},
		&gocloak.GetClientsParams{},
		&gocloak.RequestingPartyTokenOptions{},
		&gocloak.RequestingPartyPermission{},
		&gocloak.GetClientUserSessionsParams{},
		&gocloak.GetOrganizationsParams{},
		&gocloak.OrganizationDomainRepresentation{},
		&gocloak.OrganizationRepresentation{},
	}

	for _, custom := range customs {
		assert.Equal(t, "{}", prettyJSON(t, custom))
	}
}

func TestProtocolMappersConfig_IncludedCustomAudience(t *testing.T) {
	t.Parallel()
	data := []byte(`{
		"protocolMapper": "oidc-audience-mapper",
		"config": {
			"included.custom.audience": "https://example.com/api",
			"access.token.claim": "true",
			"id.token.claim": "false"
		}
	}`)

	var mapper gocloak.ProtocolMappers
	err := json.Unmarshal(data, &mapper)
	require.NoError(t, err)
	require.NotNil(t, mapper.ProtocolMappersConfig)
	assert.Equal(t, "https://example.com/api", gocloak.PString(mapper.ProtocolMappersConfig.IncludedCustomAudience))

	out, err := json.Marshal(mapper.ProtocolMappersConfig)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"included.custom.audience":"https://example.com/api"`)
}

func TestModel_GetGroupsParamsMarshal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		in       gocloak.GetGroupsParams
		expected map[string]string
	}{
		{"neither", gocloak.GetGroupsParams{}, map[string]string{}},
		{"brief only", gocloak.GetGroupsParams{BriefRepresentation: new(true)}, map[string]string{"briefRepresentation": "true", "full": "false"}},
		{"full only", gocloak.GetGroupsParams{Full: new(true)}, map[string]string{"briefRepresentation": "false", "full": "true"}},
		{"both, brief wins", gocloak.GetGroupsParams{BriefRepresentation: new(false), Full: new(false)}, map[string]string{"briefRepresentation": "false", "full": "true"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			params, err := gocloak.GetQueryParams(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, params)
		})
	}
}

func TestModel_TokenOptionsFormData(t *testing.T) {
	t.Parallel()
	opts := gocloak.TokenOptions{
		ClientID:      new("client"),
		ClientSecret:  new("secret"),
		GrantType:     new("password"),
		Scopes:        []string{"openid", "profile"},
		ResponseTypes: []string{"code", "id_token"},
	}
	assert.Equal(t, map[string]string{
		"client_id":     "client",
		"grant_type":    "password",
		"scope":         "openid profile",
		"response_type": "code id_token",
	}, opts.FormData())
	// FormData writes the joined values back into the receiver
	assert.Equal(t, "openid profile", gocloak.PString(opts.Scope))
	assert.Equal(t, "code id_token", gocloak.PString(opts.ResponseType))

	empty := gocloak.TokenOptions{}
	assert.Equal(t, map[string]string{"response_type": "token"}, empty.FormData())
}

func TestModel_OrganizationInviteUserParamsFormData(t *testing.T) {
	t.Parallel()
	params := gocloak.OrganizationInviteUserParams{Email: new("a@b.c"), FirstName: new("A")}
	assert.Equal(t, map[string]string{"email": "a@b.c", "firstName": "A"}, params.FormData())
	assert.Empty(t, (&gocloak.OrganizationInviteUserParams{}).FormData())
}

func TestModel_HTTPErrorResponse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		in       gocloak.HTTPErrorResponse
		expected string
	}{
		{"empty", gocloak.HTTPErrorResponse{}, ""},
		{"error only", gocloak.HTTPErrorResponse{Error: "invalid_grant"}, "invalid_grant"},
		{"message only", gocloak.HTTPErrorResponse{Message: "msg"}, "msg"},
		{"description only", gocloak.HTTPErrorResponse{Description: "desc"}, "desc"},
		{"all", gocloak.HTTPErrorResponse{Error: "e", Message: "m", Description: "d"}, "e: m: d"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.in.String())
			assert.Equal(t, tc.expected != "", tc.in.NotEmpty())
		})
	}
}

func TestModel_APIErrorError(t *testing.T) {
	t.Parallel()
	var err error = &gocloak.APIError{Code: 404, Message: "404 Not Found"}
	require.EqualError(t, err, "404 Not Found")
	require.EqualError(t, gocloak.APIError{Message: "by value"}, "by value")
}

// fillAllFields sets every exported pointer and slice field of the struct v points to a non-zero value.
// Input: v [any] pointer to a struct. Output: [any] the same pointer.
func fillAllFields(v any) any {
	for _, field := range reflect.ValueOf(v).Elem().Fields() {
		if !field.CanSet() {
			continue
		}
		switch field.Kind() {
		case reflect.Pointer:
			elem := reflect.New(field.Type().Elem())
			setSample(elem.Elem())
			field.Set(elem)
		case reflect.Slice:
			elem := reflect.New(field.Type().Elem()).Elem()
			setSample(elem)
			field.Set(reflect.Append(reflect.MakeSlice(field.Type(), 0, 1), elem))
		default:
			setSample(field)
		}
	}
	return v
}

// setSample sets a non-zero sample value for basic kinds. Input: v [reflect.Value] settable value.
func setSample(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		v.SetString("x")
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(1)
	default:
	}
}

// TestGetQueryParams_AllParamsTypes makes sure every struct passed to GetQueryParams by the client works with all fields set.
func TestGetQueryParams_AllParamsTypes(t *testing.T) {
	t.Parallel()
	params := []any{
		&gocloak.GetAdminEventsParams{},
		&gocloak.GetChildGroupsParams{},
		&gocloak.GetClientsParams{},
		&gocloak.GetClientUserSessionsParams{},
		&gocloak.GetComponentsParams{},
		&gocloak.GetEventsParams{},
		&gocloak.GetGroupsParams{},
		&gocloak.GetMembersParams{},
		&gocloak.GetOrganizationCountParams{},
		&gocloak.GetOrganizationGroupsParams{},
		&gocloak.GetOrganizationGroupSubgroupsParams{},
		&gocloak.GetOrganizationInvitationsParams{},
		&gocloak.GetOrganizationMembersParams{},
		&gocloak.GetOrganizationsParams{},
		&gocloak.GetPermissionParams{},
		&gocloak.GetPolicyParams{},
		&gocloak.GetResourceParams{},
		&gocloak.GetResourcePoliciesParams{},
		&gocloak.GetRoleParams{},
		&gocloak.GetScopeParams{},
		&gocloak.GetUserPermissionParams{},
		&gocloak.GetUsersByRoleParams{},
		&gocloak.GetUsersParams{},
		&gocloak.ExecuteActionsEmail{},
	}
	for _, p := range params {
		t.Run(reflect.TypeOf(p).Elem().Name(), func(t *testing.T) {
			t.Parallel()
			query, err := gocloak.GetQueryParams(reflect.ValueOf(fillAllFields(p)).Elem().Interface())
			require.NoError(t, err)
			require.NotEmpty(t, query)
		})
	}
}
