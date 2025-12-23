package domain

import "testing"

func TestNodeNewDefaults(t *testing.T) {
	n := NodeNew([]string{"L"}, "a", nil)
	if n == nil {
		t.Fatalf("NodeNew returned nil")
	}
	if n.Properties == nil {
		t.Fatalf("NodeNew should initialize Properties")
	}
}
