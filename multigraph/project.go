// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"slices"

	"github.com/intelligrit/graphwizard/densegraph"
)

// Projection errors can be inspected with errors.Is.
var (
	ErrParallelRelationships = errors.New("multigraph: parallel relationships")
	ErrMissingWeight         = errors.New("multigraph: missing weight")
	ErrUnknownWeight         = errors.New("multigraph: unknown weight")
)

// Combine controls how a projection merges selected relationships connecting
// the same pair. The zero value rejects parallel relationships.
type Combine uint8

const (
	RejectParallel Combine = iota
	Sum
	Minimum
	Maximum
)

// Missing controls selected relationships without the requested named weight.
// The zero value reports an error rather than conflating absence with zero.
type Missing uint8

const (
	ErrorMissing Missing = iota
	SkipMissing
	ZeroMissing
)

// Projection selects the meaning of one scalar-weight analytical graph.
type Projection struct {
	// Weight is a named metric. Empty selects a unit weight per relationship;
	// combined with Sum this counts relationships, independent of their weights.
	Weight string
	// Filter selects records by ID, endpoints, type, or year. Nil selects all.
	// Filtering occurs before weight lookup or parallel-edge reduction.
	Filter  func(Relationship) bool
	Combine Combine
	Missing Missing
}

type projectedEdge struct {
	from, to int32
	weight   float64
}

// ProjectUndirected materializes a compact gonum-compatible graph. All store
// nodes survive as possible isolates. Opposite orientations and self edges are
// grouped by unordered endpoint pair; self relationships count only once.
//
// Unknown weight names always return an error. By default, missing weights and
// parallel relationships also return errors. A zero weight remains an edge.
// Sum, Minimum, and Maximum use ordinary float64 arithmetic in insertion order;
// combining values is an explicit transformation, not lossless compression.
// Algorithm-specific weight restrictions still apply to the resulting graph.
//
// This allocates a new graph and temporary sorting/build buffers; it is not a
// zero-copy view. The store remains unchanged. The supplied densegraph options
// apply only to this projection. Gonum's zero self-weight convention is retained.
func (s *Store) ProjectUndirected(p Projection, opts ...densegraph.BuildOption) (*densegraph.Undirected, error) {
	if p.Combine > Maximum {
		return nil, fmt.Errorf("multigraph: invalid combination policy %d", p.Combine)
	}
	if p.Missing > ZeroMissing {
		return nil, fmt.Errorf("multigraph: invalid missing-weight policy %d", p.Missing)
	}
	var column weightColumn
	if p.Weight != "" {
		var ok bool
		column, ok = s.weights[p.Weight]
		if !ok {
			return nil, fmt.Errorf("%w %q", ErrUnknownWeight, p.Weight)
		}
	}
	var edges []projectedEdge
	if p.Filter == nil {
		capacity := len(s.rows)
		if p.Weight != "" && p.Missing == SkipMissing && column.present != nil {
			capacity = 0
			for _, mask := range column.present {
				capacity += bits.OnesCount64(mask)
			}
		}
		edges = make([]projectedEdge, 0, capacity)
	}
	for i, r := range s.rows {
		if p.Filter != nil {
			record, _ := s.Relationship(int64(i))
			if !p.Filter(record) {
				continue
			}
		}
		weight := 1.0
		if p.Weight != "" {
			var present bool
			weight, present = column.at(i)
			if !present {
				switch p.Missing {
				case ErrorMissing:
					return nil, fmt.Errorf("%w %q on relationship %d", ErrMissingWeight, p.Weight, i)
				case SkipMissing:
					continue
				case ZeroMissing:
					weight = 0
				}
			}
		}
		edges = append(edges, projectedEdge{min(r.from, r.to), max(r.from, r.to), weight})
	}
	// Stability makes floating-point accumulation independent of map ordering.
	slices.SortStableFunc(edges, func(a, b projectedEdge) int {
		if a.from != b.from {
			if a.from < b.from {
				return -1
			}
			return 1
		}
		if a.to < b.to {
			return -1
		}
		if a.to > b.to {
			return 1
		}
		return 0
	})
	count, directed := 0, 0
	for _, e := range edges {
		if count == 0 || edges[count-1].from != e.from || edges[count-1].to != e.to {
			edges[count] = e
			count++
			directed++
			if e.from != e.to {
				directed++
			}
			continue
		}
		previous := &edges[count-1].weight
		switch p.Combine {
		case RejectParallel:
			return nil, fmt.Errorf("%w between %d and %d; select a combination policy", ErrParallelRelationships, s.nodeIDs[e.from], s.nodeIDs[e.to])
		case Sum:
			*previous += e.weight
		case Minimum:
			*previous = math.Min(*previous, e.weight)
		case Maximum:
			*previous = math.Max(*previous, e.weight)
		}
	}
	b := densegraph.NewUndirectedBuilderSized(len(s.nodeIDs), directed)
	for _, id := range s.nodeIDs {
		b.AddNode(id)
	}
	for _, e := range edges[:count] {
		b.AddWeightedEdge(s.nodeIDs[e.from], s.nodeIDs[e.to], e.weight)
	}
	return b.Build(opts...), nil
}
