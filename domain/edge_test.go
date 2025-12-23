package domain

import "testing"

func TestEdgeIDs(t *testing.T) {
	src := &Node{ID: 1, Alias: "a"}
	dst := &Node{ID: 2, Alias: "b"}
	e := NewEdge("R", src, dst, nil)
	e.ID = 3
	e.SrcNodeID = 1
	e.DestNodeID = 2

	if got := e.GetSourceNodeID(); got != 1 {
		t.Fatalf("GetSourceNodeID = %d, want 1", got)
	}
	if got := e.GetDestNodeID(); got != 2 {
		t.Fatalf("GetDestNodeID = %d, want 2", got)
	}
}
