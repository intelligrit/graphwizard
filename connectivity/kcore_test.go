// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestKCore_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	// 2-core: all nodes in the triangle have degree 2.
	core := KCore(context.Background(), 2, g)
	if len(core) != 3 {
		t.Errorf("expected 3 nodes in 2-core, got %d", len(core))
	}

	// 3-core: no nodes (max degree is 2).
	core = KCore(context.Background(), 3, g)
	if len(core) != 0 {
		t.Errorf("expected 0 nodes in 3-core, got %d", len(core))
	}
}

func TestKCore_StarWithTriangle(t *testing.T) {
	// Triangle 0-1-2-0 with pendant 3 on node 0.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))

	// 1-core: all 4 nodes.
	core := KCore(context.Background(), 1, g)
	if len(core) != 4 {
		t.Errorf("expected 4 nodes in 1-core, got %d", len(core))
	}

	// 2-core: only triangle nodes (3 has degree 1).
	core = KCore(context.Background(), 2, g)
	if len(core) != 3 {
		t.Errorf("expected 3 nodes in 2-core, got %d", len(core))
	}
}

func TestDegeneracyOrdering_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	order, coreLayers := DegeneracyOrdering(context.Background(), g)
	if len(order) != 3 {
		t.Errorf("expected 3 nodes in ordering, got %d", len(order))
	}
	if len(coreLayers) == 0 {
		t.Fatal("expected at least one core layer")
	}
}
