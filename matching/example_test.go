// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package matching_test

import (
	"fmt"

	"github.com/intelligrit/graphwizard/matching"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleHopcroftKarp() {
	// Bipartite: left={0,1,2}, right={3,4,5}.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(4)))

	m, size := matching.HopcroftKarp(g, []int64{0, 1, 2})
	fmt.Printf("matching size: %d\n", size)
	_ = m
	// Output: matching size: 3
}
