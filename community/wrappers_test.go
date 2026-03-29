// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"testing"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func TestLouvain_TwoClusters(t *testing.T) {
	g := simple.NewUndirectedGraph()
	// Cluster A
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	// Cluster B
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))
	// Bridge
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	comms := Louvain(g, 1.0, nil)
	if comms[0] != comms[1] || comms[1] != comms[2] {
		t.Errorf("cluster A should be in same community")
	}
	if comms[3] != comms[4] || comms[4] != comms[5] {
		t.Errorf("cluster B should be in same community")
	}
	if comms[0] == comms[3] {
		t.Error("clusters A and B should be in different communities")
	}
}

func TestLouvainQ_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	// All in one community.
	communities := [][]graph.Node{
		{simple.Node(0), simple.Node(1), simple.Node(2)},
	}
	q := LouvainQ(g, communities, 1.0)
	if q < 0 {
		t.Errorf("Q for single-community triangle should be >= 0, got %f", q)
	}
}
