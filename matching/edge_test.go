// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package matching

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestHopcroftKarp_LargerBipartite(t *testing.T) {
	// K3,3: perfect matching of size 3.
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 3; i++ {
		for j := int64(3); j < 6; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	m, size := HopcroftKarp(g, []int64{0, 1, 2})
	if size != 3 {
		t.Fatalf("expected matching size 3, got %d", size)
	}
	// Verify each left is matched to a distinct right.
	rights := make(map[int64]bool)
	for _, r := range m {
		if rights[r] {
			t.Errorf("duplicate right match: %d", r)
		}
		rights[r] = true
	}
}

func TestHopcroftKarp_EmptyLeft(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	_, size := HopcroftKarp(g, nil)
	if size != 0 {
		t.Fatalf("expected matching size 0, got %d", size)
	}
}
