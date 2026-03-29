// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diff_test

import (
	"fmt"

	"github.com/intelligrit/graphwizard/diff"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleCompare() {
	before := simple.NewDirectedGraph()
	before.AddNode(simple.Node(0))
	before.AddNode(simple.Node(1))
	before.SetEdge(before.NewEdge(simple.Node(0), simple.Node(1)))

	after := simple.NewDirectedGraph()
	after.AddNode(simple.Node(0))
	after.AddNode(simple.Node(1))
	after.AddNode(simple.Node(2))
	after.SetEdge(after.NewEdge(simple.Node(0), simple.Node(1)))
	after.SetEdge(after.NewEdge(simple.Node(1), simple.Node(2)))

	d := diff.Compare(before, after)
	fmt.Printf("added nodes: %v\n", d.AddedNodes)
	fmt.Printf("added edges: %d\n", len(d.AddedEdges))
	// Output:
	// added nodes: [2]
	// added edges: 1
}
