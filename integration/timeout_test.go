package integration_test

import (
	"testing"

	"github.com/snowmerak/falkordb-go/graph"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	// Instantiate a new QueryOptions struct with a 1-second timeout
	options := graph.NewQueryOptions().SetTimeout(1)

	// Verify that the timeout was set properly
	assert.Equal(t, 1, options.GetTimeout())

	// Issue a long-running query with a 1-millisecond timeout.
	res, err := graphInstance.Query("UNWIND range(0, 1000000) AS v WITH v WHERE v % 2 = 1 RETURN COUNT(v)", nil, options)
	assert.Nil(t, res)
	assert.NotNil(t, err)

	params := make(map[string]interface{})
	params["ub"] = 1000000
	res, err = graphInstance.Query("UNWIND range(0, $ub) AS v RETURN v", params, options)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
