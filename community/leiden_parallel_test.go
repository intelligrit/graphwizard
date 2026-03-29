// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestLeidenParallel_TwoCliques(t *testing.T) {
	g := simple.NewUndirectedGraph()
	// Two 4-cliques connected by a single edge.
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	for i := int64(4); i < 8; i++ {
		for j := i + 1; j < 8; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))

	rng := rand.New(rand.NewSource(42))
	result := LeidenParallel(g, 1.0, rng)

	if len(result) != 8 {
		t.Fatalf("expected 8 assignments, got %d", len(result))
	}

	// Nodes within the same clique should be in the same community.
	for i := int64(1); i < 4; i++ {
		if result[i] != result[0] {
			t.Errorf("clique A: node %d in community %d, expected %d", i, result[i], result[0])
		}
	}
	for i := int64(5); i < 8; i++ {
		if result[i] != result[4] {
			t.Errorf("clique B: node %d in community %d, expected %d", i, result[i], result[4])
		}
	}
}

func TestLeidenParallel_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(1))
	result := LeidenParallel(g, 1.0, rng)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestLeidenParallel_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	rng := rand.New(rand.NewSource(1))
	result := LeidenParallel(g, 1.0, rng)
	if len(result) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(result))
	}
}
