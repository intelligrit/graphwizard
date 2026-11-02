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

// AnnualChangeOptions selects two observed annual slices of a relationship
// store. Type matches one exact relationship type; TypePrefix matches a family
// such as "payment/". Exactly one selector must be set. Year zero (undated,
// current relationships) is deliberately unavailable to historical analysis.
// Each selected type must have records in both years unless AllowEmptyYear is
// true, preventing an unavailable source year from appearing as mass edge loss.
type AnnualChangeOptions struct {
	Earlier, Later int32
	Type           string
	TypePrefix     string
	Weight         string // empty counts relationship records
	AllowEmptyYear bool
}

// AnnualMetric distinguishes no relationship, a relationship with a missing
// named value, and a recorded zero. Value sums recorded values in one year;
// Complete is false if any selected record lacked the named value.
type AnnualMetric struct {
	Observed      bool
	Present       bool
	Complete      bool
	Value         float64
	Relationships int
	Missing       int
}

// AnnualEdgeChange describes a provider-entity-type connection across years.
// AdjustedLogChange compares a continuing edge's change to its provider's
// total change. HasAdjustedChange is false for missing values or new/lost edges.
type AnnualEdgeChange struct {
	Provider, Entity             int64
	Type                         string
	Earlier, Later               AnnualMetric
	New, Lost                    bool
	HasAdjustedChange            bool
	HasShareShift                bool
	LogChange, AdjustedLogChange float64
	ShareShift                   float64
}

// AnnualProviderChange reports connection turnover, named totals, and
// concentration. PeerZ compares log total change to providers in the same
// floor(log2(max(1, earlier connection count))) degree bin. Peer values are
// descriptive, not calibrated fraud probabilities.
type AnnualProviderChange struct {
	Provider                                        int64
	EarlierDegree, LaterDegree                      int
	NewConnections, LostConnections                 int
	Earlier, Later                                  AnnualMetric
	HasConcentration                                bool
	EarlierHHI, LaterHHI, HHIShift                  float64
	HasMetricChange                                 bool
	HasPeerZ                                        bool
	LogChange, PeerMean, PeerSD, PeerZ              float64
	PeerCount                                       int
	TurnoverRate                                    float64
	PeerTurnoverMean, PeerTurnoverSD, PeerTurnoverZ float64
	PeerTurnoverCount                               int
	HasPeerTurnoverZ                                bool
	PeerHHIMean, PeerHHISD, PeerHHIZ                float64
	PeerHHICount                                    int
	HasPeerHHIZ                                     bool
}

// AnnualChangeStats records selected rows and output coverage. Scratch space
// is proportional to the rows from the two selected years plus providers,
// independent of the number of other years held in the Store.
type AnnualChangeStats struct {
	SelectedRelationships int
	Edges, Providers      int
	NewEdges, LostEdges   int
	SortedInput           bool
}

type annualProviderState struct {
	degree    [2]int
	new, lost int
	metric    [2]AnnualMetric
	squares   [2]float64
}

type annualPeer struct {
	n        int
	mean, m2 float64
}

func (p *annualPeer) add(x float64) {
	p.n++
	d := x - p.mean
	p.mean += d / float64(p.n)
	p.m2 += d * (x - p.mean)
}

func annualBin(degree int) int { return bits.Len(uint(max(1, degree))) - 1 }

