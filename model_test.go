package gocloak

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringOrArray_Unmarshal(t *testing.T) {
	t.Parallel()
	jsonString := []byte("\"123\"")
	var dataString StringOrArray
	err := json.Unmarshal(jsonString, &dataString)
	assert.NoErrorf(t, err, "Unmarshalling failed for json string: %s", jsonString)
	assert.Len(t, dataString, 1)
	assert.Equal(t, "123", dataString[0])

	jsonArray := []byte("[\"1\",\"2\",\"3\"]")
	var dataArray StringOrArray
	err = json.Unmarshal(jsonArray, &dataArray)
	assert.NoError(t, err, "Unmarshalling failed for json array of strings: %s", jsonArray)
	assert.Len(t, dataArray, 3)
	assert.EqualValues(t, []string{"1", "2", "3"}, dataArray)
}

func TestStringOrArray_Marshal(t *testing.T) {
	t.Parallel()
	dataString := StringOrArray{"123"}
	jsonString, err := json.Marshal(&dataString)
	assert.NoErrorf(t, err, "Marshalling failed for one string: %s", dataString)
	assert.Equal(t, "\"123\"", string(jsonString))

	dataArray := StringOrArray{"1", "2", "3"}
	jsonArray, err := json.Marshal(&dataArray)
	assert.NoError(t, err, "Marshalling failed for array of strings: %s", dataString)
	assert.Equal(t, "[\"1\",\"2\",\"3\"]", string(jsonArray))
}
