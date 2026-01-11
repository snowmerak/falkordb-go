package integration_test

import (
	"os"
	"testing"

	falkordb "github.com/snowmerak/falkordb-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUDFAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	addr := os.Getenv("FALKORDB_ADDR")
	if addr == "" {
		addr = "0.0.0.0:6379"
	}
	db, err := falkordb.FromURL("falkor://" + addr)
	require.NoError(t, err)

	// 1. Flush existing
	err = db.FlushUDFs()
	require.NoError(t, err)

	// Verify empty
	libs, err := db.ListUDF()
	require.NoError(t, err)
	assert.Empty(t, libs)

	// 2. Load a simple UDF
	// Using JS syntax as seen in flex.js
	// falkor.register('test.add', function(a,b){return a+b;});
	// Note: using block syntax to be safe
	simpleUDF := `
function myAdd(a, b) {
    return a + b;
}
falkor.register('test.add', myAdd);
`
	err = db.LoadUDF("testlib", simpleUDF)
	require.NoError(t, err)

	// 3. List
	libs, err = db.ListUDF()
	require.NoError(t, err)
	assert.Len(t, libs, 1)
	if len(libs) > 0 {
		assert.Equal(t, "testlib", libs[0].Name)
		assert.Contains(t, libs[0].Functions, "test.add")
		assert.Empty(t, libs[0].Code, "Code should be empty by default")
	}

	// 4. List With Code
	libs, err = db.ListUDF(falkordb.WithUDFCode())
	require.NoError(t, err)
	assert.Len(t, libs, 1)
	if len(libs) > 0 {
		assert.Equal(t, simpleUDF, libs[0].Code)
	}

	// 5. List With Name
	libs, err = db.ListUDF(falkordb.WithUDFLibrary("testlib"))
	require.NoError(t, err)
	assert.Len(t, libs, 1)

	// 6. List With Name mismatch
	// Assuming non-existent library returns empty list or error.
	// We'll see what behavior is.
	libs, err = db.ListUDF(falkordb.WithUDFLibrary("nomatch"))
	// If it returns error, we handle. If empty list, we handle.
	if err == nil {
		assert.Empty(t, libs)
	} else {
		// Verify if it's a "not found" error if that's expected
		// But usually LIST returns empty set.
		t.Logf("List mismatch returned error: %v", err)
	}

	// 7. Delete
	err = db.DeleteUDF("testlib")
	require.NoError(t, err)

	libs, err = db.ListUDF()
	require.NoError(t, err)
	assert.Empty(t, libs)

	// 8. Test Flush again with multiple
	// Note: Reuse same code string
	err = db.LoadUDF("lib1", simpleUDF)
	require.NoError(t, err)
	err = db.LoadUDF("lib2", simpleUDF)
	require.NoError(t, err)

	libs, err = db.ListUDF()
	require.NoError(t, err)
	assert.Len(t, libs, 2)

	err = db.FlushUDFs()
	require.NoError(t, err)

	libs, err = db.ListUDF()
	require.NoError(t, err)
	assert.Empty(t, libs)
}
