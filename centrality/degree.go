// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"gonum.org/v1/gonum/graph"
)

// Degree returns the normalized degree centrality for each node in an
// undirected graph, keyed by node ID.
//
// Degree centrality is the fraction of other nodes each node is connected to:
//
//	C_D(v) = deg(v) / (n - 1)
//
// For graphs with fewer than 2 nodes, all centralities are 0.
func Degree(g graph.Undirected) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	result := make(map[int64]float64, n)

	if n < 2 {
		for _, id := range ids {
			result[id] = 0
		}
		return result
	}

	denom := float64(n - 1)
	for _, id := range ids {
		neighbors := g.From(id)
		deg := 0
		for neighbors.Next() {
			deg++
		}
		result[id] = float64(deg) / denom
	}
	return result
}

// InDegree returns the normalized in-degree centrality for each node in a
// directed graph.
//
//	C_in(v) = in_deg(v) / (n - 1)
func InDegree(g graph.Directed) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	result := make(map[int64]float64, n)

	if n < 2 {
		for _, id := range ids {
			result[id] = 0
		}
		return result
	}

	denom := float64(n - 1)
	for _, id := range ids {
		to := g.To(id)
		deg := 0
		for to.Next() {
			deg++
		}
		result[id] = float64(deg) / denom
	}
	return result
}

// OutDegree returns the normalized out-degree centrality for each node in a
// directed graph.
//
//	C_out(v) = out_deg(v) / (n - 1)
func OutDegree(g graph.Directed) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	result := make(map[int64]float64, n)

	if n < 2 {
		for _, id := range ids {
			result[id] = 0
		}
		return result
	}

	denom := float64(n - 1)
	for _, id := range ids {
		from := g.From(id)
		deg := 0
		for from.Next() {
			deg++
		}
		result[id] = float64(deg) / denom
	}
	return result
}
