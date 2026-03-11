package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlowLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	// Reset the slow log first so we start clean
	err := graphInstance.SlowLogReset(ctx)
	assert.Nil(t, err)

	// Run some queries to populate the slow log
	_, err = graphInstance.Query(ctx, "UNWIND range(0, 1000) AS v RETURN v", nil, nil)
	assert.Nil(t, err)

	_, err = graphInstance.Query(ctx, "MATCH (n) RETURN n", nil, nil)
	assert.Nil(t, err)

	// Small wait to let the slow log populate
	time.Sleep(100 * time.Millisecond)

	// Get slow log
	entries, err := graphInstance.SlowLog(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, entries)

	fmt.Printf("Slow log entries: %d\n", len(entries))
	for _, e := range entries {
		fmt.Printf("  timestamp=%s command=%s query=%s duration=%s\n",
			e.Timestamp, e.Command, e.Query, e.Duration)
	}

	// Reset and verify empty
	err = graphInstance.SlowLogReset(ctx)
	assert.Nil(t, err)

	entries, err = graphInstance.SlowLog(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entries), "Slow log should be empty after reset")
}
