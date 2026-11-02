// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"slices"
	"strings"
)

const maxTrajectoryYears = 16

// TrajectoryOptions selects a contiguous annual series. Exactly one of Type
// and TypePrefix must be set. Every requested year must have source records;
// this guards against treating an unavailable year as zero activity.
type TrajectoryOptions struct {
	FirstYear, LastYear int32
	Type, TypePrefix    string
	Weight              string // empty counts relationship records
	MinSegmentYears     int    // zero selects two complete years on each side
	MinPeerCount        int    // zero selects 30 providers in every peer year
}

// TrajectoryYear distinguishes no selected relationship, a missing named
// value, and a recorded zero. Value sums recorded values for this provider
// and year; the immutable Store still retains exact individual values.
type TrajectoryYear struct {
	Year          int32
	Observed      bool
	Present       bool
	Complete      bool
	Value         float64
	Relationships uint32
	Missing       uint32
}

// ProviderTrajectory reports a sustained change in an annual provider series.
// BreakYear is the first year after the best split, with at least
// MinSegmentYears complete observations on each side. RawShift is the change
// in mean log1p annual value; PeerAdjustedShift subtracts the same-year mean
// of providers in the same first-year activity bin. PeerScale is their pooled
// year-to-year log-change standard deviation. Score is abs(PeerAdjustedShift)
// divided by PeerScale; it is a descriptive ranking feature, not a p-value.
// HasChangePoint is false when any requested provider year is absent or has a
// missing named value, or when peer support is insufficient.
type ProviderTrajectory struct {
	Provider          int64
	Years             []TrajectoryYear
	FirstYearActivity uint32
	CompleteYears     int
	HasChangePoint    bool
	BreakYear         int32
	RawShift          float64
	PeerAdjustedShift float64
	PeerScale         float64
	Score             float64
	MinPeerCount      int
}

// TrajectoryStats reports annual coverage and scored providers.
type TrajectoryStats struct {
	SelectedRelationships int
	Providers             int
	CompleteProviders     int
	ScoredProviders       int
	Years                 int
}

type trajectoryState struct {
	value    [maxTrajectoryYears]float64
	count    [maxTrajectoryYears]uint32
	missing  [maxTrajectoryYears]uint32
	observed uint16
	present  uint16
}

func (s *trajectoryState) complete(i int) bool {
	return s.observed&(uint16(1)<<i) != 0 && s.missing[i] == 0
}

