// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package flow

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestMaxFlow_Simple(t *testing.T) {
	// 0 --(cap 3)--> 1 --(cap 2)--> 2
	// Max flow 0->2 = 2 (bottleneck at edge 1->2).
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))

	f := MaxFlow(g, 0, 2, 1e-9)
	if math.Abs(f-2.0) > 1e-9 {
		t.Errorf("expected max flow 2.0, got %f", f)
	}
}

func TestMaxFlow_Parallel(t *testing.T) {
	// 0 --(3)--> 2, 0 --(2)--> 1 --(2)--> 2
	// Max flow 0->2 = 3 + 2 = 5.
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))

	f := MaxFlow(g, 0, 2, 1e-9)
	if math.Abs(f-5.0) > 1e-9 {
		t.Errorf("expected max flow 5.0, got %f", f)
	}
}

func TestMaxFlow_NoPath(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 5))
	g.AddNode(simple.Node(2))

	f := MaxFlow(g, 0, 2, 1e-9)
	if f != 0 {
		t.Errorf("expected max flow 0, got %f", f)
	}
}
