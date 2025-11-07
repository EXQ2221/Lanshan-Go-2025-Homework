package main

import "testing"

func TestFreq(t *testing.T) {
	result := Freq([]int{1, 1, 2})
	want := map[int]int{1: 2, 2: 1}

	for k, v := range want {
		if result[k] != v {
			t.Errorf("Freq([]int{1, 1, 2})[%d] = %d; want %d", k, result[k], v)
		}
	}
}
