package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProfileResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		want        []string
		wantErr     bool
		errContains string
	}{
		{
			name:  "Success",
			input: []interface{}{"Line 1", "Line 2"},
			want:  []string{"Line 1", "Line 2"},
		},
		{
			name:        "Not Array",
			input:       "not-array",
			wantErr:     true,
			errContains: "unexpected profile response type",
		},
		{
			name:        "Array with Non-String",
			input:       []interface{}{"Line 1", 123},
			wantErr:     true,
			errContains: "profile entry 1 not string",
		},
		{
			name:  "Empty Array",
			input: []interface{}{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseProfileResponse(tt.input)
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
