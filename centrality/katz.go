// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"

	"gonum.org/v1/gonum/graph"
)

// Katz returns the Katz centrality for each node in a directed graph, keyed
// by node ID.
//
// Katz centrality measures influence by summing all paths from every node,
// with longer paths attenuated by alpha^k. The parameter alpha must be less
// than 1/lambda_max (the largest eigenvalue of the adjacency matrix) for
// convergence; typical values are 0.01–0.1. Beta is the base score given to
// every node (usually 1.0).
//
// The algorithm uses power iteration with the given tolerance and maximum
// iterations as stopping criteria.
//
// Reference: L. Katz, "A New Status Index Derived from Sociometric Analysis",
// Psychometrika, 1953.
func Katz(g graph.Directed, alpha, beta, tol float64, maxIter int) map[int64]float64 {
	// Collect node IDs.
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return nil
	}

	// Index mapping for fast lookup.
	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	// Build predecessor lists (who points to each node).
	preds := make([][]int, n)
	for i := range preds {
		preds[i] = []int{}
	}
	for _, id := range ids {
		to := g.From(id)
		for to.Next() {
			j, ok := idx[to.Node().ID()]
			if ok {
				preds[j] = append(preds[j], idx[id])
			}
		}
	}

	// Power iteration: x_new[i] = alpha * sum(x_old[j] for j in predecessors(i)) + beta
	x := make([]float64, n)
	for i := range x {
		x[i] = 0
	}

	for iter := 0; iter < maxIter; iter++ {
		xNew := make([]float64, n)
		for i := 0; i < n; i++ {
			sum := 0.0
			for _, j := range preds[i] {
				sum += x[j]
			}
			xNew[i] = alpha*sum + beta
		}

		// Check convergence.
		diff := 0.0
		for i := 0; i < n; i++ {
			diff += math.Abs(xNew[i] - x[i])
		}
		x = xNew
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

// KatzUndirected returns the Katz centrality for each node in an undirected
// graph by treating each undirected edge as two directed edges.
func KatzUndirected(g graph.Undirected, alpha, beta, tol float64, maxIter int) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return nil
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	adj := make([][]int, n)
	for i := range adj {
		adj[i] = []int{}
	}
	for _, id := range ids {
		neighbors := g.From(id)
		for neighbors.Next() {
			j, ok := idx[neighbors.Node().ID()]
			if ok {
				adj[idx[id]] = append(adj[idx[id]], j)
			}
		}
	}

	x := make([]float64, n)

	for iter := 0; iter < maxIter; iter++ {
		xNew := make([]float64, n)
		for i := 0; i < n; i++ {
			sum := 0.0
			for _, j := range adj[i] {
				sum += x[j]
			}
			xNew[i] = alpha*sum + beta
		}

		diff := 0.0
		for i := 0; i < n; i++ {
			diff += math.Abs(xNew[i] - x[i])
		}
		x = xNew
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
