package gocloak_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Nerzal/gocloak/v13"

	"github.com/stretchr/testify/assert"
)

func TestStringOrArray_Unmarshal(t *testing.T) {
	t.Parallel()
	jsonString := []byte("\"123\"")
	var dataString gocloak.StringOrArray
	err := json.Unmarshal(jsonString, &dataString)
	assert.NoErrorf(t, err, "Unmarshalling failed for json string: %s", jsonString)
	assert.Len(t, dataString, 1)
	assert.Equal(t, "123", dataString[0])

	jsonArray := []byte("[\"1\",\"2\",\"3\"]")
	var dataArray gocloak.StringOrArray
	err = json.Unmarshal(jsonArray, &dataArray)
	assert.NoError(t, err, "Unmarshalling failed for json array of strings: %s", jsonArray)
	assert.Len(t, dataArray, 3)
	assert.EqualValues(t, []string{"1", "2", "3"}, dataArray)
}

func TestStringOrArray_Marshal(t *testing.T) {
	t.Parallel()
	dataString := gocloak.StringOrArray{"123"}
	jsonString, err := json.Marshal(&dataString)
	assert.NoErrorf(t, err, "Marshaling failed for one string: %s", dataString)
	assert.Equal(t, "\"123\"", string(jsonString))

	dataArray := gocloak.StringOrArray{"1", "2", "3"}
	jsonArray, err := json.Marshal(&dataArray)
	assert.NoError(t, err, "Marshaling failed for array of strings: %s", dataArray)
	assert.Equal(t, "[\"1\",\"2\",\"3\"]", string(jsonArray))
}

func TestEnforcedString_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	type testData struct {
		In  []byte
		Out gocloak.EnforcedString
	}

	data := []testData{{
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
		var val gocloak.EnforcedString
		err := json.Unmarshal(d.In, &val)
		assert.NoErrorf(t, err, "Unmarshalling failed with data: %v", d.In)
		assert.Equal(t, d.Out, val)
	}
}

func TestEnforcedString_MarshalJSON(t *testing.T) {
	t.Parallel()

	data := gocloak.EnforcedString("foo")
	jsonString, err := json.Marshal(&data)
	assert.NoErrorf(t, err, "Unmarshalling failed with data: %v", data)
	assert.Equal(t, `"foo"`, string(jsonString))
}

func TestGetQueryParams(t *testing.T) {
	t.Parallel()

	type TestParams struct {
		IntField    *int    `json:"int_field,string,omitempty"`
		StringField *string `json:"string_field,omitempty"`
		BoolField   *bool   `json:"bool_field,string,omitempty"`
	}

	params, err := gocloak.GetQueryParams(TestParams{})
	assert.NoError(t, err)
	assert.True(
		t,
		len(params) == 0,
		"Params must be empty, but got: %+v",
		params,
	)

	params, err = gocloak.GetQueryParams(TestParams{
		IntField:    gocloak.IntP(1),
		StringField: gocloak.StringP("fake"),
		BoolField:   gocloak.BoolP(true),
	})
	assert.NoError(t, err)
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
		StringField: gocloak.StringP("fake"),
		BoolField:   gocloak.BoolP(false),
	})
	assert.NoError(t, err)
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
			Error:    errors.New("something something invalid_grant something"),
			Expected: gocloak.APIErrTypeInvalidGrant,
		},
		{
			Name:     "other error",
			Error:    errors.New("something something unsupported_grant_type something"),
			Expected: gocloak.APIErrTypeUnknown,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := gocloak.ParseAPIErrType(testCase.Error)
			if result != testCase.Expected {
				t.Fatalf("expected %s but received %s", testCase.Expected, result)
			}
		})
	}
}

func TestStringer(t *testing.T) {
	// nested structs
	actions := []string{"someAction", "anotherAction"}
	access := gocloak.AccessRepresentation{
		Manage:      gocloak.BoolP(true),
		Impersonate: gocloak.BoolP(false),
	}
	v := gocloak.PermissionTicketDescriptionRepresentation{
		ID:               gocloak.StringP("someID"),
		CreatedTimeStamp: gocloak.Int64P(1607702613),
		Enabled:          gocloak.BoolP(true),
		RequiredActions:  &actions,
		Access:           &access,
	}

	str := v.String()

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
			Name:   gocloak.StringP("someMapper"),
			Config: &config,
		},
	}
	clients := []gocloak.Client{
		{
			Name:            gocloak.StringP("someClient"),
			ProtocolMappers: &pmappers,
		},
		{
			Name: gocloak.StringP("AnotherClient"),
		},
	}

	realmRep := gocloak.RealmRepresentation{
		DisplayName: gocloak.StringP("someRealm"),
		Clients:     &clients,
	}

	str = realmRep.String()
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

type Stringable interface {
	String() string
}

func TestStringerOmitEmpty(t *testing.T) {
	customs := []Stringable{
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
		&gocloak.Attributes{},
		&gocloak.Access{},
		&gocloak.UserGroup{},
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
	}

	for _, custom := range customs {
		assert.Equal(t, "{}", custom.(Stringable).String())
	}
}
