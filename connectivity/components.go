// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
)

// ConnectedComponents returns the weakly connected components of an undirected
// graph as slices of node IDs.
//
// Wraps gonum/graph/topo.ConnectedComponents.
func ConnectedComponents(ctx context.Context, g graph.Undirected) [][]int64 {
	raw := topo.ConnectedComponents(g)
	result := make([][]int64, len(raw))
	for i, comp := range raw {
		ids := make([]int64, len(comp))
		for j, n := range comp {
			ids[j] = n.ID()
		}
		result[i] = ids
	}
	return result
}

// StronglyConnectedComponents returns the strongly connected components of a
// directed graph as slices of node IDs.
//
// Wraps gonum/graph/topo.TarjanSCC.
func StronglyConnectedComponents(ctx context.Context, g graph.Directed) [][]int64 {
	raw := topo.TarjanSCC(g)
	result := make([][]int64, len(raw))
	for i, comp := range raw {
		ids := make([]int64, len(comp))
		for j, n := range comp {
			ids[j] = n.ID()
		}
		result[i] = ids
	}
	return result
}
