// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package subgraph_test

import (
	"fmt"
	"sort"

	"github.com/intelligrit/graphwizard/subgraph"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleNHopNeighborhoodUndirected() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	sub := subgraph.NHopNeighborhoodUndirected(g, 0, 1)
	var ids []int64
	nodes := sub.Nodes()
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	fmt.Println(ids)
	// Output: [0 1]
}

func ExampleFilterNodes() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	sub := subgraph.FilterNodes(g, func(id int64) bool { return id <= 1 })
	var ids []int64
	nodes := sub.Nodes()
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	fmt.Println(ids)
	// Output: [0 1]
}
