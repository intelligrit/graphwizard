// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// AStar returns the shortest path from source to target using the A* algorithm
// with the given heuristic function. The heuristic must never overestimate the
// actual cost (admissible). Returns nil and +Inf if no path exists.
//
// Wraps gonum/graph/path.AStar.
func AStar(ctx context.Context, g graph.Graph, source, target graph.Node, h path.Heuristic) ([]graph.Node, float64) {
	shortest, _ := path.AStar(source, target, g, h)
	nodes, weight := shortest.To(target.ID())
	return nodes, weight
}
