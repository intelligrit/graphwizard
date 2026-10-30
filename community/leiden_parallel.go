// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/intelligrit/graphwizard/progress"
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
func LeidenParallel(ctx context.Context, g graph.Undirected, resolution float64, rng *rand.Rand) map[int64]int64 {
	progress.Report(ctx, progress.Progress{Phase: "build", Step: 0, Total: 1})
	origIDs, adj, degree, totalWeight := buildWeightedAdj(g)
	n := len(origIDs)
	if n == 0 {
		return make(map[int64]int64)
	}

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

	edgeWeightsBuf := make(map[edgeKey]float64)
	remapBuf := make(map[int]int)

	for iter := 0; iter < 100; iter++ {
		progress.Report(ctx, progress.Progress{Phase: "iterate", Step: iter, Total: -1})
		moved := localMove(adj, degree, selfLoops, comm, curN, totalWeight, resolution, defaultMaxLocalSweeps, rng)
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
		comm, adj, degree, selfLoops, curN, aggMap = aggregate(refined, adj, degree, selfLoops, curN, remapBuf, edgeWeightsBuf)

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

// refineParallel runs the Leiden refinement phase with disjoint communities
// distributed across a bounded pool of workers.
func refineParallel(adj weightedAdj, degree []float64, comm []int, n int, rng *rand.Rand) []int {
	refined := make([]int, n)
	for i := range refined {
		refined[i] = i
	}

	commMembers := make(map[int][]int)
	for i := 0; i < n; i++ {
		commMembers[comm[i]] = append(commMembers[comm[i]], i)
	}

	// Sort community IDs so per-community seeds are consumed in a deterministic
	// order regardless of Go map iteration randomization.
	cids := make([]int, 0, len(commMembers))
	for cid := range commMembers {
		cids = append(cids, cid)
	}
	sort.Ints(cids)

	type commWork struct {
		commID  int
		members []int
		seed    int64
	}
	var work []commWork
	for _, cid := range cids {
		members := commMembers[cid]
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

	// Each work item owns a disjoint set of nodes. Refined labels start at
	// their node index and can only move to another member of that community.
	// Checking comm[j] before reading refined[j] therefore confines every
	// read/write of refined and subWeights to slots owned by one worker.
	// Share these arrays instead of allocating two graph-sized arrays per worker.
	subWeights := make([]float64, n)
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
			cursor := adj.cursor()
			// The touched-slot list and RNG are worker-local. Clearing slots from
			// the previous community is safe: each community is processed once.
			dirty := make([]int, 0, 64)
			localRNG := rand.New(rand.NewSource(1))
			for cw := range ch {
				localRNG.Seed(cw.seed)
				perm := localRNG.Perm(len(cw.members))
				for _, pi := range perm {
					i := cw.members[pi]
					for _, d := range dirty {
						subWeights[d] = 0
					}
					dirty = dirty[:0]
					cursor.reset(i)
					for cursor.next() {
						j, w := cursor.value()
						if comm[j] == comm[i] {
							r := refined[j]
							if subWeights[r] == 0 {
								dirty = append(dirty, r)
							}
							subWeights[r] += w
						}
					}
					bestRef := refined[i]
					bestW := 0.0
					for _, r := range dirty {
						if r != refined[i] && subWeights[r] > bestW {
							bestW = subWeights[r]
							bestRef = r
						}
					}
					if bestW > 0 {
						refined[i] = bestRef
					}
				}
			}
		}()
	}
	wg.Wait()

	return refined
}
