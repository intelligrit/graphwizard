// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestApproximateBetweenness_Star(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for i := int64(1); i <= 4; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}

	rng := rand.New(rand.NewSource(42))
	scores := ApproximateBetweenness(context.Background(), g, 5, rng)
	if scores[0] <= scores[1] {
		t.Errorf("center should have highest approx betweenness: 0=%f, 1=%f", scores[0], scores[1])
	}
}

func TestApproximateBetweenness_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(42))
	scores := ApproximateBetweenness(context.Background(), g, 10, rng)
	if len(scores) != 0 {
		t.Errorf("expected empty, got %d", len(scores))
	}
}

func TestApproximateBetweenness_KZero(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	rng := rand.New(rand.NewSource(42))
	scores := ApproximateBetweenness(context.Background(), g, 0, rng)
	if len(scores) != 0 {
		t.Errorf("expected empty for k=0, got %d", len(scores))
	}
}

func TestApproximateBetweenness_KExceedsN(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	rng := rand.New(rand.NewSource(42))
	scores := ApproximateBetweenness(context.Background(), g, 100, rng)
	if len(scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(scores))
	}
}

func TestApproximateBetweenness_Weighted(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 1))

	rng := rand.New(rand.NewSource(42))
	scores := ApproximateBetweenness(context.Background(), g, 4, rng)
	// Middle nodes should have higher betweenness.
	if scores[1] <= scores[0] {
		t.Errorf("node 1 should have higher betweenness than endpoint 0")
	}
}

func TestEccentricityParallel_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	ecc := EccentricityParallel(context.Background(), g)
	if math.Abs(ecc[0]-2.0) > epsilon {
		t.Errorf("node 0: expected ecc 2.0, got %f", ecc[0])
	}
	if math.Abs(ecc[1]-1.0) > epsilon {
		t.Errorf("node 1: expected ecc 1.0, got %f", ecc[1])
	}
}

func TestEccentricityParallel_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	ecc := EccentricityParallel(context.Background(), g)
	if len(ecc) != 0 {
		t.Errorf("expected empty, got %d", len(ecc))
	}
}

func TestDiameterParallel_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	d := DiameterParallel(context.Background(), g)
	if math.Abs(d-2.0) > epsilon {
		t.Errorf("expected diameter 2.0, got %f", d)
	}
}
