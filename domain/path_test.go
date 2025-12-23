package domain

import "testing"

func TestNewPathValidation(t *testing.T) {
	p := NewPath([]interface{}{"x"}, []interface{}{})
	if len(p.Nodes) != 0 || len(p.Edges) != 0 {
		t.Fatalf("expected empty path on invalid input")
	}

	p = NewPath([]interface{}{&Node{ID: 1}}, []interface{}{&Edge{ID: 2}})
	if len(p.Nodes) != 1 || len(p.Edges) != 1 {
		t.Fatalf("expected path with 1 node and 1 edge")
	}
}
