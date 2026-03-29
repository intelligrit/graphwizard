// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

func TestPageRank_Triangle(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	scores := PageRank(g, 0.85, 1e-6)
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}
	if math.Abs(scores[0]-scores[1]) > epsilon || math.Abs(scores[1]-scores[2]) > epsilon {
		t.Errorf("cycle should have equal PageRank: %v", scores)
	}
}

func TestPageRankSparse_Triangle(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	scores := PageRankSparse(g, 0.85, 1e-6)
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}
}

func TestBetweenness_Star(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for i := int64(1); i <= 4; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}

	scores := Betweenness(g)
	for i := int64(1); i <= 4; i++ {
		if scores[0] <= scores[i] {
			t.Errorf("center should have higher betweenness than leaf %d", i)
		}
	}
}

func TestBetweennessWeighted_Chain(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))

	allPaths := path.DijkstraAllPaths(g)
	scores := BetweennessWeighted(g, allPaths)
	if scores[1] <= scores[0] {
		t.Errorf("node 1 should have higher betweenness than node 0")
	}
}

func TestEdgeBetweenness_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	scores := EdgeBetweenness(g)
	if len(scores) == 0 {
		t.Fatal("expected non-empty edge betweenness")
	}
}

func TestEdgeBetweennessWeighted_Chain(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))

	allPaths := path.DijkstraAllPaths(g)
	scores := EdgeBetweennessWeighted(g, allPaths)
	if len(scores) == 0 {
		t.Fatal("expected non-empty edge betweenness")
	}
}

func TestCloseness_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	allPaths := path.DijkstraAllPaths(g)
	scores := Closeness(g, allPaths)
	if len(scores) != 3 {
		t.Fatalf("expected 3 closeness scores, got %d", len(scores))
	}
	if math.Abs(scores[0]-scores[1]) > epsilon {
		t.Errorf("symmetric graph should have equal closeness")
	}
}

func TestHarmonic_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	allPaths := path.DijkstraAllPaths(g)
	scores := Harmonic(g, allPaths)
	if len(scores) != 3 {
		t.Fatalf("expected 3 harmonic scores, got %d", len(scores))
	}
}

func TestHITS_Star(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))

	result := HITS(g, 1e-6)
	if result.Authority[0] <= result.Authority[1] {
		t.Errorf("node 0 should have highest authority")
	}
	if result.Hub[1] <= result.Hub[0] {
		t.Errorf("node 1 should have higher hub score than node 0")
	}
}

func TestInDegree_SingleNode(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.AddNode(simple.Node(0))
	scores := InDegree(g)
	if scores[0] != 0 {
		t.Errorf("expected 0, got %f", scores[0])
	}
}

func TestOutDegree_SingleNode(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.AddNode(simple.Node(0))
	scores := OutDegree(g)
	if scores[0] != 0 {
		t.Errorf("expected 0, got %f", scores[0])
	}
}

func TestKatzUndirected_EmptyGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	scores := KatzUndirected(g, 0.1, 1.0, 1e-8, 100)
	if scores != nil {
		t.Errorf("expected nil for empty graph, got %v", scores)
	}
}
