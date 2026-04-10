// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"runtime"
	"sort"
	"sync"

	"gonum.org/v1/gonum/graph"
)

// JaccardAllParallel is a concurrent version of JaccardAll that distributes
// pair computation across available CPU cores.
//
// WARNING: This computes V*(V-1)/2 pairs and is only practical for graphs
// with fewer than ~50K nodes. For larger graphs, use a sparse approach that
// only computes similarity for nodes sharing at least one neighbor.
//
// The graph implementation must be safe for concurrent reads.
func JaccardAllParallel(ctx context.Context, g graph.Undirected, threshold float64) []NodePairScore {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)

	// Pre-build neighbor sets (shared, read-only).
	nsets := make(map[int64]map[int64]struct{}, n)
	for _, id := range ids {
		nsets[id] = neighborSet(g, id)
	}

	workers := runtime.GOMAXPROCS(0)
	var mu sync.Mutex
	var results []NodePairScore

	// Range-based work distribution: each worker gets a contiguous row range.
	var wg sync.WaitGroup
	rowCh := make(chan int, workers*4)

	go func() {
		for i := 0; i < n; i++ {
			rowCh <- i
		}
		close(rowCh)
	}()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local []NodePairScore
			for i := range rowCh {
				nu := nsets[ids[i]]
				for j := i + 1; j < n; j++ {
					nv := nsets[ids[j]]
					if len(nu) == 0 && len(nv) == 0 {
						continue
					}
					intersection := 0
					for id := range nu {
						if _, ok := nv[id]; ok {
							intersection++
						}
					}
					union := len(nu) + len(nv) - intersection
					if union == 0 {
						continue
					}
					score := float64(intersection) / float64(union)
					if score >= threshold {
						local = append(local, NodePairScore{
							A:     g.Node(ids[i]),
							B:     g.Node(ids[j]),
							Score: score,
						})
					}
				}
			}
			mu.Lock()
			results = append(results, local...)
			mu.Unlock()
		}()
	}
	wg.Wait()

	return results
}

// PredictLinksParallel is a concurrent version of PredictLinks.
//
// WARNING: This evaluates all V*(V-1)/2 non-adjacent pairs. Only practical
// for graphs with fewer than ~50K nodes.
//
// The graph implementation must be safe for concurrent reads.
func PredictLinksParallel(ctx context.Context, g graph.Undirected, k int, scorer func(graph.Undirected, int64, int64) float64) []PredictedLink {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)

	workers := runtime.GOMAXPROCS(0)
	var mu sync.Mutex
	var results []PredictedLink

	var wg sync.WaitGroup
	rowCh := make(chan int, workers*4)

	go func() {
		for i := 0; i < n; i++ {
			rowCh <- i
		}
		close(rowCh)
	}()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local []PredictedLink
			for i := range rowCh {
				for j := i + 1; j < n; j++ {
					if g.HasEdgeBetween(ids[i], ids[j]) {
						continue
					}
					score := scorer(g, ids[i], ids[j])
					if score > 0 {
						local = append(local, PredictedLink{A: ids[i], B: ids[j], Score: score})
					}
				}
			}
			mu.Lock()
			results = append(results, local...)
			mu.Unlock()
		}()
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if k > 0 && k < len(results) {
		results = results[:k]
	}
	return results
}
