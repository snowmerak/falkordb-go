package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSlowLogResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		want        []SlowLogEntry
		wantErr     bool
		errContains string
	}{
		{
			name:  "Empty Array",
			input: []interface{}{},
			want:  []SlowLogEntry{},
		},
		{
			name: "Single Entry",
			input: []interface{}{
				[]interface{}{"1581932396", "GRAPH.QUERY", "MATCH (n) RETURN n", "0.831"},
			},
			want: []SlowLogEntry{
				{Timestamp: "1581932396", Command: "GRAPH.QUERY", Query: "MATCH (n) RETURN n", Duration: "0.831"},
			},
		},
		{
			name: "Multiple Entries",
			input: []interface{}{
				[]interface{}{"1581932396", "GRAPH.QUERY", "MATCH (a)-[:FRIEND]->(e) RETURN e", "0.831"},
				[]interface{}{"1581932397", "GRAPH.RO_QUERY", "MATCH (n) RETURN n", "0.288"},
			},
			want: []SlowLogEntry{
				{Timestamp: "1581932396", Command: "GRAPH.QUERY", Query: "MATCH (a)-[:FRIEND]->(e) RETURN e", Duration: "0.831"},
				{Timestamp: "1581932397", Command: "GRAPH.RO_QUERY", Query: "MATCH (n) RETURN n", Duration: "0.288"},
			},
		},
		{
			name:        "Not Array",
			input:       "not-array",
			wantErr:     true,
			errContains: "unexpected slowlog response type",
		},
		{
			name:        "Entry Not Array",
			input:       []interface{}{"not-an-entry"},
			wantErr:     true,
			errContains: "slowlog entry is not array",
		},
		{
			name:  "Partial Entry",
			input: []interface{}{[]interface{}{"1581932396", "GRAPH.QUERY"}},
			want: []SlowLogEntry{
				{Timestamp: "1581932396", Command: "GRAPH.QUERY"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSlowLogResponse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
