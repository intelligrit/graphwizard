// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/network"
)

// PageRank returns the PageRank centrality for each node in a directed graph,
// keyed by node ID.
//
// The damping factor (typically 0.85) is the probability of following an edge
// rather than jumping to a random node. Tolerance controls convergence.
//
// Wraps gonum/graph/network.PageRank.
func PageRank(ctx context.Context, g graph.Directed, damping, tol float64) map[int64]float64 {
	return network.PageRank(g, damping, tol)
}

// PageRankSparse is identical to PageRank but optimized for sparse graphs.
//
// Wraps gonum/graph/network.PageRankSparse.
func PageRankSparse(ctx context.Context, g graph.Directed, damping, tol float64) map[int64]float64 {
	return network.PageRankSparse(g, damping, tol)
}
