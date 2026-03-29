// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
)

// DirectedCycles returns all elementary cycles in a directed graph.
// Each cycle is a slice of node IDs.
//
// Wraps gonum/graph/topo.DirectedCyclesIn (Johnson's algorithm).
func DirectedCycles(g graph.Directed) [][]int64 {
	raw := topo.DirectedCyclesIn(g)
	result := make([][]int64, len(raw))
	for i, cycle := range raw {
		ids := make([]int64, len(cycle))
		for j, n := range cycle {
			ids[j] = n.ID()
		}
		result[i] = ids
	}
	return result
}

// UndirectedCycles returns a cycle basis for an undirected graph.
// Each cycle is a slice of node IDs.
//
// Wraps gonum/graph/topo.UndirectedCyclesIn (Paton's algorithm).
func UndirectedCycles(g graph.Undirected) [][]int64 {
	raw := topo.UndirectedCyclesIn(g)
	result := make([][]int64, len(raw))
	for i, cycle := range raw {
		ids := make([]int64, len(cycle))
		for j, n := range cycle {
			ids[j] = n.ID()
		}
		result[i] = ids
	}
	return result
}
