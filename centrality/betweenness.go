// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/network"
	"gonum.org/v1/gonum/graph/path"
)

// Betweenness returns the betweenness centrality for each node in an
// unweighted graph, keyed by node ID.
//
// Betweenness centrality measures how often a node lies on shortest paths
// between other node pairs.
//
// Wraps gonum/graph/network.Betweenness.
func Betweenness(g graph.Graph) map[int64]float64 {
	return network.Betweenness(g)
}

// BetweennessWeighted returns the betweenness centrality for each node in a
// weighted graph, keyed by node ID. The path.AllShortest argument provides
// precomputed shortest paths (use AllShortestPaths from the paths package).
//
// Wraps gonum/graph/network.BetweennessWeighted.
func BetweennessWeighted(g graph.Weighted, allPaths path.AllShortest) map[int64]float64 {
	return network.BetweennessWeighted(g, allPaths)
}

// EdgeBetweenness returns the betweenness centrality for each edge in an
// unweighted graph. The result maps [from, to] node ID pairs to scores.
//
// Wraps gonum/graph/network.EdgeBetweenness.
func EdgeBetweenness(g graph.Graph) map[[2]int64]float64 {
	return network.EdgeBetweenness(g)
}

// EdgeBetweennessWeighted returns the betweenness centrality for each edge in
// a weighted graph.
//
// Wraps gonum/graph/network.EdgeBetweennessWeighted.
func EdgeBetweennessWeighted(g graph.Weighted, allPaths path.AllShortest) map[[2]int64]float64 {
	return network.EdgeBetweennessWeighted(g, allPaths)
}

// Closeness returns the closeness centrality for each node, keyed by node ID.
//
// Wraps gonum/graph/network.Closeness.
func Closeness(g graph.Graph, allPaths path.AllShortest) map[int64]float64 {
	return network.Closeness(g, allPaths)
}

// Harmonic returns the harmonic centrality for each node, keyed by node ID.
//
// Wraps gonum/graph/network.Harmonic.
func Harmonic(g graph.Graph, allPaths path.AllShortest) map[int64]float64 {
	return network.Harmonic(g, allPaths)
}
