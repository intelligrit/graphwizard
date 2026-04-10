// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestTSP_NoImprovement2Opt(t *testing.T) {
	// 3-node complete graph with equal weights: 2-opt can't improve.
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 1))

	result := TSP(context.Background(), g)
	if math.Abs(result.Weight-3.0) > epsilon {
		t.Errorf("expected weight 3.0, got %f", result.Weight)
	}
}

func TestTSP_FiveNodeComplete(t *testing.T) {
	// Complete K5 with varied weights to exercise 2-opt swaps.
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	weights := map[[2]int64]float64{
		{0, 1}: 2, {0, 2}: 9, {0, 3}: 10, {0, 4}: 7,
		{1, 2}: 6, {1, 3}: 4, {1, 4}: 3,
		{2, 3}: 8, {2, 4}: 5,
		{3, 4}: 1,
	}
	for k, w := range weights {
		g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(k[0]), simple.Node(k[1]), w))
	}

	result := TSP(context.Background(), g)
	if len(result.Tour) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(result.Tour))
	}
	// Optimal tour for this graph: 0-1-3-4-2-0 = 2+4+1+5+9=21
	// or 0-1-4-3-2-0 = 2+3+1+8+9=23... nearest neighbor+2-opt should
	// find something reasonable (under 25).
	if result.Weight > 25 {
		t.Errorf("TSP heuristic produced poor tour: weight %f", result.Weight)
	}
}

func TestTSP_EmptyGraph(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	result := TSP(context.Background(), g)
	if len(result.Tour) != 0 {
		t.Errorf("expected empty tour, got %d nodes", len(result.Tour))
	}
}
