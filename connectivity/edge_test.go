// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestBiconnectedComponents_Path(t *testing.T) {
	// Path 0-1-2-3: each edge is its own biconnected component (all bridges).
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	comps := BiconnectedComponents(context.Background(), g)
	if len(comps) != 3 {
		t.Fatalf("expected 3 biconnected components in path, got %d", len(comps))
	}
}

func TestArticulationPoints_Path(t *testing.T) {
	// Path 0-1-2: node 1 is an articulation point.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	aps := ArticulationPoints(context.Background(), g)
	if len(aps) != 1 || aps[0] != 1 {
		t.Errorf("expected [1], got %v", aps)
	}
}

func TestArticulationPoints_Star(t *testing.T) {
	// Star: center is the only AP.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))

	aps := ArticulationPoints(context.Background(), g)
	if len(aps) != 1 || aps[0] != 0 {
		t.Errorf("expected [0], got %v", aps)
	}
}

func TestUnionFind_RankTie(t *testing.T) {
	// Force a rank tie to cover the rank increment branch.
	uf := NewUnionFind()
	uf.Find(1)
	uf.Find(2)
	uf.Union(1, 2) // rank tie: winner gets rank 1
	uf.Find(3)
	uf.Find(4)
	uf.Union(3, 4)    // rank tie again: winner gets rank 1
	uf.Union(1, 3)    // rank tie at rank 1
	if !uf.Connected(1, 4) {
		t.Error("1 and 4 should be connected")
	}
}

func TestUnionFind_RankDifference(t *testing.T) {
	// Force rank[rx] < rank[ry] to cover the swap branch.
	uf := NewUnionFind()
	uf.Find(1)
	uf.Find(2)
	uf.Union(1, 2) // rank of root becomes 1
	uf.Find(3)     // singleton, rank 0
	// Union singleton (rank 0) with group (rank 1): triggers rx/ry swap.
	uf.Union(3, 1)
	if !uf.Connected(2, 3) {
		t.Error("2 and 3 should be connected")
	}
}

func TestBiconnectedComponents_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	comps := BiconnectedComponents(context.Background(), g)
	if len(comps) != 0 {
		t.Errorf("expected 0 components, got %d", len(comps))
	}
}

func TestArticulationPoints_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	aps := ArticulationPoints(context.Background(), g)
	if len(aps) != 0 {
		t.Errorf("expected 0 APs, got %d", len(aps))
	}
}

func TestBiconnectedComponents_Star(t *testing.T) {
	// Star: root has 3 children. Root is AP, should pop multiple components.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))

	comps := BiconnectedComponents(context.Background(), g)
	if len(comps) != 3 {
		t.Fatalf("expected 3 biconnected components in star, got %d", len(comps))
	}
}

func TestBiconnectedComponents_SingleEdge(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	comps := BiconnectedComponents(context.Background(), g)
	if len(comps) != 1 {
		t.Fatalf("expected 1 component, got %d", len(comps))
	}
	if len(comps[0]) != 1 {
		t.Errorf("expected 1 edge in component, got %d", len(comps[0]))
	}
}
