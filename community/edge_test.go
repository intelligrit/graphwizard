// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestLeiden_LargerGraph(t *testing.T) {
	// 3 dense clusters of 5 nodes each, connected by single bridges.
	g := simple.NewUndirectedGraph()

	// Cluster A: nodes 0-4 (complete).
	for i := int64(0); i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	// Cluster B: nodes 5-9 (complete).
	for i := int64(5); i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	// Cluster C: nodes 10-14 (complete).
	for i := int64(10); i < 15; i++ {
		for j := i + 1; j < 15; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	// Bridges.
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(9), simple.Node(10)))

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)

	// Count distinct communities.
	seen := make(map[int64]bool)
	for _, c := range comms {
		seen[c] = true
	}
	if len(seen) < 2 {
		t.Errorf("expected at least 2 communities, got %d", len(seen))
	}
}

func TestLeiden_TotalWeightZero(t *testing.T) {
	// Two isolated nodes, no edges. totalWeight = 0.
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)
	if len(comms) != 2 {
		t.Errorf("expected 2 community assignments, got %d", len(comms))
	}
}
