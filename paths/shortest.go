// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// ShortestPath returns the shortest path from source to target in a weighted
// graph, along with the total weight. Returns nil and +Inf if no path exists.
//
// Wraps gonum/graph/path.DijkstraFrom.
func ShortestPath(ctx context.Context, g graph.Graph, source, target int64) ([]graph.Node, float64) {
	shortest := path.DijkstraFrom(g.Node(source), g)
	nodes, weight := shortest.To(target)
	return nodes, weight
}

// AllShortestPaths computes all-pairs shortest paths for a graph.
// Returns a gonum path.AllShortest that can be queried for any pair and
// passed to BetweennessWeighted, Closeness, etc.
//
// Wraps gonum/graph/path.DijkstraAllPaths.
func AllShortestPaths(ctx context.Context, g graph.Graph) path.AllShortest {
	return path.DijkstraAllPaths(g)
}

// BellmanFord returns shortest paths from a single source, supporting
// negative edge weights. Returns the path and weight to the target, or
// nil and +Inf if unreachable. Returns ok=false if a negative cycle exists.
//
// Wraps gonum/graph/path.BellmanFordFrom.
func BellmanFord(ctx context.Context, g graph.Graph, source, target int64) (nodes []graph.Node, weight float64, ok bool) {
	shortest, ok := path.BellmanFordFrom(g.Node(source), g)
	if !ok {
		return nil, 0, false
	}
	nodes, weight = shortest.To(target)
	return nodes, weight, true
}

// FloydWarshall computes all-pairs shortest paths, supporting negative edge
// weights. Returns ok=false if a negative cycle exists.
//
// Wraps gonum/graph/path.FloydWarshall.
func FloydWarshall(ctx context.Context, g graph.Graph) (path.AllShortest, bool) {
	return path.FloydWarshall(g)
}
