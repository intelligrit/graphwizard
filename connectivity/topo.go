// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
)

// TopologicalSort returns a topological ordering of a directed acyclic graph
// (DAG). Nodes are returned in dependency order: if edge u->v exists, u
// appears before v.
//
// Returns the ordered node IDs and ok=true if the graph is a DAG. Returns
// nil and ok=false if the graph contains a cycle.
//
// Wraps gonum/graph/topo.Sort.
func TopologicalSort(g graph.Directed) (order []int64, ok bool) {
	sorted, err := topo.Sort(g)
	if err != nil {
		return nil, false
	}
	order = make([]int64, len(sorted))
	for i, n := range sorted {
		order[i] = n.ID()
	}
	return order, true
}
