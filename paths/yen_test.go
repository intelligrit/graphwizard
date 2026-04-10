// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths

import (
	"context"
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

const epsilon = 1e-9

func TestYenKShortest_Diamond(t *testing.T) {
	// Diamond graph: 0->1 (w=1), 0->2 (w=2), 1->3 (w=3), 2->3 (w=1)
	// Shortest 0->3: 0->2->3 (w=3), then 0->1->3 (w=4)
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 1))

	paths := YenKShortest(context.Background(), g, 0, 3, 3)
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}

	if math.Abs(paths[0].Weight-3.0) > epsilon {
		t.Errorf("first path weight: expected 3.0, got %f", paths[0].Weight)
	}
	if math.Abs(paths[1].Weight-4.0) > epsilon {
		t.Errorf("second path weight: expected 4.0, got %f", paths[1].Weight)
	}
}

func TestYenKShortest_Chain(t *testing.T) {
	// Simple chain: 0->1->2, only 1 path
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))

	paths := YenKShortest(context.Background(), g, 0, 2, 5)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	if math.Abs(paths[0].Weight-2.0) > epsilon {
		t.Errorf("expected weight 2.0, got %f", paths[0].Weight)
	}
}

func TestYenKShortest_NoPath(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	paths := YenKShortest(context.Background(), g, 0, 1, 3)
	if len(paths) != 0 {
		t.Fatalf("expected 0 paths, got %d", len(paths))
	}
}

func TestYenKShortest_MultipleParallel(t *testing.T) {
	// 0->1 (w=1), 0->2 (w=1), 1->3 (w=1), 2->3 (w=2), 0->3 (w=5)
	// Paths: 0->1->3 (w=2), 0->2->3 (w=3), 0->3 (w=5)
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(3), 5))

	paths := YenKShortest(context.Background(), g, 0, 3, 3)
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
	if math.Abs(paths[0].Weight-2.0) > epsilon {
		t.Errorf("path 0 weight: expected 2.0, got %f", paths[0].Weight)
	}
	if math.Abs(paths[1].Weight-3.0) > epsilon {
		t.Errorf("path 1 weight: expected 3.0, got %f", paths[1].Weight)
	}
	if math.Abs(paths[2].Weight-5.0) > epsilon {
		t.Errorf("path 2 weight: expected 5.0, got %f", paths[2].Weight)
	}
}
