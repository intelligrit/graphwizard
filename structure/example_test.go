// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure_test

import (
	"fmt"
	"math"

	"github.com/intelligrit/graphwizard/structure"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleClusteringCoefficient() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	coeffs := structure.ClusteringCoefficient(g)
	fmt.Printf("C(0) = %.1f\n", coeffs[0])
	// Output: C(0) = 1.0
}

func ExampleTriangleCount() {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	_, total := structure.TriangleCount(g)
	fmt.Printf("triangles in K4: %d\n", total)
	// Output: triangles in K4: 4
}

func ExampleKruskal() {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 3))

	mst := structure.Kruskal(g)
	fmt.Printf("MST edges: %d, weight: %.0f\n", len(mst.Edges), mst.Weight)
	// Output: MST edges: 2, weight: 3
}

func ExampleSetCover() {
	universe := []int64{1, 2, 3, 4, 5}
	sets := [][]int64{
		{1, 2, 3},
		{2, 4},
		{3, 4, 5},
	}

	result := structure.SetCover(universe, sets)
	fmt.Printf("sets used: %d\n", len(result))
	// Output: sets used: 2
}

func ExampleMaximalCliques() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	cliques := structure.MaximalCliques(g)
	fmt.Printf("cliques: %d, size: %d\n", len(cliques), len(cliques[0]))
	// Output: cliques: 1, size: 3
}