// VisitAnnualTrajectories scans one loaded Store without building graph
// projections. It keeps up to 16 yearly values per provider, then streams
// results in provider-ID order. Additional memory is O(providers * years).
// Peers share the floor(log2(max(1, first-year relationship count))) bin;
// source rows grouped to one provider-entity-type-year relationship make that
// count a graph-degree proxy. From must be the provider for selected types.
//
// The split-score approach follows the change-point problem for evolving
// bipartite graphs described by Sun et al., "GraphScope", KDD 2007. It is a
// simpler, bounded-memory provider trajectory statistic, not GraphScope's
// community segmentation algorithm. Compare it with raw annual changes and
// reviewed outcomes before using it as an investigative signal.
func (s *Store) VisitAnnualTrajectories(ctx context.Context, o TrajectoryOptions, visit func(ProviderTrajectory) error) (TrajectoryStats, error) {
	var stats TrajectoryStats
	if s == nil || o.FirstYear <= 0 || o.LastYear < o.FirstYear || (o.Type == "") == (o.TypePrefix == "") {
		return stats, errors.New("multigraph: select positive, increasing years and exactly one type selector")
	}
	years := int(o.LastYear-o.FirstYear) + 1
	if years > maxTrajectoryYears {
		return stats, fmt.Errorf("multigraph: trajectory supports at most %d years", maxTrajectoryYears)
	}
	if o.MinSegmentYears == 0 {
		o.MinSegmentYears = 2
	}
	if o.MinPeerCount == 0 {
		o.MinPeerCount = 30
	}
	if o.MinSegmentYears < 1 || years < 2*o.MinSegmentYears || o.MinPeerCount < 2 {
		return stats, errors.New("multigraph: invalid trajectory segment or peer support")
	}
	var column weightColumn
	if o.Weight != "" {
		var ok bool
		column, ok = s.weights[o.Weight]
		if !ok {
			return stats, fmt.Errorf("%w %q", ErrUnknownWeight, o.Weight)
		}
	}
	selectedType := make([]bool, len(s.types))
	for i, name := range s.types {
		selectedType[i] = o.Type != "" && name == o.Type || o.TypePrefix != "" && strings.HasPrefix(name, o.TypePrefix)
	}
	states := make(map[int32]*trajectoryState)
	yearCounts := make([]int, years)
	for i, r := range s.rows {
		if i&0xfffff == 0 {
			if err := ctx.Err(); err != nil {
				return stats, err
			}
		}
		if r.year < o.FirstYear || r.year > o.LastYear || !selectedType[r.kind] {
			continue
		}
		j := int(r.year - o.FirstYear)
		yearCounts[j]++
		state := states[r.from]
		if state == nil {
			state = &trajectoryState{}
			states[r.from] = state
		}
		if state.count[j] == ^uint32(0) {
			return stats, errors.New("multigraph: too many annual relationships for one provider")
		}
		state.count[j]++
		state.observed |= uint16(1) << j
		value, present := 1.0, true
		if o.Weight != "" {
			value, present = column.at(i)
		}
		if present {
			if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				return stats, fmt.Errorf("multigraph: trajectory weight %q on relationship %d must be finite and nonnegative", o.Weight, i)
			}
			state.present |= uint16(1) << j
			state.value[j] += value
		} else {
			state.missing[j]++
		}
		stats.SelectedRelationships++
	}
	for j, count := range yearCounts {
		if count == 0 {
			return stats, fmt.Errorf("multigraph: selected type has no relationships in %d", o.FirstYear+int32(j))
		}
	}
	ids := make([]int32, 0, len(states))
	for id := range states {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	stats.Providers, stats.Years = len(ids), years
	var peerLevel, peerChange [33][maxTrajectoryYears]annualPeer
	for _, id := range ids {
		state := states[id]
		complete := true
		for j := 0; j < years; j++ {
			if !state.complete(j) {
				complete = false
				break
			}
		}
		if !complete {
			continue
		}
		stats.CompleteProviders++
		bin := bits.Len32(max(1, state.count[0])) - 1
		for j := 0; j < years; j++ {
			peerLevel[bin][j].add(math.Log1p(state.value[j]))
			if j > 0 {
				peerChange[bin][j].add(math.Log1p(state.value[j]) - math.Log1p(state.value[j-1]))
			}
		}
	}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return stats, err
		}
		state := states[id]
		result := ProviderTrajectory{Provider: s.nodeIDs[id], Years: make([]TrajectoryYear, years), FirstYearActivity: state.count[0]}
		for j := 0; j < years; j++ {
			result.Years[j] = TrajectoryYear{Year: o.FirstYear + int32(j), Observed: state.observed&(uint16(1)<<j) != 0, Present: state.present&(uint16(1)<<j) != 0, Complete: state.complete(j), Value: state.value[j], Relationships: state.count[j], Missing: state.missing[j]}
			if state.complete(j) {
				result.CompleteYears++
			}
		}
		if result.CompleteYears == years {
			bin := bits.Len32(max(1, state.count[0])) - 1
			peerEnough := true
			minPeers := int(^uint(0) >> 1)
			varianceSum := 0.0
			for j := 0; j < years; j++ {
				minPeers = min(minPeers, peerLevel[bin][j].n)
				if peerLevel[bin][j].n < o.MinPeerCount {
					peerEnough = false
				}
				if j > 0 {
					p := peerChange[bin][j]
					if p.n < o.MinPeerCount {
						peerEnough = false
					}
					if p.n > 1 {
						varianceSum += p.m2 / float64(p.n-1)
					}
				}
			}
			result.MinPeerCount = minPeers
			if peerEnough {
				result.PeerScale = math.Sqrt(varianceSum / float64(years-1))
				if result.PeerScale > 0 {
					var raw, adjusted [maxTrajectoryYears]float64
					for j := 0; j < years; j++ {
						raw[j] = math.Log1p(state.value[j])
						adjusted[j] = raw[j] - peerLevel[bin][j].mean
					}
					best := -1.0
					for split := o.MinSegmentYears; split <= years-o.MinSegmentYears; split++ {
						rawShift := meanSlice(raw[split:years]) - meanSlice(raw[:split])
						peerShift := meanSlice(adjusted[split:years]) - meanSlice(adjusted[:split])
						score := math.Abs(peerShift) / result.PeerScale
						if score > best {
							best = score
							result.BreakYear = o.FirstYear + int32(split)
							result.RawShift = rawShift
							result.PeerAdjustedShift = peerShift
							result.Score = score
						}
					}
					result.HasChangePoint = true
					stats.ScoredProviders++
				}
			}
		}
		if visit != nil {
			if err := visit(result); err != nil {
				return stats, err
			}
		}
	}
	return stats, nil
}

func meanSlice(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total / float64(len(values))
}
