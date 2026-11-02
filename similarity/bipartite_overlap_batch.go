// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"

	"gonum.org/v1/gonum/graph"
)

// BatchOverlapOptions bounds pair discovery for each left-side node. A pair
// must share a bridge of degree at most MaxBridgeDegree to be considered.
// Candidates are admitted in bridge-degree, bridge-ID, then provider-ID order until the
// per-provider cap is reached. This deterministic cap may omit better pairs;
// DroppedPairEncounters reports skipped bridge-to-pair encounters (the same
// omitted pair may be counted more than once). No inferential p-values are reported
// for this selected set of pairs.
type BatchOverlapOptions struct {
	MaxBridgeDegree      int
	MaxCandidatesPerLeft int
	TopK                 int
	MinSeedShared        int
	// LeftIDs limits the query to these left-side nodes. Nil scans all of them.
	LeftIDs []int64
}

// SharedBridge is an entity shared by a reported pair. Eligible says whether
// it participated in candidate discovery and ranking; popular bridges remain
// visible as evidence but cannot create a candidate by themselves.
type SharedBridge struct {
	ID       int64
	Degree   int
	Rarity   float64
	Eligible bool
}

// BatchOverlapMatch is one ranked provider pair. Score is the rarity sum of
// eligible shared entities divided by the geometric mean of provider degrees.
// Shared and RarityScore include every shared entity, including hubs.
// CandidateSearchTruncated means this provider's top-K may be incomplete.
type BatchOverlapMatch struct {
	A, B                     int64
	Score                    float64
	SeedShared               int
	SeedRarity               float64
	Shared                   int
	RarityScore              float64
	DegreeA                  int
	DegreeB                  int
	CandidateSearchTruncated bool
	Evidence                 []SharedBridge
}

// BatchOverlapStats describes coverage and deterministic candidate truncation.
type BatchOverlapStats struct {
	LeftNodes             int64
	MatchedLeftNodes      int64
	Candidates            int64
	DroppedPairEncounters int64
	TruncatedLeftNodes    int64
	Matches               int64
}

type overlapCandidate struct {
	id     int64
	shared int
	rarity float64
	score  float64
}

