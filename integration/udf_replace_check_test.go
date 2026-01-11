package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUDFReplace(t *testing.T) {
	ctx := context.Background()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	// 1. Download flex.js
	url := "https://github.com/FalkorDB/flex/releases/latest/download/flex.js"
	tmpFile, err := os.CreateTemp("", "flex-replace-*.js")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := downloadFile(tmpPath, url); err != nil {
		t.Skipf("Skipping UDF test: failed to download flex.js: %v", err)
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	code := string(content)

	// 2. Load UDF normally (might fail if already there, so ignore 'already registered')
	// We use raw command to assume clean state or ignore error
	db.Conn.Do(ctx, "GRAPH.UDF", "LOAD", "flex", code)

	// 3. Load again with REPLACE
	fmt.Println("Reloading UDF with REPLACE...")
	// The syntax is GRAPH.UDF LOAD [REPLACE] <library_name> <library_script>
	err = db.LoadUDFReplace("flex", code)
	assert.NoError(t, err, "GRAPH.UDF LOAD REPLACE should succeed even if library exists")
}
