// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package flow

import (
	"gonum.org/v1/gonum/graph"
)

// MinCutResult holds the minimum s-t cut: the partition of nodes into the
// source side and target side, and the cut weight.
type MinCutResult struct {
	SourceSide []int64 // Nodes reachable from source in residual graph.
	TargetSide []int64 // Remaining nodes.
	Weight     float64 // Total weight of cut edges (equals max flow).
}

// MinCut computes the minimum s-t cut in a weighted directed graph.
//
// Uses Edmonds-Karp (BFS-based Ford-Fulkerson) to compute max flow, then
// finds reachable nodes from source in the residual graph. The cut weight
// equals the max flow value (max-flow min-cut theorem).
//
// Reference: Ford and Fulkerson, "Maximal Flow Through a Network",
// Canadian Journal of Mathematics, 1956.
func MinCut(g graph.WeightedDirected, source, target int64, eps float64) MinCutResult {
	// Collect nodes.
	allNodes := make(map[int64]bool)
	nodes := g.Nodes()
	for nodes.Next() {
		allNodes[nodes.Node().ID()] = true
	}

	// Build adjacency with capacities.
	adj := make(map[int64][]int64)
	cap := make(map[[2]int64]float64)

	for id := range allNodes {
		it := g.From(id)
		for it.Next() {
			to := it.Node().ID()
			w, ok := g.Weight(id, to)
			if !ok {
				continue
			}
			adj[id] = append(adj[id], to)
			cap[[2]int64{id, to}] = w
			// Ensure reverse edge exists in adjacency (for residual).
			if cap[[2]int64{to, id}] == 0 {
				if !hasEdge(adj[to], id) {
					adj[to] = append(adj[to], id)
				}
			}
		}
	}

	// Edmonds-Karp: BFS-based augmenting paths.
	flow := make(map[[2]int64]float64)
	totalFlow := 0.0

	for {
		// BFS to find augmenting path.
		parent := map[int64]int64{source: -1}
		queue := []int64{source}
		found := false
		for len(queue) > 0 && !found {
			u := queue[0]
			queue = queue[1:]
			for _, v := range adj[u] {
				if _, seen := parent[v]; seen {
					continue
				}
				residual := cap[[2]int64{u, v}] - flow[[2]int64{u, v}] + flow[[2]int64{v, u}]
				if residual > eps {
					parent[v] = u
					if v == target {
						found = true
						break
					}
					queue = append(queue, v)
				}
			}
		}
		if !found {
			break
		}

		// Find bottleneck.
		bottleneck := cap[[2]int64{parent[target], target}] // start large
		for v := target; v != source; v = parent[v] {
			u := parent[v]
			residual := cap[[2]int64{u, v}] - flow[[2]int64{u, v}] + flow[[2]int64{v, u}]
			if residual < bottleneck {
				bottleneck = residual
			}
		}

		// Update flow.
		for v := target; v != source; v = parent[v] {
			u := parent[v]
			flow[[2]int64{u, v}] += bottleneck
		}
		totalFlow += bottleneck
	}

	// Find source-side: BFS on residual graph from source.
	reachable := map[int64]bool{source: true}
	queue := []int64{source}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range adj[u] {
			if reachable[v] {
				continue
			}
			residual := cap[[2]int64{u, v}] - flow[[2]int64{u, v}] + flow[[2]int64{v, u}]
			if residual > eps {
				reachable[v] = true
				queue = append(queue, v)
			}
		}
	}

	var sourceSide, targetSide []int64
	for id := range allNodes {
		if reachable[id] {
			sourceSide = append(sourceSide, id)
		} else {
			targetSide = append(targetSide, id)
		}
	}

	return MinCutResult{
		SourceSide: sourceSide,
		TargetSide: targetSide,
		Weight:     totalFlow,
	}
}

func hasEdge(adj []int64, target int64) bool {
	for _, id := range adj {
		if id == target {
			return true
		}
	}
	return false
}
