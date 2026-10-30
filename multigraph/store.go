// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph

import (
	"iter"
	"math"
	"slices"

	"github.com/intelligrit/graphwizard/internal/compact"
)

// Attributes identifies a relationship's category and observation year. Empty
// Type and zero Year may be used for untyped and undated relationships.
type Attributes struct {
	Type string
	Year int32
}

// Relationship is a copy of one stored relationship's identity and metadata.
// ID is unique within its Store, including among otherwise identical records.
type Relationship struct {
	ID       int64
	From, To int64
	Attributes
}

// Endpoints and type names are interned once. IDs are implicit row positions.
type relation struct {
	from, to int32
	kind     uint32
	year     int32
}

type weightColumn struct {
	values  compact.Weights
	present []uint64 // nil means all relationships have this weight
}

// Store preserves every relationship and named weight, including missing values
// and exact float64 bit patterns. Read operations and independent projections
// are safe concurrently. Build a Store with NewBuilder.
type Store struct {
	nodeIDs []int64
	rows    []relation
	types   []string
	weights map[string]weightColumn
}

// NumNodes includes isolated nodes registered with Builder.AddNode.
func (s *Store) NumNodes() int { return len(s.nodeIDs) }

// NumRelationships counts records, not distinct endpoint pairs.
func (s *Store) NumRelationships() int { return len(s.rows) }

// NodeIDs returns the sorted original IDs. The caller must not modify the slice.
func (s *Store) NodeIDs() []int64 { return s.nodeIDs }

