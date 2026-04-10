// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
)

// ApproximateBetweenness estimates betweenness centrality by running Brandes'
// algorithm from k randomly sampled source nodes instead of all V nodes.
// Results are scaled by V/k to approximate the true values.
//
// This reduces complexity from O(VE) to O(kE) and runs source BFS passes in
// parallel across available CPU cores.
//
// For large graphs, k=1000 typically gives a good approximation of the
// relative ranking. Higher k improves accuracy at the cost of runtime.
//
// Reference: U. Brandes and C. Pich, "Centrality Estimation in Large
// Networks", International Journal of Bifurcation and Chaos, 2007.
func ApproximateBetweenness(ctx context.Context, g graph.Graph, k int, rng *rand.Rand) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 || k <= 0 {
		return make(map[int64]float64)
	}
	if k > n {
		k = n
	}

	// Sample k source nodes.
	perm := rng.Perm(n)
	sources := make([]int64, k)
	for i := 0; i < k; i++ {
		sources[i] = ids[perm[i]]
	}

	// Collect all node IDs for BFS.
	idSet := make(map[int64]bool, n)
	for _, id := range ids {
		idSet[id] = true
	}

	// Run Brandes from each source in parallel.
	workers := runtime.GOMAXPROCS(0)
	if workers > k {
		workers = k
	}

	type localResult struct {
		scores map[int64]float64
	}
	resultsCh := make(chan localResult, workers)

	var wg sync.WaitGroup
	ch := make(chan int64, len(sources))
	for _, s := range sources {
		ch <- s
	}
	close(ch)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make(map[int64]float64, n)
			for source := range ch {
				brandesSingleSource(g, source, ids, local)
			}
			resultsCh <- localResult{scores: local}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Aggregate.
	result := make(map[int64]float64, n)
	completed := 0
	for lr := range resultsCh {
		completed++
		progress.Report(ctx, progress.Progress{Phase: "sources", Step: completed, Total: workers})
		for id, v := range lr.scores {
			result[id] += v
		}
	}

	// Scale by V/k to approximate full betweenness.
	scale := float64(n) / float64(k)
	for id := range result {
		result[id] *= scale
	}

	return result
}

// brandesSingleSource computes dependency scores from a single source using
// BFS-based Brandes algorithm and accumulates into the local map.
func brandesSingleSource(g graph.Graph, source int64, allIDs []int64, accum map[int64]float64) {
	// Dijkstra-based Brandes: uses a priority queue to handle weighted edges
	// correctly, processing nodes in non-decreasing distance order.
	dist := make(map[int64]float64)
	sigma := make(map[int64]float64)
	pred := make(map[int64][]int64)
	dist[source] = 0
	sigma[source] = 1

	var stack []int64
	visited := make(map[int64]bool)

	// Simple priority queue via linear scan (sufficient for sampled sources).
	pending := map[int64]bool{source: true}

	for len(pending) > 0 {
		// Extract minimum distance node.
		v := int64(-1)
		vDist := math.Inf(1)
		for id := range pending {
			if dist[id] < vDist {
				vDist = dist[id]
				v = id
			}
		}
		delete(pending, v)
		if visited[v] {
			continue
		}
		visited[v] = true
		stack = append(stack, v)

		neighbors := g.From(v)
		for neighbors.Next() {
			w := neighbors.Node().ID()
			vw := 1.0
			if wg, ok := g.(graph.Weighted); ok {
				if ew, ok := wg.Weight(v, w); ok {
					vw = ew
				}
			}

			newDist := dist[v] + vw
			if dw, seen := dist[w]; !seen {
				dist[w] = newDist
				sigma[w] = sigma[v]
				pred[w] = []int64{v}
				pending[w] = true
			} else if math.Abs(newDist-dw) < 1e-10 {
				sigma[w] += sigma[v]
				pred[w] = append(pred[w], v)
			} else if newDist < dw {
				dist[w] = newDist
				sigma[w] = sigma[v]
				pred[w] = []int64{v}
				pending[w] = true
			}
		}
	}

	// Back-propagation of dependencies.
	delta := make(map[int64]float64)
	for i := len(stack) - 1; i >= 0; i-- {
		w := stack[i]
		for _, v := range pred[w] {
			delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
		}
		if w != source {
			accum[w] += delta[w]
		}
	}
}
