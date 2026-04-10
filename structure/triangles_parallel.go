// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"gonum.org/v1/gonum/graph"
)

// TriangleCountParallel is a concurrent version of TriangleCount that
// distributes per-node triangle counting across available CPU cores.
//
// The graph implementation must be safe for concurrent reads (e.g.,
// simple.UndirectedGraph).
func TriangleCountParallel(ctx context.Context, g graph.Undirected) (perNode map[int64]int, total int) {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)

	// O(1) index lookup.
	idIdx := make(map[int64]int, n)
	for i, id := range ids {
		idIdx[id] = i
	}

	// Build neighbor sets (shared, read-only during counting).
	neighborSets := make(map[int64]map[int64]struct{}, n)
	for _, id := range ids {
		s := make(map[int64]struct{})
		it := g.From(id)
		for it.Next() {
			s[it.Node().ID()] = struct{}{}
		}
		neighborSets[id] = s
	}

	counts := make([]int64, n)
	var totalAtomic int64

	workers := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	ch := make(chan int, workers*4)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range ch {
				u := ids[idx]
				localCount := int64(0)
				for v := range neighborSets[u] {
					if v <= u {
						continue
					}
					for w := range neighborSets[u] {
						if w <= v {
							continue
						}
						if _, ok := neighborSets[v][w]; ok {
							localCount++
							atomic.AddInt64(&counts[idIdx[v]], 1)
							atomic.AddInt64(&counts[idIdx[w]], 1)
						}
					}
				}
				atomic.AddInt64(&counts[idx], localCount)
				atomic.AddInt64(&totalAtomic, localCount)
			}
		}()
	}

	for i := 0; i < n; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()

	perNode = make(map[int64]int, n)
	for i, id := range ids {
		perNode[id] = int(counts[i])
	}
	return perNode, int(totalAtomic)
}
