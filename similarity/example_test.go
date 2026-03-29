// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity_test

import (
	"fmt"
	"math"

	"github.com/intelligrit/graphwizard/similarity"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleJaccard() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))

	score := similarity.Jaccard(g, 0, 1)
	fmt.Printf("J(0,1) = %.4f\n", score)
	// Output: J(0,1) = 0.3333
}

func ExampleOverlap() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	score := similarity.Overlap(g, 0, 1)
	fmt.Printf("O(0,1) = %.1f\n", score)
	// Output: O(0,1) = 1.0
}

func ExampleCosine() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))

	c := similarity.Cosine(g, 0, 1)
	fmt.Printf("Cosine(0,1) = %.1f\n", c)
	// Output: Cosine(0,1) = 1.0
}

func ExampleAdamicAdar() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	aa := similarity.AdamicAdar(g, 0, 1)
	fmt.Printf("AA(0,1) = %.4f\n", aa)
	// Output: AA(0,1) = 0.9102
}

func ExampleCommonNeighbors() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))

	cn := similarity.CommonNeighbors(g, 0, 1)
	fmt.Printf("CN(0,1) = %d\n", cn)
	// Output: CN(0,1) = 2
}

func ExamplePreferentialAttachment() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))

	pa := similarity.PreferentialAttachment(g, 0, 1)
	fmt.Printf("PA(0,1) = %d\n", pa)
	// Output: PA(0,1) = 2
}

func ExamplePredictLinks() {
	// Triangle with one missing edge: 0-2 is the clear prediction.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	scorer := func(g graph.Undirected, u, v int64) float64 {
		return float64(similarity.CommonNeighbors(g, u, v))
	}
	preds := similarity.PredictLinks(g, 1, scorer)
	fmt.Printf("predicted: %d-%d (CN=%d)\n", preds[0].A, preds[0].B, int(preds[0].Score))
	// Output: predicted: 0-2 (CN=1)
}

func ExampleSimRank() {
	// Nodes 0 and 1 both receive edges from the same sources (2 and 3).
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(1)))

	sim := similarity.SimRank(g, 0.8, 10)
	// Nodes 0 and 1 have the same in-neighbors, so they are structurally similar.
	s := sim[[2]int64{0, 1}]
	fmt.Printf("sim(0,1) = %.2f\n", s)
	// Output: sim(0,1) = 0.40
}

// Ensure unused imports.
var _ = math.Inf
