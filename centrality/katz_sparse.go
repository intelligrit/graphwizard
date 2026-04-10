// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
)

// KatzSparse computes Katz centrality for a directed graph using O(N) peak
// memory. Instead of pre-building a predecessor list (which is O(N+E)), each
// power iteration scatters contributions by walking outgoing edges on the fly.
//
// Use this when the edge count E is large relative to available memory.
// The trade-off is CPU: edges are traversed maxIter times rather than once.
//
// Parameters and semantics are identical to Katz.
func KatzSparse(ctx context.Context, g graph.Directed, alpha, beta, tol float64, maxIter int) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return make(map[int64]float64)
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	x := make([]float64, n)
	xNew := make([]float64, n)

	for iter := 0; iter < maxIter; iter++ {
		progress.Report(ctx, progress.Progress{Phase: "iterate", Step: iter, Total: maxIter})

		// Base score for every node.
		for i := range xNew {
			xNew[i] = beta
		}
		// Scatter: each source j adds alpha*x[j] to every successor.
		for j, jid := range ids {
			contrib := alpha * x[j]
			succs := g.From(jid)
			for succs.Next() {
				if i, ok := idx[succs.Node().ID()]; ok {
					xNew[i] += contrib
				}
			}
		}

		diff := 0.0
		for i := range x {
			diff += math.Abs(xNew[i] - x[i])
		}
		x, xNew = xNew, x
		if diff < tol {
			break
		}
	}

	result := make(map[int64]float64, n)
	for i, id := range ids {
		result[id] = x[i]
	}
	return result
}

// KatzUndirectedSparse computes Katz centrality for an undirected graph using
// O(N) peak memory by scanning edges each iteration rather than caching adj.
//
// Parameters and semantics are identical to KatzUndirected.
func KatzUndirectedSparse(ctx context.Context, g graph.Undirected, alpha, beta, tol float64, maxIter int) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return make(map[int64]float64)
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	x := make([]float64, n)
	xNew := make([]float64, n)

	for iter := 0; iter < maxIter; iter++ {
		progress.Report(ctx, progress.Progress{Phase: "iterate", Step: iter, Total: maxIter})

		for i := range xNew {
			xNew[i] = beta
		}
		// Each undirected edge (j,i) is traversed once per direction.
		for j, jid := range ids {
			contrib := alpha * x[j]
			neighbors := g.From(jid)
			for neighbors.Next() {
				if i, ok := idx[neighbors.Node().ID()]; ok {
					xNew[i] += contrib
				}
			}
		}

		diff := 0.0
		for i := range x {
			diff += math.Abs(xNew[i] - x[i])
		}
		x, xNew = xNew, x
		if diff < tol {
			break
		}
	}

	result := make(map[int64]float64, n)
	for i, id := range ids {
		result[id] = x[i]
	}
	return result
}
