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

func ExamplePrim() {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 3))

	mst := structure.Prim(g, 0)
	fmt.Printf("MST edges: %d, weight: %.0f\n", len(mst.Edges), mst.Weight)
	// Output: MST edges: 2, weight: 3
}

func ExampleBipartiteProject() {
	// Providers 0,1 share organizations 10,11.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(10)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(11)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(10)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(11)))

	proj := structure.BipartiteProject(g, []int64{0, 1})
	w, _ := proj.Weight(0, 1)
	fmt.Printf("co-affiliations: %.0f\n", w)
	// Output: co-affiliations: 2
}

func ExampleGraphColoring() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	colors, k, _ := structure.GraphColoring(g)
	fmt.Printf("chromatic number: %d, colors: %d\n", k, len(colors))
	// Output: chromatic number: 2, colors: 3
}

func ExampleTSP() {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(0), 1))

	result := structure.TSP(g)
	fmt.Printf("tour weight: %.0f\n", result.Weight)
	// Output: tour weight: 3
}

func ExampleAverageClusteringCoefficient() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	avg := structure.AverageClusteringCoefficient(g)
	fmt.Printf("avg CC: %.1f\n", avg)
	// Output: avg CC: 1.0
}

func ExampleTriangleCountParallel() {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	_, total := structure.TriangleCountParallel(g)
	fmt.Printf("triangles: %d\n", total)
	// Output: triangles: 4
}

func ExampleClusteringCoefficientParallel() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	coeffs := structure.ClusteringCoefficientParallel(g)
	fmt.Printf("CC(0) = %.1f\n", coeffs[0])
	// Output: CC(0) = 1.0
}
