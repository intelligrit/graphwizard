// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"
	"runtime"
	"sort"
	"sync"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
)

// BridgesParallel returns all bridge edges in an undirected graph, with
// bridge-finding DFS running independently per connected component.
//
// Results are identical to Bridges but computed in parallel across components.
func BridgesParallel(ctx context.Context, g graph.Undirected) []Bridge {
	components := topo.ConnectedComponents(g)

	if len(components) <= 1 {
		// Fall back to sequential for single component.
		return Bridges(ctx, g)
	}

	workers := runtime.GOMAXPROCS(0)
	if workers > len(components) {
		workers = len(components)
	}

	type result struct {
		bridges []Bridge
	}

	ch := make(chan []graph.Node, len(components))
	for _, comp := range components {
		ch <- comp
	}
	close(ch)

	results := make([][]Bridge, len(components))
	var mu sync.Mutex
	idx := 0

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for comp := range ch {
				if len(comp) < 2 {
					continue
				}

				// Run bridge-finding on this component.
				disc := make(map[int64]int)
				low := make(map[int64]int)
				visited := make(map[int64]bool)
				var bridges []Bridge
				timer := 0

				root := comp[0]
				bridgeDFS(g, root, nil, visited, disc, low, &timer, &bridges)

				if len(bridges) > 0 {
					mu.Lock()
					results[idx] = bridges
					idx++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	// Flatten results.
	var all []Bridge
	for _, b := range results {
		all = append(all, b...)
	}

	// Sort for deterministic output.
	sort.Slice(all, func(i, j int) bool {
		if all[i].From.ID() != all[j].From.ID() {
			return all[i].From.ID() < all[j].From.ID()
		}
		return all[i].To.ID() < all[j].To.ID()
	})

	return all
}
