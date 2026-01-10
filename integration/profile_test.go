package integration_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	queries := []string{
		"MATCH (n) RETURN n",
		"CREATE (:ProfileNode)",
	}

	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			profile, err := graphInstance.Profile(q, nil, nil)
			assert.Nil(t, err)
			assert.NotEmpty(t, profile, "Profile should not be empty")

			// Check content format briefly
			// Note: The root operation name differs depending on query.
			// e.g. "Results | ..." for read queries or "Create | ..." for write queries.
			assert.Contains(t, profile[0], "Records produced", "First line should contain execution metrics")

			for _, line := range profile {
				fmt.Println(line)
			}
		})
	}
}
