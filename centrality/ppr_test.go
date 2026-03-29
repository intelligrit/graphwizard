// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestPersonalizedPageRank_Star(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))

	scores := PersonalizedPageRank(g, 0, 0.85, 1e-6, 100)
	if scores[0] <= scores[1] {
		t.Errorf("seed node should have highest PPR: 0=%f, 1=%f", scores[0], scores[1])
	}
}

func TestPersonalizedPageRank_Empty(t *testing.T) {
	g := simple.NewDirectedGraph()
	scores := PersonalizedPageRank(g, 0, 0.85, 1e-6, 100)
	if len(scores) != 0 {
		t.Errorf("expected empty map for empty graph")
	}
}

func TestPersonalizedPageRank_MissingSeed(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.AddNode(simple.Node(0))
	scores := PersonalizedPageRank(g, 99, 0.85, 1e-6, 100)
	if len(scores) != 0 {
		t.Errorf("expected empty map for missing seed")
	}
}

func TestPersonalizedPageRankUndirected_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	scores := PersonalizedPageRankUndirected(g, 0, 0.85, 1e-6, 100)
	// Seed and its neighbor should have the highest scores.
	// Far end (node 3) should have the lowest score.
	if scores[3] >= scores[0] {
		t.Errorf("seed (0) should score higher than farthest node (3): %v", scores)
	}
	if scores[3] >= scores[1] {
		t.Errorf("node 1 (neighbor of seed) should score higher than node 3: %v", scores)
	}
}

func TestPersonalizedPageRankUndirected_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	scores := PersonalizedPageRankUndirected(g, 0, 0.85, 1e-6, 100)
	if len(scores) != 0 {
		t.Errorf("expected empty map for empty graph")
	}
}

func TestPersonalizedPageRankUndirected_MissingSeed(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	scores := PersonalizedPageRankUndirected(g, 99, 0.85, 1e-6, 100)
	if len(scores) != 0 {
		t.Errorf("expected empty map for missing seed")
	}
}
