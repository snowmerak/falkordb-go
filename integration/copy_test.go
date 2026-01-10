package integration_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCopyGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	// Ensure destination graph does not exist
	destGraphName := "social_copy"
	destGraph := db.SelectGraph(destGraphName)
	destGraph.Delete()

	// Copy "social" to "social_copy"
	err := db.CopyGraph("social", destGraphName)
	assert.Nil(t, err)

	// Verify data in copied graph
	res, err := destGraph.Query("MATCH (n) RETURN count(n)", nil, nil)
	assert.Nil(t, err)
	assert.False(t, res.Empty())
	
	// Expecting same number of nodes/etc.
	// In createGraph(), we create 2 nodes.
	val := res.Results()[0].Values()[0]
	assert.Equal(t, int64(2), val)

	// Cleanup
	destGraph.Delete()
}
