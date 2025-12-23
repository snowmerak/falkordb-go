package strs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils(t *testing.T) {
	res := RandomString(10)
	assert.Equal(t, len(res), 10)

	res = ToString("test_string")
	assert.Equal(t, res, "\"test_string\"")

	res = ToString(10)
	assert.Equal(t, res, "10")

	res = ToString(1.2)
	assert.Equal(t, res, "1.2")

	res = ToString(true)
	assert.Equal(t, res, "true")

	var arr = []interface{}{1, 2, 3, "boom"}
	res = ToString(arr)
	assert.Equal(t, res, "[1,2,3,\"boom\"]")

	jsonMap := make(map[string]interface{})
	jsonMap["object"] = map[string]interface{}{"foo": 1}
	res = ToString(jsonMap)
	assert.Equal(t, res, "{object: {foo: 1}}")
}
