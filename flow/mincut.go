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
	Weight     float64 // Total weight of cut edges.
}

// MinCut computes the minimum s-t cut in a weighted directed graph using the
// max-flow min-cut theorem: run max flow, then find reachable nodes from
// source in the residual graph.
//
// The cut weight equals the max flow value. The cut edges are those from
// SourceSide to TargetSide in the original graph.
//
// Reference: Ford and Fulkerson, "Maximal Flow Through a Network",
// Canadian Journal of Mathematics, 1956.
func MinCut(g graph.WeightedDirected, source, target int64, eps float64) MinCutResult {
	maxFlow := MaxFlow(g, source, target, eps)

	// Build residual graph: edge has residual capacity if flow < capacity.
	// After max flow, BFS from source on edges with remaining capacity.
	// Since we don't have access to the flow assignment from gonum's Dinic,
	// we reconstruct: an edge (u,v) is saturated if removing it and re-running
	// would decrease flow. Instead, use the simpler approach: BFS from source
	// using only edges where we can push more flow (approximate via weight check).
	//
	// Practical approach: compute max flow, then find source-side via BFS on
	// the original graph, but skip edges that are "bottleneck" saturated.
	// For exact min-cut, we build a simple residual using flow decomposition.

	// Simpler exact approach: after max flow, the source side is the set of
	// nodes reachable from source when we only traverse edges with residual > 0.
	// We approximate this by: for each edge (u,v), compute max flow from u to v
	// directly... that's too expensive.
	//
	// Most practical: BFS from source, only follow edges where individual
	// max-flow(u,v) < weight(u,v). But that's also expensive.
	//
	// Correct approach: we manually implement a max-flow that exposes residuals.
	// For now, use the standard BFS trick on a capacity-based residual.

	residual := buildResidual(g, source, target)
	reachable := bfsReachable(residual, source)

	var sourceSide, targetSide []int64
	nodes := g.Nodes()
	for nodes.Next() {
		id := nodes.Node().ID()
		if reachable[id] {
			sourceSide = append(sourceSide, id)
		} else {
			targetSide = append(targetSide, id)
		}
	}

	return MinCutResult{
		SourceSide: sourceSide,
		TargetSide: targetSide,
		Weight:     maxFlow,
	}
}

// residualGraph implements a simple residual graph for BFS reachability.
type residualGraph struct {
	nodes    map[int64]bool
	adj      map[int64][]int64
	capacity map[[2]int64]float64
}

func buildResidual(g graph.WeightedDirected, source, target int64) *residualGraph {
	rg := &residualGraph{
		nodes:    make(map[int64]bool),
		adj:      make(map[int64][]int64),
		capacity: make(map[[2]int64]float64),
	}

	// Collect all nodes and edges.
	nodes := g.Nodes()
	for nodes.Next() {
		rg.nodes[nodes.Node().ID()] = true
	}

	// Build adjacency with capacities.
	for id := range rg.nodes {
		it := g.From(id)
		for it.Next() {
			to := it.Node().ID()
			w, ok := g.Weight(id, to)
			if !ok {
				continue
			}
			rg.adj[id] = append(rg.adj[id], to)
			rg.capacity[[2]int64{id, to}] = w
		}
	}

	// Run Edmonds-Karp (BFS-based Ford-Fulkerson) to compute flow and residual.
	flow := make(map[[2]int64]float64)
	// Add reverse edges for residual.
	for id := range rg.nodes {
		for _, to := range rg.adj[id] {
			if rg.capacity[[2]int64{to, id}] == 0 {
				// Add reverse edge with 0 capacity if not present.
				rg.adj[to] = append(rg.adj[to], id)
			}
		}
	}

	for {
		// BFS to find augmenting path.
		parent := map[int64]int64{source: -1}
		queue := []int64{source}
		found := false
		for len(queue) > 0 && !found {
			u := queue[0]
			queue = queue[1:]
			for _, v := range rg.adj[u] {
				if _, seen := parent[v]; seen {
					continue
				}
				residCap := rg.capacity[[2]int64{u, v}] - flow[[2]int64{u, v}] + flow[[2]int64{v, u}]
				if residCap > 1e-10 {
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
		bottleneck := 1e18
		for v := target; v != source; v = parent[v] {
			u := parent[v]
			residCap := rg.capacity[[2]int64{u, v}] - flow[[2]int64{u, v}] + flow[[2]int64{v, u}]
			if residCap < bottleneck {
				bottleneck = residCap
			}
		}

		// Update flow.
		for v := target; v != source; v = parent[v] {
			u := parent[v]
			flow[[2]int64{u, v}] += bottleneck
		}
	}

	// Build final residual for BFS reachability.
	finalAdj := make(map[int64][]int64)
	for id := range rg.nodes {
		for _, to := range rg.adj[id] {
			residCap := rg.capacity[[2]int64{id, to}] - flow[[2]int64{id, to}] + flow[[2]int64{to, id}]
			if residCap > 1e-10 {
				finalAdj[id] = append(finalAdj[id], to)
			}
		}
	}
	rg.adj = finalAdj
	return rg
}

func bfsReachable(rg *residualGraph, source int64) map[int64]bool {
	reachable := map[int64]bool{source: true}
	queue := []int64{source}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range rg.adj[u] {
			if !reachable[v] {
				reachable[v] = true
				queue = append(queue, v)
			}
		}
	}
	return reachable
}

