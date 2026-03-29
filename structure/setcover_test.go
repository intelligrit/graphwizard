// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"testing"
)

func TestSetCover_Simple(t *testing.T) {
	universe := []int64{1, 2, 3, 4, 5}
	sets := [][]int64{
		{1, 2, 3},    // 0
		{2, 4},        // 1
		{3, 4},        // 2
		{4, 5},        // 3
	}

	result := SetCover(universe, sets)
	// Greedy should pick set 0 (covers 3), then set 3 (covers 2 more) = full cover
	covered := make(map[int64]bool)
	for _, idx := range result {
		for _, e := range sets[idx] {
			covered[e] = true
		}
	}
	for _, e := range universe {
		if !covered[e] {
			t.Errorf("element %d not covered", e)
		}
	}
}

func TestSetCover_AlreadyCovered(t *testing.T) {
	universe := []int64{1, 2}
	sets := [][]int64{
		{1, 2},
	}

	result := SetCover(universe, sets)
	if len(result) != 1 || result[0] != 0 {
		t.Errorf("expected [0], got %v", result)
	}
}

func TestSetCover_Empty(t *testing.T) {
	result := SetCover(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestSetCover_Impossible(t *testing.T) {
	universe := []int64{1, 2, 3}
	sets := [][]int64{
		{1, 2},
	}

	result := SetCover(universe, sets)
	// Should return what it can cover.
	if len(result) != 1 {
		t.Errorf("expected 1 set used, got %d", len(result))
	}
}
