// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diff

import (
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

// erdosRenyi generates a random undirected graph. Inlined to avoid circular
// imports with benchmark/.
func erdosRenyi(n int, p float64, rng *rand.Rand) *simple.UndirectedGraph {
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

// barabasiAlbert generates a scale-free undirected graph. Inlined to avoid
// circular imports with benchmark/.
func barabasiAlbert(n, m int, rng *rand.Rand) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	if n <= 0 || m <= 0 {
		return g
	}

	m0 := m + 1
	if m0 > n {
		m0 = n
	}

	for i := int64(0); i < int64(m0); i++ {
		for j := i + 1; j < int64(m0); j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	degree := make([]int, n)
	totalDegree := 0
	for i := 0; i < m0; i++ {
		degree[i] = m0 - 1
		totalDegree += m0 - 1
	}

	for i := m0; i < n; i++ {
		g.AddNode(simple.Node(int64(i)))
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
			g.SetEdge(g.NewEdge(simple.Node(int64(i)), simple.Node(int64(t))))
			degree[i]++
			degree[t]++
			totalDegree += 2
		}
	}

	return g
}

func BenchmarkCompare_1K(b *testing.B) {
	rng1 := rand.New(rand.NewSource(42))
	rng2 := rand.New(rand.NewSource(99))
	g1 := erdosRenyi(1000, 0.01, rng1)
	g2 := erdosRenyi(1000, 0.01, rng2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compare(g1, g2)
	}
}

func BenchmarkCompare_10K(b *testing.B) {
	rng1 := rand.New(rand.NewSource(42))
	rng2 := rand.New(rand.NewSource(99))
	g1 := barabasiAlbert(10000, 3, rng1)
	g2 := barabasiAlbert(10000, 3, rng2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compare(g1, g2)
	}
}
