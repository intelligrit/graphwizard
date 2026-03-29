// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestSpectralClustering_TwoClusters(t *testing.T) {
	g := simple.NewUndirectedGraph()
	// Two disconnected triangles — unambiguous 2-cluster structure.
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))

	clusters := SpectralClustering(g, 2)
	if len(clusters) != 6 {
		t.Fatalf("expected 6 assignments, got %d", len(clusters))
	}

	// Nodes in the same triangle must share a cluster.
	if clusters[0] != clusters[1] || clusters[1] != clusters[2] {
		t.Errorf("triangle A should be together: %v", clusters)
	}
	if clusters[3] != clusters[4] || clusters[4] != clusters[5] {
		t.Errorf("triangle B should be together: %v", clusters)
	}
	// Must produce 2 distinct labels.
	labels := make(map[int]bool)
	for _, c := range clusters {
		labels[c] = true
	}
	if len(labels) != 2 {
		t.Errorf("expected 2 distinct clusters, got %d: %v", len(labels), clusters)
	}
}

func TestSpectralClustering_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	clusters := SpectralClustering(g, 2)
	if len(clusters) != 0 {
		t.Errorf("expected empty, got %v", clusters)
	}
}

func TestSpectralClustering_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	clusters := SpectralClustering(g, 1)
	if len(clusters) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(clusters))
	}
}
