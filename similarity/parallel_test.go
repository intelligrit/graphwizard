// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"testing"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func TestJaccardAllParallel_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	results := JaccardAllParallel(g, 0.1)
	if len(results) != 3 {
		t.Errorf("expected 3 pairs, got %d", len(results))
	}
}

func TestJaccardAllParallel_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	results := JaccardAllParallel(g, 0.1)
	if len(results) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(results))
	}
}

func TestPredictLinksParallel_Square(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))

	scorer := func(g graph.Undirected, u, v int64) float64 {
		return float64(CommonNeighbors(g, u, v))
	}
	preds := PredictLinksParallel(g, 5, scorer)
	if len(preds) != 2 {
		t.Errorf("expected 2 predictions, got %d", len(preds))
	}
}

func TestPredictLinksParallel_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	scorer := func(g graph.Undirected, u, v int64) float64 { return 0 }
	preds := PredictLinksParallel(g, 5, scorer)
	if len(preds) != 0 {
		t.Errorf("expected 0, got %d", len(preds))
	}
}
