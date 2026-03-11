package integration_test

import (
	"context"
	"os"
	"testing"

	falkordb "github.com/snowmerak/falkordb-go"
	"github.com/snowmerak/falkordb-go/graph"
)

var graphInstance *graph.Graph
var db *falkordb.FalkorDB
var ctx = context.Background()

func createGraph() {
	addr := os.Getenv("FALKORDB_ADDR")
	if addr == "" {
		addr = "0.0.0.0:6379"
	}

	var err error

	if os.Getenv("FALKORDB_TEST_MODE") == "cluster" {
		db, err = falkordb.NewCluster(&falkordb.ConnectionClusterOption{
			Addrs: []string{addr},
		})
	} else {
		db, err = falkordb.FromURL("falkor://" + addr)
	}

	if err != nil {
		panic(err)
	}

	graphInstance = db.SelectGraph("social")
	graphInstance.DeleteContext(ctx)

	_, err = graphInstance.QueryContext(ctx, "CREATE (:Person {name: 'John Doe', age: 33, gender: 'male', status: 'single'})-[:Visited {year: 2017}]->(c:Country {name: 'Japan', population: 126800000})", nil, nil)
	if err != nil {
		panic(err)
	}
}

func setup() {
	createGraph()
}

func shutdown() {
	graphInstance.Conn.Close()
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}
