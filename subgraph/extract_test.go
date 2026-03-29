// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package subgraph

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestNHopNeighborhood_Directed(t *testing.T) {
	// Graph: 0->1->2->3->4
	g := simple.NewDirectedGraph()
	for i := int64(0); i < 5; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < 4; i++ {
		g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(i+1)))
	}

	sub := NHopNeighborhood(g, 1, 2)

	// From node 1 with 2 hops: should reach 0 (incoming), 2 (outgoing),
	// then from 0 nothing new, from 2 reach 3 (outgoing), from 2 reach 1 (incoming, already visited).
	// So nodes: {0, 1, 2, 3}
	nodeCount := 0
	nodes := sub.Nodes()
	for nodes.Next() {
		nodeCount++
	}
	if nodeCount != 4 {
		t.Errorf("expected 4 nodes, got %d", nodeCount)
	}

	// Node 4 should not be included.
	if sub.Node(4) != nil {
		t.Error("node 4 should not be in 2-hop neighborhood of 1")
	}
}

func TestNHopNeighborhood_MissingCenter(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.AddNode(simple.Node(0))

	sub := NHopNeighborhood(g, 99, 2)
	nodeCount := 0
	nodes := sub.Nodes()
	for nodes.Next() {
		nodeCount++
	}
	if nodeCount != 0 {
		t.Errorf("expected 0 nodes for missing center, got %d", nodeCount)
	}
}

func TestNHopNeighborhood_ZeroHops(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	sub := NHopNeighborhood(g, 0, 0)
	nodeCount := 0
	nodes := sub.Nodes()
	for nodes.Next() {
		nodeCount++
	}
	if nodeCount != 1 {
		t.Errorf("expected 1 node for 0-hop, got %d", nodeCount)
	}
}

func TestNHopNeighborhoodUndirected(t *testing.T) {
	// Star graph: center 0 connected to 1,2,3,4; also 1-5.
	g := simple.NewUndirectedGraph()
	for i := int64(0); i <= 5; i++ {
		g.AddNode(simple.Node(i))
	}
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(5)))

	// 1-hop from 0: {0,1,2,3,4}
	sub1 := NHopNeighborhoodUndirected(g, 0, 1)
	count1 := 0
	n1 := sub1.Nodes()
	for n1.Next() {
		count1++
	}
	if count1 != 5 {
		t.Errorf("1-hop: expected 5 nodes, got %d", count1)
	}

	// 2-hop from 0: {0,1,2,3,4,5}
	sub2 := NHopNeighborhoodUndirected(g, 0, 2)
	count2 := 0
	n2 := sub2.Nodes()
	for n2.Next() {
		count2++
	}
	if count2 != 6 {
		t.Errorf("2-hop: expected 6 nodes, got %d", count2)
	}

	// Edges should be preserved.
	if !sub2.HasEdgeBetween(0, 1) {
		t.Error("missing edge 0-1")
	}
	if !sub2.HasEdgeBetween(1, 5) {
		t.Error("missing edge 1-5")
	}
}

func TestNHopNeighborhoodUndirected_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	sub := NHopNeighborhoodUndirected(g, 0, 3)
	count := 0
	nodes := sub.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 nodes for empty graph, got %d", count)
	}
}
