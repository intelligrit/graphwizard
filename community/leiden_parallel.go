// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"math/rand"
	"runtime"
	"sync"

	"gonum.org/v1/gonum/graph"
)

// LeidenParallel performs community detection using the Leiden algorithm with
// the refinement phase parallelized across communities. Each community's
// refinement is independent, so they run concurrently.
//
// Results are identical to the sequential Leiden for a given RNG seed. The rng
// parameter seeds per-community RNGs deterministically.
//
// Reference: V. Traag, L. Waltman, N.J. van Eck, "From Louvain to Leiden:
// guaranteeing well-connected communities", Scientific Reports, 2019.
func LeidenParallel(g graph.Undirected, resolution float64, rng *rand.Rand) map[int64]int64 {
	nodes := g.Nodes()
	var origIDs []int64
	for nodes.Next() {
		origIDs = append(origIDs, nodes.Node().ID())
	}
	n := len(origIDs)
	if n == 0 {
		return make(map[int64]int64)
	}

	idx := make(map[int64]int, n)
	for i, id := range origIDs {
		idx[id] = i
	}

	adj := make([][]neighbor, n)
	degree := make([]float64, n)
	totalWeight := 0.0

	for i, id := range origIDs {
		it := g.From(id)
		for it.Next() {
			j, ok := idx[it.Node().ID()]
			if !ok {
				continue
			}
			w := 1.0
			if wg, ok := g.(graph.Weighted); ok {
				if ew, ok := wg.Weight(id, origIDs[j]); ok {
					w = ew
				}
			}
			adj[i] = append(adj[i], neighbor{node: j, weight: w})
			degree[i] += w
			totalWeight += w
		}
	}
	totalWeight /= 2

	membership := make([]int, n)
	for i := range membership {
		membership[i] = i
	}

	comm := make([]int, n)
	for i := range comm {
		comm[i] = i
	}
	curN := n
	selfLoops := make([]float64, n)

	for iter := 0; iter < 100; iter++ {
		moved := localMove(adj, degree, selfLoops, comm, curN, totalWeight, resolution, rng)
		refined := refineParallel(adj, degree, comm, curN, rng)

		if iter == 0 {
			for i := 0; i < n; i++ {
				membership[i] = refined[i]
			}
		} else {
			for i := 0; i < n; i++ {
				membership[i] = refined[membership[i]]
			}
		}

		if !moved {
			break
		}

		var aggMap []int
		comm, adj, degree, selfLoops, curN, aggMap = aggregate(refined, adj, degree, selfLoops, curN)

		for i := 0; i < n; i++ {
			membership[i] = aggMap[membership[i]]
		}

		if curN <= 1 {
			break
		}
	}

	remap := make(map[int]int64)
	nextID := int64(0)
	result := make(map[int64]int64, n)
	for i, id := range origIDs {
		c := membership[i]
		if _, ok := remap[c]; !ok {
			remap[c] = nextID
			nextID++
		}
		result[id] = remap[c]
	}
	return result
}

// refineParallel runs the Leiden refinement phase with each community
// processed in a separate goroutine.
func refineParallel(adj [][]neighbor, degree []float64, comm []int, n int, rng *rand.Rand) []int {
	refined := make([]int, n)
	for i := range refined {
		refined[i] = i
	}

	commMembers := make(map[int][]int)
	for i := 0; i < n; i++ {
		commMembers[comm[i]] = append(commMembers[comm[i]], i)
	}

	// Create deterministic per-community seeds from the main RNG.
	type commWork struct {
		commID  int
		members []int
		seed    int64
	}
	var work []commWork
	for cid, members := range commMembers {
		if len(members) <= 1 {
			continue
		}
		work = append(work, commWork{cid, members, rng.Int63()})
	}

	workers := runtime.GOMAXPROCS(0)
	if workers > len(work) {
		workers = len(work)
	}
	if workers == 0 {
		return refined
	}

	var mu sync.Mutex
	ch := make(chan commWork, len(work))
	for _, w := range work {
		ch <- w
	}
	close(ch)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for cw := range ch {
				localRNG := rand.New(rand.NewSource(cw.seed))
				perm := localRNG.Perm(len(cw.members))
				localRefined := make(map[int]int)
				for _, i := range cw.members {
					localRefined[i] = i
				}
				for _, pi := range perm {
					i := cw.members[pi]
					subWeights := make(map[int]float64)
					for _, nb := range adj[i] {
						if comm[nb.node] == comm[i] {
							subWeights[localRefined[nb.node]] += nb.weight
						}
					}

					bestRef := localRefined[i]
					bestW := 0.0
					for ref, w := range subWeights {
						if ref != localRefined[i] && w > bestW {
							bestW = w
							bestRef = ref
						}
					}
					if bestW > 0 {
						localRefined[i] = bestRef
					}
				}
				mu.Lock()
				for i, ref := range localRefined {
					refined[i] = ref
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return refined
}
