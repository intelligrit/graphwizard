// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestDegree_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	scores := Degree(g)
	// Each node has degree 2, n-1=2, so C_D = 1.0
	for id, c := range scores {
		if math.Abs(c-1.0) > epsilon {
			t.Errorf("node %d: expected C_D=1.0, got %f", id, c)
		}
	}
}

func TestDegree_Star(t *testing.T) {
	// Star: center=0 connected to 1,2,3,4
	g := simple.NewUndirectedGraph()
	for i := int64(1); i <= 4; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}

	scores := Degree(g)
	// Center: deg=4, n-1=4, C_D=1.0
	if math.Abs(scores[0]-1.0) > epsilon {
		t.Errorf("center: expected C_D=1.0, got %f", scores[0])
	}
	// Leaf: deg=1, n-1=4, C_D=0.25
	if math.Abs(scores[1]-0.25) > epsilon {
		t.Errorf("leaf: expected C_D=0.25, got %f", scores[1])
	}
}

func TestDegree_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	scores := Degree(g)
	if scores[0] != 0 {
		t.Errorf("single node: expected C_D=0, got %f", scores[0])
	}
}

func TestInDegree_Chain(t *testing.T) {
	// 0->1->2
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	scores := InDegree(g)
	if scores[0] != 0 {
		t.Errorf("node 0: expected in-degree 0, got %f", scores[0])
	}
	if math.Abs(scores[1]-0.5) > epsilon {
		t.Errorf("node 1: expected in-degree 0.5, got %f", scores[1])
	}
	if math.Abs(scores[2]-0.5) > epsilon {
		t.Errorf("node 2: expected in-degree 0.5, got %f", scores[2])
	}
}

func TestOutDegree_Chain(t *testing.T) {
	// 0->1->2
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	scores := OutDegree(g)
	if math.Abs(scores[0]-0.5) > epsilon {
		t.Errorf("node 0: expected out-degree 0.5, got %f", scores[0])
	}
	if math.Abs(scores[1]-0.5) > epsilon {
		t.Errorf("node 1: expected out-degree 0.5, got %f", scores[1])
	}
	if scores[2] != 0 {
		t.Errorf("node 2: expected out-degree 0, got %f", scores[2])
	}
}
