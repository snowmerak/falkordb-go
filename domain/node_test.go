package domain

import "testing"

func TestNewNodeDefaults(t *testing.T) {
	n := NewNode([]string{"L"}, "a", nil)
	if n == nil {
		t.Fatalf("NewNode returned nil")
	}
	if n.Properties == nil {
		t.Fatalf("NewNode should initialize Properties")
	}
}
