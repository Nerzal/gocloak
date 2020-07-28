package gocloak_test

import (
	"encoding/json"
	"testing"

	"github.com/Nerzal/gocloak/v7"

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
	assert.NoError(t, err, "Marshaling failed for array of strings: %s", dataString)
	assert.Equal(t, "[\"1\",\"2\",\"3\"]", string(jsonArray))
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
