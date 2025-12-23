package integration_test

import (
	"testing"

	"github.com/snowmerak/falkordb-go/domain"
	"github.com/stretchr/testify/assert"
)

func TestMultiLabelNode(t *testing.T) {
	// clear database
	err := graphInstance.Delete()
	assert.Nil(t, err)

	// create a multi label node
	_, err = graphInstance.Query("CREATE (:A:B)", nil, nil)
	assert.Nil(t, err)

	// fetch node
	res, err := graphInstance.Query("MATCH (n) RETURN n", nil, nil)
	assert.Nil(t, err)

	res.Next()
	r := res.Record()
	n := r.GetByIndex(0).(*domain.Node)

	// expecting 2 labels
	assert.Equal(t, len(n.Labels), 2, "expecting 2 labels")
	assert.Equal(t, n.Labels[0], "A")
	assert.Equal(t, n.Labels[1], "B")
}

func TestNodeMapDatatype(t *testing.T) {
	err := graphInstance.Delete()
	assert.Nil(t, err)

	// Create 2 nodes connect via a single edge.
	res, err := graphInstance.Query("CREATE (:Person {name: 'John Doe', age: 33, gender: 'male', status: 'single'})-[:Visited {year: 2017}]->(c:Country {name: 'Japan', population: 126800000, states: ['Kanto', 'Chugoku']})", nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 2, res.NodesCreated(), "Expecting 2 node created")
	assert.Equal(t, 0, res.NodesDeleted(), "Expecting 0 nodes deleted")
	assert.Equal(t, 8, res.PropertiesSet(), "Expecting 8 properties set")
	assert.Equal(t, 1, res.RelationshipsCreated(), "Expecting 1 relationships created")
	assert.Equal(t, 0, res.RelationshipsDeleted(), "Expecting 0 relationships deleted")
	assert.Greater(t, res.InternalExecutionTime(), 0.0, "Expecting internal execution time not to be 0.0")
	assert.Equal(t, true, res.Empty(), "Expecting empty resultset")
	res, err = graphInstance.Query("MATCH p = (:Person)-[:Visited]->(:Country) RETURN p", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, len(res.Results()), 1, "expecting 1 result record")
	assert.Equal(t, false, res.Empty(), "Expecting resultset to have records")
	res, err = graphInstance.Query("MATCH ()-[r]-() DELETE r", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, res.RelationshipsDeleted(), "Expecting 1 relationships deleted")
}
