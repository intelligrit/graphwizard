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

// NearBicliqueSeed names two provider nodes, typically a pair returned by
// similarity.VisitBatchBipartiteOverlap on the same annual graph.
type NearBicliqueSeed struct{ A, B int64 }

// NearBicliqueOptions bounds every local search. A zero value selects the
// documented defaults through DefaultNearBicliqueOptions. The algorithm
// accepts at most 10,000 seeds per call.
type NearBicliqueOptions struct {
	MaxBridgeDegree       int
	MaxCandidateProviders int
	MaxCandidateBridges   int
	MaxProviders          int
	MaxBridges            int
	MinProviders          int
	MinBridges            int
	MinCoverage           float64 // minimum fraction of opposite-side nodes adjacent to each vertex
	MinDensity            float64 // minimum fraction of possible block edges present
	TopK                  int
	MaxBlockSimilarity    float64 // maximum mean Jaccard overlap across both partitions
}

// DefaultNearBicliqueOptions limits each seed to local, bounded neighborhoods.
func DefaultNearBicliqueOptions() NearBicliqueOptions {
	return NearBicliqueOptions{MaxBridgeDegree: 1024, MaxCandidateProviders: 16_384,
		MaxCandidateBridges: 16_384, MaxProviders: 64, MaxBridges: 16,
		MinProviders: 3, MinBridges: 2, MinCoverage: 0.75,
		MinDensity: 0.8, TopK: 10, MaxBlockSimilarity: 0.7}
}

// NearBicliqueBridge records global popularity and support within a block.
type NearBicliqueBridge struct {
	ID             int64
	GlobalDegree   int
	BlockProviders int
	Rarity         float64 // 1/log(GlobalDegree+5), as in FRAUDAR
}

// NearBicliqueBlock is a bounded, explainable provider/entity group. Every
// retained vertex meets MinCoverage; Density also meets MinDensity. This is a
// seeded heuristic, not an exact maximum biclique or a fraud probability.
type NearBicliqueBlock struct {
	Seed            NearBicliqueSeed
	ProviderIDs     []int64
	Bridges         []NearBicliqueBridge
	Edges           int
	Density         float64
	WeightedDensity float64
	SearchTruncated bool
}

// NearBicliqueStats reports coverage of the supplied seeds. Truncated counts
// searches that hit a cap even if they produced no accepted block.
type NearBicliqueStats struct {
	Seeds, Accepted, Truncated, Returned int
}

type nearSupport struct {
	id     int64
	count  int
	rarity float64
}

