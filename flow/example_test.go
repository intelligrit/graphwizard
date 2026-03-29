// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package flow_test

import (
	"fmt"
	"math"

	"github.com/intelligrit/graphwizard/flow"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleMaxFlow() {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 3))

	f := flow.MaxFlow(g, 0, 3, 1e-9)
	fmt.Printf("max flow: %.0f\n", f)
	// Output: max flow: 4
}

func ExampleMinCut() {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))

	result := flow.MinCut(g, 0, 2, 1e-9)
	fmt.Printf("cut weight: %.0f, partitions: %d+%d\n",
		result.Weight, len(result.SourceSide), len(result.TargetSide))
	// Output: cut weight: 2, partitions: 2+1
}
