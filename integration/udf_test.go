package integration_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/snowmerak/falkordb-go"
	"github.com/stretchr/testify/assert"
)

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func TestLoadUDFAndBitwise(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	createGraph()

	// 1. Download flex.js to a temporary file
	url := "https://github.com/FalkorDB/flex/releases/latest/download/flex.js"

	tmpFile, err := os.CreateTemp("", "flex-*.js")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	fmt.Printf("Downloading flex.js from %s to %s...\n", url, tmpPath)
	if err := downloadFile(tmpPath, url); err != nil {
		t.Skipf("Skipping UDF test: failed to download flex.js: %v", err)
	}

	// 2. Load UDF into FalkorDB
	fmt.Println("Loading UDF into FalkorDB...")
	err = db.LoadUDFFromFile("flex", tmpPath)
	if err != nil && !falkordb.IsUdfAlreadyRegisteredError(err) {
		assert.NoError(t, err)
	} else if err != nil {
		fmt.Println("UDF 'flex' already registered, proceeding...")
	}

	// 3. Execute Query using UDF
	// 3 & 1 = 1
	query := "RETURN flex.bitwise.and(3, 1)"
	fmt.Printf("Executing query: %s\n", query)

	res, err := graphInstance.Query(query, nil, nil)
	assert.NoError(t, err)
	if err == nil {
		assert.False(t, res.Empty(), "Result should not be empty")
		if !res.Empty() {
			val := res.Results()[0].Values()[0]
			fmt.Printf("flex.bitwise.and(3, 1) result: %v (Type: %T)\n", val, val)

			// Depending on how custom types are returned, checking value.
			// Usually integer returns int64.
			assert.Equal(t, int64(1), val)
		}
	}
}
