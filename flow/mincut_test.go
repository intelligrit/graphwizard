// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package flow

import (
	"context"
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestMinCut_Simple(t *testing.T) {
	// 0 --(3)--> 1 --(2)--> 2. Min cut = 2 (edge 1->2).
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))

	result := MinCut(context.Background(), g, 0, 2, 1e-9)
	if math.Abs(result.Weight-2.0) > 1e-9 {
		t.Errorf("expected min cut weight 2.0, got %f", result.Weight)
	}
	if len(result.SourceSide) == 0 || len(result.TargetSide) == 0 {
		t.Error("expected non-empty partitions")
	}
}

func TestMinCut_Parallel(t *testing.T) {
	// Two parallel paths: 0->1->3 (cap 2) and 0->2->3 (cap 3). Min cut = 5.
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 3))

	result := MinCut(context.Background(), g, 0, 3, 1e-9)
	if math.Abs(result.Weight-5.0) > 1e-9 {
		t.Errorf("expected min cut weight 5.0, got %f", result.Weight)
	}
}

func TestMinCut_SourceInSourceSide(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))

	result := MinCut(context.Background(), g, 0, 1, 1e-9)
	found := false
	for _, id := range result.SourceSide {
		if id == 0 {
			found = true
		}
	}
	if !found {
		t.Error("source should be in SourceSide")
	}
}
