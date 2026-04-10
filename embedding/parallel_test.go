// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package embedding

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestNode2VecWalksParallel_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	walks := Node2VecWalksParallel(context.Background(), g, WalkParams{
		WalkLength:   5,
		WalksPerNode: 2,
		P:            1.0,
		Q:            1.0,
	}, 42)

	if len(walks) != 6 {
		t.Fatalf("expected 6 walks, got %d", len(walks))
	}
	for i, walk := range walks {
		if len(walk) != 5 {
			t.Errorf("walk %d: expected length 5, got %d", i, len(walk))
		}
	}
}

func TestNode2VecWalksParallel_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	walks := Node2VecWalksParallel(context.Background(), g, WalkParams{WalkLength: 5, WalksPerNode: 1, P: 1, Q: 1}, 42)
	if walks != nil {
		t.Errorf("expected nil for empty graph, got %d walks", len(walks))
	}
}

func TestNode2VecWalksParallel_Isolated(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	walks := Node2VecWalksParallel(context.Background(), g, WalkParams{WalkLength: 5, WalksPerNode: 1, P: 1, Q: 1}, 42)
	if len(walks) != 1 {
		t.Fatalf("expected 1 walk, got %d", len(walks))
	}
	if len(walks[0]) != 1 {
		t.Errorf("isolated node walk should be length 1, got %d", len(walks[0]))
	}
}

func TestDeepWalkWalksParallel_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	walks := DeepWalkWalksParallel(context.Background(), g, 5, 2, 42)
	if len(walks) != 6 {
		t.Fatalf("expected 6 walks, got %d", len(walks))
	}
}
