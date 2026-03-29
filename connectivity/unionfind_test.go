// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"testing"
)

func TestUnionFind_Basic(t *testing.T) {
	uf := NewUnionFind()

	// Initially, each element is its own set.
	if uf.Connected(1, 2) {
		t.Error("1 and 2 should not be connected initially")
	}

	// Union 1 and 2.
	if !uf.Union(1, 2) {
		t.Error("Union(1,2) should return true (different sets)")
	}
	if !uf.Connected(1, 2) {
		t.Error("1 and 2 should be connected after union")
	}

	// Union same set.
	if uf.Union(1, 2) {
		t.Error("Union(1,2) should return false (same set)")
	}
}

func TestUnionFind_Transitivity(t *testing.T) {
	uf := NewUnionFind()
	uf.Union(1, 2)
	uf.Union(2, 3)
	uf.Union(4, 5)

	if !uf.Connected(1, 3) {
		t.Error("1 and 3 should be connected transitively")
	}
	if uf.Connected(1, 4) {
		t.Error("1 and 4 should not be connected")
	}

	// Merge the two groups.
	uf.Union(3, 4)
	if !uf.Connected(1, 5) {
		t.Error("1 and 5 should be connected after merging groups")
	}
}

func TestUnionFind_Sets(t *testing.T) {
	uf := NewUnionFind()
	uf.Union(1, 2)
	uf.Union(2, 3)
	uf.Union(4, 5)
	uf.Find(6) // singleton

	sets := uf.Sets()
	if len(sets) != 3 {
		t.Errorf("expected 3 sets, got %d", len(sets))
	}

	// Find the set containing 1.
	root := uf.Find(1)
	members := sets[root]
	if len(members) != 3 {
		t.Errorf("set containing 1 should have 3 members, got %d", len(members))
	}
}