// VisitAnnualChanges streams annual provider and edge results from one loaded
// Store. It selects only the two explicit nonzero years and never includes
// undated affiliations. Edge grouping is sorted if the selected store rows
// were not already ordered by endpoint and type. Two passes over the selected
// row IDs allow provider and degree-peer controls without retaining edge
// results. From must be the provider and To the entity for the selected types.
// Callbacks may be nil; results are emitted in stable ID order.
//
// The caller should compare turnover, raw log change, and HHI against peer
// adjusted features and known review outcomes before using any as a signal.
func (s *Store) VisitAnnualChanges(ctx context.Context, o AnnualChangeOptions, visitProvider func(AnnualProviderChange) error, visitEdge func(AnnualEdgeChange) error) (AnnualChangeStats, error) {
	var stats AnnualChangeStats
	if s == nil || o.Earlier <= 0 || o.Later <= o.Earlier || (o.Type == "") == (o.TypePrefix == "") {
		return stats, errors.New("multigraph: select increasing, positive years and exactly one type selector")
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
		selectedType[i] = name == o.Type && o.Type != "" || o.TypePrefix != "" && strings.HasPrefix(name, o.TypePrefix)
	}
	count := 0
	var yearCounts [2]int
	for i, r := range s.rows {
		if i&0xfffff == 0 {
			if err := ctx.Err(); err != nil {
				return stats, err
			}
		}
		if (r.year == o.Earlier || r.year == o.Later) && selectedType[r.kind] {
			count++
			if r.year == o.Earlier {
				yearCounts[0]++
			} else {
				yearCounts[1]++
			}
		}
	}
	if !o.AllowEmptyYear && (yearCounts[0] == 0 || yearCounts[1] == 0) {
		return stats, errors.New("multigraph: selected type has no relationships in one requested year")
	}
	if uint64(len(s.rows)) > uint64(^uint32(0)) {
		return stats, errors.New("multigraph: annual analysis supports up to 2^32-1 stored relationships")
	}
	ids := make([]uint32, count)
	j := 0
	sorted := true
	var previous relation
	for i, r := range s.rows {
		if (r.year != o.Earlier && r.year != o.Later) || !selectedType[r.kind] {
			continue
		}
		ids[j] = uint32(i)
		if j > 0 && annualRelationCompare(previous, r, s.types) > 0 {
			sorted = false
		}
		previous = r
		j++
	}
	if !sorted {
		slices.SortFunc(ids, func(a, b uint32) int {
			return annualRelationCompare(s.rows[a], s.rows[b], s.types)
		})
	}
	stats.SelectedRelationships, stats.SortedInput = count, sorted
	providers := make(map[int32]*annualProviderState)
	process := func(fn func(int32, int32, uint32, [2]AnnualMetric) error) error {
		for i := 0; i < len(ids); {
			if i&0xfffff == 0 {
				if err := ctx.Err(); err != nil {
					return err
				}
			}
			r := s.rows[ids[i]]
			var metric [2]AnnualMetric
			for i < len(ids) {
				row := s.rows[ids[i]]
				if row.from != r.from || row.to != r.to || row.kind != r.kind {
					break
				}
				k := 0
				if row.year == o.Later {
					k = 1
				}
				m := &metric[k]
				m.Observed = true
				m.Relationships++
				value, present := 1.0, true
				if o.Weight != "" {
					value, present = column.at(int(ids[i]))
				}
				if present {
					if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
						return fmt.Errorf("multigraph: annual weight %q on relationship %d must be finite and nonnegative", o.Weight, ids[i])
					}
					m.Present = true
					m.Value += value
				} else {
					m.Missing++
				}
				i++
			}
			for k := range metric {
				metric[k].Complete = metric[k].Observed && metric[k].Missing == 0
			}
			if err := fn(r.from, r.to, r.kind, metric); err != nil {
				return err
			}
		}
		return nil
	}
	err := process(func(from, to int32, kind uint32, m [2]AnnualMetric) error {
		p := providers[from]
		if p == nil {
			p = &annualProviderState{}
			providers[from] = p
		}
		for k := range m {
			if m[k].Observed {
				p.degree[k]++
				p.metric[k].Observed = true
				p.metric[k].Relationships += m[k].Relationships
				p.metric[k].Missing += m[k].Missing
				if m[k].Present {
					p.metric[k].Present = true
					p.metric[k].Value += m[k].Value
				}
				p.squares[k] += m[k].Value * m[k].Value
			}
		}
		if !m[0].Observed {
			p.new++
			stats.NewEdges++
		}
		if !m[1].Observed {
			p.lost++
			stats.LostEdges++
		}
		stats.Edges++
		return nil
	})
	if err != nil {
		return stats, err
	}
	providerIDs := make([]int32, 0, len(providers))
	for id := range providers {
		providerIDs = append(providerIDs, id)
	}
	slices.Sort(providerIDs)
	var peers, turnoverPeers, hhiPeers [33]annualPeer
	for _, id := range providerIDs {
		p := providers[id]
		bin := annualBin(p.degree[0])
		turnoverPeers[bin].add(float64(p.new+p.lost) / float64(p.degree[0]+p.degree[1]))
		for k := range p.metric {
			p.metric[k].Complete = p.metric[k].Observed && p.metric[k].Missing == 0
		}
		if p.metric[0].Complete && p.metric[1].Complete {
			peers[bin].add(math.Log1p(p.metric[1].Value) - math.Log1p(p.metric[0].Value))
			if p.metric[0].Value > 0 && p.metric[1].Value > 0 {
				earlierHHI := p.squares[0] / (p.metric[0].Value * p.metric[0].Value)
				laterHHI := p.squares[1] / (p.metric[1].Value * p.metric[1].Value)
				hhiPeers[bin].add(laterHHI - earlierHHI)
			}
		}
	}
	stats.Providers = len(providerIDs)
	if visitProvider != nil {
		for _, id := range providerIDs {
			if err := ctx.Err(); err != nil {
				return stats, err
			}
			p := providers[id]
			result := AnnualProviderChange{Provider: s.nodeIDs[id], EarlierDegree: p.degree[0], LaterDegree: p.degree[1], NewConnections: p.new, LostConnections: p.lost, Earlier: p.metric[0], Later: p.metric[1]}
			bin := annualBin(p.degree[0])
			result.TurnoverRate = float64(p.new+p.lost) / float64(p.degree[0]+p.degree[1])
			turnoverPeer := turnoverPeers[bin]
			result.PeerTurnoverCount, result.PeerTurnoverMean = turnoverPeer.n, turnoverPeer.mean
			if turnoverPeer.n > 1 {
				result.PeerTurnoverSD = math.Sqrt(turnoverPeer.m2 / float64(turnoverPeer.n-1))
			}
			if result.PeerTurnoverSD > 0 {
				result.HasPeerTurnoverZ = true
				result.PeerTurnoverZ = (result.TurnoverRate - result.PeerTurnoverMean) / result.PeerTurnoverSD
			}
			if result.Earlier.Complete && result.Later.Complete {
				result.HasMetricChange = true
				result.LogChange = math.Log1p(result.Later.Value) - math.Log1p(result.Earlier.Value)
				peer := peers[annualBin(p.degree[0])]
				result.PeerCount, result.PeerMean = peer.n, peer.mean
				if peer.n > 1 {
					result.PeerSD = math.Sqrt(peer.m2 / float64(peer.n-1))
				}
				if result.PeerSD > 0 {
					result.HasPeerZ = true
					result.PeerZ = (result.LogChange - result.PeerMean) / result.PeerSD
				}
				if result.Earlier.Value > 0 && result.Later.Value > 0 {
					result.HasConcentration = true
					result.EarlierHHI = p.squares[0] / (result.Earlier.Value * result.Earlier.Value)
					result.LaterHHI = p.squares[1] / (result.Later.Value * result.Later.Value)
					result.HHIShift = result.LaterHHI - result.EarlierHHI
					hhiPeer := hhiPeers[bin]
					result.PeerHHICount, result.PeerHHIMean = hhiPeer.n, hhiPeer.mean
					if hhiPeer.n > 1 {
						result.PeerHHISD = math.Sqrt(hhiPeer.m2 / float64(hhiPeer.n-1))
					}
					if result.PeerHHISD > 0 {
						result.HasPeerHHIZ = true
						result.PeerHHIZ = (result.HHIShift - result.PeerHHIMean) / result.PeerHHISD
					}
				}
			}
			if err := visitProvider(result); err != nil {
				return stats, err
			}
		}
	}
	if visitEdge != nil {
		err = process(func(from, to int32, kind uint32, m [2]AnnualMetric) error {
			p := providers[from]
			result := AnnualEdgeChange{Provider: s.nodeIDs[from], Entity: s.nodeIDs[to], Type: s.types[kind], Earlier: m[0], Later: m[1], New: !m[0].Observed, Lost: !m[1].Observed}
			if m[0].Complete && m[1].Complete && p.metric[0].Complete && p.metric[1].Complete {
				result.HasAdjustedChange = true
				result.LogChange = math.Log1p(m[1].Value) - math.Log1p(m[0].Value)
				providerChange := math.Log1p(p.metric[1].Value) - math.Log1p(p.metric[0].Value)
				result.AdjustedLogChange = result.LogChange - providerChange
				if p.metric[0].Value > 0 && p.metric[1].Value > 0 {
					result.HasShareShift = true
					result.ShareShift = m[1].Value/p.metric[1].Value - m[0].Value/p.metric[0].Value
				}
			}
			return visitEdge(result)
		})
	}
	return stats, err
}

func annualRelationCompare(a, b relation, types []string) int {
	if a.from != b.from {
		if a.from < b.from {
			return -1
		}
		return 1
	}
	if a.to != b.to {
		if a.to < b.to {
			return -1
		}
		return 1
	}
	if a.kind != b.kind {
		if types[a.kind] < types[b.kind] {
			return -1
		}
		return 1
	}
	if a.year != b.year {
		if a.year < b.year {
			return -1
		}
		return 1
	}
	return 0
}
