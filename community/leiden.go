// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"math/rand"
	"sort"

	"github.com/intelligrit/graphwizard"
	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
)

type neighbor struct {
	node   int
	weight float64
}

// Leiden performs community detection using the Leiden algorithm, returning a
// map from node ID to community ID.
//
// The Leiden algorithm improves on Louvain by guaranteeing well-connected
// communities through a refinement phase. The resolution parameter controls
// community granularity: higher values produce more, smaller communities.
//
// Reference: V. Traag, L. Waltman, N.J. van Eck, "From Louvain to Leiden:
// guaranteeing well-connected communities", Scientific Reports, 2019.
func Leiden(ctx context.Context, g graph.Undirected, resolution float64, rng *rand.Rand) map[int64]int64 {
	progress.Report(ctx, progress.Progress{Phase: "build", Step: 0, Total: 1})
	origIDs, adj, degree, totalWeight := buildWeightedAdj(g)
	n := len(origIDs)
	if n == 0 {
		return make(map[int64]int64)
	}

	// membership[i] = current community label for original node i.
	membership := make([]int, n)
	for i := range membership {
		membership[i] = i
	}

	comm := make([]int, n)
	for i := range comm {
		comm[i] = i
	}
	curN := n
	// selfLoops tracks intra-community weight for each aggregate node.
	selfLoops := make([]float64, n)

	// Reusable buffers for aggregate() — allocated once, cleared each iteration.
	edgeWeightsBuf := make(map[edgeKey]float64)
	remapBuf := make(map[int]int)

	for iter := 0; iter < 100; iter++ {
		progress.Report(ctx, progress.Progress{Phase: "iterate", Step: iter, Total: -1})
		moved := localMove(adj, degree, selfLoops, comm, curN, totalWeight, resolution, defaultMaxLocalSweeps, rng)
		refined := refine(adj, degree, comm, curN, rng)

		// Update membership.
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

// defaultMaxLocalSweeps bounds the number of full node sweeps inside
// localMove. Three shuffled sweeps break oscillation cycles on hub-heavy
// graphs while keeping per-iteration cost proportional to edges, not sweeps²
// (the prior value of 10 caused 3+ hour runtimes on 3M-node bipartite graphs).
const defaultMaxLocalSweeps = 3

func localMove(adj [][]neighbor, degree, selfLoops []float64, comm []int, n int, totalWeight, resolution float64, maxSweeps int, rng *rand.Rand) bool {
	moved := false
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}

	// Use slices instead of maps: comm[i] ∈ [0, n) is guaranteed —
	// initial comm[i] = i, and aggregate() remaps to [0, newN) before each call.
	// Slice access is ~50x faster than map for hub nodes with 100K+ neighbors.
	sigmaTot := make([]float64, n)
	for i := 0; i < n; i++ {
		sigmaTot[comm[i]] += degree[i]
	}

	// commWeights[c] = total edge weight from the current node to community c.
	// dirty tracks which entries are non-zero; cleared at the start of each node
	// to avoid O(n) zeroing on every iteration.
	commWeights := make([]float64, n)
	dirty := make([]int, 0, 64)

	for sweep := 0; sweep < maxSweeps; sweep++ {
		// Reshuffle each sweep to break deterministic oscillation cycles.
		rng.Shuffle(n, func(i, j int) { order[i], order[j] = order[j], order[i] })
		changed := false
		for _, i := range order {
			// Clear only entries written by the previous node.
			for _, d := range dirty {
				commWeights[d] = 0
			}
			dirty = dirty[:0]

			for _, nb := range adj[i] {
				c := comm[nb.node]
				if commWeights[c] == 0 {
					dirty = append(dirty, c)
				}
				commWeights[c] += nb.weight
			}

			oldComm := comm[i]
			bestComm := oldComm
			bestDelta := 0.0

			wOld := commWeights[oldComm]
			oldSigmaTot := sigmaTot[oldComm]

			m := totalWeight
			if m == 0 {
				continue
			}

			for _, c := range dirty {
				if c == oldComm {
					continue
				}
				wc := commWeights[c]
				cSigmaTot := sigmaTot[c]
				delta := (wc-wOld)/m - resolution*degree[i]*(cSigmaTot-(oldSigmaTot-degree[i]))/(2*m*m)
				if delta > bestDelta {
					bestDelta = delta
					bestComm = c
				}
			}

			if bestComm != oldComm {
				sigmaTot[oldComm] -= degree[i]
				sigmaTot[bestComm] += degree[i]
				comm[i] = bestComm
				changed = true
				moved = true
			}
		}
		if !changed {
			break
		}
	}
	return moved
}

func refine(adj [][]neighbor, degree []float64, comm []int, n int, rng *rand.Rand) []int {
	refined := make([]int, n)
	for i := range refined {
		refined[i] = i
	}

	commMembers := make(map[int][]int)
	for i := 0; i < n; i++ {
		commMembers[comm[i]] = append(commMembers[comm[i]], i)
	}

	// Sort community IDs so RNG is consumed in a deterministic order across runs.
	cids := make([]int, 0, len(commMembers))
	for cid := range commMembers {
		cids = append(cids, cid)
	}
	sort.Ints(cids)

	// Slice-based subWeights: refined[nb.node] ∈ [0, n) since refined[i] = i initially.
	subWeights := make([]float64, n)
	dirty := make([]int, 0, 64)

	for _, cid := range cids {
		members := commMembers[cid]
		if len(members) <= 1 {
			continue
		}
		// Seed a per-community RNG from the main RNG. This mirrors refineParallel's
		// approach so that Leiden and LeidenParallel consume the top-level RNG
		// identically and produce the same partition for a given seed.
		localRNG := rand.New(rand.NewSource(rng.Int63()))
		perm := localRNG.Perm(len(members))
		for _, pi := range perm {
			i := members[pi]
			for _, d := range dirty {
				subWeights[d] = 0
			}
			dirty = dirty[:0]

			for _, nb := range adj[i] {
				if comm[nb.node] == comm[i] {
					r := refined[nb.node]
					if subWeights[r] == 0 {
						dirty = append(dirty, r)
					}
					subWeights[r] += nb.weight
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

	return refined
}

// edgeKey identifies a directed community→community edge for aggregation.
type edgeKey struct{ from, to int }

func aggregate(refined []int, adj [][]neighbor, degree, selfLoops []float64, n int, remap map[int]int, edgeWeights map[edgeKey]float64) ([]int, [][]neighbor, []float64, []float64, int, []int) {
	// Clear reusable buffers.
	for k := range remap {
		delete(remap, k)
	}

	newN := 0
	for i := 0; i < n; i++ {
		if _, ok := remap[refined[i]]; !ok {
			remap[refined[i]] = newN
			newN++
		}
	}

	aggMap := make([]int, n)
	for i := 0; i < n; i++ {
		aggMap[i] = remap[refined[i]]
	}

	newComm := make([]int, newN)
	for i := range newComm {
		newComm[i] = i
	}

	// Aggregate degrees: sum of original degrees of all nodes in each community.
	newDegree := make([]float64, newN)
	for i := 0; i < n; i++ {
		newDegree[remap[refined[i]]] += degree[i]
	}

	// Aggregate self-loops: existing self-loops + new intra-community edges.
	newSelfLoops := make([]float64, newN)
	for i := 0; i < n; i++ {
		ci := remap[refined[i]]
		newSelfLoops[ci] += selfLoops[i]
		for _, nb := range adj[i] {
			cj := remap[refined[nb.node]]
			if ci == cj {
				newSelfLoops[ci] += nb.weight // counted once per direction
			}
		}
	}
	// Each intra-community edge was counted from both endpoints.
	for i := range newSelfLoops {
		newSelfLoops[i] /= 2
	}

	// Aggregate inter-community edges using reusable map.
	for k := range edgeWeights {
		delete(edgeWeights, k)
	}
	for i := 0; i < n; i++ {
		ci := remap[refined[i]]
		for _, nb := range adj[i] {
			cj := remap[refined[nb.node]]
			if ci != cj {
				key := edgeKey{ci, cj}
				edgeWeights[key] += nb.weight
			}
		}
	}

	// Sort edge keys so adjacency list order is deterministic across runs.
	// Non-deterministic order here would propagate through localMove tie-breaking
	// in all subsequent iterations.
	keys := make([]edgeKey, 0, len(edgeWeights))
	for k := range edgeWeights {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].from != keys[j].from {
			return keys[i].from < keys[j].from
		}
		return keys[i].to < keys[j].to
	})
	newAdj := make([][]neighbor, newN)
	for _, k := range keys {
		newAdj[k.from] = append(newAdj[k.from], neighbor{node: k.to, weight: edgeWeights[k]})
	}

	return newComm, newAdj, newDegree, newSelfLoops, newN, aggMap
}

// buildWeightedAdj constructs dense-indexed weighted adjacency lists.
// It picks the fastest available path:
//  1. DenseAdjacency (preloaded CSR) — zero SQL, shares memory.
//  2. EdgeScanner (unpreloaded diskgraph) — one sequential table scan.
//  3. Fallback iter — per-node From()+Weight() queries.
func buildWeightedAdj(g graph.Undirected) (origIDs []int64, adj [][]neighbor, degree []float64, totalWeight float64) {
	if da, ok := g.(graphwizard.DenseAdjacency); ok && da.NodeIDs() != nil {
		return buildWeightedAdjFromDense(g, da)
	}
	if es, ok := g.(graphwizard.EdgeScanner); ok {
		return buildWeightedAdjFromScan(g, es)
	}
	return buildWeightedAdjFromIter(g)
}

func buildWeightedAdjFromScan(g graph.Undirected, es graphwizard.EdgeScanner) ([]int64, [][]neighbor, []float64, float64) {
	nodes := g.Nodes()
	var origIDs []int64
	for nodes.Next() {
		origIDs = append(origIDs, nodes.Node().ID())
	}
	// Sort node IDs for a deterministic index mapping (g.Nodes() iterates a map).
	sort.Slice(origIDs, func(i, j int) bool { return origIDs[i] < origIDs[j] })

	n := len(origIDs)
	if n == 0 {
		return nil, nil, nil, 0
	}

	idx := make(map[int64]int, n)
	for i, id := range origIDs {
		idx[id] = i
	}

	adj := make([][]neighbor, n)
	degree := make([]float64, n)
	totalWeight := 0.0

	es.ScanWeightedEdges(func(src, dst int64, w float64) {
		i, ok1 := idx[src]
		j, ok2 := idx[dst]
		if !ok1 || !ok2 {
			return
		}
		adj[i] = append(adj[i], neighbor{node: j, weight: w})
		degree[i] += w
		totalWeight += w
	})
	// Sort each adjacency list by node index for deterministic tie-breaking.
	for i := range adj {
		sort.Slice(adj[i], func(a, b int) bool { return adj[i][a].node < adj[i][b].node })
	}
	totalWeight /= 2
	return origIDs, adj, degree, totalWeight
}

func buildWeightedAdjFromDense(g graph.Undirected, da graphwizard.DenseAdjacency) ([]int64, [][]neighbor, []float64, float64) {
	origIDs := da.NodeIDs()
	n := da.NumNodes()
	adj := make([][]neighbor, n)
	degree := make([]float64, n)
	totalWeight := 0.0

	wg, isWeighted := g.(graph.Weighted)

	for i := 0; i < n; i++ {
		nbs := da.DenseNeighbors(i)
		if len(nbs) == 0 {
			continue
		}
		adj[i] = make([]neighbor, len(nbs))
		for k, j := range nbs {
			w := 1.0
			if isWeighted {
				if ew, ok := wg.Weight(origIDs[i], origIDs[j]); ok {
					w = ew
				}
			}
			adj[i][k] = neighbor{node: int(j), weight: w}
			degree[i] += w
			totalWeight += w
		}
		// Sort by node index for deterministic tie-breaking. DenseNeighbors
		// is documented to return dense indices, but custom implementations
		// are not required to return them sorted.
		sort.Slice(adj[i], func(a, b int) bool { return adj[i][a].node < adj[i][b].node })
	}
	totalWeight /= 2
	return origIDs, adj, degree, totalWeight
}

func buildWeightedAdjFromIter(g graph.Undirected) ([]int64, [][]neighbor, []float64, float64) {
	nodes := g.Nodes()
	var origIDs []int64
	for nodes.Next() {
		origIDs = append(origIDs, nodes.Node().ID())
	}
	// Sort node IDs for a deterministic index mapping regardless of g.Nodes() order.
	// g.Nodes() and g.From() iterate internal Go maps whose order varies per run.
	sort.Slice(origIDs, func(i, j int) bool { return origIDs[i] < origIDs[j] })

	n := len(origIDs)
	if n == 0 {
		return nil, nil, nil, 0
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
		// Sort neighbors by node index for deterministic adjacency list ordering.
		// Non-deterministic order here leaks into localMove and refine tie-breaking.
		sort.Slice(adj[i], func(a, b int) bool { return adj[i][a].node < adj[i][b].node })
	}
	totalWeight /= 2
	return origIDs, adj, degree, totalWeight
}

