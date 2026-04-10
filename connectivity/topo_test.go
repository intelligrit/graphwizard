// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestTopologicalSort_DAG(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))

	order, ok := TopologicalSort(context.Background(), g)
	if !ok {
		t.Fatal("expected ok=true for DAG")
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}

	// Verify order: 0 must come before 1 and 2, 1 must come before 2.
	pos := make(map[int64]int)
	for i, id := range order {
		pos[id] = i
	}
	if pos[0] > pos[1] || pos[0] > pos[2] || pos[1] > pos[2] {
		t.Errorf("invalid topological order: %v", order)
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))

	_, ok := TopologicalSort(context.Background(), g)
	if ok {
		t.Error("expected ok=false for cycle")
	}
}

func TestTopologicalSort_Empty(t *testing.T) {
	g := simple.NewDirectedGraph()
	order, ok := TopologicalSort(context.Background(), g)
	if !ok {
		t.Error("expected ok=true for empty graph")
	}
	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}
}
