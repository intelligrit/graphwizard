// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity_test

import (
	"fmt"

	"github.com/intelligrit/graphwizard/similarity"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleJaccard() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))

	score := similarity.Jaccard(g, 0, 1)
	fmt.Printf("J(0,1) = %.4f\n", score) // {2}∩ / {2,3,4}∪ = 1/3
	// Output: J(0,1) = 0.3333
}

func ExampleOverlap() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	score := similarity.Overlap(g, 0, 1)
	fmt.Printf("O(0,1) = %.1f\n", score) // {2} / min(2,1) = 1.0
	// Output: O(0,1) = 1.0
}
