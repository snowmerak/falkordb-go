package integration_test

import (
	"testing"

	"github.com/snowmerak/falkordb-go/domain"
	"github.com/snowmerak/falkordb-go/graph"
	"github.com/stretchr/testify/assert"
)

func TestMatchQuery(t *testing.T) {
	q := "MATCH (s)-[e]->(d) RETURN s,e,d"
	res, err := graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	checkQueryResults(t, res)
}

func TestMatchROQuery(t *testing.T) {
	q := "MATCH (s)-[e]->(d) RETURN s,e,d"
	res, err := graphInstance.ROQuery(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	checkQueryResults(t, res)
}

func checkQueryResults(t *testing.T, res *graph.QueryResult) {
	assert.Equal(t, len(res.Results()), 1, "expecting 1 result record")

	res.Next()
	r := res.Record()

	s, ok := r.GetByIndex(0).(*domain.Node)
	assert.True(t, ok, "First column should contain nodes.")
	e, ok := r.GetByIndex(1).(*domain.Edge)
	assert.True(t, ok, "Second column should contain edges.")
	d, ok := r.GetByIndex(2).(*domain.Node)
	assert.True(t, ok, "Third column should contain nodes.")

	assert.Equal(t, s.Labels[0], "Person", "Node should be of type 'Person'")
	assert.Equal(t, e.Relation, "Visited", "Edge should be of relation type 'Visited'")
	assert.Equal(t, d.Labels[0], "Country", "Node should be of type 'Country'")

	assert.Equal(t, len(s.Properties), 4, "Person node should have 4 properties")

	assert.Equal(t, s.GetProperty("name"), "John Doe", "Unexpected property value.")
	assert.Equal(t, s.GetProperty("age"), int64(33), "Unexpected property value.")
	assert.Equal(t, s.GetProperty("gender"), "male", "Unexpected property value.")
	assert.Equal(t, s.GetProperty("status"), "single", "Unexpected property value.")

	assert.Equal(t, e.GetProperty("year"), int64(2017), "Unexpected property value.")

	assert.Equal(t, d.GetProperty("name"), "Japan", "Unexpected property value.")
	assert.Equal(t, d.GetProperty("population"), int64(126800000), "Unexpected property value.")
}
