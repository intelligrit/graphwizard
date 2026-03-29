// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diff

import (
	"sort"

	"gonum.org/v1/gonum/graph"
)

// DiffResult holds the structural differences between two graphs.
type DiffResult struct {
	// AddedNodes contains IDs of nodes present in after but not before.
	AddedNodes []int64
	// RemovedNodes contains IDs of nodes present in before but not after.
	RemovedNodes []int64
	// AddedEdges contains edges present in after but not before. Each entry
	// is [from, to].
	AddedEdges [][2]int64
	// RemovedEdges contains edges present in before but not after. Each entry
	// is [from, to].
	RemovedEdges [][2]int64
}

// Compare computes the structural diff between before and after graphs.
// Both graphs must implement graph.Graph. For undirected graphs, each edge
// is reported once with the smaller ID first.
//
// All result slices are sorted for deterministic output.
func Compare(before, after graph.Graph) DiffResult {
	beforeNodes := collectNodes(before)
	afterNodes := collectNodes(after)
	beforeEdges := collectEdges(before)
	afterEdges := collectEdges(after)

	var result DiffResult

	// Find added and removed nodes.
	for id := range afterNodes {
		if !beforeNodes[id] {
			result.AddedNodes = append(result.AddedNodes, id)
		}
	}
	for id := range beforeNodes {
		if !afterNodes[id] {
			result.RemovedNodes = append(result.RemovedNodes, id)
		}
	}

	// Find added and removed edges.
	for e := range afterEdges {
		if !beforeEdges[e] {
			result.AddedEdges = append(result.AddedEdges, e)
		}
	}
	for e := range beforeEdges {
		if !afterEdges[e] {
			result.RemovedEdges = append(result.RemovedEdges, e)
		}
	}

	// Sort for deterministic output.
	sort.Slice(result.AddedNodes, func(i, j int) bool { return result.AddedNodes[i] < result.AddedNodes[j] })
	sort.Slice(result.RemovedNodes, func(i, j int) bool { return result.RemovedNodes[i] < result.RemovedNodes[j] })
	sortEdges(result.AddedEdges)
	sortEdges(result.RemovedEdges)

	return result
}

func collectNodes(g graph.Graph) map[int64]bool {
	m := make(map[int64]bool)
	nodes := g.Nodes()
	for nodes.Next() {
		m[nodes.Node().ID()] = true
	}
	return m
}

func collectEdges(g graph.Graph) map[[2]int64]bool {
	m := make(map[[2]int64]bool)
	nodes := g.Nodes()
	for nodes.Next() {
		u := nodes.Node().ID()
		it := g.From(u)
		for it.Next() {
			v := it.Node().ID()
			m[[2]int64{u, v}] = true
		}
	}
	return m
}

func sortEdges(edges [][2]int64) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i][0] != edges[j][0] {
			return edges[i][0] < edges[j][0]
		}
		return edges[i][1] < edges[j][1]
	})
}
