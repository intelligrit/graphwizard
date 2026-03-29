// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package subgraph

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// FilterNodes extracts an induced subgraph containing only nodes for which
// the keep predicate returns true. All edges between kept nodes are preserved.
func FilterNodes(g graph.Undirected, keep func(int64) bool) *simple.UndirectedGraph {
	result := simple.NewUndirectedGraph()

	// Collect kept node IDs.
	kept := make(map[int64]bool)
	nodes := g.Nodes()
	for nodes.Next() {
		id := nodes.Node().ID()
		if keep(id) {
			kept[id] = true
			result.AddNode(simple.Node(id))
		}
	}

	// Add edges between kept nodes.
	for id := range kept {
		it := g.From(id)
		for it.Next() {
			nid := it.Node().ID()
			if kept[nid] && id < nid {
				result.SetEdge(result.NewEdge(simple.Node(id), simple.Node(nid)))
			}
		}
	}

	return result
}
