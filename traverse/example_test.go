// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package traverse_test

import (
	"context"
	"fmt"

	"github.com/intelligrit/graphwizard/traverse"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleBFS() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	visited := traverse.BFS(context.Background(), g, 0)
	fmt.Printf("visited: %v\n", visited)
	// Output: visited: [0 1 2 3]
}

func ExampleBFSPath() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	path := traverse.BFSPath(context.Background(), g, 0, 3)
	fmt.Printf("hops: %d\n", len(path)-1)
	// Output: hops: 2
}

func ExampleDFS() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	visited := traverse.DFS(context.Background(), g, 0)
	fmt.Printf("nodes reached: %d\n", len(visited))
	// Output: nodes reached: 3
}
