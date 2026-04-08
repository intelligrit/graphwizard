// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

// unpreloadedUndirected wraps a simple.UndirectedGraph and implements
// DenseAdjacency with nil returns, simulating a diskgraph that has not
// been preloaded. KatzUndirected must fall back to the generic path.
type unpreloadedUndirected struct {
	*simple.UndirectedGraph
}

func (u unpreloadedUndirected) NodeIDs() []int64         { return nil }
func (u unpreloadedUndirected) DenseNeighbors(int) []int32 { return nil }
func (u unpreloadedUndirected) NumNodes() int {
	return u.UndirectedGraph.Nodes().Len()
}

const epsilon = 1e-6

func TestKatz_StarGraph(t *testing.T) {
	// Directed star: 1->0, 2->0, 3->0. Node 0 should have highest Katz.
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))

	scores := Katz(g, 0.1, 1.0, 1e-8, 100)
	if scores[0] <= scores[1] {
		t.Errorf("node 0 (hub) should have highest Katz; got 0=%f, 1=%f", scores[0], scores[1])
	}
}

func TestKatz_Chain(t *testing.T) {
	// Chain: 0->1->2->3. Node 3 accumulates all paths.
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	scores := Katz(g, 0.1, 1.0, 1e-8, 100)
	// Node 3 receives from 2, and transitively from 1 and 0.
	if scores[3] <= scores[0] {
		t.Errorf("node 3 should have highest Katz; got 3=%f, 0=%f", scores[3], scores[0])
	}
}

func TestKatz_EmptyGraph(t *testing.T) {
	g := simple.NewDirectedGraph()
	scores := Katz(g, 0.1, 1.0, 1e-8, 100)
	if len(scores) != 0 {
		t.Errorf("expected empty map for empty graph, got %v", scores)
	}
}

func TestKatzUndirected_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	scores := KatzUndirected(g, 0.1, 1.0, 1e-8, 100)
	// All nodes symmetric, should have equal Katz.
	if math.Abs(scores[0]-scores[1]) > epsilon || math.Abs(scores[1]-scores[2]) > epsilon {
		t.Errorf("symmetric triangle should have equal Katz; got %v", scores)
	}
}

// Regression test: a graph that satisfies DenseAdjacency but returns nil
// from NodeIDs() (unpreloaded diskgraph) must fall back to the generic
// path and produce correct non-zero results.
func TestKatzUndirected_UnpreloadedDenseAdjacency(t *testing.T) {
	inner := simple.NewUndirectedGraph()
	inner.SetEdge(inner.NewEdge(simple.Node(0), simple.Node(1)))
	inner.SetEdge(inner.NewEdge(simple.Node(1), simple.Node(2)))
	inner.SetEdge(inner.NewEdge(simple.Node(2), simple.Node(0)))

	g := unpreloadedUndirected{inner}

	// Verify the wrapper satisfies DenseAdjacency with nil NodeIDs.
	type denseAdj interface {
		NodeIDs() []int64
		DenseNeighbors(int) []int32
		NumNodes() int
	}
	var ug interface{} = g
	da, ok := ug.(denseAdj)
	if !ok {
		t.Fatal("test wrapper must satisfy DenseAdjacency")
	}
	if da.NodeIDs() != nil {
		t.Fatal("test wrapper NodeIDs must return nil")
	}

	scores := KatzUndirected(g, 0.1, 1.0, 1e-8, 100)
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}
	for id, v := range scores {
		if v == 0 {
			t.Errorf("node %d has zero Katz score; expected non-zero", id)
		}
	}
}
