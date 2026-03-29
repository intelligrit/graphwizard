// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package embedding

import (
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestEmbed_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	rng := rand.New(rand.NewSource(42))
	walks := DeepWalkWalks(g, 10, 20, rng)
	nodeIDs := []int64{0, 1, 2}

	emb := Embed(walks, nodeIDs, 2, 3)
	if len(emb) != 3 {
		t.Fatalf("expected 3 embeddings, got %d", len(emb))
	}
	for _, id := range nodeIDs {
		vec := emb[id]
		if len(vec) != 2 {
			t.Errorf("node %d: expected dim 2, got %d", id, len(vec))
		}
	}
}

func TestEmbed_Empty(t *testing.T) {
	emb := Embed(nil, nil, 2, 3)
	if len(emb) != 0 {
		t.Errorf("expected empty embedding for empty input")
	}
}

func TestEmbed_NoWalks(t *testing.T) {
	nodeIDs := []int64{0, 1}
	emb := Embed(nil, nodeIDs, 2, 3)
	if len(emb) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(emb))
	}
	// With no walks, all embeddings should be zero vectors.
	for _, id := range nodeIDs {
		for _, v := range emb[id] {
			if v != 0 {
				t.Errorf("node %d: expected zero vector, got non-zero", id)
				break
			}
		}
	}
}

func TestEmbed_DimLargerThanNodes(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	rng := rand.New(rand.NewSource(42))
	walks := DeepWalkWalks(g, 5, 10, rng)
	nodeIDs := []int64{0, 1}

	// dim=10 but only 2 nodes — should clamp to 2.
	emb := Embed(walks, nodeIDs, 10, 3)
	if len(emb) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(emb))
	}
	// Vector length should be clamped to min(dim, n) = 2.
	if len(emb[0]) != 2 {
		t.Errorf("expected dim clamped to 2, got %d", len(emb[0]))
	}
}
