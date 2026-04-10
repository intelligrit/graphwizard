// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
)

// MaximalCliques returns all maximal cliques in an undirected graph using the
// Bron-Kerbosch algorithm. Each clique is a slice of node IDs.
//
// Wraps gonum/graph/topo.BronKerbosch.
func MaximalCliques(ctx context.Context, g graph.Undirected) [][]int64 {
	raw := topo.BronKerbosch(g)
	result := make([][]int64, len(raw))
	for i, clique := range raw {
		ids := make([]int64, len(clique))
		for j, n := range clique {
			ids[j] = n.ID()
		}
		result[i] = ids
	}
	return result
}