// WeightNames returns an independent, sorted list of the stored weight names.
func (s *Store) WeightNames() []string {
	names := make([]string, 0, len(s.weights))
	for name := range s.weights {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Relationship returns a record by ID. IDs are stable for the life of the store.
func (s *Store) Relationship(id int64) (Relationship, bool) {
	if id < 0 || id >= int64(len(s.rows)) {
		return Relationship{}, false
	}
	r := s.rows[int(id)]
	return Relationship{
		ID: id, From: s.nodeIDs[r.from], To: s.nodeIDs[r.to],
		Attributes: Attributes{Type: s.types[r.kind], Year: r.year},
	}, true
}

// Relationships visits records in insertion order. The sequence is reusable;
// breaking early does not change subsequent reads.
func (s *Store) Relationships() iter.Seq[Relationship] {
	return func(yield func(Relationship) bool) {
		for i := range s.rows {
			r, _ := s.Relationship(int64(i))
			if !yield(r) {
				return
			}
		}
	}
}

// Weight returns the exact value for a relationship and name. A missing name,
// missing value, or invalid ID returns (0, false); a stored zero returns true.
func (s *Store) Weight(id int64, name string) (float64, bool) {
	c, ok := s.weights[name]
	if !ok || id < 0 || id >= int64(len(s.rows)) {
		return 0, false
	}
	return c.at(int(id))
}

func (c *weightColumn) at(i int) (float64, bool) {
	if c.present != nil && c.present[i/64]&(uint64(1)<<uint(i%64)) == 0 {
		return 0, false
	}
	return c.values.At(i), true
}

// StorageBytes reports retained numeric array payloads. It excludes strings,
// dictionaries, object headers, builder buffers, and projected graphs. Each
// relationship uses a 16-byte metadata row and stores each weight only once,
// even when its eventual projection is undirected.
func (s *Store) StorageBytes() int64 {
	size := int64(cap(s.nodeIDs))*8 + int64(cap(s.rows))*16
	for _, c := range s.weights {
		size += c.values.StorageBytes() + int64(cap(c.present))*8
	}
	return size
}

type columnBuilder struct {
	values  []float64
	present []uint64
	count   int
	first   float64
}

// Builder collects relationships and named weights. Every AddRelationship call
// creates a distinct record, even if endpoints, attributes, and weights match.
// It is not safe for concurrent writes and must not be reused after Build.
// The zero value is ready to use.
type Builder struct {
	nodeIndex map[int64]int32
	rows      []relation
	types     []string
	typeIndex map[string]uint32
	columns   map[string]*columnBuilder
	edgeHint  int
	built     bool
}

// NewBuilder creates an empty relationship builder.
func NewBuilder() *Builder { return NewBuilderSized(0, 0) }

// NewBuilderSized preallocates for nodeHint nodes and relationshipHint records.
// Count each relationship once, not twice for an undirected projection. Each
// encountered weight name also preallocates its column using this hint.
func NewBuilderSized(nodeHint, relationshipHint int) *Builder {
	return &Builder{
		nodeIndex: make(map[int64]int32, nodeHint),
		rows:      make([]relation, 0, relationshipHint), edgeHint: relationshipHint,
		typeIndex: make(map[string]uint32), columns: make(map[string]*columnBuilder),
	}
}

func (b *Builder) init() {
	if b.built {
		panic("multigraph: builder reused after Build")
	}
	if b.nodeIndex == nil {
		b.nodeIndex = make(map[int64]int32)
		b.typeIndex = make(map[string]uint32)
		b.columns = make(map[string]*columnBuilder)
	}
}

func (b *Builder) index(id int64) int32 {
	if i, ok := b.nodeIndex[id]; ok {
		return i
	}
	if len(b.nodeIndex) >= math.MaxInt32 {
		panic("multigraph: more than MaxInt32 nodes")
	}
	i := int32(len(b.nodeIndex))
	b.nodeIndex[id] = i
	return i
}

// AddNode retains a node even if it has no relationships or a projection filters
// out all of its relationships.
func (b *Builder) AddNode(id int64) {
	b.init()
	b.index(id)
}

// DefineWeight registers a named metric even if all its values are missing.
// Existing relationships receive missing values. Repeated calls are harmless.
// Names passed to AddRelationship are also registered automatically.
func (b *Builder) DefineWeight(name string) {
	b.init()
	if name == "" {
		panic("multigraph: empty weight name")
	}
	if _, ok := b.columns[name]; ok {
		return
	}
	n := len(b.rows)
	b.columns[name] = &columnBuilder{
		values:  make([]float64, n, max(n, b.edgeHint)),
		present: make([]uint64, (n+63)/64),
	}
}

// AddRelationship copies the supplied weights and returns the new record ID.
// An omitted weight is missing, not zero. A nil map is valid. Empty weight names
// panic because the empty name is reserved for unit-weight projections.
// Nonfinite values and signed zeros are preserved exactly in the store.
func (b *Builder) AddRelationship(from, to int64, attrs Attributes, weights map[string]float64) int64 {
	b.init()
	if _, bad := weights[""]; bad {
		panic("multigraph: empty weight name")
	}
	for name := range weights {
		b.DefineWeight(name)
	}
	kind, ok := b.typeIndex[attrs.Type]
	if !ok {
		if uint64(len(b.types)) > math.MaxUint32 {
			panic("multigraph: too many relationship types")
		}
		kind = uint32(len(b.types))
		b.typeIndex[attrs.Type] = kind
		b.types = append(b.types, attrs.Type)
	}
	i := len(b.rows)
	b.rows = append(b.rows, relation{from: b.index(from), to: b.index(to), kind: kind, year: attrs.Year})
	for name, c := range b.columns {
		value, present := weights[name]
		c.values = append(c.values, value)
		if len(c.present) < (i+64)/64 {
			c.present = append(c.present, 0)
		}
		if present {
			if c.count == 0 {
				c.first = value
			}
			c.present[i/64] |= uint64(1) << uint(i%64)
			c.count++
		}
	}
	return int64(i)
}

type buildConfig struct{ blockWeights bool }

// BuildOption selects the immutable store's weight representation.
type BuildOption func(*buildConfig)

// BlockWeights opts into exact per-128-record weight encoding independently for
// every name. A column keeps its global encoding unless blocks save payload.
func BlockWeights(c *buildConfig) { c.blockWeights = true }

// Build freezes the relationship table. Weight columns use the same exact
// constant/integer/float64 encodings as densegraph. Missing values use a bitmap
// only in columns that need one. Build consumes the builder.
func (b *Builder) Build(opts ...BuildOption) *Store {
	b.init()
	var cfg buildConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	s := &Store{rows: b.rows, types: b.types, weights: make(map[string]weightColumn, len(b.columns))}
	s.nodeIDs = make([]int64, 0, len(b.nodeIndex))
	for id := range b.nodeIndex {
		s.nodeIDs = append(s.nodeIDs, id)
	}
	slices.Sort(s.nodeIDs)
	remap := make([]int32, len(s.nodeIDs))
	for i, id := range s.nodeIDs {
		remap[b.nodeIndex[id]] = int32(i)
	}
	for i := range s.rows {
		r := &s.rows[i]
		r.from, r.to = remap[r.from], remap[r.to]
	}
	b.nodeIndex = nil
	encode := compact.NewWeights
	if cfg.blockWeights {
		encode = compact.NewBlockWeights
	}
	for name, c := range b.columns {
		col := weightColumn{values: encode(len(s.rows), func(i int) float64 {
			// Absent slots are never exposed. Fill them with an existing value
			// so missing entries cannot widen an otherwise constant column.
			if c.present[i/64]&(uint64(1)<<uint(i%64)) == 0 {
				return c.first
			}
			return c.values[i]
		})}
		if c.count != len(s.rows) {
			col.present = make([]uint64, len(c.present))
			copy(col.present, c.present)
		}
		s.weights[name] = col
		delete(b.columns, name)
	}
	*b = Builder{built: true}
	return s
}
