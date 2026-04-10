// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestTriangleCount_SingleTriangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	perNode, total := TriangleCount(context.Background(), g)
	if total != 1 {
		t.Errorf("expected 1 triangle, got %d", total)
	}
	for id, count := range perNode {
		if count != 1 {
			t.Errorf("node %d: expected 1 triangle, got %d", id, count)
		}
	}
}

func TestTriangleCount_K4(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	perNode, total := TriangleCount(context.Background(), g)
	// K4 has C(4,3) = 4 triangles.
	if total != 4 {
		t.Errorf("expected 4 triangles in K4, got %d", total)
	}
	// Each node participates in 3 triangles.
	for id, count := range perNode {
		if count != 3 {
			t.Errorf("node %d: expected 3 triangles, got %d", id, count)
		}
	}
}

func TestTriangleCount_NoTriangles(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	_, total := TriangleCount(context.Background(), g)
	if total != 0 {
		t.Errorf("expected 0 triangles in path, got %d", total)
	}
}

func TestTriangleCount_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	perNode, total := TriangleCount(context.Background(), g)
	if total != 0 || len(perNode) != 0 {
		t.Errorf("expected empty results for empty graph")
	}
}
