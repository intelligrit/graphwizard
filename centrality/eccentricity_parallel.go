// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math"
	"runtime"
	"sync"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// EccentricityParallel is a concurrent version of Eccentricity that computes
// per-node shortest paths in parallel across available CPU cores.
func EccentricityParallel(ctx context.Context, g graph.Graph) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return make(map[int64]float64)
	}

	results := make([]float64, n)

	workers := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	ch := make(chan int, n)
	for i := 0; i < n; i++ {
		ch <- i
	}
	close(ch)

	progress.Report(ctx, progress.Progress{Phase: "nodes", Step: 0, Total: n})
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range ch {
				uid := ids[idx]
				shortest := path.DijkstraFrom(g.Node(uid), g)
				maxDist := 0.0
				for _, vid := range ids {
					if vid == uid {
						continue
					}
					_, d := shortest.To(vid)
					if !math.IsInf(d, 1) && d > maxDist {
						maxDist = d
					}
				}
				results[idx] = maxDist
			}
		}()
	}
	wg.Wait()
	progress.Report(ctx, progress.Progress{Phase: "nodes", Step: n, Total: n})

	result := make(map[int64]float64, n)
	for i, id := range ids {
		result[id] = results[i]
	}
	return result
}

// DiameterParallel computes the graph diameter using parallel eccentricity.
func DiameterParallel(ctx context.Context, g graph.Graph) float64 {
	ecc := EccentricityParallel(ctx, g)
	d := 0.0
	for _, e := range ecc {
		if e > d {
			d = e
		}
	}
	return d
}
