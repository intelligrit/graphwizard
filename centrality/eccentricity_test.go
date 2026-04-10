// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestEccentricity_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	ecc := Eccentricity(context.Background(), g)
	// Node 0: max dist = 2 (to node 2). Node 1: max dist = 1. Node 2: max dist = 2.
	if math.Abs(ecc[0]-2.0) > epsilon {
		t.Errorf("node 0: expected ecc 2.0, got %f", ecc[0])
	}
	if math.Abs(ecc[1]-1.0) > epsilon {
		t.Errorf("node 1: expected ecc 1.0, got %f", ecc[1])
	}
}

func TestDiameter_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	d := Diameter(context.Background(), g)
	if math.Abs(d-2.0) > epsilon {
		t.Errorf("expected diameter 2.0, got %f", d)
	}
}

func TestDiameter_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	d := Diameter(context.Background(), g)
	if d != 0 {
		t.Errorf("expected diameter 0 for empty graph, got %f", d)
	}
}

func TestRadius_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	r := Radius(context.Background(), g)
	if math.Abs(r-1.0) > epsilon {
		t.Errorf("expected radius 1.0, got %f", r)
	}
}

func TestRadius_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	r := Radius(context.Background(), g)
	if !math.IsInf(r, 1) {
		t.Errorf("expected +Inf for empty graph, got %f", r)
	}
}
