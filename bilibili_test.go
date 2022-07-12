package main

import (
	"testing"
)

func TestBv2av(t *testing.T) {
	tests := []struct {
		name string
		bv   string
		av   int64
	}{
		{"case1", "BV17x411w7KC", 170001},
		{"case2", "BV1RT411E7mF", 470835963},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := bv2av(test.bv); got != test.av {
				t.Errorf("bv=%s, want %v, got %v", test.bv, test.av, got)
			}
		})
	}
}
