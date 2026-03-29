// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package embedding

import (
	"math/rand"

	"gonum.org/v1/gonum/graph"
)

// WalkParams configures random walk generation.
type WalkParams struct {
	WalkLength   int     // Length of each random walk.
	WalksPerNode int     // Number of walks starting from each node.
	P            float64 // Return parameter (1.0 = no bias toward returning).
	Q            float64 // In-out parameter (1.0 = no bias; <1 = BFS-like; >1 = DFS-like).
}

// Node2VecWalks generates biased random walks from every node in an undirected
// graph using the Node2Vec strategy.
//
// The return parameter p controls the likelihood of revisiting the previous
// node (low p = high return probability). The in-out parameter q controls
// exploration: low q favors BFS-like local exploration, high q favors
// DFS-like outward exploration.
//
// Reference: A. Grover and J. Leskovec, "node2vec: Scalable Feature Learning
// for Networks", KDD 2016.
func Node2VecWalks(g graph.Undirected, params WalkParams, rng *rand.Rand) [][]int64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}

	// Build adjacency lists.
	adj := make(map[int64][]int64)
	adjSet := make(map[int64]map[int64]bool)
	for _, id := range ids {
		var neighbors []int64
		neighborSet := make(map[int64]bool)
		it := g.From(id)
		for it.Next() {
			nid := it.Node().ID()
			neighbors = append(neighbors, nid)
			neighborSet[nid] = true
		}
		adj[id] = neighbors
		adjSet[id] = neighborSet
	}

	var walks [][]int64
	for w := 0; w < params.WalksPerNode; w++ {
		perm := rng.Perm(len(ids))
		for _, idx := range perm {
			start := ids[idx]
			walk := make([]int64, 0, params.WalkLength)
			walk = append(walk, start)

			if len(adj[start]) == 0 {
				walks = append(walks, walk)
				continue
			}

			// First step: uniform random neighbor.
			cur := adj[start][rng.Intn(len(adj[start]))]
			walk = append(walk, cur)
			prev := start

			for step := 2; step < params.WalkLength; step++ {
				neighbors := adj[cur]
				if len(neighbors) == 0 {
					break
				}

				// Compute unnormalized transition weights.
				weights := make([]float64, len(neighbors))
				total := 0.0
				for i, next := range neighbors {
					if next == prev {
						weights[i] = 1.0 / params.P
					} else if adjSet[prev][next] {
						weights[i] = 1.0
					} else {
						weights[i] = 1.0 / params.Q
					}
					total += weights[i]
				}

				// Sample proportionally.
				r := rng.Float64() * total
				cumulative := 0.0
				chosen := neighbors[len(neighbors)-1]
				for i, w := range weights {
					cumulative += w
					if r <= cumulative {
						chosen = neighbors[i]
						break
					}
				}

				walk = append(walk, chosen)
				prev = cur
				cur = chosen
			}

			walks = append(walks, walk)
		}
	}

	return walks
}

// DeepWalkWalks generates uniform random walks from every node. This is
// equivalent to Node2Vec with p=1, q=1.
//
// Reference: B. Perozzi, R. Al-Rfou, S. Skiena, "DeepWalk: Online Learning
// of Social Representations", KDD 2014.
func DeepWalkWalks(g graph.Undirected, walkLength, walksPerNode int, rng *rand.Rand) [][]int64 {
	return Node2VecWalks(g, WalkParams{
		WalkLength:   walkLength,
		WalksPerNode: walksPerNode,
		P:            1.0,
		Q:            1.0,
	}, rng)
}
