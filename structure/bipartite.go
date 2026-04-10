// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// BipartiteProject projects a bipartite undirected graph onto one partition.
// Two nodes in the specified partition are connected in the projection if they
// share at least one neighbor in the other partition.
//
// The edge weight in the returned graph equals the number of shared neighbors
// (co-occurrence count). For example, if providers A and B both affiliate with
// organizations X and Y, the projected edge A-B has weight 2.
//
// The partition parameter specifies the node IDs to project onto. All other
// nodes are treated as the "bridge" partition.
//
// Returns a new weighted undirected graph containing only the projected nodes.
func BipartiteProject(ctx context.Context, g graph.Undirected, partition []int64) *simple.WeightedUndirectedGraph {
	proj := simple.NewWeightedUndirectedGraph(0, 0)
	partSet := make(map[int64]bool, len(partition))
	for _, id := range partition {
		partSet[id] = true
		proj.AddNode(simple.Node(id))
	}

	// For each bridge node (not in partition), connect all its partition
	// neighbors to each other.
	weights := make(map[[2]int64]float64)
	nodes := g.Nodes()
	for nodes.Next() {
		bridge := nodes.Node().ID()
		if partSet[bridge] {
			continue
		}
		// Collect partition neighbors of this bridge node.
		var pNeighbors []int64
		it := g.From(bridge)
		for it.Next() {
			nid := it.Node().ID()
			if partSet[nid] {
				pNeighbors = append(pNeighbors, nid)
			}
		}
		// All pairs of partition neighbors share this bridge node.
		for i := 0; i < len(pNeighbors); i++ {
			for j := i + 1; j < len(pNeighbors); j++ {
				a, b := pNeighbors[i], pNeighbors[j]
				if a > b {
					a, b = b, a
				}
				weights[[2]int64{a, b}]++
			}
		}
	}

	for pair, w := range weights {
		proj.SetWeightedEdge(proj.NewWeightedEdge(
			simple.Node(pair[0]),
			simple.Node(pair[1]),
			w,
		))
	}

	return proj
}
