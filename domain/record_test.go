package domain

import "testing"

func TestRecordLookup(t *testing.T) {
	r := NewRecord([]interface{}{int64(10), "v"}, []string{"a", "b"})

	if val, ok := r.Get("a"); !ok || val != int64(10) {
		t.Fatalf("Get a = %v ok=%v, want 10 true", val, ok)
	}
	if val := r.GetByIndex(1); val != "v" {
		t.Fatalf("GetByIndex(1) = %v, want v", val)
	}
	if val, ok := r.Get("missing"); ok || val != nil {
		t.Fatalf("Get missing should be nil,false")
	}

	r.indices = nil
	if val, ok := r.Get("b"); !ok || val != "v" {
		t.Fatalf("lazy index rebuild failed: %v %v", val, ok)
	}
}
