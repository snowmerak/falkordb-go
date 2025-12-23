package falkordb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryResultNew_EdgeCases(t *testing.T) {
	// Setup a graph with a pre-populated schema for testing
	g := &Graph{
		schema: GraphSchema{
			labels:        []string{"L0", "L1"},
			relationships: []string{"R0", "R1"},
			properties:    []string{"P0", "P1"},
		},
	}

	tests := []struct {
		name        string
		response    interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:        "Nil response",
			response:    nil,
			wantErr:     true,
			errContains: "unexpected response type",
		},
		{
			name:        "Wrong type response",
			response:    "error string",
			wantErr:     true,
			errContains: "unexpected response type",
		},
		{
			name:        "Empty array",
			response:    []interface{}{},
			wantErr:     true,
			errContains: "empty response payload",
		},
		{
			name:     "Short array (stats only empty)",
			response: []interface{}{[]interface{}{}}, // Empty stats array allowed
			wantErr:  false,
		},
		{
			name: "Malformed Header - Not Array",
			response: []interface{}{
				"not-array-header",
				[]interface{}{},
				[]interface{}{"Labels added: 0"},
			},
			wantErr:     true,
			errContains: "header payload is not array",
		},
		{
			name: "Malformed Header - Invalid Column",
			response: []interface{}{
				[]interface{}{"bad-col"},
				[]interface{}{},
				[]interface{}{"Labels added: 0"},
			},
			wantErr:     true,
			errContains: "invalid header column format",
		},
		{
			name: "Malformed Statistics - Not Array",
			response: []interface{}{
				[]interface{}{},
				[]interface{}{},
				"bad-stats",
			},
			wantErr:     true,
			errContains: "statistics payload is not array",
		},
		{
			name: "Malformed Statistics - Invalid Format",
			response: []interface{}{
				[]interface{}{},
				[]interface{}{},
				[]interface{}{"Invalid Stat String"},
			},
			wantErr:     true,
			errContains: "invalid statistic format",
		},
		{
			name: "Malformed Records - Not Array",
			response: []interface{}{
				[]interface{}{},
				"bad-records",
				[]interface{}{"Labels added: 0"},
			},
			wantErr:     true,
			errContains: "records payload is not array",
		},
		{
			name: "Record Column Count Mismatch",
			response: []interface{}{
				[]interface{}{
					[]interface{}{int64(COLUMN_SCALAR), "col1"},
				},
				[]interface{}{
					[]interface{}{}, // Empty record, expected 1 column
				},
				[]interface{}{"Labels added: 0"},
			},
			wantErr:     true,
			errContains: "column count mismatch",
		},
		{
			name: "Record Invalid Scalar",
			response: []interface{}{
				[]interface{}{
					[]interface{}{int64(COLUMN_SCALAR), "col1"},
				},
				[]interface{}{
					[]interface{}{"not-a-scalar-cell"}, // Should be []interface{}
				},
				[]interface{}{"Labels added: 0"},
			},
			wantErr:     true,
			errContains: "not scalar payload",
		},
		{
			name: "Record Unknown Column Type",
			response: []interface{}{
				[]interface{}{
					[]interface{}{int64(999), "col1"}, // Unknown type
				},
				[]interface{}{
					[]interface{}{[]interface{}{int64(VALUE_INTEGER), int64(1)}},
				},
				[]interface{}{"Labels added: 0"},
			},
			wantErr:     true,
			errContains: "unknown column type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := QueryResultNew(g, tt.response)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPathNew_Safety(t *testing.T) {
	// Test with invalid inputs to ensure no panics
	p := PathNew([]interface{}{"not a node"}, []interface{}{})
	assert.Empty(t, p.Nodes)
	assert.Empty(t, p.Edges)

	p = PathNew([]interface{}{}, []interface{}{"not an edge"})
	assert.Empty(t, p.Nodes)
	assert.Empty(t, p.Edges)

	// Valid empty path
	p = PathNew([]interface{}{}, []interface{}{})
	assert.NotNil(t, p.Nodes)
	assert.NotNil(t, p.Edges)
	assert.Equal(t, 0, len(p.Nodes))
	assert.Equal(t, 0, len(p.Edges))
}

func TestParseScalar_EdgeCases(t *testing.T) {
	qr := &QueryResult{}

	// Helper to wrap scalar cell
	makeCell := func(typ ResultSetScalarTypes, val interface{}) []interface{} {
		return []interface{}{int64(typ), val}
	}

	tests := []struct {
		name        string
		cell        []interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:        "Short Cell",
			cell:        []interface{}{int64(VALUE_INTEGER)},
			wantErr:     true,
			errContains: "scalar cell too short",
		},
		{
			name:        "Bad Type ID",
			cell:        []interface{}{"not-int", "val"},
			wantErr:     true,
			errContains: "scalar type not int64",
		},
		{
			name:        "String - Wrong Type",
			cell:        makeCell(VALUE_STRING, 123),
			wantErr:     true,
			errContains: "string scalar not string",
		},
		{
			name:        "Integer - Wrong Type",
			cell:        makeCell(VALUE_INTEGER, "not-int"),
			wantErr:     true,
			errContains: "integer scalar not int64",
		},
		{
			name:        "Boolean - Wrong Type",
			cell:        makeCell(VALUE_BOOLEAN, 123),
			wantErr:     true,
			errContains: "boolean scalar not string",
		},
		{
			name:        "Double - Wrong Type",
			cell:        makeCell(VALUE_DOUBLE, 123),
			wantErr:     true,
			errContains: "double scalar not string",
		},
		{
			name:        "Unknown Scalar Type",
			cell:        makeCell(VALUE_UNKNOWN, "val"),
			wantErr:     true,
			errContains: "unknown scalar type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := qr.parseScalar(tt.cell)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
