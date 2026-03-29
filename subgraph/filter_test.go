// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package subgraph

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestFilterNodes_KeepEven(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 6; i++ {
		g.AddNode(simple.Node(i))
	}
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))

	// Keep only even nodes.
	sub := FilterNodes(g, func(id int64) bool { return id%2 == 0 })

	count := 0
	nodes := sub.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 even nodes, got %d", count)
	}

	// Edge 0-2 should exist, edge 2-4 should exist.
	if !sub.HasEdgeBetween(0, 2) {
		t.Error("missing edge 0-2")
	}
	if !sub.HasEdgeBetween(2, 4) {
		t.Error("missing edge 2-4")
	}
	// Edge 0-1 should not exist (1 is odd).
	if sub.HasEdgeBetween(0, 1) {
		t.Error("unexpected edge 0-1")
	}
}

func TestFilterNodes_KeepNone(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	sub := FilterNodes(g, func(id int64) bool { return false })
	count := 0
	nodes := sub.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 nodes, got %d", count)
	}
}

func TestFilterNodes_KeepAll(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	sub := FilterNodes(g, func(id int64) bool { return true })
	count := 0
	nodes := sub.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 nodes, got %d", count)
	}
	if !sub.HasEdgeBetween(0, 1) {
		t.Error("missing edge 0-1")
	}
}
