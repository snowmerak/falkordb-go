package strs

import (
	"crypto/rand"
	"strconv"
	"strings"
)

func arrayToString(arr []interface{}) string {
	strArray := make([]string, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		strArray = append(strArray, ToString(arr[i]))
	}
	return "[" + strings.Join(strArray, ",") + "]"
}

func strArrayToString(arr []string) string {
	strArray := make([]string, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		strArray = append(strArray, ToString(arr[i]))
	}
	return "[" + strings.Join(strArray, ",") + "]"
}

func mapToString(data map[string]interface{}) string {
	pairsArray := make([]string, 0, len(data))
	for k, v := range data {
		pairsArray = append(pairsArray, k+": "+ToString(v))
	}
	return "{" + strings.Join(pairsArray, ",") + "}"
}

// ToString converts supported Go values to Cypher-friendly strings.
func ToString(i interface{}) string {
	if i == nil {
		return "null"
	}

	switch v := i.(type) {
	case string:
		return strconv.Quote(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []interface{}:
		return arrayToString(v)
	case map[string]interface{}:
		return mapToString(v)
	case []string:
		return strArrayToString(v)
	default:
		panic("Unrecognized type to convert to string")
	}
}

// RandomString generates a random alphanumeric string of length n.
func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	output := make([]byte, n)
	randomness := make([]byte, n)
	if _, err := rand.Read(randomness); err != nil {
		panic(err)
	}
	l := len(letterBytes)
	for pos := range output {
		random := uint8(randomness[pos])
		randomPos := random % uint8(l)
		output[pos] = letterBytes[randomPos]
	}
	return string(output)
}
