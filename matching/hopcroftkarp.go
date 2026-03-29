// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package matching

import (
	"math"

	"gonum.org/v1/gonum/graph"
)

const inf = math.MaxInt64

// Matching represents a maximum bipartite matching as a map from left-side
// node IDs to their matched right-side node IDs.
type Matching map[int64]int64

// HopcroftKarp finds a maximum cardinality matching in a bipartite undirected
// graph. The caller must provide the left partition node IDs; all other nodes
// are assumed to be in the right partition.
//
// Returns a Matching from left to right node IDs and the matching size.
//
// The algorithm runs in O(E * sqrt(V)) time.
//
// Reference: J. Hopcroft and R. Karp, "An n^(5/2) Algorithm for Maximum
// Matchings in Bipartite Graphs", SIAM Journal on Computing, 1973.
func HopcroftKarp(g graph.Undirected, left []int64) (Matching, int) {
	leftSet := make(map[int64]bool, len(left))
	for _, id := range left {
		leftSet[id] = true
	}

	// matchL[u] = matched right node for left node u (or -1)
	// matchR[v] = matched left node for right node v (or -1)
	matchL := make(map[int64]int64)
	matchR := make(map[int64]int64)
	for _, u := range left {
		matchL[u] = -1
	}
	// Initialize right side.
	nodes := g.Nodes()
	for nodes.Next() {
		id := nodes.Node().ID()
		if !leftSet[id] {
			matchR[id] = -1
		}
	}

	dist := make(map[int64]int)
	size := 0

	for bfs(g, left, matchL, matchR, dist) {
		for _, u := range left {
			if matchL[u] == -1 {
				if dfs(g, u, leftSet, matchL, matchR, dist) {
					size++
				}
			}
		}
	}

	result := make(Matching)
	for u, v := range matchL {
		if v != -1 {
			result[u] = v
		}
	}
	return result, size
}

func bfs(g graph.Undirected, left []int64, matchL, matchR map[int64]int64, dist map[int64]int) bool {
	var queue []int64
	for _, u := range left {
		if matchL[u] == -1 {
			dist[u] = 0
			queue = append(queue, u)
		} else {
			dist[u] = inf
		}
	}

	found := false
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]

		neighbors := g.From(u)
		for neighbors.Next() {
			v := neighbors.Node().ID()
			w := matchR[v]
			if w == -1 {
				found = true
			} else if dist[w] == inf {
				dist[w] = dist[u] + 1
				queue = append(queue, w)
			}
		}
	}

	return found
}

func dfs(g graph.Undirected, u int64, leftSet map[int64]bool, matchL, matchR map[int64]int64, dist map[int64]int) bool {
	neighbors := g.From(u)
	for neighbors.Next() {
		v := neighbors.Node().ID()
		if leftSet[v] {
			continue
		}
		w := matchR[v]
		if w == -1 || (dist[w] == dist[u]+1 && dfs(g, w, leftSet, matchL, matchR, dist)) {
			matchL[u] = v
			matchR[v] = u
			return true
		}
	}
	dist[u] = inf
	return false
}
