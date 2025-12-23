package falkordb

import "github.com/snowmerak/falkordb-go/util/strs"

// Deprecated: use util/strs.ToString.
func ToString(i interface{}) string { return strs.ToString(i) }

// Deprecated: use util/strs.RandomString.
func RandomString(n int) string { return strs.RandomString(n) }
