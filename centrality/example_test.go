// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality_test

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/intelligrit/graphwizard/centrality"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

func ExamplePageRank() {
	// Directed cycle: all nodes have equal PageRank.
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	scores := centrality.PageRank(g, 0.85, 1e-6)
	fmt.Printf("%.4f\n", scores[0])
	// Output: 0.3333
}

func ExampleBetweenness() {
	// Star graph: center has highest betweenness.
	g := simple.NewUndirectedGraph()
	for i := int64(1); i <= 3; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}

	scores := centrality.Betweenness(g)
	fmt.Printf("center=%.1f leaf=%.1f\n", scores[0], scores[1])
	// Output: center=6.0 leaf=0.0
}

func ExampleCloseness() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	allPaths := path.DijkstraAllPaths(g)
	scores := centrality.Closeness(g, allPaths)
	fmt.Printf("center=%.4f endpoint=%.4f\n", scores[1], scores[0])
	// Output: center=0.5000 endpoint=0.3333
}

func ExampleDegree() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))

	scores := centrality.Degree(g)
	fmt.Printf("hub=%.2f leaf=%.2f\n", scores[0], scores[1])
	// Output: hub=1.00 leaf=0.50
}

func ExampleKatz() {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	scores := centrality.Katz(g, 0.1, 1.0, 1e-8, 100)
	fmt.Printf("node2 > node0: %v\n", scores[2] > scores[0])
	// Output: node2 > node0: true
}

func ExamplePersonalizedPageRank() {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(1)))

	scores := centrality.PersonalizedPageRank(g, 0, 0.85, 1e-6, 100)
	fmt.Printf("seed highest: %v\n", scores[0] > scores[2])
	// Output: seed highest: true
}

func ExampleDiameter() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	d := centrality.Diameter(g)
	fmt.Printf("diameter=%.0f\n", d)
	// Output: diameter=2
}

func ExampleRadius() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	r := centrality.Radius(g)
	fmt.Printf("radius=%.0f\n", r)
	// Output: radius=1
}

func ExampleHITS() {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	result := centrality.HITS(g, 1e-6)
	fmt.Printf("node0 is authority: %v\n", result.Authority[0] > result.Authority[1])
	// Output: node0 is authority: true
}

func ExampleEccentricity() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	ecc := centrality.Eccentricity(g)
	fmt.Printf("center ecc=%.0f, endpoint ecc=%.0f\n", ecc[1], ecc[0])
	// Output: center ecc=1, endpoint ecc=2
}

func ExampleInfluenceMaximization() {
	g := simple.NewUndirectedGraph()
	for i := int64(1); i <= 4; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}

	rng := rand.New(rand.NewSource(42))
	seeds, influence := centrality.InfluenceMaximization(g, 1, 0.5, 200, rng)
	fmt.Printf("seed: %d, influence>1: %v\n", seeds[0], influence > 1.0)
	// Output: seed: 0, influence>1: true
}

func ExamplePersonalizedPageRankUndirected() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	scores := centrality.PersonalizedPageRankUndirected(g, 0, 0.85, 1e-6, 100)
	fmt.Printf("seed > far: %v\n", scores[0] > scores[2])
	// Output: seed > far: true
}

func ExampleInDegree() {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(1)))

	scores := centrality.InDegree(g)
	fmt.Printf("node1=%.2f\n", scores[1])
	// Output: node1=1.00
}

func ExampleOutDegree() {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))

	scores := centrality.OutDegree(g)
	fmt.Printf("node0=%.2f\n", scores[0])
	// Output: node0=1.00
}

func ExampleApproximateBetweenness() {
	g := simple.NewUndirectedGraph()
	for i := int64(1); i <= 4; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}

	rng := rand.New(rand.NewSource(42))
	scores := centrality.ApproximateBetweenness(g, 5, rng)
	fmt.Printf("center highest: %v\n", scores[0] > scores[1])
	// Output: center highest: true
}

func ExampleEccentricityParallel() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	ecc := centrality.EccentricityParallel(g)
	fmt.Printf("endpoint=%.0f center=%.0f\n", ecc[0], ecc[1])
	// Output: endpoint=2 center=1
}

func ExampleDiameterParallel() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	d := centrality.DiameterParallel(g)
	fmt.Printf("diameter=%.0f\n", d)
	// Output: diameter=2
}

// Ensure unused imports are valid.
var _ = math.Inf
