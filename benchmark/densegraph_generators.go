// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"math/rand"

	"github.com/intelligrit/graphwizard/densegraph"
)

// DenseBarabasiAlbert generates a scale-free graph using densegraph.
func DenseBarabasiAlbert(n, m int, rng *rand.Rand) *densegraph.Undirected {
	b := densegraph.NewUndirectedBuilder()
	if n <= 0 || m <= 0 {
		return b.Build()
	}

	m0 := min(m+1, n)

	for i := int64(0); i < int64(m0); i++ {
		for j := i + 1; j < int64(m0); j++ {
			b.AddEdge(i, j)
		}
	}

	degree := make([]int, n)
	totalDegree := 0
	for i := range m0 {
		degree[i] = m0 - 1
		totalDegree += m0 - 1
	}

	for i := m0; i < n; i++ {
		b.AddNode(int64(i))
		targets := make(map[int]bool)
		for len(targets) < m && len(targets) < i {
			r := rng.Intn(totalDegree)
			cumulative := 0
			for j := 0; j < i; j++ {
				cumulative += degree[j]
				if r < cumulative {
					targets[j] = true
					break
				}
			}
		}
		for t := range targets {
			b.AddEdge(int64(i), int64(t))
			degree[i]++
			degree[t]++
			totalDegree += 2
		}
	}

	return b.Build()
}

// DenseTwoClusterGraph generates a two-cluster graph using densegraph.
func DenseTwoClusterGraph(clusterSize int, pIn, pOut float64, rng *rand.Rand) *densegraph.Undirected {
	b := densegraph.NewUndirectedBuilder()
	n := int64(clusterSize)

	for i := range n {
		for j := int64(i) + 1; j < n; j++ {
			if rng.Float64() < pIn {
				b.AddEdge(int64(i), j)
			}
		}
	}
	for i := range n {
		for j := n + int64(i) + 1; j < 2*n; j++ {
			if rng.Float64() < pIn {
				b.AddEdge(n+int64(i), j)
			}
		}
	}
	for i := range n {
		for j := n; j < 2*n; j++ {
			if rng.Float64() < pOut {
				b.AddEdge(int64(i), j)
			}
		}
	}

	return b.Build()
}
