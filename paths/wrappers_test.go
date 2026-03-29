// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph"
	gonumpath "gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

func TestShortestPath_Chain(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 4))

	nodes, weight := ShortestPath(g, 0, 2)
	if len(nodes) != 3 {
		t.Fatalf("expected path of 3 nodes, got %d", len(nodes))
	}
	if math.Abs(weight-7.0) > epsilon {
		t.Errorf("expected weight 7.0, got %f", weight)
	}
}

func TestShortestPath_NoPath(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	nodes, weight := ShortestPath(g, 0, 1)
	if len(nodes) != 0 {
		t.Fatalf("expected empty path, got %d nodes", len(nodes))
	}
	if !math.IsInf(weight, 1) {
		t.Errorf("expected +Inf, got %f", weight)
	}
}

func TestAllShortestPaths_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	allPaths := AllShortestPaths(g)
	w := allPaths.Weight(0, 2)
	if math.IsInf(w, 1) {
		t.Fatal("expected finite weight between 0 and 2")
	}
}

func TestBellmanFord_Chain(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 3))

	nodes, weight, ok := BellmanFord(g, 0, 2)
	if !ok {
		t.Fatal("expected ok=true (no negative cycle)")
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	if math.Abs(weight-5.0) > epsilon {
		t.Errorf("expected weight 5.0, got %f", weight)
	}
}

func TestBellmanFord_NegativeCycle(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), -3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 1))

	_, _, ok := BellmanFord(g, 0, 2)
	if ok {
		t.Error("expected ok=false for negative cycle")
	}
}

func TestFloydWarshall_Triangle(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 5))

	allPaths, ok := FloydWarshall(g)
	if !ok {
		t.Fatal("expected ok=true")
	}
	w := allPaths.Weight(0, 2)
	if math.Abs(w-3.0) > epsilon {
		t.Errorf("expected 0->2 weight 3.0, got %f", w)
	}
}

func TestAStar_Chain(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))

	// Trivial heuristic (always 0) — degenerates to Dijkstra.
	h := func(a, b graph.Node) float64 { return 0 }
	nodes, weight := AStar(g, g.Node(0), g.Node(2), gonumpath.Heuristic(h))
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	if math.Abs(weight-2.0) > epsilon {
		t.Errorf("expected weight 2.0, got %f", weight)
	}
}
