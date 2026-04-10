// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package embedding

import (
	"context"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestNode2VecWalks_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	rng := rand.New(rand.NewSource(42))
	walks := Node2VecWalks(context.Background(), g, WalkParams{
		WalkLength:   5,
		WalksPerNode: 2,
		P:            1.0,
		Q:            1.0,
	}, rng)

	// 3 nodes * 2 walks = 6 walks.
	if len(walks) != 6 {
		t.Fatalf("expected 6 walks, got %d", len(walks))
	}
	for i, walk := range walks {
		if len(walk) != 5 {
			t.Errorf("walk %d: expected length 5, got %d", i, len(walk))
		}
	}
}

func TestNode2VecWalks_BFSBias(t *testing.T) {
	// Low Q = BFS-like. Walks should stay local.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	rng := rand.New(rand.NewSource(42))
	walks := Node2VecWalks(context.Background(), g, WalkParams{
		WalkLength:   10,
		WalksPerNode: 1,
		P:            1.0,
		Q:            0.1, // strong BFS bias
	}, rng)

	if len(walks) != 4 {
		t.Fatalf("expected 4 walks, got %d", len(walks))
	}
}

func TestNode2VecWalks_IsolatedNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	rng := rand.New(rand.NewSource(42))
	walks := Node2VecWalks(context.Background(), g, WalkParams{
		WalkLength:   5,
		WalksPerNode: 1,
		P:            1.0,
		Q:            1.0,
	}, rng)

	// 3 nodes, 1 walk each.
	if len(walks) != 3 {
		t.Fatalf("expected 3 walks, got %d", len(walks))
	}
	// The isolated node's walk should be length 1.
	found := false
	for _, walk := range walks {
		if walk[0] == 0 && len(walk) == 1 {
			found = true
		}
	}
	if !found {
		t.Error("isolated node should produce a walk of length 1")
	}
}

func TestDeepWalkWalks_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	rng := rand.New(rand.NewSource(42))
	walks := DeepWalkWalks(context.Background(), g, 5, 2, rng)

	if len(walks) != 6 {
		t.Fatalf("expected 6 walks, got %d", len(walks))
	}
}

func TestNode2VecWalks_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(42))
	walks := Node2VecWalks(context.Background(), g, WalkParams{WalkLength: 5, WalksPerNode: 1, P: 1, Q: 1}, rng)
	if len(walks) != 0 {
		t.Errorf("expected 0 walks for empty graph, got %d", len(walks))
	}
}
