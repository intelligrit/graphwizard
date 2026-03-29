// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestTSP_Triangle(t *testing.T) {
	// Equilateral triangle: all edges weight 1. Optimal tour = 3.
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 1))

	result := TSP(g)
	if len(result.Tour) != 3 {
		t.Fatalf("expected tour of 3 nodes, got %d", len(result.Tour))
	}
	if math.Abs(result.Weight-3.0) > epsilon {
		t.Errorf("expected tour weight 3.0, got %f", result.Weight)
	}
}

func TestTSP_Square(t *testing.T) {
	// Square: 0-1=1, 1-2=1, 2-3=1, 3-0=1, diagonals=1.5
	// Optimal tour: 0-1-2-3-0 = 4
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(3), simple.Node(0), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 1.5))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 1.5))

	result := TSP(g)
	if len(result.Tour) != 4 {
		t.Fatalf("expected tour of 4 nodes, got %d", len(result.Tour))
	}
	// Should find optimal tour of weight 4
	if math.Abs(result.Weight-4.0) > epsilon {
		t.Errorf("expected tour weight 4.0, got %f", result.Weight)
	}
}

func TestTSP_SingleNode(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.AddNode(simple.Node(0))

	result := TSP(g)
	if len(result.Tour) != 1 {
		t.Fatalf("expected tour of 1 node, got %d", len(result.Tour))
	}
	if result.Weight != 0 {
		t.Errorf("expected weight 0, got %f", result.Weight)
	}
}

func TestTSP_TwoNodes(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 5))

	result := TSP(g)
	if len(result.Tour) != 2 {
		t.Fatalf("expected tour of 2 nodes, got %d", len(result.Tour))
	}
	if math.Abs(result.Weight-10.0) > epsilon {
		t.Errorf("expected tour weight 10.0 (round trip), got %f", result.Weight)
	}
}
