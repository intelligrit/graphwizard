// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestBipartiteProject_Simple(t *testing.T) {
	// Providers: 0, 1, 2. Organizations: 10, 11.
	// 0-10, 1-10, 1-11, 2-11
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(10)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(10)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(11)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(11)))

	proj := BipartiteProject(context.Background(), g, []int64{0, 1, 2})

	// 0-1 share org 10 (weight 1).
	if !proj.HasEdgeBetween(0, 1) {
		t.Error("expected edge 0-1")
	}
	w01, _ := proj.Weight(0, 1)
	if math.Abs(w01-1.0) > epsilon {
		t.Errorf("expected weight 1.0 for 0-1, got %f", w01)
	}

	// 1-2 share org 11 (weight 1).
	if !proj.HasEdgeBetween(1, 2) {
		t.Error("expected edge 1-2")
	}

	// 0-2 share no org.
	if proj.HasEdgeBetween(0, 2) {
		t.Error("expected no edge 0-2")
	}
}

func TestBipartiteProject_MultipleShared(t *testing.T) {
	// Providers 0 and 1 share orgs 10 and 11. Weight should be 2.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(10)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(11)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(10)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(11)))

	proj := BipartiteProject(context.Background(), g, []int64{0, 1})
	w, _ := proj.Weight(0, 1)
	if math.Abs(w-2.0) > epsilon {
		t.Errorf("expected weight 2.0, got %f", w)
	}
}

func TestBipartiteProject_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	proj := BipartiteProject(context.Background(), g, nil)
	if proj.Nodes().Len() != 0 {
		t.Error("expected empty projection")
	}
}
