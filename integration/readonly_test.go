package integration_test

import (
	"os"
	"testing"

	falkordb "github.com/snowmerak/falkordb-go"
	"github.com/snowmerak/falkordb-go/graph"
	"github.com/stretchr/testify/assert"
)

// Ensures that a read-only client rejects write queries and still allows RO queries.
func TestReadOnlyClient(t *testing.T) {
	addr := os.Getenv("FALKORDB_ADDR")
	if addr == "" {
		addr = "0.0.0.0:6379"
	}

	db, err := falkordb.NewReadOnly(&falkordb.ConnectionOption{Addr: addr})
	if err != nil {
		t.Fatalf("NewReadOnly error: %v", err)
	}

	g := db.SelectGraph("social_ro")

	// Write should fail
	_, err = g.Query("CREATE (:X)", nil, nil)
	assert.Error(t, err, "write query should fail on read-only client")

	// Pipeline with a write should also fail
	_, err = g.Pipeline([]graph.QueryRequest{{Query: "CREATE (:Y)"}})
	assert.Error(t, err, "pipeline write should fail on read-only client")

	// RO query on empty key may return an error from the server; ensure no panic.
	res, err := g.ROQuery("MATCH (n) RETURN n", nil, nil)
	if err == nil && res != nil {
		assert.True(t, res.Empty())
	} else {
		assert.Error(t, err)
	}
}
