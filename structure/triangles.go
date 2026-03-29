// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"gonum.org/v1/gonum/graph"
)

// TriangleCount returns the number of triangles each node participates in,
// keyed by node ID, and the total number of triangles in the graph.
//
// A triangle is a set of three mutually connected nodes. Each triangle is
// counted once in the total but contributes +1 to each of its three nodes.
//
// Time complexity: O(V * E) in the worst case.
func TriangleCount(g graph.Undirected) (perNode map[int64]int, total int) {
	perNode = make(map[int64]int)
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		id := nodes.Node().ID()
		ids = append(ids, id)
		perNode[id] = 0
	}

	// Build neighbor sets for fast lookup.
	neighborSets := make(map[int64]map[int64]struct{})
	for _, id := range ids {
		s := make(map[int64]struct{})
		it := g.From(id)
		for it.Next() {
			s[it.Node().ID()] = struct{}{}
		}
		neighborSets[id] = s
	}

	// For each edge (u, v) where u < v, count common neighbors w > v.
	totalTriangles := 0
	for _, u := range ids {
		for v := range neighborSets[u] {
			if v <= u {
				continue
			}
			for w := range neighborSets[u] {
				if w <= v {
					continue
				}
				if _, ok := neighborSets[v][w]; ok {
					totalTriangles++
					perNode[u]++
					perNode[v]++
					perNode[w]++
				}
			}
		}
	}

	return perNode, totalTriangles
}
