// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community_test

import (
	"fmt"
	"math/rand"

	"github.com/intelligrit/graphwizard/community"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleLeiden() {
	g := simple.NewUndirectedGraph()
	// Cluster A: triangle 0-1-2.
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	// Cluster B: triangle 3-4-5.
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))
	// Bridge.
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	rng := rand.New(rand.NewSource(42))
	comms := community.Leiden(g, 1.0, rng)
	fmt.Printf("same cluster: %v\n", comms[0] == comms[1])
	fmt.Printf("different clusters: %v\n", comms[0] != comms[3])
	// Output:
	// same cluster: true
	// different clusters: true
}

func ExampleLabelPropagation() {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	rng := rand.New(rand.NewSource(42))
	labels := community.LabelPropagation(g, 100, rng)
	// Fully connected: all nodes should converge to one label.
	fmt.Printf("all same: %v\n", labels[0] == labels[1] && labels[1] == labels[2])
	// Output: all same: true
}

func ExampleLouvain() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	comms := community.Louvain(g, 1.0, nil)
	fmt.Printf("nodes: %d\n", len(comms))
	// Output: nodes: 3
}

func ExampleSpectralClustering() {
	g := simple.NewUndirectedGraph()
	// Two disconnected triangles.
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))

	clusters := community.SpectralClustering(g, 2)
	// Count distinct labels.
	labels := make(map[int]bool)
	for _, c := range clusters {
		labels[c] = true
	}
	fmt.Printf("clusters: %d\n", len(labels))
	// Output: clusters: 2
}

func ExampleLeidenParallel() {
	g := simple.NewUndirectedGraph()
	// Cluster A: triangle 0-1-2.
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	// Cluster B: triangle 3-4-5.
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))
	// Bridge.
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	rng := rand.New(rand.NewSource(42))
	comms := community.LeidenParallel(g, 1.0, rng)
	fmt.Printf("same cluster: %v\n", comms[0] == comms[1])
	fmt.Printf("different clusters: %v\n", comms[0] != comms[3])
	// Output:
	// same cluster: true
	// different clusters: true
}

func ExampleLouvainQ() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	q := community.LouvainQ(g, [][]graph.Node{
		{simple.Node(0), simple.Node(1), simple.Node(2)},
	}, 1.0)
	fmt.Printf("Q >= 0: %v\n", q >= 0)
	// Output: Q >= 0: true
}
