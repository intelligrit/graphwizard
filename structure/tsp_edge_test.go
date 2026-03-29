// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestTSP_ForceTwoOptImprovement(t *testing.T) {
	// 4-node graph where nearest-neighbor from some starts gives a bad tour
	// that 2-opt must fix.
	// Nodes arranged: 0 and 3 are close, 1 and 2 are close, but NN might
	// visit 0->1->2->3 (crossing) instead of 0->1->3->2 (non-crossing).
	//
	// Distances:
	// 0-1: 1, 0-2: 10, 0-3: 2
	// 1-2: 1, 1-3: 10
	// 2-3: 1
	//
	// NN from 0: 0->1->2->3->0 = 1+1+1+2 = 5 (optimal)
	// NN from 2: 2->1->0->3->2 = 1+1+2+1 = 5
	// But let's create a case where NN gives a bad tour:
	//
	// 0-1: 5, 0-2: 1, 0-3: 1
	// 1-2: 5, 1-3: 5
	// 2-3: 1
	// Optimal: 0-2-3-1-0 = 1+1+5+5 = 12
	//          0-3-2-1-0 = 1+1+5+5 = 12
	// NN from 0: picks 2 (cost 1), then 3 (cost 1), then 1 (cost 5), back = 5. Total 12.
	// NN from 1: picks 0 (cost 5), then 2 (cost 1), then 3 (cost 1), back = 5. Total 12.
	// All tours have same cost, 2-opt won't improve. Need asymmetry.

	// Better setup: force NN to cross edges then 2-opt fixes it.
	// Square: 0=(0,0), 1=(1,0), 2=(1,1), 3=(0,1)
	// Edges: 0-1=1, 1-2=1, 2-3=1, 3-0=1, 0-2=1.41, 1-3=1.41
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(3), simple.Node(0), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 1.42))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 1.42))

	result := TSP(g)
	// Optimal tour = 4.0 (follow the square). 2-opt should find it.
	if math.Abs(result.Weight-4.0) > 0.01 {
		t.Errorf("expected tour weight ~4.0, got %f", result.Weight)
	}
}

func TestTSP_SixNodeForce2Opt(t *testing.T) {
	// 6-node complete graph with weights designed so NN produces a crossing
	// that 2-opt must reverse.
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	// Create a hexagon: sequential edges cost 1, skip edges cost 3
	for i := int64(0); i < 6; i++ {
		next := (i + 1) % 6
		g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(i), simple.Node(next), 1))
	}
	// Diagonals cost 3.
	for i := int64(0); i < 6; i++ {
		for j := i + 2; j < 6; j++ {
			if (j+1)%6 == i {
				continue // already set as sequential
			}
			if !g.HasEdgeBetween(i, j) {
				g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(i), simple.Node(j), 3))
			}
		}
	}

	result := TSP(g)
	// Optimal tour follows the hexagon: weight = 6.
	if result.Weight > 6.01 {
		t.Errorf("expected tour weight ~6.0, got %f", result.Weight)
	}
}
