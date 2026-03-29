// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestInfluenceMaximization_Star(t *testing.T) {
	// Star: center should be the most influential seed.
	g := simple.NewUndirectedGraph()
	for i := int64(1); i <= 5; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}

	rng := rand.New(rand.NewSource(42))
	seeds, influence := InfluenceMaximization(g, 1, 0.5, 100, rng)

	if len(seeds) != 1 {
		t.Fatalf("expected 1 seed, got %d", len(seeds))
	}
	if seeds[0] != 0 {
		t.Errorf("expected center node 0 as seed, got %d", seeds[0])
	}
	if influence < 2.0 {
		t.Errorf("influence too low: %f", influence)
	}
}

func TestInfluenceMaximization_MultipleSeeds(t *testing.T) {
	g := simple.NewUndirectedGraph()
	// Two disconnected stars.
	for i := int64(1); i <= 3; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}
	for i := int64(5); i <= 7; i++ {
		g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(i)))
	}

	rng := rand.New(rand.NewSource(42))
	seeds, _ := InfluenceMaximization(g, 2, 0.5, 100, rng)

	if len(seeds) != 2 {
		t.Fatalf("expected 2 seeds, got %d", len(seeds))
	}
}

func TestInfluenceMaximization_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(42))
	seeds, _ := InfluenceMaximization(g, 1, 0.5, 100, rng)
	if seeds != nil {
		t.Errorf("expected nil seeds for empty graph")
	}
}

func TestInfluenceMaximization_KZero(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	rng := rand.New(rand.NewSource(42))
	seeds, _ := InfluenceMaximization(g, 0, 0.5, 100, rng)
	if seeds != nil {
		t.Errorf("expected nil seeds for k=0")
	}
}