// FindNearBicliqueBlocks grows bounded local groups from provider-pair seeds,
// prunes vertices that lack enough links to the opposite partition, and keeps
// the highest-scoring diverse blocks. The score reuses FRAUDAR's logarithmic
// bridge-degree discount. The graph must be one annual, typed bipartite view;
// the algorithm does not infer historical affiliation dates.
//
// Working memory is O(nodes + MaxCandidateProviders + MaxCandidateBridges +
// seeds*(MaxProviders+MaxBridges)); no provider-pair or edge projection is
// materialized. SearchTruncated reports cap hits. Results depend on the seeds
// and are not exhaustive. Near-biclique search is motivated by top-k biplex
// methods (Yu and Long, PACMMOD 2023) and multi-block mining (Shin et al.,
// ECML-PKDD 2016); this bounded local heuristic is not either paper's solver.
func FindNearBicliqueBlocks(ctx context.Context, g graph.Undirected, isProvider func(int64) bool, seeds []NearBicliqueSeed, o NearBicliqueOptions) ([]NearBicliqueBlock, NearBicliqueStats, error) {
	var stats NearBicliqueStats
	if g == nil || isProvider == nil || len(seeds) > 10_000 {
		return nil, stats, errors.New("anomaly: near-biclique requires a graph, partition, and at most 10000 seeds")
	}
	if o == (NearBicliqueOptions{}) {
		o = DefaultNearBicliqueOptions()
	}
	if o.MaxBridgeDegree < 2 || o.MaxCandidateProviders < 2 || o.MaxCandidateBridges < 2 ||
		o.MaxProviders < o.MinProviders || o.MaxBridges < o.MinBridges || o.MinProviders < 2 || o.MinBridges < 2 ||
		o.MinCoverage <= 0 || o.MinCoverage > 1 || o.MinDensity <= 0 || o.MinDensity > 1 ||
		o.TopK < 1 || o.MaxBlockSimilarity < 0 || o.MaxBlockSimilarity > 1 {
		return nil, stats, errors.New("anomaly: invalid near-biclique limits")
	}
	if len(seeds) == 0 {
		return nil, stats, nil
	}
	degree := make(map[int64]int)
	providerDegree := make(map[int64]int)
	nodes := g.Nodes()
	for nodes.Next() {
		if err := ctx.Err(); err != nil {
			return nil, stats, err
		}
		id := nodes.Node().ID()
		left := isProvider(id)
		n := 0
		neighbors := g.From(id)
		for neighbors.Next() {
			if left == isProvider(neighbors.Node().ID()) {
				return nil, stats, fmt.Errorf("anomaly: edge from %d does not cross bipartition", id)
			}
			n++
		}
		if left {
			providerDegree[id] = n
		} else {
			degree[id] = n
		}
	}
	ordered := slices.Clone(seeds)
	for i := range ordered {
		if ordered[i].A > ordered[i].B {
			ordered[i].A, ordered[i].B = ordered[i].B, ordered[i].A
		}
		if ordered[i].A == ordered[i].B {
			return nil, stats, errors.New("anomaly: seed requires distinct providers")
		}
		if _, ok := providerDegree[ordered[i].A]; !ok {
			return nil, stats, fmt.Errorf("anomaly: seed %d is not a provider", ordered[i].A)
		}
		if _, ok := providerDegree[ordered[i].B]; !ok {
			return nil, stats, fmt.Errorf("anomaly: seed %d is not a provider", ordered[i].B)
		}
	}
	slices.SortFunc(ordered, func(a, b NearBicliqueSeed) int {
		if a.A != b.A {
			return compareInt64(a.A, b.A)
		}
		return compareInt64(a.B, b.B)
	})
	ordered = slices.Compact(ordered)
	stats.Seeds = len(ordered)
	blocks := make([]NearBicliqueBlock, 0, min(len(ordered), o.TopK*4))
	for _, seed := range ordered {
		if err := ctx.Err(); err != nil {
			return nil, stats, err
		}
		block, ok, err := nearBlockFromSeed(ctx, g, degree, providerDegree, seed, o)
		if err != nil {
			return nil, stats, err
		}
		if block.SearchTruncated {
			stats.Truncated++
		}
		if !ok {
			continue
		}
		stats.Accepted++
		blocks = append(blocks, block)
	}
	slices.SortFunc(blocks, func(a, b NearBicliqueBlock) int {
		if a.WeightedDensity != b.WeightedDensity {
			if a.WeightedDensity > b.WeightedDensity {
				return -1
			}
			return 1
		}
		if a.Density != b.Density {
			if a.Density > b.Density {
				return -1
			}
			return 1
		}
		if a.Edges != b.Edges {
			if a.Edges > b.Edges {
				return -1
			}
			return 1
		}
		if a.Seed.A != b.Seed.A {
			return compareInt64(a.Seed.A, b.Seed.A)
		}
		return compareInt64(a.Seed.B, b.Seed.B)
	})
	selected := make([]NearBicliqueBlock, 0, min(o.TopK, len(blocks)))
	for _, b := range blocks {
		diverse := true
		for _, prior := range selected {
			if blockSimilarity(b, prior) > o.MaxBlockSimilarity {
				diverse = false
				break
			}
		}
		if diverse {
			selected = append(selected, b)
			if len(selected) == o.TopK {
				break
			}
		}
	}
	stats.Returned = len(selected)
	return selected, stats, nil
}

