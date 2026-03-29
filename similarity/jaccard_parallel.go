// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"runtime"
	"sync"

	"gonum.org/v1/gonum/graph"
)

// JaccardAllParallel is a concurrent version of JaccardAll that distributes
// pair computation across available CPU cores.
//
// For a graph with V nodes, this computes V*(V-1)/2 pairs. On an 11-core
// machine, expect ~8-10x speedup over the sequential version.
func JaccardAllParallel(g graph.Undirected, threshold float64) []NodePairScore {
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

	// Generate pair tasks.
	type pair struct{ i, j int }
	pairs := make([]pair, 0, n*(n-1)/2)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			pairs = append(pairs, pair{i, j})
		}
	}

	workers := runtime.GOMAXPROCS(0)
	var mu sync.Mutex
	var results []NodePairScore

	var wg sync.WaitGroup
	ch := make(chan pair, len(pairs))
	for _, p := range pairs {
		ch <- p
	}
	close(ch)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local []NodePairScore
			for p := range ch {
				nu := nsets[ids[p.i]]
				nv := nsets[ids[p.j]]
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
						A:     g.Node(ids[p.i]),
						B:     g.Node(ids[p.j]),
						Score: score,
					})
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
func PredictLinksParallel(g graph.Undirected, k int, scorer func(graph.Undirected, int64, int64) float64) []PredictedLink {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)

	type pair struct{ i, j int }
	var pairs []pair
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if !g.HasEdgeBetween(ids[i], ids[j]) {
				pairs = append(pairs, pair{i, j})
			}
		}
	}

	workers := runtime.GOMAXPROCS(0)
	var mu sync.Mutex
	var results []PredictedLink

	var wg sync.WaitGroup
	ch := make(chan pair, len(pairs))
	for _, p := range pairs {
		ch <- p
	}
	close(ch)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local []PredictedLink
			for p := range ch {
				score := scorer(g, ids[p.i], ids[p.j])
				if score > 0 {
					local = append(local, PredictedLink{A: ids[p.i], B: ids[p.j], Score: score})
				}
			}
			mu.Lock()
			results = append(results, local...)
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Sort by score descending.
	sortPredictions(results)
	if k > 0 && k < len(results) {
		results = results[:k]
	}
	return results
}

func sortPredictions(preds []PredictedLink) {
	for i := 1; i < len(preds); i++ {
		j := i
		for j > 0 && preds[j].Score > preds[j-1].Score {
			preds[j], preds[j-1] = preds[j-1], preds[j]
			j--
		}
	}
}
