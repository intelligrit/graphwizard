// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestSimRank_Symmetric(t *testing.T) {
	// 0->2, 1->2: nodes 0 and 1 point to the same target.
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	sim := SimRank(context.Background(), g, 0.8, 10)

	// Self-similarity = 1.
	if sim[[2]int64{0, 0}] != 1.0 {
		t.Errorf("sim(0,0) should be 1.0, got %f", sim[[2]int64{0, 0}])
	}

	// Nodes 0 and 1 are structurally identical (both point to 2 only).
	// They have no in-neighbors, so sim(0,1) should be 0 (no one points to them).
	// But they DO have in-neighbors of 2: {0, 1}. SimRank is based on in-neighbors.
	// Actually: in-neighbors of 0 = {} (nothing points to 0), so sim(0,1) = 0.
	if sim[[2]int64{0, 1}] != 0 {
		t.Errorf("sim(0,1) should be 0 (no in-neighbors), got %f", sim[[2]int64{0, 1}])
	}
}

func TestSimRank_WithInLinks(t *testing.T) {
	// 2->0, 2->1, 3->0, 3->1: nodes 0 and 1 have identical in-neighbors.
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(1)))

	sim := SimRank(context.Background(), g, 0.8, 10)
	// Nodes 0 and 1 share in-neighbors {2, 3}, so sim(0,1) > 0.
	key := [2]int64{0, 1}
	if sim[key] <= 0 {
		t.Errorf("sim(0,1) should be > 0 with shared in-neighbors, got %f", sim[key])
	}
}

func TestSimRank_Empty(t *testing.T) {
	g := simple.NewDirectedGraph()
	sim := SimRank(context.Background(), g, 0.8, 10)
	if len(sim) != 0 {
		t.Errorf("expected empty, got %d entries", len(sim))
	}
}
