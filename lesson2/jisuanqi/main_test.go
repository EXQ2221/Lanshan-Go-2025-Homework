package main

import "testing"

func TestCclt(t *testing.T) {
	if got := cclt(2, 3, 1)(); got != 5 {
		t.Errorf("cclt(2,3,1)() = %d, want 5", got)
	}
	if got := cclt(3, 2, 2)(); got != 1 {
		t.Errorf("cclt(2,3,1)() = %d, want 1", got)
	}
	if got := cclt(2, 3, 3)(); got != 6 {
		t.Errorf("cclt(2,3,1)() = %d, want 6", got)
	}
	if got := cclt(4, 2, 4)(); got != 2 {
		t.Errorf("cclt(2,3,1)() = %d, want 2", got)
	}
}