// VisitBatchBipartiteOverlap streams each left node's top matches to visit.
// It uses O(nodes + MaxCandidatesPerLeft + MaxBridgeDegree) working memory,
// excluding the caller's graph and output. Only bridges up to the configured
// degree generate pairs, preventing popular drugs or manufacturers from
// producing quadratic pair expansions. The graph must stay unchanged.
//
// Scores are descriptive ranking features. Candidate selection and truncation
// invalidate naive pairwise p-values; use an independently specified testing
// universe and multiple-testing correction for statistical claims. Compare
// results with raw shared count and degree-matched review baselines before
// treating them as investigative signals.
func VisitBatchBipartiteOverlap(ctx context.Context, g graph.Undirected, isLeft func(int64) bool, o BatchOverlapOptions, visit func(BatchOverlapMatch) error) (BatchOverlapStats, error) {
	var stats BatchOverlapStats
	if g == nil || isLeft == nil || visit == nil || o.MaxBridgeDegree < 2 || o.MaxCandidatesPerLeft < 1 || o.TopK < 1 || o.MinSeedShared < 1 {
		return stats, errors.New("similarity: invalid batch overlap graph, partition, callback, or limits")
	}
	degrees := make(map[int64]int)
	leftDegrees := make(map[int64]int)
	leftIDs := make([]int64, 0)
	nodes := g.Nodes()
	for nodes.Next() {
		if err := ctx.Err(); err != nil {
			return stats, err
		}
		id := nodes.Node().ID()
		left := isLeft(id)
		degree := 0
		neighbors := g.From(id)
		for neighbors.Next() {
			if left == isLeft(neighbors.Node().ID()) {
				return stats, fmt.Errorf("similarity: edge from %d does not cross bipartite partition", id)
			}
			degree++
			if degree&1023 == 0 {
				if err := ctx.Err(); err != nil {
					return stats, err
				}
			}
		}
		if left {
			leftDegrees[id] = degree
			if o.LeftIDs == nil {
				leftIDs = append(leftIDs, id)
			}
		} else {
			degrees[id] = degree
		}
	}
	if o.LeftIDs != nil {
		leftIDs = slices.Clone(o.LeftIDs)
		for _, id := range leftIDs {
			if _, ok := leftDegrees[id]; !ok {
				return stats, fmt.Errorf("similarity: %d is not a left-side node", id)
			}
		}
	}
	slices.Sort(leftIDs)
	leftIDs = slices.Compact(leftIDs)
	for _, a := range leftIDs {
		if err := ctx.Err(); err != nil {
			return stats, err
		}
		stats.LeftNodes++
		bridges := make([]int64, 0, leftDegrees[a])
		neighbors := g.From(a)
		for neighbors.Next() {
			id := neighbors.Node().ID()
			degree := degrees[id]
			if degree >= 2 && degree <= o.MaxBridgeDegree {
				bridges = append(bridges, id)
			}
		}
		if len(bridges) == 0 {
			continue
		}
		slices.SortFunc(bridges, func(x, y int64) int {
			if degrees[x] < degrees[y] {
				return -1
			}
			if degrees[x] > degrees[y] {
				return 1
			}
			if x < y {
				return -1
			}
			if x > y {
				return 1
			}
			return 0
		})
		candidates := make(map[int64]overlapCandidate, min(o.MaxCandidatesPerLeft, 16))
		truncated := false
		for _, bridge := range bridges {
			if err := ctx.Err(); err != nil {
				return stats, err
			}
			providers := make([]int64, 0, degrees[bridge])
			neighbors = g.From(bridge)
			for neighbors.Next() {
				providers = append(providers, neighbors.Node().ID())
			}
			slices.Sort(providers)
			for _, b := range providers {
				if b == a {
					continue
				}
				c, present := candidates[b]
				if !present {
					if len(candidates) == o.MaxCandidatesPerLeft {
						stats.DroppedPairEncounters++
						truncated = true
						continue
					}
					c = overlapCandidate{id: b}
					stats.Candidates++
				}
				c.shared++
				c.rarity += bridgeRarity(degrees[bridge])
				candidates[b] = c
			}
		}
		if truncated {
			stats.TruncatedLeftNodes++
		}
		ranked := make([]overlapCandidate, 0, len(candidates))
		for _, c := range candidates {
			if c.shared < o.MinSeedShared {
				continue
			}
			c.score = c.rarity / math.Sqrt(float64(leftDegrees[a])*float64(leftDegrees[c.id]))
			ranked = append(ranked, c)
		}
		slices.SortFunc(ranked, func(x, y overlapCandidate) int {
			if x.score > y.score {
				return -1
			}
			if x.score < y.score {
				return 1
			}
			if x.shared > y.shared {
				return -1
			}
			if x.shared < y.shared {
				return 1
			}
			if x.id < y.id {
				return -1
			}
			if x.id > y.id {
				return 1
			}
			return 0
		})
		if len(ranked) == 0 {
			continue
		}
		stats.MatchedLeftNodes++
		allA := make(map[int64]int, leftDegrees[a])
		neighbors = g.From(a)
		for neighbors.Next() {
			id := neighbors.Node().ID()
			allA[id] = degrees[id]
		}
		for _, c := range ranked[:min(o.TopK, len(ranked))] {
			m := BatchOverlapMatch{A: a, B: c.id, Score: c.score, SeedShared: c.shared, SeedRarity: c.rarity, DegreeA: leftDegrees[a], DegreeB: leftDegrees[c.id], CandidateSearchTruncated: truncated}
			neighbors = g.From(c.id)
			for neighbors.Next() {
				id := neighbors.Node().ID()
				degree, common := allA[id]
				if !common {
					continue
				}
				e := SharedBridge{ID: id, Degree: degree, Rarity: bridgeRarity(degree), Eligible: degree <= o.MaxBridgeDegree}
				m.Evidence = append(m.Evidence, e)
				m.RarityScore += e.Rarity
				m.Shared++
			}
			slices.SortFunc(m.Evidence, func(x, y SharedBridge) int {
				if x.ID < y.ID {
					return -1
				}
				if x.ID > y.ID {
					return 1
				}
				return 0
			})
			if err := visit(m); err != nil {
				return stats, err
			}
			stats.Matches++
		}
	}
	return stats, nil
}
