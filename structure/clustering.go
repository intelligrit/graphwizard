// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"gonum.org/v1/gonum/graph"
)

// ClusteringCoefficient returns the local clustering coefficient for each node
// in an undirected graph, keyed by node ID.
//
// The local clustering coefficient of a node v measures the fraction of pairs
// of v's neighbors that are themselves connected. For a node with degree k:
//
//	C(v) = 2 * (edges between neighbors) / (k * (k - 1))
//
// Nodes with fewer than 2 neighbors have a clustering coefficient of 0.
//
// Reference: D. Watts and S. Strogatz, "Collective dynamics of 'small-world'
// networks", Nature, 1998.
func ClusteringCoefficient(g graph.Undirected) map[int64]float64 {
	result := make(map[int64]float64)

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		neighbors := neighborSet(g, n.ID())
		k := len(neighbors)
		if k < 2 {
			result[n.ID()] = 0
			continue
		}

		// Count edges between neighbors.
		edges := 0
		ids := make([]int64, 0, k)
		for id := range neighbors {
			ids = append(ids, id)
		}
		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				if g.HasEdgeBetween(ids[i], ids[j]) {
					edges++
				}
			}
		}

		result[n.ID()] = 2.0 * float64(edges) / float64(k*(k-1))
	}

	return result
}

// AverageClusteringCoefficient returns the mean of all local clustering
// coefficients in the graph.
func AverageClusteringCoefficient(g graph.Undirected) float64 {
	coeffs := ClusteringCoefficient(g)
	if len(coeffs) == 0 {
		return 0
	}
	var sum float64
	for _, c := range coeffs {
		sum += c
	}
	return sum / float64(len(coeffs))
}

func neighborSet(g graph.Undirected, id int64) map[int64]struct{} {
	s := make(map[int64]struct{})
	it := g.From(id)
	for it.Next() {
		s[it.Node().ID()] = struct{}{}
	}
	return s
}
