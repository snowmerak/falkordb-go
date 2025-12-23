package falkordb

import (
	"github.com/redis/go-redis/v9"

	"github.com/snowmerak/falkordb-go/graph"
)

type (
	Graph        = graph.Graph
	QueryResult  = graph.QueryResult
	QueryOptions = graph.QueryOptions
	GraphSchema  = graph.GraphSchema
)

func NewQueryOptions() *QueryOptions { return graph.NewQueryOptions() }

func QueryResultNew(g *Graph, response interface{}) (*QueryResult, error) {
	return graph.QueryResultNew(g, response)
}

func BuildParamsHeader(params map[string]interface{}) string {
	return graph.BuildParamsHeader(params)
}

func graphNew(id string, conn redis.UniversalClient) *Graph {
	return graph.New(id, conn)
}

// Test helpers passthroughs.
func GraphSchemaWithData(labels, relationships, properties []string) GraphSchema {
	return graph.GraphSchemaWithData(labels, relationships, properties)
}

func NewGraphWithSchema(schema GraphSchema) *Graph {
	return graph.NewGraphWithSchema(schema)
}
