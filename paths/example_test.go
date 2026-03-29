// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths_test

import (
	"fmt"
	"math"

	"github.com/intelligrit/graphwizard/paths"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleShortestPath() {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 4))

	nodes, weight := paths.ShortestPath(g, 0, 2)
	fmt.Printf("hops: %d, weight: %.0f\n", len(nodes)-1, weight)
	// Output: hops: 2, weight: 7
}

func ExampleYenKShortest() {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(3), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(2), simple.Node(3), 1))

	kpaths := paths.YenKShortest(g, 0, 3, 2)
	for i, p := range kpaths {
		fmt.Printf("path %d: weight=%.0f\n", i+1, p.Weight)
	}
	// Output:
	// path 1: weight=3
	// path 2: weight=4
}

func ExampleAStar() {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(2), 10))

	// Zero heuristic makes A* behave like Dijkstra.
	h := func(x, y graph.Node) float64 { return 0 }
	nodes, weight := paths.AStar(g, simple.Node(0), simple.Node(2), h)
	fmt.Printf("hops: %d, weight: %.0f\n", len(nodes)-1, weight)
	// Output: hops: 2, weight: 2
}

func ExampleAllShortestPaths() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	allPaths := paths.AllShortestPaths(g)
	p, _, unique := allPaths.Between(0, 2)
	fmt.Printf("path length: %d, unique: %v\n", len(p), unique)
	// Output: path length: 3, unique: true
}

func ExampleFloydWarshall() {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 3))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), 4))

	allPaths, ok := paths.FloydWarshall(g)
	_, w, _ := allPaths.Between(0, 2)
	fmt.Printf("weight: %.0f, no negative cycle: %v\n", w, ok)
	// Output: weight: 7, no negative cycle: true
}

func ExampleBellmanFord() {
	g := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(0), simple.Node(1), 2))
	g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(1), simple.Node(2), -1))

	_, weight, ok := paths.BellmanFord(g, 0, 2)
	fmt.Printf("weight: %.0f, ok: %v\n", weight, ok)
	// Output: weight: 1, ok: true
}
