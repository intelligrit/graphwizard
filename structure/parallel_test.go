// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestTriangleCountParallel_K4(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	perNode, total := TriangleCountParallel(g)
	if total != 4 {
		t.Errorf("expected 4 triangles in K4, got %d", total)
	}
	for id, count := range perNode {
		if count != 3 {
			t.Errorf("node %d: expected 3 triangles, got %d", id, count)
		}
	}
}

func TestTriangleCountParallel_Path(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	_, total := TriangleCountParallel(g)
	if total != 0 {
		t.Errorf("expected 0 triangles, got %d", total)
	}
}

func TestTriangleCountParallel_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	perNode, total := TriangleCountParallel(g)
	if total != 0 || len(perNode) != 0 {
		t.Errorf("expected empty results")
	}
}

func TestClusteringCoefficientParallel_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	coeffs := ClusteringCoefficientParallel(g)
	for id, c := range coeffs {
		if math.Abs(c-1.0) > 1e-9 {
			t.Errorf("node %d: expected CC=1.0, got %f", id, c)
		}
	}
}

func TestClusteringCoefficientParallel_Star(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))

	coeffs := ClusteringCoefficientParallel(g)
	for _, c := range coeffs {
		if c != 0 {
			t.Errorf("star graph should have CC=0, got %f", c)
		}
	}
}
