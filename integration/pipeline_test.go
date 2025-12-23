package integration_test

import (
	"testing"

	"github.com/snowmerak/falkordb-go/domain"
	"github.com/snowmerak/falkordb-go/graph"
	"github.com/stretchr/testify/assert"
)

func TestPipelineQueries(t *testing.T) {
	createGraph()

	reqs := []graph.QueryRequest{
		{
			Query:   "MATCH (p:Person) RETURN p",
			Options: graph.NewQueryOptions().SetTimeout(50),
		},
		{
			Command: graph.CmdROQuery,
			Query:   "MATCH (c:Country {name:$name}) RETURN c",
			Params:  map[string]interface{}{"name": "Japan"},
		},
	}

	results, err := graphInstance.Pipeline(reqs)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// First result: single Person node
	assert.Equal(t, 1, len(results[0].Results()))
	assert.True(t, results[0].Next())
	rec0 := results[0].Record()
	pNode, ok := rec0.GetByIndex(0).(*domain.Node)
	assert.True(t, ok)
	assert.Equal(t, "Person", pNode.Labels[0])

	// Second result: single Country node filtered by param
	assert.Equal(t, 1, len(results[1].Results()))
	assert.True(t, results[1].Next())
	rec1 := results[1].Record()
	cNode, ok := rec1.GetByIndex(0).(*domain.Node)
	assert.True(t, ok)
	assert.Equal(t, "Country", cNode.Labels[0])
	assert.Equal(t, "Japan", cNode.GetProperty("name"))
}
