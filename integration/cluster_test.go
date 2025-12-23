package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterRoutingAndComplexQueries(t *testing.T) {
	if db == nil {
		t.Skip("Database connection not initialized")
	}

	// Define multiple graphs that likely map to different slots
	graphs := []string{"marketing", "engineering", "sales", "hr"}

	for _, gName := range graphs {
		g := db.SelectGraph(gName)
		g.Delete() // Ensure clean state

		// Create a chain of nodes
		query := `
            CREATE (n1:Node {id: 1})-[:NEXT]->(n2:Node {id: 2}),
                   (n2)-[:NEXT]->(n3:Node {id: 3}),
                   (n3)-[:NEXT]->(n4:Node {id: 4}),
                   (n4)-[:NEXT]->(n5:Node {id: 5})
        `
		_, err := g.Query(query, nil, nil)
		assert.Nil(t, err, "Failed to create data in graph %s", gName)
	}

	// Query each graph to verify routing and data integrity
	for _, gName := range graphs {
		g := db.SelectGraph(gName)

		// Variable length path query
		query := "MATCH (n1:Node {id: 1})-[:NEXT*]->(target) RETURN count(target) as count"
		res, err := g.Query(query, nil, nil)
		assert.Nil(t, err, "Failed to query graph %s", gName)

		assert.True(t, res.Next())
		record := res.Record()

		val := record.GetByIndex(0)
		assert.EqualValues(t, 4, val, "Incorrect path count in graph %s", gName)

		g.Delete()
	}
}
