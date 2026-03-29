// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"
	"runtime"
	"sync"

	"gonum.org/v1/gonum/graph"
)

// KatzParallel returns the Katz centrality for each node in a directed graph,
// with the per-node score computation within each power iteration parallelized
// across available CPU cores.
//
// Parameters and semantics are identical to Katz.
func KatzParallel(g graph.Directed, alpha, beta, tol float64, maxIter int) map[int64]float64 {
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

	workers := runtime.GOMAXPROCS(0)
	x := make([]float64, n)

	for iter := 0; iter < maxIter; iter++ {
		xNew := make([]float64, n)

		var wg sync.WaitGroup
		chunkSize := (n + workers - 1) / workers
		for start := 0; start < n; start += chunkSize {
			end := start + chunkSize
			if end > n {
				end = n
			}
			wg.Add(1)
			go func(lo, hi int) {
				defer wg.Done()
				for i := lo; i < hi; i++ {
					sum := 0.0
					for _, j := range preds[i] {
						sum += x[j]
					}
					xNew[i] = alpha*sum + beta
				}
			}(start, end)
		}
		wg.Wait()

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

// KatzUndirectedParallel returns the Katz centrality for each node in an
// undirected graph, with power iteration parallelized.
func KatzUndirectedParallel(g graph.Undirected, alpha, beta, tol float64, maxIter int) map[int64]float64 {
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

	workers := runtime.GOMAXPROCS(0)
	x := make([]float64, n)

	for iter := 0; iter < maxIter; iter++ {
		xNew := make([]float64, n)

		var wg sync.WaitGroup
		chunkSize := (n + workers - 1) / workers
		for start := 0; start < n; start += chunkSize {
			end := start + chunkSize
			if end > n {
				end = n
			}
			wg.Add(1)
			go func(lo, hi int) {
				defer wg.Done()
				for i := lo; i < hi; i++ {
					sum := 0.0
					for _, j := range adj[i] {
						sum += x[j]
					}
					xNew[i] = alpha*sum + beta
				}
			}(start, end)
		}
		wg.Wait()

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
