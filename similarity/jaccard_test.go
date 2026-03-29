// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

const epsilon = 1e-9

func TestJaccard_IdenticalNeighbors(t *testing.T) {
	// 0--2, 1--2: nodes 0 and 1 share neighbor 2, and only have neighbor 2
	// J(0,1) = |{2}| / |{2}| = 1.0
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	score := Jaccard(g, 0, 1)
	if math.Abs(score-1.0) > epsilon {
		t.Errorf("expected J=1.0, got %f", score)
	}
}

func TestJaccard_NoOverlap(t *testing.T) {
	// 0--2, 1--3: no shared neighbors
	// J(0,1) = 0 / |{2,3}| = 0
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))

	score := Jaccard(g, 0, 1)
	if score != 0 {
		t.Errorf("expected J=0, got %f", score)
	}
}

func TestJaccard_PartialOverlap(t *testing.T) {
	// 0--2, 0--3, 1--2, 1--4
	// N(0) = {2,3}, N(1) = {2,4}
	// intersection = {2}, union = {2,3,4}
	// J(0,1) = 1/3
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))

	score := Jaccard(g, 0, 1)
	if math.Abs(score-1.0/3.0) > epsilon {
		t.Errorf("expected J=1/3, got %f", score)
	}
}

func TestJaccard_BothIsolated(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	score := Jaccard(g, 0, 1)
	if score != 0 {
		t.Errorf("expected J=0 for isolated nodes, got %f", score)
	}
}

func TestJaccardAll_Threshold(t *testing.T) {
	// Triangle plus isolated node
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.AddNode(simple.Node(3))

	// In a triangle, N(0)={1,2}, N(1)={0,2}, N(2)={0,1}
	// J(0,1) = |{2}| / |{0,1,2}| = 1/3
	// All pairs have J=1/3
	results := JaccardAll(g, 1.0/3.0-epsilon)
	if len(results) != 3 {
		t.Fatalf("expected 3 pairs at threshold 1/3, got %d", len(results))
	}
	for _, r := range results {
		if math.Abs(r.Score-1.0/3.0) > epsilon {
			t.Errorf("pair %d-%d: expected J=1/3, got %f", r.A.ID(), r.B.ID(), r.Score)
		}
	}

	// Nothing at threshold 0.5
	results = JaccardAll(g, 0.5)
	if len(results) != 0 {
		t.Errorf("expected 0 pairs at threshold 0.5, got %d", len(results))
	}
}

func TestOverlap_PartialOverlap(t *testing.T) {
	// N(0) = {2,3}, N(1) = {2}
	// intersection = {2}, min(2,1) = 1
	// O(0,1) = 1/1 = 1.0
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	score := Overlap(g, 0, 1)
	if math.Abs(score-1.0) > epsilon {
		t.Errorf("expected O=1.0, got %f", score)
	}
}

func TestOverlap_NoNeighbors(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	score := Overlap(g, 0, 1)
	if score != 0 {
		t.Errorf("expected O=0 for isolated nodes, got %f", score)
	}
}