func compareInt64(a, b int64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func nearBlockFromSeed(ctx context.Context, g graph.Undirected, degree, providerDegree map[int64]int, seed NearBicliqueSeed, o NearBicliqueOptions) (NearBicliqueBlock, bool, error) {
	block := NearBicliqueBlock{Seed: seed}
	shared := make([]int64, 0)
	aNeighbors := make(map[int64]bool, providerDegree[seed.A])
	for neighbors := g.From(seed.A); neighbors.Next(); {
		aNeighbors[neighbors.Node().ID()] = true
	}
	for neighbors := g.From(seed.B); neighbors.Next(); {
		id := neighbors.Node().ID()
		if aNeighbors[id] && degree[id] >= 2 && degree[id] <= o.MaxBridgeDegree {
			shared = append(shared, id)
		}
	}
	if len(shared) < o.MinBridges {
		return block, false, nil
	}
	slices.SortFunc(shared, func(a, b int64) int {
		if degree[a] != degree[b] {
			if degree[a] < degree[b] {
				return -1
			}
			return 1
		}
		return compareInt64(a, b)
	})
	seedLimit := min(o.MaxBridges, max(o.MinBridges, o.MaxBridges/2))
	if len(shared) > seedLimit {
		shared = shared[:seedLimit]
		block.SearchTruncated = true
	}
	providerSupport := make(map[int64]nearSupport, min(o.MaxCandidateProviders, 1024))
	for _, bridge := range shared {
		if err := ctx.Err(); err != nil {
			return block, false, err
		}
		for neighbors := g.From(bridge); neighbors.Next(); {
			id := neighbors.Node().ID()
			s, exists := providerSupport[id]
			if !exists && len(providerSupport) == o.MaxCandidateProviders {
				block.SearchTruncated = true
				continue
			}
			s.id, s.count, s.rarity = id, s.count+1, s.rarity+1/math.Log(float64(degree[bridge]+5))
			providerSupport[id] = s
		}
	}
	providers := []int64{seed.A, seed.B}
	others := make([]nearSupport, 0, len(providerSupport))
	for id, s := range providerSupport {
		if id != seed.A && id != seed.B {
			others = append(others, s)
		}
	}
	slices.SortFunc(others, func(a, b nearSupport) int {
		if a.count != b.count {
			if a.count > b.count {
				return -1
			}
			return 1
		}
		if a.rarity != b.rarity {
			if a.rarity > b.rarity {
				return -1
			}
			return 1
		}
		if providerDegree[a.id] != providerDegree[b.id] {
			if providerDegree[a.id] < providerDegree[b.id] {
				return -1
			}
			return 1
		}
		return compareInt64(a.id, b.id)
	})
	for _, s := range others {
		if len(providers) == o.MaxProviders {
			break
		}
		if s.count >= o.MinBridges {
			providers = append(providers, s.id)
		}
	}
	if len(providers) < o.MinProviders {
		return block, false, nil
	}
	entitySupport := make(map[int64]nearSupport, min(o.MaxCandidateBridges, 1024))
	for _, p := range providers {
		if err := ctx.Err(); err != nil {
			return block, false, err
		}
		for neighbors := g.From(p); neighbors.Next(); {
			id := neighbors.Node().ID()
			if degree[id] > o.MaxBridgeDegree {
				continue
			}
			s, exists := entitySupport[id]
			if !exists && len(entitySupport) == o.MaxCandidateBridges {
				block.SearchTruncated = true
				continue
			}
			s.id, s.count, s.rarity = id, s.count+1, s.rarity+1/math.Log(float64(degree[id]+5))
			entitySupport[id] = s
		}
	}
	seedEntities := make(map[int64]bool, len(shared))
	for _, id := range shared {
		seedEntities[id] = true
	}
	entities := slices.Clone(shared)
	remaining := make([]nearSupport, 0, len(entitySupport))
	for id, s := range entitySupport {
		if !seedEntities[id] && s.count >= o.MinProviders {
			remaining = append(remaining, s)
		}
	}
	slices.SortFunc(remaining, func(a, b nearSupport) int {
		if a.count != b.count {
			if a.count > b.count {
				return -1
			}
			return 1
		}
		if a.rarity != b.rarity {
			if a.rarity > b.rarity {
				return -1
			}
			return 1
		}
		return compareInt64(a.id, b.id)
	})
	for _, s := range remaining {
		if len(entities) == o.MaxBridges {
			break
		}
		entities = append(entities, s.id)
	}
	for len(providers) >= o.MinProviders && len(entities) >= o.MinBridges {
		worstSide, worstIndex, worstCoverage, worstID := -1, -1, 1.0, int64(0)
		for i, p := range providers {
			links := 0
			for _, e := range entities {
				if g.HasEdgeBetween(p, e) {
					links++
				}
			}
			coverage := float64(links) / float64(len(entities))
			if coverage < o.MinCoverage && (coverage < worstCoverage || coverage == worstCoverage && (worstSide < 0 || p < worstID)) {
				worstSide, worstIndex, worstCoverage, worstID = 0, i, coverage, p
			}
		}
		for i, e := range entities {
			links := 0
			for _, p := range providers {
				if g.HasEdgeBetween(p, e) {
					links++
				}
			}
			coverage := float64(links) / float64(len(providers))
			if coverage < o.MinCoverage && (coverage < worstCoverage || coverage == worstCoverage && (worstSide < 0 || e < worstID)) {
				worstSide, worstIndex, worstCoverage, worstID = 1, i, coverage, e
			}
		}
		if worstSide < 0 {
			break
		}
		if worstSide == 0 {
			providers = append(providers[:worstIndex], providers[worstIndex+1:]...)
		} else {
			entities = append(entities[:worstIndex], entities[worstIndex+1:]...)
		}
	}
	if len(providers) < o.MinProviders || len(entities) < o.MinBridges || !slices.Contains(providers, seed.A) || !slices.Contains(providers, seed.B) {
		return block, false, nil
	}
	slices.Sort(providers)
	slices.Sort(entities)
	weighted := 0.0
	for _, e := range entities {
		links := 0
		for _, p := range providers {
			if g.HasEdgeBetween(p, e) {
				links++
			}
		}
		block.Bridges = append(block.Bridges, NearBicliqueBridge{ID: e, GlobalDegree: degree[e], BlockProviders: links, Rarity: 1 / math.Log(float64(degree[e]+5))})
		block.Edges += links
		weighted += float64(links) / math.Log(float64(degree[e]+5))
	}
	block.Density = float64(block.Edges) / float64(len(providers)*len(entities))
	if block.Density < o.MinDensity {
		return block, false, nil
	}
	block.ProviderIDs = providers
	block.WeightedDensity = weighted / float64(len(providers)+len(entities))
	return block, true, nil
}

func blockSimilarity(a, b NearBicliqueBlock) float64 {
	providerOverlap := sortedJaccard(a.ProviderIDs, b.ProviderIDs)
	aBridges := make([]int64, len(a.Bridges))
	bBridges := make([]int64, len(b.Bridges))
	for i, e := range a.Bridges {
		aBridges[i] = e.ID
	}
	for i, e := range b.Bridges {
		bBridges[i] = e.ID
	}
	return (providerOverlap + sortedJaccard(aBridges, bBridges)) / 2
}

func sortedJaccard(a, b []int64) float64 {
	i, j, common := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			common++
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return float64(common) / float64(len(a)+len(b)-common)
}
