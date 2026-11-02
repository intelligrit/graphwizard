// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"

	"gonum.org/v1/gonum/graph"
)

// FRAUDARBlock is the highest-scoring bipartite block encountered during
// weighted minimum-degree peeling. Edge weight is 1/log(bridge degree+5),
// using the bridge degree in the original graph. WeightedDensity is the sum
// of these edge weights divided by the total number of block nodes. Density
// is the ordinary fraction of possible provider-bridge edges present.
// ProviderIDs and BridgeIDs are capped by MaxReturnedNodes; counts and scores
// describe the full block even when IDsTruncated is true.
type FRAUDARBlock struct {
	ProviderIDs, BridgeIDs []int64
	Providers, Bridges     int
	Edges                  int64
	WeightedDensity        float64
	Density                float64
	IDsTruncated           bool
}

type fraudarHeap struct {
	ids      []int64
	weight   []float64
	position []int
	items    []int
}

func (h *fraudarHeap) less(i, j int) bool {
	a, b := h.items[i], h.items[j]
	if h.weight[a] != h.weight[b] {
		return h.weight[a] < h.weight[b]
	}
	return h.ids[a] < h.ids[b]
}
func (h *fraudarHeap) swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.position[h.items[i]], h.position[h.items[j]] = i, j
}
func (h *fraudarHeap) down(i int) {
	n := len(h.items)
	for {
		left := i*2 + 1
		if left >= n {
			return
		}
		best := left
		if left+1 < n && h.less(left+1, left) {
			best++
		}
		if !h.less(best, i) {
			return
		}
		h.swap(i, best)
		i = best
	}
}
func (h *fraudarHeap) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(i, parent) {
			return
		}
		h.swap(i, parent)
		i = parent
	}
}
func (h *fraudarHeap) init() {
	for i := len(h.items)/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}
func (h *fraudarHeap) pop() int {
	first := h.items[0]
	last := len(h.items) - 1
	h.swap(0, last)
	h.items = h.items[:last]
	h.position[first] = -1
	if len(h.items) > 0 {
		h.down(0)
	}
	return first
}

// FRAUDAR finds one dense bipartite block with the logarithmic column weight
// proposed by Hooi et al., "FRAUDAR: Bounding Graph Fraud in the Face of
// Camouflage", KDD 2016. It scans the input graph but does not project all
// provider pairs or copy its edges. Working memory is O(nodes), plus at most
// maxReturnedNodes returned IDs; heap updates cost O(edges log nodes).
// isProvider identifies the row partition; every edge must cross partitions.
// A nil isProvider or nonpositive maxReturnedNodes is invalid. The graph must
// remain unchanged during the call. Use a single annual and relationship-type
// view: mixing years or current affiliations would corrupt historical meaning.
//
// The block is a structural lead, not a fraud probability. Drug popularity,
// common clinical practice, and manufacturer market structure can also form
// dense blocks; independent labels and matched baselines are required.
func FRAUDAR(ctx context.Context, g graph.Undirected, isProvider func(int64) bool, maxReturnedNodes int) (FRAUDARBlock, error) {
	var out FRAUDARBlock
	if g == nil || isProvider == nil || maxReturnedNodes < 1 {
		return out, errors.New("anomaly: FRAUDAR requires a graph, partition, and positive output limit")
	}
	ids := make([]int64, 0)
	provider := make([]bool, 0)
	index := make(map[int64]int)
	nodes := g.Nodes()
	for nodes.Next() {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		id := nodes.Node().ID()
		index[id] = len(ids)
		ids = append(ids, id)
		provider = append(provider, isProvider(id))
	}
	n := len(ids)
	if n == 0 {
		return out, nil
	}
	degree := make([]int, n)
	for i, id := range ids {
		if i&0xfff == 0 {
			if err := ctx.Err(); err != nil {
				return out, err
			}
		}
		neighbors := g.From(id)
		for neighbors.Next() {
			j, ok := index[neighbors.Node().ID()]
			if !ok || provider[i] == provider[j] {
				return out, fmt.Errorf("anomaly: edge from %d does not cross bipartite partition", id)
			}
			degree[i]++
		}
	}
	weights := make([]float64, n)
	total := 0.0
	for i, id := range ids {
		if !provider[i] {
			continue
		}
		neighbors := g.From(id)
		for neighbors.Next() {
			j := index[neighbors.Node().ID()]
			w := 1 / math.Log(float64(degree[j]+5))
			weights[i] += w
			weights[j] += w
			total += w
		}
	}
	if total == 0 {
		return out, nil
	}
	h := fraudarHeap{ids: ids, weight: weights, position: make([]int, n), items: make([]int, n)}
	for i := range ids {
		h.items[i] = i
		h.position[i] = i
	}
	h.init()
	removedAt := make([]int, n)
	bestStep, bestScore := 0, total/float64(n)
	for step := 1; step < n; step++ {
		if step&0xfff == 0 {
			if err := ctx.Err(); err != nil {
				return out, err
			}
		}
		i := h.pop()
		removedAt[i] = step
		total -= weights[i]
		neighbors := g.From(ids[i])
		for neighbors.Next() {
			j := index[neighbors.Node().ID()]
			if h.position[j] < 0 {
				continue
			}
			bridge := j
			if !provider[i] {
				bridge = i
			}
			w := 1 / math.Log(float64(degree[bridge]+5))
			weights[j] -= w
			h.up(h.position[j])
		}
		if total < 0 && total > -1e-8 {
			total = 0
		}
		score := total / float64(n-step)
		if score > bestScore+1e-12 {
			bestStep, bestScore = step, score
		}
	}
	last := h.pop()
	removedAt[last] = n
	active := make([]bool, n)
	for i := range ids {
		if removedAt[i] <= bestStep {
			continue
		}
		active[i] = true
		if provider[i] {
			out.Providers++
		} else {
			out.Bridges++
		}
	}
	providerLimit := min(out.Providers, (maxReturnedNodes+1)/2)
	bridgeLimit := min(out.Bridges, maxReturnedNodes-providerLimit)
	providerLimit += min(out.Providers-providerLimit, maxReturnedNodes-providerLimit-bridgeLimit)
	out.ProviderIDs = make([]int64, 0, providerLimit)
	out.BridgeIDs = make([]int64, 0, bridgeLimit)
	for i, id := range ids {
		if !active[i] {
			continue
		}
		if provider[i] && len(out.ProviderIDs) < providerLimit {
			out.ProviderIDs = append(out.ProviderIDs, id)
		}
		if !provider[i] && len(out.BridgeIDs) < bridgeLimit {
			out.BridgeIDs = append(out.BridgeIDs, id)
		}
	}
	out.IDsTruncated = len(out.ProviderIDs) < out.Providers || len(out.BridgeIDs) < out.Bridges
	weightedEdges := 0.0
	for i, id := range ids {
		if !active[i] || !provider[i] {
			continue
		}
		neighbors := g.From(id)
		for neighbors.Next() {
			j := index[neighbors.Node().ID()]
			if active[j] {
				out.Edges++
				weightedEdges += 1 / math.Log(float64(degree[j]+5))
			}
		}
	}
	if out.Providers+out.Bridges > 0 {
		out.WeightedDensity = weightedEdges / float64(out.Providers+out.Bridges)
	}
	if out.Providers > 0 && out.Bridges > 0 {
		out.Density = float64(out.Edges) / (float64(out.Providers) * float64(out.Bridges))
	}
	slices.Sort(out.ProviderIDs)
	slices.Sort(out.BridgeIDs)
	return out, nil
}
