// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func TestCommonNeighbors_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	cn := CommonNeighbors(g, 0, 1)
	if cn != 1 { // shared neighbor: 2
		t.Errorf("expected 1 common neighbor, got %d", cn)
	}
}

func TestCommonNeighbors_NoShared(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	cn := CommonNeighbors(g, 0, 2)
	if cn != 0 {
		t.Errorf("expected 0, got %d", cn)
	}
}

func TestAdamicAdar_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	aa := AdamicAdar(g, 0, 1)
	// Common neighbor is 2, degree(2)=2, so 1/log(2)
	expected := 1.0 / math.Log(2)
	if math.Abs(aa-expected) > epsilon {
		t.Errorf("expected %.4f, got %.4f", expected, aa)
	}
}

func TestAdamicAdar_HighDegreeNeighbor(t *testing.T) {
	// Hub node 2 has degree 4 — less informative as shared neighbor.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(4)))

	aa := AdamicAdar(g, 0, 1)
	expected := 1.0 / math.Log(4)
	if math.Abs(aa-expected) > epsilon {
		t.Errorf("expected %.4f, got %.4f", expected, aa)
	}
}

func TestPreferentialAttachment(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))

	pa := PreferentialAttachment(g, 0, 1)
	if pa != 2 { // deg(0)=2 * deg(1)=1
		t.Errorf("expected 2, got %d", pa)
	}
}

func TestPredictLinks_Square(t *testing.T) {
	// Square 0-1-2-3-0: missing diagonals 0-2 and 1-3.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))

	scorer := func(g graph.Undirected, u, v int64) float64 {
		return float64(CommonNeighbors(g, u, v))
	}
	preds := PredictLinks(g, 5, scorer)
	if len(preds) != 2 {
		t.Fatalf("expected 2 predictions, got %d", len(preds))
	}
	// Both diagonals should have CN=2.
	for _, p := range preds {
		if p.Score != 2 {
			t.Errorf("expected score 2, got %.0f for %d-%d", p.Score, p.A, p.B)
		}
	}
}

func TestCosine_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	c := Cosine(g, 0, 1)
	// N(0)={1,2}, N(1)={0,2}, intersection={2}, cos = 1/sqrt(2*2) = 0.5
	if math.Abs(c-0.5) > epsilon {
		t.Errorf("expected 0.5, got %f", c)
	}
}

func TestCosine_Isolated(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	c := Cosine(g, 0, 1)
	if c != 0 {
		t.Errorf("expected 0 for isolated nodes, got %f", c)
	}
}
