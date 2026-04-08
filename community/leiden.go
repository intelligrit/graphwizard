// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"math/rand"

	"github.com/intelligrit/graphwizard"
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
func Leiden(g graph.Undirected, resolution float64, rng *rand.Rand) map[int64]int64 {
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
		moved := localMove(adj, degree, selfLoops, comm, curN, totalWeight, resolution, rng)
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

func localMove(adj [][]neighbor, degree, selfLoops []float64, comm []int, n int, totalWeight, resolution float64, rng *rand.Rand) bool {
	moved := false
	order := rng.Perm(n)

	// Maintain sigmaTot incrementally: O(1) lookup instead of O(n) scan.
	sigmaTot := make(map[int]float64)
	for i := 0; i < n; i++ {
		sigmaTot[comm[i]] += degree[i]
	}

	commWeights := make(map[int]float64)
	changed := true
	for changed {
		changed = false
		for _, i := range order {
			for k := range commWeights {
				delete(commWeights, k)
			}
			for _, nb := range adj[i] {
				commWeights[comm[nb.node]] += nb.weight
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

			for c, wc := range commWeights {
				if c == oldComm {
					continue
				}
				cSigmaTot := sigmaTot[c]
				delta := (wc-wOld)/m - resolution*degree[i]*(cSigmaTot-(oldSigmaTot-degree[i]))/(2*m*m)
				if delta > bestDelta {
					bestDelta = delta
					bestComm = c
				}
			}

			if bestComm != oldComm {
				// Update sigmaTot incrementally.
				sigmaTot[oldComm] -= degree[i]
				sigmaTot[bestComm] += degree[i]
				comm[i] = bestComm
				changed = true
				moved = true
			}
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

	subWeights := make(map[int]float64)
	for _, members := range commMembers {
		if len(members) <= 1 {
			continue
		}
		perm := rng.Perm(len(members))
		for _, pi := range perm {
			i := members[pi]
			for k := range subWeights {
				delete(subWeights, k)
			}
			for _, nb := range adj[i] {
				if comm[nb.node] == comm[i] {
					subWeights[refined[nb.node]] += nb.weight
				}
			}

			bestRef := refined[i]
			bestW := 0.0
			for ref, w := range subWeights {
				if ref != refined[i] && w > bestW {
					bestW = w
					bestRef = ref
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

	newAdj := make([][]neighbor, newN)
	for key, w := range edgeWeights {
		newAdj[key.from] = append(newAdj[key.from], neighbor{node: key.to, weight: w})
	}

	return newComm, newAdj, newDegree, newSelfLoops, newN, aggMap
}

// buildWeightedAdj constructs dense-indexed weighted adjacency lists.
// If the graph implements DenseAdjacency (e.g. diskgraph with preloaded CSR),
// it reads directly from the shared structure — zero SQL queries and no
// duplicate adjacency copy for unweighted graphs.
func buildWeightedAdj(g graph.Undirected) (origIDs []int64, adj [][]neighbor, degree []float64, totalWeight float64) {
	if da, ok := g.(graphwizard.DenseAdjacency); ok && da.NodeIDs() != nil {
		return buildWeightedAdjFromDense(g, da)
	}
	return buildWeightedAdjFromIter(g)
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
	}
	totalWeight /= 2
	return origIDs, adj, degree, totalWeight
}

