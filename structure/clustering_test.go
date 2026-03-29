// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

const epsilon = 1e-9

func TestClusteringCoefficient_Triangle(t *testing.T) {
	// Triangle: every node has C=1.0
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	coeffs := ClusteringCoefficient(g)
	for id, c := range coeffs {
		if math.Abs(c-1.0) > epsilon {
			t.Errorf("node %d: expected C=1.0, got %f", id, c)
		}
	}
}

func TestClusteringCoefficient_Star(t *testing.T) {
	// Star: center node has C=0 (neighbors not connected to each other)
	// Leaf nodes have degree 1, so C=0
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))

	coeffs := ClusteringCoefficient(g)
	for id, c := range coeffs {
		if c != 0 {
			t.Errorf("node %d: expected C=0, got %f", id, c)
		}
	}
}

func TestClusteringCoefficient_Square(t *testing.T) {
	// Square: 0-1-2-3-0. Each node has 2 neighbors, 0 edges between them.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))

	coeffs := ClusteringCoefficient(g)
	for id, c := range coeffs {
		if c != 0 {
			t.Errorf("node %d: expected C=0, got %f", id, c)
		}
	}
}

func TestClusteringCoefficient_SquareWithDiagonal(t *testing.T) {
	// Square with one diagonal: 0-1-2-3-0, plus 0-2
	// Node 0: neighbors {1,2,3}, edges between: 1-2? yes, 2-3? yes, 1-3? no => 2
	//   C(0) = 2*2 / (3*2) = 2/3
	// Node 2: neighbors {1,3,0}, edges between: 1-0? yes, 3-0? yes, 1-3? no => 2
	//   C(2) = 2*2 / (3*2) = 2/3
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))

	coeffs := ClusteringCoefficient(g)
	if math.Abs(coeffs[0]-2.0/3.0) > epsilon {
		t.Errorf("node 0: expected C=2/3, got %f", coeffs[0])
	}
	if math.Abs(coeffs[2]-2.0/3.0) > epsilon {
		t.Errorf("node 2: expected C=2/3, got %f", coeffs[2])
	}
}

func TestAverageClusteringCoefficient_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	avg := AverageClusteringCoefficient(g)
	if math.Abs(avg-1.0) > epsilon {
		t.Errorf("expected average C=1.0, got %f", avg)
	}
}

func TestClusteringCoefficient_EmptyGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	coeffs := ClusteringCoefficient(g)
	if len(coeffs) != 0 {
		t.Errorf("expected empty map, got %d entries", len(coeffs))
	}
}
