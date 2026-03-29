// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

const epsilon = 1e-9

// buildStarGraph creates a star with a center connected to n leaves.
func buildStarGraph(center int64, leaves int) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(center))
	for i := 0; i < leaves; i++ {
		id := center + int64(i) + 1
		g.AddNode(simple.Node(id))
		g.SetEdge(g.NewEdge(simple.Node(center), simple.Node(id)))
	}
	return g
}

func TestDegreeZScore_StarGraph(t *testing.T) {
	g := buildStarGraph(0, 4)
	scores := DegreeZScore(g)

	// Center has degree 4, leaves have degree 1. Mean=1.6, stddev>0.
	if scores[0] <= 0 {
		t.Errorf("center should have positive z-score, got %f", scores[0])
	}
	for i := int64(1); i <= 4; i++ {
		if scores[i] >= 0 {
			t.Errorf("leaf %d should have negative z-score, got %f", i, scores[i])
		}
	}
}

func TestDegreeZScore_RegularGraph(t *testing.T) {
	// Complete graph K4: all degrees are 3.
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	scores := DegreeZScore(g)
	for i := int64(0); i < 4; i++ {
		if math.Abs(scores[i]) > epsilon {
			t.Errorf("node %d should have z-score 0 in regular graph, got %f", i, scores[i])
		}
	}
}

func TestDegreeZScore_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	scores := DegreeZScore(g)
	if len(scores) != 0 {
		t.Errorf("expected empty map, got %v", scores)
	}
}

func TestDegreeZScore_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	scores := DegreeZScore(g)
	if scores[0] != 0 {
		t.Errorf("single node should have z-score 0, got %f", scores[0])
	}
}

func TestIsolationScore_StarGraph(t *testing.T) {
	g := buildStarGraph(0, 4)
	scores := IsolationScore(g)

	// The center node (high degree, different structure) should have a
	// non-zero isolation score.
	if scores[0] == 0 {
		t.Error("center of star should have non-zero isolation score")
	}

	// All leaves should have the same score.
	leafScore := scores[1]
	for i := int64(2); i <= 4; i++ {
		if math.Abs(scores[i]-leafScore) > epsilon {
			t.Errorf("leaves should have equal scores: leaf 1=%f, leaf %d=%f", leafScore, i, scores[i])
		}
	}
}

func TestIsolationScore_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	scores := IsolationScore(g)
	if len(scores) != 0 {
		t.Errorf("expected empty map, got %v", scores)
	}
}

func TestStructuralOutliers_StarGraph(t *testing.T) {
	g := buildStarGraph(0, 4)
	outliers := StructuralOutliers(g, 1)
	if len(outliers) != 1 {
		t.Fatalf("expected 1 outlier, got %d", len(outliers))
	}
	// The center or a leaf could be most anomalous depending on the
	// combination. Just verify we get a valid node.
	if outliers[0] < 0 || outliers[0] > 4 {
		t.Errorf("unexpected outlier ID: %d", outliers[0])
	}
}

func TestStructuralOutliers_KLargerThanGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))

	outliers := StructuralOutliers(g, 10)
	if len(outliers) != 2 {
		t.Errorf("expected 2 outliers (graph size), got %d", len(outliers))
	}
}

// TestIsolationScore_HubAndSpoke tests a graph with a clear structural
// anomaly: a hub node connecting two otherwise separate cliques.
func TestIsolationScore_HubAndSpoke(t *testing.T) {
	g := simple.NewUndirectedGraph()
	// Clique 1: nodes 0,1,2
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	// Clique 2: nodes 3,4,5
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))
	// Bridge node 6 connects both cliques.
	g.SetEdge(g.NewEdge(simple.Node(6), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(6), simple.Node(3)))

	scores := IsolationScore(g)
	// Node 6 bridges two cliques and should have a notable score.
	if scores[6] == 0 {
		t.Error("bridge node should have non-zero isolation score")
	}
}
