package domain

import (
	"fmt"
	"strconv"
)

// toString converts supported types to their Cypher string representation.
func toString(i interface{}) string {
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
	default:
		return strconv.Quote(fmt.Sprint(v))
	}
}
