package integration_test

import (
	"testing"

	"github.com/snowmerak/falkordb-go/domain"
	"github.com/stretchr/testify/assert"
)

func TestCreateQuery(t *testing.T) {
	q := "CREATE (w:WorkPlace {name:'FalkorDB'})"
	res, err := graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, res.Empty(), "Expecting empty result-set")

	// Validate statistics.
	assert.Equal(t, res.NodesCreated(), 1, "Expecting a single node to be created.")
	assert.Equal(t, res.PropertiesSet(), 1, "Expecting a songle property to be added.")

	q = "MATCH (w:WorkPlace) RETURN w"
	res, err = graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	assert.False(t, res.Empty(), "Expecting resultset to include a single node.")
	res.Next()
	r := res.Record()
	w := r.GetByIndex(0).(*domain.Node)
	assert.Equal(t, w.Labels[0], "WorkPlace", "Unexpected node label.")
}

func TestCreateROQueryFailure(t *testing.T) {
	q := "CREATE (w:WorkPlace {name:'FalkorDB'})"
	_, err := graphInstance.ROQuery(q, nil, nil)
	assert.NotNil(t, err, "error should not be nil")
}

func TestErrorReporting(t *testing.T) {
	q := "RETURN toupper(5)"
	res, err := graphInstance.Query(q, nil, nil)
	assert.Nil(t, res)
	assert.NotNil(t, err)

	q = "MATCH (p:Person) RETURN toupper(p.age)"
	res, err = graphInstance.Query(q, nil, nil)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
