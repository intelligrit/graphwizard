// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"runtime"
	"sync"

	"gonum.org/v1/gonum/graph"
)

// ClusteringCoefficientParallel is a concurrent version of
// ClusteringCoefficient that distributes per-node computation across
// available CPU cores.
func ClusteringCoefficientParallel(g graph.Undirected) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)

	results := make([]float64, n)

	workers := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	ch := make(chan int, n)
	for i := 0; i < n; i++ {
		ch <- i
	}
	close(ch)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range ch {
				id := ids[idx]
				neighbors := neighborSet(g, id)
				k := len(neighbors)
				if k < 2 {
					results[idx] = 0
					continue
				}

				nids := make([]int64, 0, k)
				for nid := range neighbors {
					nids = append(nids, nid)
				}
				edges := 0
				for i := 0; i < len(nids); i++ {
					for j := i + 1; j < len(nids); j++ {
						if g.HasEdgeBetween(nids[i], nids[j]) {
							edges++
						}
					}
				}
				results[idx] = 2.0 * float64(edges) / float64(k*(k-1))
			}
		}()
	}
	wg.Wait()

	result := make(map[int64]float64, n)
	for i, id := range ids {
		result[id] = results[i]
	}
	return result
}
