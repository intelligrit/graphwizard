// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package matching

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestHopcroftKarp_PerfectMatching(t *testing.T) {
	// Bipartite: left={0,1,2}, right={3,4,5}
	// 0-3, 0-4, 1-3, 1-4, 2-5
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(5)))

	m, size := HopcroftKarp(g, []int64{0, 1, 2})
	if size != 3 {
		t.Fatalf("expected matching size 3, got %d", size)
	}
	if len(m) != 3 {
		t.Errorf("expected 3 matched pairs, got %d", len(m))
	}
	// Verify each left node matched to a right node.
	for l, r := range m {
		if r < 3 || r > 5 {
			t.Errorf("left %d matched to invalid right %d", l, r)
		}
	}
}

func TestHopcroftKarp_PartialMatching(t *testing.T) {
	// Left={0,1,2}, right={3,4}. Only 2 can be matched.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(4)))

	_, size := HopcroftKarp(g, []int64{0, 1, 2})
	if size != 2 {
		t.Fatalf("expected matching size 2, got %d", size)
	}
}

func TestHopcroftKarp_NoEdges(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	_, size := HopcroftKarp(g, []int64{0})
	if size != 0 {
		t.Fatalf("expected matching size 0, got %d", size)
	}
}

func TestHopcroftKarp_SingleEdge(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	m, size := HopcroftKarp(g, []int64{0})
	if size != 1 {
		t.Fatalf("expected matching size 1, got %d", size)
	}
	if m[0] != 1 {
		t.Errorf("expected 0->1, got 0->%d", m[0])
	}
}
