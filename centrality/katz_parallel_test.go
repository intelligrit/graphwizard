// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestKatzParallel_MatchesSequential(t *testing.T) {
	g := simple.NewDirectedGraph()
	// Chain: 0->1->2->3
	for i := int64(0); i < 4; i++ {
		g.AddNode(simple.Node(i))
	}
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	seq := Katz(g, 0.1, 1.0, 1e-10, 200)
	par := KatzParallel(g, 0.1, 1.0, 1e-10, 200)

	for id, sv := range seq {
		pv := par[id]
		if math.Abs(sv-pv) > 1e-8 {
			t.Errorf("node %d: sequential=%f parallel=%f", id, sv, pv)
		}
	}
}

func TestKatzParallel_Empty(t *testing.T) {
	g := simple.NewDirectedGraph()
	scores := KatzParallel(g, 0.1, 1.0, 1e-8, 100)
	if len(scores) != 0 {
		t.Errorf("expected empty map, got %v", scores)
	}
}

func TestKatzUndirectedParallel_MatchesSequential(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	seq := KatzUndirected(g, 0.1, 1.0, 1e-10, 200)
	par := KatzUndirectedParallel(g, 0.1, 1.0, 1e-10, 200)

	for id, sv := range seq {
		pv := par[id]
		if math.Abs(sv-pv) > 1e-8 {
			t.Errorf("node %d: sequential=%f parallel=%f", id, sv, pv)
		}
	}
}
