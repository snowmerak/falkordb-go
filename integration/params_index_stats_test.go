package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParameterizedQuery(t *testing.T) {
	createGraph()
	params := []interface{}{int64(1), 2.3, "str", true, false, nil, []interface{}{int64(0), int64(1), int64(2)}, []interface{}{"0", "1", "2"}}
	q := "RETURN $param"
	params_map := make(map[string]interface{})
	for index, param := range params {
		params_map["param"] = param
		res, err := graphInstance.Query(q, params_map, nil)
		if err != nil {
			t.Error(err)
		}
		res.Next()
		assert.Equal(t, res.Record().GetByIndex(0), params[index], "Unexpected parameter value")
	}
}

func TestCreateIndex(t *testing.T) {
	res, err := graphInstance.Query("CREATE INDEX FOR (u:user) ON (u.name)", nil, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, res.IndicesCreated(), "Expecting 1 index created")

	_, err = graphInstance.Query("CREATE INDEX FOR (u:user) ON (u.name)", nil, nil)
	if err == nil {
		t.Error("expecting error")
	}

	res, err = graphInstance.Query("DROP INDEX FOR (u:user) ON (u.name)", nil, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, res.IndicesDeleted(), "Expecting 1 index deleted")

	_, err = graphInstance.Query("DROP INDEX FOR (u:user) ON (u.name)", nil, nil)
	assert.Equal(t, err.Error(), "ERR Unable to drop index on :user(name): no such index.")
}

func TestQueryStatistics(t *testing.T) {
	err := graphInstance.Delete()
	assert.Nil(t, err)

	q := "CREATE (:Person{name:'a',age:32,array:[0,1,2]})"
	res, err := graphInstance.Query(q, nil, nil)
	assert.Nil(t, err)

	assert.Equal(t, 1, res.NodesCreated(), "Expecting 1 node created")
	assert.Equal(t, 0, res.NodesDeleted(), "Expecting 0 nodes deleted")
	assert.Greater(t, res.InternalExecutionTime(), 0.0, "Expecting internal execution time not to be 0.0")
	assert.Equal(t, true, res.Empty(), "Expecting empty resultset")

	res, err = graphInstance.Query("MATCH (n) DELETE n", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, res.NodesDeleted(), "Expecting 1 nodes deleted")

	res, err = graphInstance.Query("CREATE (:Person {name: 'John Doe', age: 33, gender: 'male', status: 'single'})-[:Visited {year: 2017}]->(c:Country {name: 'Japan', population: 126800000})", nil, nil)

	assert.Nil(t, err)
	assert.Equal(t, 2, res.NodesCreated(), "Expecting 2 node created")
	assert.Equal(t, 0, res.NodesDeleted(), "Expecting 0 nodes deleted")
	assert.Equal(t, 7, res.PropertiesSet(), "Expecting 7 properties set")
	assert.Equal(t, 1, res.RelationshipsCreated(), "Expecting 1 relationships created")
	assert.Equal(t, 0, res.RelationshipsDeleted(), "Expecting 0 relationships deleted")
	assert.Greater(t, res.InternalExecutionTime(), 0.0, "Expecting internal execution time not to be 0.0")
	assert.Equal(t, true, res.Empty(), "Expecting empty resultset")
	q = "MATCH p = (:Person)-[:Visited]->(:Country) RETURN p"
	res, err = graphInstance.Query(q, nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, len(res.Results()), 1, "expecting 1 result record")
	assert.Equal(t, false, res.Empty(), "Expecting resultset to have records")
	res, err = graphInstance.Query("MATCH ()-[r]-() DELETE r", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, res.RelationshipsDeleted(), "Expecting 1 relationships deleted")
}
