// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestKruskal_Triangle(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 3))

	result := Kruskal(g)
	if len(result.Edges) != 2 {
		t.Fatalf("expected 2 MST edges, got %d", len(result.Edges))
	}
	if math.Abs(result.Weight-3.0) > epsilon {
		t.Errorf("expected MST weight 3.0, got %f", result.Weight)
	}
}

func TestKruskal_Disconnected(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 2))

	result := Kruskal(g)
	// MSF: 2 edges (one per component).
	if len(result.Edges) != 2 {
		t.Fatalf("expected 2 MSF edges, got %d", len(result.Edges))
	}
	if math.Abs(result.Weight-3.0) > epsilon {
		t.Errorf("expected MSF weight 3.0, got %f", result.Weight)
	}
}

func TestKruskal_Empty(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	result := Kruskal(g)
	if len(result.Edges) != 0 || result.Weight != 0 {
		t.Errorf("expected empty MST for empty graph")
	}
}

func TestPrim_Triangle(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 3))

	result := Prim(g, 0)
	if len(result.Edges) != 2 {
		t.Fatalf("expected 2 MST edges, got %d", len(result.Edges))
	}
	if math.Abs(result.Weight-3.0) > epsilon {
		t.Errorf("expected MST weight 3.0, got %f", result.Weight)
	}
}

func TestPrim_Empty(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	result := Prim(g, 0)
	if len(result.Edges) != 0 {
		t.Errorf("expected empty MST for empty graph")
	}
}

func TestPrim_MissingSource(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))

	result := Prim(g, 99)
	if len(result.Edges) != 0 {
		t.Errorf("expected empty MST for missing source")
	}
}
