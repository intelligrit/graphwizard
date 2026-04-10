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
			// Per-goroutine slice buffers avoid map allocation and map iteration
			// order non-determinism. Tie-breaking now follows adjacency list order,
			// which is deterministic after aggregate() sorts its edge keys.
			//
			// localRefined is a flat slice (indexed by node ID in [0,n)) that
			// replaces the per-community map. Using a slice eliminates Go's
			// per-process map hash randomization from the hot path entirely.
			// Stale entries from prior communities are never read: communities
			// are disjoint, so a community only accesses its own members' slots.
			subWeights := make([]float64, n)
			dirty := make([]int, 0, 64)
			localRefined := make([]int, n)
			for cw := range ch {
				// Initialise this community's slots. Prior communities' entries
				// may linger in other slots but are never accessed (disjoint membership).
				for _, i := range cw.members {
					localRefined[i] = i
				}
				localRNG := rand.New(rand.NewSource(cw.seed))
				perm := localRNG.Perm(len(cw.members))
				for _, pi := range perm {
					i := cw.members[pi]
					for _, d := range dirty {
						subWeights[d] = 0
					}
					dirty = dirty[:0]
					for _, nb := range adj[i] {
						if comm[nb.node] == comm[i] {
							r := localRefined[nb.node]
							if subWeights[r] == 0 {
								dirty = append(dirty, r)
							}
							subWeights[r] += nb.weight
						}
					}
					bestRef := localRefined[i]
					bestW := 0.0
					for _, r := range dirty {
						if r != localRefined[i] && subWeights[r] > bestW {
							bestW = subWeights[r]
							bestRef = r
						}
					}
					if bestW > 0 {
						localRefined[i] = bestRef
					}
				}
				// Write back using the member list (sorted ascending) rather than
				// map iteration, so the write order is deterministic across runs.
				mu.Lock()
				for _, i := range cw.members {
					refined[i] = localRefined[i]
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return refined
}
