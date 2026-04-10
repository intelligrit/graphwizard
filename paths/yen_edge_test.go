// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths

import (
	"context"
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestYenKShortest_KZero(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))

	paths := YenKShortest(context.Background(), g, 0, 1, 0)
	if len(paths) != 0 {
		t.Errorf("expected 0 paths for k=0, got %d", len(paths))
	}
}

func TestYenKShortest_KNegative(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))

	paths := YenKShortest(context.Background(), g, 0, 1, -1)
	if len(paths) != 0 {
		t.Errorf("expected 0 paths for k=-1, got %d", len(paths))
	}
}

func TestYenKShortest_DuplicatePaths(t *testing.T) {
	// Graph where spur paths can produce duplicate candidates.
	// 0->1 (1), 0->2 (1), 1->3 (1), 2->3 (1), 1->2 (1)
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))

	paths := YenKShortest(context.Background(), g, 0, 3, 10)
	// Should have: 0-1-3 (2), 0-2-3 (2), 0-1-2-3 (3) = 3 distinct paths.
	if len(paths) != 3 {
		t.Errorf("expected 3 distinct paths, got %d", len(paths))
	}
	// Verify no duplicates.
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if len(paths[i].Nodes) == len(paths[j].Nodes) {
				same := true
				for k := range paths[i].Nodes {
					if paths[i].Nodes[k].ID() != paths[j].Nodes[k].ID() {
						same = false
						break
					}
				}
				if same {
					t.Errorf("duplicate path at indices %d and %d", i, j)
				}
			}
		}
	}
}

func TestYenKShortest_MoreKThanPaths(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))

	paths := YenKShortest(context.Background(), g, 0, 1, 5)
	if len(paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(paths))
	}
}
