package integration_test

import (
	"testing"

	"github.com/snowmerak/falkordb-go/domain"
	"github.com/stretchr/testify/assert"
)

func TestArray(t *testing.T) {
	graphInstance.Query("MATCH (n) DELETE n", nil, nil)

	q := "CREATE (:person{name:'a',age:32,array:[0,1,2]})"
	res, err := graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	q = "CREATE (:person{name:'b',age:30,array:[3,4,5]})"
	res, err = graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	q = "WITH [0,1,2] as x return x"
	res, err = graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	res.Next()
	r := res.Record()
	assert.Equal(t, len(res.Results()), 1, "expecting 1 result record")
	assert.Equal(t, []interface{}{int64(0), int64(1), int64(2)}, r.GetByIndex(0))

	q = "unwind([0,1,2]) as x return x"
	res, err = graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(res.Results()), 3, "expecting 3 result record")

	i := 0
	for res.Next() {
		r = res.Record()
		assert.Equal(t, int64(i), r.GetByIndex(0))
		i++
	}

	q = "MATCH(n) return collect(n) as x"
	res, err = graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	a := domain.NodeNew([]string{"person"}, "", nil)
	b := domain.NodeNew([]string{"person"}, "", nil)

	a.SetProperty("name", "a")
	a.SetProperty("age", int64(32))
	a.SetProperty("array", []interface{}{int64(0), int64(1), int64(2)})

	b.SetProperty("name", "b")
	b.SetProperty("age", int64(30))
	b.SetProperty("array", []interface{}{int64(3), int64(4), int64(5)})

	assert.Equal(t, 1, len(res.Results()), "expecting 1 results record")

	res.Next()
	r = res.Record()
	arr := r.GetByIndex(0).([]interface{})

	assert.Equal(t, 2, len(arr))

	resA := arr[0].(*domain.Node)
	resB := arr[1].(*domain.Node)
	// the order of values in the array returned by collect operation is not defined
	// check for the node that contains the name "a" and set it to be resA
	if resA.GetProperty("name") != "a" {
		resA = arr[1].(*domain.Node)
		resB = arr[0].(*domain.Node)
	}

	assert.Equal(t, a.GetProperty("name"), resA.GetProperty("name"), "Unexpected property value.")
	assert.Equal(t, a.GetProperty("age"), resA.GetProperty("age"), "Unexpected property value.")
	assert.Equal(t, a.GetProperty("array"), resA.GetProperty("array"), "Unexpected property value.")

	assert.Equal(t, b.GetProperty("name"), resB.GetProperty("name"), "Unexpected property value.")
	assert.Equal(t, b.GetProperty("age"), resB.GetProperty("age"), "Unexpected property value.")
	assert.Equal(t, b.GetProperty("array"), resB.GetProperty("array"), "Unexpected property value.")
}

func TestMap(t *testing.T) {
	createGraph()

	q := "RETURN {val_1: 5, val_2: 'str', inner: {x: [1]}}"
	res, err := graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}
	res.Next()
	r := res.Record()
	mapval := r.GetByIndex(0).(map[string]interface{})

	inner_map := map[string]interface{}{"x": []interface{}{int64(1)}}
	expected := map[string]interface{}{"val_1": int64(5), "val_2": "str", "inner": inner_map}
	assert.Equal(t, mapval, expected, "expecting a map literal")

	q = "MATCH (a:Country) RETURN a { .name }"
	res, err = graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}
	res.Next()
	r = res.Record()
	mapval = r.GetByIndex(0).(map[string]interface{})

	expected = map[string]interface{}{"name": "Japan"}
	assert.Equal(t, mapval, expected, "expecting a map projection")
}

func TestPath(t *testing.T) {
	createGraph()
	q := "MATCH p = (:Person)-[:Visited]->(:Country) RETURN p"
	res, err := graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(res.Results()), 1, "expecting 1 result record")

	res.Next()
	r := res.Record()

	p, ok := r.GetByIndex(0).(domain.Path)
	assert.True(t, ok, "First column should contain path.")

	assert.Equal(t, 2, p.NodesCount(), "Path should contain two nodes")
	assert.Equal(t, 1, p.EdgeCount(), "Path should contain one edge")

	s := p.FirstNode()
	e := p.GetEdge(0)
	d := p.LastNode()

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

func TestPoint(t *testing.T) {
	q := "RETURN point({latitude: 37.0, longitude: -122.0})"
	res, err := graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}
	res.Next()
	r := res.Record()
	point := r.GetByIndex(0).(map[string]interface{})
	assert.Equal(t, point["latitude"], 37.0, "Unexpected latitude value")
	assert.Equal(t, point["longitude"], -122.0, "Unexpected longitude value")
}

func TestVectorF32(t *testing.T) {
	q := "RETURN vecf32([1.0, 2.0, 3.0])"
	res, err := graphInstance.Query(q, nil, nil)
	if err != nil {
		t.Error(err)
	}
	res.Next()
	r := res.Record()
	vec := r.GetByIndex(0).([]float32)
	assert.Equal(t, vec, []float32{1.0, 2.0, 3.0}, "Unexpected vector value")
}
