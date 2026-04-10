// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"
	"math"
	"sort"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
)

// MSTEdge represents an edge in a minimum spanning tree.
type MSTEdge struct {
	From, To int64
	Weight   float64
}

// MSTResult holds the minimum spanning tree edges and total weight.
type MSTResult struct {
	Edges  []MSTEdge
	Weight float64
}

// Kruskal returns the minimum spanning tree of a weighted undirected graph
// using Kruskal's algorithm with union-find.
//
// If the graph is disconnected, returns a minimum spanning forest (the MST
// of each connected component).
//
// Time complexity: O(E log E).
//
// Reference: J. Kruskal, "On the Shortest Spanning Subtree of a Graph and
// the Traveling Salesman Problem", Proceedings of the AMS, 1956.
func Kruskal(ctx context.Context, g graph.WeightedUndirected) MSTResult {
	// Collect all edges.
	var edges []MSTEdge
	seen := make(map[[2]int64]bool)
	nodes := g.Nodes()
	for nodes.Next() {
		u := nodes.Node()
		it := g.From(u.ID())
		for it.Next() {
			v := it.Node()
			key := [2]int64{u.ID(), v.ID()}
			rev := [2]int64{v.ID(), u.ID()}
			if seen[key] || seen[rev] {
				continue
			}
			seen[key] = true
			w, ok := g.Weight(u.ID(), v.ID())
			if !ok {
				w = math.Inf(1)
			}
			edges = append(edges, MSTEdge{From: u.ID(), To: v.ID(), Weight: w})
		}
	}

	// Sort by weight.
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Weight < edges[j].Weight
	})

	// Union-find.
	parent := make(map[int64]int64)
	rank := make(map[int64]int)
	find := func(x int64) int64 {
		if _, ok := parent[x]; !ok {
			parent[x] = x
		}
		for parent[x] != x {
			parent[x] = parent[parent[x]] // path halving
			x = parent[x]
		}
		return x
	}

	var result []MSTEdge
	totalWeight := 0.0

	for i, e := range edges {
		progress.Report(ctx, progress.Progress{Phase: "edges", Step: i, Total: len(edges)})
		ru := find(e.From)
		rv := find(e.To)
		if ru == rv {
			continue
		}
		// Union by rank.
		if rank[ru] < rank[rv] {
			ru, rv = rv, ru
		}
		parent[rv] = ru
		if rank[ru] == rank[rv] {
			rank[ru]++
		}
		result = append(result, e)
		totalWeight += e.Weight
	}

	return MSTResult{Edges: result, Weight: totalWeight}
}

// Prim returns the minimum spanning tree of a weighted undirected graph
// starting from the given source node using Prim's algorithm.
//
// Only returns the MST of the connected component containing the source.
//
// Time complexity: O(V² ) with adjacency scan (suitable for dense graphs).
//
// Reference: R. Prim, "Shortest Connection Networks and Some Generalizations",
// Bell System Technical Journal, 1957.
func Prim(ctx context.Context, g graph.WeightedUndirected, source int64) MSTResult {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return MSTResult{}
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	if _, ok := idx[source]; !ok {
		return MSTResult{}
	}

	inMST := make([]bool, n)
	key := make([]float64, n)   // minimum edge weight to reach node
	from := make([]int64, n)    // which MST node connects
	for i := range key {
		key[i] = math.Inf(1)
		from[i] = -1
	}
	key[idx[source]] = 0

	var result []MSTEdge
	totalWeight := 0.0

	for count := 0; count < n; count++ {
		progress.Report(ctx, progress.Progress{Phase: "nodes", Step: count, Total: n})
		// Pick minimum key node not in MST.
		u := -1
		minKey := math.Inf(1)
		for i := 0; i < n; i++ {
			if !inMST[i] && key[i] < minKey {
				u = i
				minKey = key[i]
			}
		}
		if u == -1 {
			break // remaining nodes unreachable
		}

		inMST[u] = true
		if from[u] != -1 {
			result = append(result, MSTEdge{From: from[u], To: ids[u], Weight: key[u]})
			totalWeight += key[u]
		}

		// Update neighbors.
		it := g.From(ids[u])
		for it.Next() {
			v := it.Node()
			vi, ok := idx[v.ID()]
			if !ok || inMST[vi] {
				continue
			}
			w, ok := g.Weight(ids[u], v.ID())
			if ok && w < key[vi] {
				key[vi] = w
				from[vi] = ids[u]
			}
		}
	}

	return MSTResult{Edges: result, Weight: totalWeight}
}
