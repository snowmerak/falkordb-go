package integration_test

import (
	"fmt"
	"testing"

	falkordb "github.com/snowmerak/falkordb-go"
	"github.com/stretchr/testify/assert"
)

func TestInfoAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	info, err := db.InfoContext(ctx, falkordb.InfoAll)
	assert.Nil(t, err)
	assert.NotNil(t, info)

	fmt.Printf("Running queries: %d\n", len(info.RunningQueries))
	fmt.Printf("Waiting queries: %d\n", len(info.WaitingQueries))
}

func TestInfoRunningQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	info, err := db.InfoContext(ctx, falkordb.InfoRunningQueries)
	assert.Nil(t, err)
	assert.NotNil(t, info)

	// When filtering by RunningQueries, WaitingQueries should be nil
	assert.Nil(t, info.WaitingQueries)
}

func TestInfoWaitingQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	info, err := db.InfoContext(ctx, falkordb.InfoWaitingQueries)
	assert.Nil(t, err)
	assert.NotNil(t, info)

	// When filtering by WaitingQueries, RunningQueries should be nil
	assert.Nil(t, info.RunningQueries)
}
