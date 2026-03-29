// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
)

// KCore returns the k-core of an undirected graph: the maximal subgraph where
// every node has degree at least k. Returns the node IDs in the k-core.
//
// The k-core is useful for identifying densely connected subgraphs and
// filtering out peripheral nodes.
//
// Wraps gonum/graph/topo.KCore.
func KCore(k int, g graph.Undirected) []int64 {
	core := topo.KCore(k, g)
	ids := make([]int64, len(core))
	for i, n := range core {
		ids[i] = n.ID()
	}
	return ids
}

// DegeneracyOrdering returns the degeneracy ordering of an undirected graph.
//
// Returns:
//   - order: node IDs in degeneracy order (least connected first)
//   - coreLayers: each element i is the set of node IDs in core layer i
//     (the k-core for k = index). The degeneracy of the graph equals
//     len(coreLayers) - 1.
//
// Wraps gonum/graph/topo.DegeneracyOrdering.
func DegeneracyOrdering(g graph.Undirected) (order []int64, coreLayers [][]int64) {
	rawOrder, rawCores := topo.DegeneracyOrdering(g)
	order = make([]int64, len(rawOrder))
	for i, n := range rawOrder {
		order[i] = n.ID()
	}
	coreLayers = make([][]int64, len(rawCores))
	for i, layer := range rawCores {
		ids := make([]int64, len(layer))
		for j, n := range layer {
			ids[j] = n.ID()
		}
		coreLayers[i] = ids
	}
	return order, coreLayers
}
