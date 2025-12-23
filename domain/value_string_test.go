package domain

import "testing"

func TestToString(t *testing.T) {
	cases := []struct {
		in  interface{}
		out string
	}{
		{"s", "\"s\""},
		{int(1), "1"},
		{int64(2), "2"},
		{float64(3.5), "3.5"},
		{true, "true"},
		{nil, "null"},
	}

	for _, tt := range cases {
		if got := toString(tt.in); got != tt.out {
			t.Fatalf("toString(%v) = %q, want %q", tt.in, got, tt.out)
		}
	}
}
