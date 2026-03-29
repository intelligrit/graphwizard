// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
)

// simulateSpreadParallel runs Monte Carlo cascade simulations in parallel.
func simulateSpreadParallel(adj map[int64][]int64, seeds map[int64]bool, prob float64, sims int, baseSeed int64) float64 {
	workers := runtime.GOMAXPROCS(0)
	if workers > sims {
		workers = sims
	}

	var totalActivated int64
	var wg sync.WaitGroup
	ch := make(chan int, sims)
	for s := 0; s < sims; s++ {
		ch <- s
	}
	close(ch)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(baseSeed + int64(workerID)*1000))
			for range ch {
				activated := make(map[int64]bool)
				var queue []int64
				for id := range seeds {
					activated[id] = true
					queue = append(queue, id)
				}
				for len(queue) > 0 {
					node := queue[0]
					queue = queue[1:]
					for _, neighbor := range adj[node] {
						if !activated[neighbor] && rng.Float64() < prob {
							activated[neighbor] = true
							queue = append(queue, neighbor)
						}
					}
				}
				atomic.AddInt64(&totalActivated, int64(len(activated)))
			}
		}(w)
	}
	wg.Wait()

	return float64(totalActivated) / float64(sims)
}
