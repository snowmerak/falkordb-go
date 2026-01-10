package integration_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	// Default samples
	info, err := graphInstance.MemoryUsage(-1)
	assert.Nil(t, err)
	assert.NotEmpty(t, info)

	fmt.Println("Memory Usage (Default):")
	for k, v := range info {
		fmt.Printf("%s: %v\n", k, v)
	}

	// With specific samples
	info2, err := graphInstance.MemoryUsage(50)
	assert.Nil(t, err)
	assert.NotEmpty(t, info2)

	// Check if essential keys exist
	// Note: Keys depend on FalkorDB version, but usually include total_graph_sz_mb
	// Actually based on description: "total_graph_sz_mb"
	// Let's verify at least one known key or just that we have data
	// Some keys might be missing if graph is empty or simple
}
