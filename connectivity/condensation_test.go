// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestCondensation_TwoSCCs(t *testing.T) {
	// Cycle 0->1->2->0 and chain 2->3.
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	comps, edges, nodeToSCC := Condensation(context.Background(), g)
	if len(comps) != 2 {
		t.Fatalf("expected 2 SCCs, got %d", len(comps))
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 condensation edge, got %d", len(edges))
	}
	// Nodes 0,1,2 should be in the same SCC.
	if nodeToSCC[0] != nodeToSCC[1] || nodeToSCC[1] != nodeToSCC[2] {
		t.Error("nodes 0,1,2 should be in the same SCC")
	}
	// Node 3 should be in a different SCC.
	if nodeToSCC[0] == nodeToSCC[3] {
		t.Error("node 3 should be in a different SCC")
	}
}

func TestCondensation_DAG(t *testing.T) {
	// Already a DAG: 0->1->2.
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	comps, edges, _ := Condensation(context.Background(), g)
	if len(comps) != 3 {
		t.Fatalf("expected 3 SCCs (each node), got %d", len(comps))
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 condensation edges, got %d", len(edges))
	}
}

func TestCondensation_SingleSCC(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))

	comps, edges, _ := Condensation(context.Background(), g)
	if len(comps) != 1 {
		t.Fatalf("expected 1 SCC, got %d", len(comps))
	}
	if len(edges) != 0 {
		t.Fatalf("expected 0 edges, got %d", len(edges))
	}
}
