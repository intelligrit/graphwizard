// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"math/rand/v2"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/community"
)

// Louvain performs community detection using the Louvain algorithm, returning
// a map from node ID to community ID.
//
// The resolution parameter controls community granularity: higher values
// produce more, smaller communities. Use 1.0 for standard modularity.
//
// Wraps gonum/graph/community.Modularize.
func Louvain(ctx context.Context, g graph.Graph, resolution float64, src rand.Source) map[int64]int64 {
	reduced := community.Modularize(g, resolution, src)
	return extractCommunities(reduced)
}

// LouvainQ returns the modularity Q score of a graph partitioned into the
// given communities at the given resolution.
//
// Wraps gonum/graph/community.Q.
func LouvainQ(ctx context.Context, g graph.Graph, communities [][]graph.Node, resolution float64) float64 {
	return community.Q(g, communities, resolution)
}

func extractCommunities(reduced community.ReducedGraph) map[int64]int64 {
	result := make(map[int64]int64)
	comms := reduced.Communities()
	for cID, members := range comms {
		for _, n := range members {
			result[n.ID()] = int64(cID)
		}
	}
	return result
}
