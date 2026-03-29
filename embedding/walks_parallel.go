// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package embedding

import (
	"math/rand"
	"runtime"
	"sync"

	"gonum.org/v1/gonum/graph"
)

// Node2VecWalksParallel generates biased random walks using multiple goroutines.
// Each worker gets its own RNG derived from the provided seed to ensure
// reproducibility while enabling parallelism.
//
// This is the recommended version for graphs with >1K nodes.
func Node2VecWalksParallel(g graph.Undirected, params WalkParams, seed int64) [][]int64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return nil
	}

	// Build adjacency (shared, read-only).
	adj := make(map[int64][]int64, n)
	adjSet := make(map[int64]map[int64]bool, n)
	for _, id := range ids {
		var neighbors []int64
		neighborSet := make(map[int64]bool)
		it := g.From(id)
		for it.Next() {
			nid := it.Node().ID()
			neighbors = append(neighbors, nid)
			neighborSet[nid] = true
		}
		adj[id] = neighbors
		adjSet[id] = neighborSet
	}

	// Generate all (walkIndex, startNode) tasks.
	type walkTask struct {
		walkIdx int
		startID int64
	}
	totalWalks := params.WalksPerNode * n
	tasks := make([]walkTask, 0, totalWalks)

	// Use a master RNG for permutation ordering.
	masterRng := rand.New(rand.NewSource(seed))
	for w := 0; w < params.WalksPerNode; w++ {
		perm := masterRng.Perm(n)
		for _, idx := range perm {
			tasks = append(tasks, walkTask{walkIdx: len(tasks), startID: ids[idx]})
		}
	}

	// Results array (indexed, no lock needed).
	results := make([][]int64, totalWalks)

	workers := runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	ch := make(chan walkTask, workers*4)

	// Feed tasks from a separate goroutine to avoid buffering all at once.
	go func() {
		for _, t := range tasks {
			ch <- t
		}
		close(ch)
	}()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerSeed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(workerSeed))

			for task := range ch {
				walk := make([]int64, 0, params.WalkLength)
				walk = append(walk, task.startID)

				if len(adj[task.startID]) == 0 {
					results[task.walkIdx] = walk
					continue
				}

				cur := adj[task.startID][rng.Intn(len(adj[task.startID]))]
				walk = append(walk, cur)
				prev := task.startID

				for step := 2; step < params.WalkLength; step++ {
					neighbors := adj[cur]
					if len(neighbors) == 0 {
						break
					}

					weights := make([]float64, len(neighbors))
					total := 0.0
					for i, next := range neighbors {
						if next == prev {
							weights[i] = 1.0 / params.P
						} else if adjSet[prev][next] {
							weights[i] = 1.0
						} else {
							weights[i] = 1.0 / params.Q
						}
						total += weights[i]
					}

					r := rng.Float64() * total
					cumulative := 0.0
					chosen := neighbors[len(neighbors)-1]
					for i, wt := range weights {
						cumulative += wt
						if r <= cumulative {
							chosen = neighbors[i]
							break
						}
					}

					walk = append(walk, chosen)
					prev = cur
					cur = chosen
				}

				results[task.walkIdx] = walk
			}
		}(seed + int64(w) + 1)
	}
	wg.Wait()

	return results
}

// DeepWalkWalksParallel generates uniform random walks in parallel.
func DeepWalkWalksParallel(g graph.Undirected, walkLength, walksPerNode int, seed int64) [][]int64 {
	return Node2VecWalksParallel(g, WalkParams{
		WalkLength:   walkLength,
		WalksPerNode: walksPerNode,
		P:            1.0,
		Q:            1.0,
	}, seed)
}
