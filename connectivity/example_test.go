// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity_test

import (
	"fmt"

	"github.com/intelligrit/graphwizard/connectivity"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleBridges() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3))) // bridge

	bridges := connectivity.Bridges(g)
	fmt.Printf("bridges: %d\n", len(bridges))
	// Output: bridges: 1
}

func ExampleConnectedComponents() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	comps := connectivity.ConnectedComponents(g)
	fmt.Printf("components: %d\n", len(comps))
	// Output: components: 2
}

func ExampleCondensation() {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	comps, edges, _ := connectivity.Condensation(g)
	fmt.Printf("SCCs: %d, DAG edges: %d\n", len(comps), len(edges))
	// Output: SCCs: 2, DAG edges: 1
}

func ExampleKCore() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3))) // pendant

	core := connectivity.KCore(2, g)
	fmt.Printf("2-core size: %d\n", len(core))
	// Output: 2-core size: 3
}

func ExampleUnionFind() {
	uf := connectivity.NewUnionFind()
	uf.Union(1, 2)
	uf.Union(2, 3)

	fmt.Printf("1-3 connected: %v\n", uf.Connected(1, 3))
	fmt.Printf("1-4 connected: %v\n", uf.Connected(1, 4))
	// Output:
	// 1-3 connected: true
	// 1-4 connected: false
}

func ExampleArticulationPoints() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	aps := connectivity.ArticulationPoints(g)
	fmt.Printf("cut vertices: %d\n", len(aps))
	// Output: cut vertices: 1
}
