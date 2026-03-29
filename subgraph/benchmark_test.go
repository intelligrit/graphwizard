// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package subgraph

import (
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

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

// toDirected converts an undirected graph to a directed graph with edges in
// both directions.
func toDirected(g *simple.UndirectedGraph) *simple.DirectedGraph {
	dg := simple.NewDirectedGraph()
	nodes := g.Nodes()
	for nodes.Next() {
		dg.AddNode(simple.Node(nodes.Node().ID()))
	}
	nodes = g.Nodes()
	for nodes.Next() {
		u := nodes.Node().ID()
		it := g.From(u)
		for it.Next() {
			v := it.Node().ID()
			dg.SetEdge(dg.NewEdge(simple.Node(u), simple.Node(v)))
		}
	}
	return dg
}

func BenchmarkNHopNeighborhood_1K_2hop(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	ug := barabasiAlbert(1000, 3, rng)
	dg := toDirected(ug)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NHopNeighborhood(dg, 0, 2)
	}
}

func BenchmarkNHopNeighborhood_10K_3hop(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	ug := barabasiAlbert(10000, 3, rng)
	dg := toDirected(ug)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NHopNeighborhood(dg, 0, 3)
	}
}

func BenchmarkFilterNodes_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := barabasiAlbert(10000, 3, rng)

	// Precompute degrees to find the median for filtering.
	degrees := make(map[int64]int)
	nodes := g.Nodes()
	for nodes.Next() {
		id := nodes.Node().ID()
		deg := 0
		it := g.From(id)
		for it.Next() {
			deg++
		}
		degrees[id] = deg
	}

	// Find median degree for the 50% filter.
	allDegs := make([]int, 0, len(degrees))
	for _, d := range degrees {
		allDegs = append(allDegs, d)
	}
	// Simple selection: sort and pick middle.
	sortInts(allDegs)
	median := allDegs[len(allDegs)/2]

	keep := func(id int64) bool {
		return degrees[id] >= median
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterNodes(g, keep)
	}
}

// sortInts sorts a slice of ints in ascending order.
func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}
