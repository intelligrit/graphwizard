// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestMaximalCliques_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	cliques := MaximalCliques(context.Background(), g)
	if len(cliques) != 1 {
		t.Fatalf("expected 1 maximal clique in triangle, got %d", len(cliques))
	}
	if len(cliques[0]) != 3 {
		t.Errorf("expected clique of size 3, got %d", len(cliques[0]))
	}
}

func TestMaximalCliques_Path(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	cliques := MaximalCliques(context.Background(), g)
	// Path has 2 maximal cliques: {0,1} and {1,2}.
	if len(cliques) != 2 {
		t.Fatalf("expected 2 maximal cliques in path, got %d", len(cliques))
	}
}

func TestGraphColoring_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	colors, k, err := GraphColoring(context.Background(), g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != 3 {
		t.Errorf("expected chromatic number 3, got %d", k)
	}
	// Verify no two adjacent nodes share a color.
	if colors[0] == colors[1] || colors[1] == colors[2] || colors[0] == colors[2] {
		t.Errorf("adjacent nodes share colors: %v", colors)
	}
}

func TestGraphColoring_Bipartite(t *testing.T) {
	// K2,2: bipartite, chromatic number = 2.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))

	_, k, err := GraphColoring(context.Background(), g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != 2 {
		t.Errorf("expected chromatic number 2, got %d", k)
	}
}

func TestAverageClusteringCoefficient_EmptyGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	avg := AverageClusteringCoefficient(context.Background(), g)
	if avg != 0 {
		t.Errorf("expected 0 for empty graph, got %f", avg)
	}
}
