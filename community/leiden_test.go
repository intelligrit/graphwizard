// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestLeiden_TwoClusters(t *testing.T) {
	// Two dense clusters connected by a single weak link.
	// Cluster A: 0-1-2-0 (triangle)
	// Cluster B: 3-4-5-3 (triangle)
	// Bridge: 2-3
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

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)

	// Nodes in the same cluster should be in the same community.
	if comms[0] != comms[1] || comms[1] != comms[2] {
		t.Errorf("cluster A nodes should be in same community: %d, %d, %d", comms[0], comms[1], comms[2])
	}
	if comms[3] != comms[4] || comms[4] != comms[5] {
		t.Errorf("cluster B nodes should be in same community: %d, %d, %d", comms[3], comms[4], comms[5])
	}
	// The two clusters should be in different communities.
	if comms[0] == comms[3] {
		t.Error("clusters A and B should be in different communities")
	}
}

func TestLeiden_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)
	if len(comms) != 1 {
		t.Errorf("expected 1 community assignment, got %d", len(comms))
	}
}

func TestLeiden_CompleteGraph(t *testing.T) {
	// K4: all nodes should be in the same community at resolution 1.0
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)

	first := comms[0]
	for id, c := range comms {
		if c != first {
			t.Errorf("K4: all nodes should be in same community; node %d in %d, expected %d", id, c, first)
		}
	}
}

func TestLeiden_EmptyGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)
	if len(comms) != 0 {
		t.Errorf("expected empty result for empty graph, got %v", comms)
	}
}
