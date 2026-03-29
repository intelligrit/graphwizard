// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"math"
	"math/rand"

	"gonum.org/v1/gonum/graph/simple"
)

// ErdosRenyi generates a random undirected graph with n nodes where each
// possible edge exists independently with probability p.
//
// Reference: P. Erdos and A. Renyi, "On Random Graphs", 1959.
func ErdosRenyi(n int, p float64, rng *rand.Rand) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < int64(n); i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < int64(n); i++ {
		for j := i + 1; j < int64(n); j++ {
			if rng.Float64() < p {
				g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
			}
		}
	}
	return g
}

// BarabasiAlbert generates a scale-free undirected graph using preferential
// attachment. Starts with m0 fully connected nodes, then adds nodes one at a
// time, each connecting to m existing nodes with probability proportional to
// their current degree.
//
// This produces power-law degree distributions similar to real-world social
// and provider networks.
//
// Reference: A. Barabasi and R. Albert, "Emergence of Scaling in Random
// Networks", Science, 1999.
func BarabasiAlbert(n, m int, rng *rand.Rand) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	if n <= 0 || m <= 0 {
		return g
	}

	m0 := m + 1
	if m0 > n {
		m0 = n
	}

	// Start with m0 fully connected nodes.
	for i := int64(0); i < int64(m0); i++ {
		for j := i + 1; j < int64(m0); j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	// Track degrees for preferential attachment.
	degree := make([]int, n)
	totalDegree := 0
	for i := 0; i < m0; i++ {
		degree[i] = m0 - 1
		totalDegree += m0 - 1
	}

	// Add remaining nodes.
	for i := m0; i < n; i++ {
		g.AddNode(simple.Node(int64(i)))
		targets := make(map[int]bool)

		for len(targets) < m && len(targets) < i {
			// Select target with probability proportional to degree.
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
			g.SetEdge(g.NewEdge(simple.Node(int64(i)), simple.Node(int64(t))))
			degree[i]++
			degree[t]++
			totalDegree += 2
		}
	}

	return g
}

// WeightedErdosRenyi generates a random weighted undirected graph.
func WeightedErdosRenyi(n int, p float64, maxWeight float64, rng *rand.Rand) *simple.WeightedUndirectedGraph {
	g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
	for i := int64(0); i < int64(n); i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < int64(n); i++ {
		for j := i + 1; j < int64(n); j++ {
			if rng.Float64() < p {
				w := rng.Float64() * maxWeight
				g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(i), simple.Node(j), w))
			}
		}
	}
	return g
}

// TwoClusterGraph generates an undirected graph with two dense clusters
// connected by a sparse bridge. Each cluster is Erdos-Renyi with intra-
// probability pIn, and inter-cluster edges have probability pOut.
func TwoClusterGraph(clusterSize int, pIn, pOut float64, rng *rand.Rand) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	n := int64(clusterSize)

	// Cluster A: nodes 0..n-1
	for i := int64(0); i < n; i++ {
		for j := i + 1; j < n; j++ {
			if rng.Float64() < pIn {
				g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
			}
		}
	}

	// Cluster B: nodes n..2n-1
	for i := n; i < 2*n; i++ {
		for j := i + 1; j < 2*n; j++ {
			if rng.Float64() < pIn {
				g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
			}
		}
	}

	// Inter-cluster edges.
	for i := int64(0); i < n; i++ {
		for j := n; j < 2*n; j++ {
			if rng.Float64() < pOut {
				g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
			}
		}
	}

	return g
}
