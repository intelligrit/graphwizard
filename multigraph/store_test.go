// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph

import (
	"math"
	"slices"
	"testing"
)

func TestRelationshipsPreserveIdentityAndNamedWeights(t *testing.T) {
	b := NewBuilder()
	b.AddNode(99)
	weights := map[string]float64{"claims": 12, "dollars": 400, "risk": 0}
	attrs := Attributes{Type: "prescribed", Year: 2023}
	first := b.AddRelationship(math.MaxInt64, math.MinInt64, attrs, weights)
	second := b.AddRelationship(math.MaxInt64, math.MinInt64, attrs, weights)
	third := b.AddRelationship(math.MinInt64, math.MaxInt64, Attributes{Type: "paid", Year: 2022}, nil)
	weights["claims"] = 999 // the builder must not retain caller maps
	s := b.Build()
	if first != 0 || second != 1 || third != 2 || s.NumRelationships() != 3 || s.NumNodes() != 3 {
		t.Fatal("distinct relationships or isolated node lost")
	}
	if !slices.Equal(s.NodeIDs(), []int64{math.MinInt64, 99, math.MaxInt64}) {
		t.Fatal("original node IDs changed")
	}
	if got := s.WeightNames(); !slices.Equal(got, []string{"claims", "dollars", "risk"}) {
		t.Fatal("weight names not sorted")
	}
	names := s.WeightNames()
	names[0] = "changed"
	if s.WeightNames()[0] != "claims" {
		t.Fatal("weight name list aliases the store")
	}
	for _, id := range []int64{first, second} {
		r, ok := s.Relationship(id)
		if !ok || r.ID != id || r.From != math.MaxInt64 || r.To != math.MinInt64 || r.Attributes != attrs {
			t.Fatal("relationship metadata changed")
		}
		for name, want := range map[string]float64{"claims": 12, "dollars": 400, "risk": 0} {
			if got, ok := s.Weight(id, name); !ok || got != want {
				t.Fatalf("%s: got %g, present=%t", name, got, ok)
			}
		}
	}
	if _, ok := s.Weight(third, "risk"); ok {
		t.Fatal("missing weight confused with zero")
	}
	if r, _ := s.Relationship(third); r.From != math.MinInt64 || r.To != math.MaxInt64 || r.Type != "paid" || r.Year != 2022 {
		t.Fatal("record orientation or attributes lost")
	}
	for _, id := range []int64{-1, 3, math.MaxInt64} {
		if _, ok := s.Relationship(id); ok {
			t.Fatal("invalid relationship exists")
		}
		if _, ok := s.Weight(id, "claims"); ok {
			t.Fatal("invalid weight exists")
		}
	}
	if _, ok := s.Weight(first, "unknown"); ok {
		t.Fatal("unknown weight exists")
	}
	seq := s.Relationships()
	for r := range seq {
		if r.ID != first {
			t.Fatal("iteration order changed")
		}
		break
	}
	var ids []int64
	for r := range seq {
		ids = append(ids, r.ID)
	}
	if !slices.Equal(ids, []int64{0, 1, 2}) {
		t.Fatal("sequence not reusable")
	}
}

func TestExactColumnsAndMissingBoundaries(t *testing.T) {
	values := []float64{math.Copysign(0, -1), 0, -5, 0.125, math.Inf(1), math.Inf(-1), math.Float64frombits(0x7ff8000000000012), math.Float64frombits(0x7ff0000000000001)}
	for _, blockwise := range []bool{false, true} {
		var b Builder
		for i := range 1025 {
			weights := map[string]float64{"special": values[i%len(values)]}
			if i%3 == 0 {
				weights["sparse"] = values[(i/3)%len(values)]
			}
			if i >= 129 && i%2 != 0 {
				weights["late"] = float64(i)
			}
			b.AddRelationship(int64(i%17), int64(i%23), Attributes{Year: int32(i)}, weights)
		}
		var opts []BuildOption
		if blockwise {
			opts = []BuildOption{BlockWeights}
		}
		s := b.Build(opts...)
		for i := 1024; i >= 0; i-- {
			for name, present := range map[string]bool{"special": true, "sparse": i%3 == 0, "late": i >= 129 && i%2 != 0} {
				got, ok := s.Weight(int64(i), name)
				if ok != present {
					t.Fatalf("%s/%d: missing bitmap changed", name, i)
				}
				if !present {
					continue
				}
				want := values[i%len(values)]
				if name == "sparse" {
					want = values[(i/3)%len(values)]
				}
				if name == "late" {
					want = float64(i)
				}
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("%s/%d: bits changed", name, i)
				}
			}
		}
	}
}

func TestCompactColumnStorage(t *testing.T) {
	const n = 1025
	b := NewBuilderSized(2, n)
	for i := range n {
		w := map[string]float64{"unit": 1, "byte": float64(i % 256)}
		if i%2 == 0 {
			w["partial"] = 1.25
		}
		b.AddRelationship(1, 2, Attributes{}, w)
	}
	s := b.Build()
	// One 16-byte row per relationship; two IDs; one byte-valued column;
	// constants need no value array; only the partial column needs a bitmap.
	want := int64(16*n + 2*8 + n + ((n+63)/64)*8)
	if got := s.StorageBytes(); got != want {
		t.Fatalf("payload=%d want %d", got, want)
	}
	if b.rows != nil || b.nodeIndex != nil || b.columns != nil || b.types != nil {
		t.Fatal("consumed builder retained buffers")
	}
}

func TestBlockColumnsAreOptional(t *testing.T) {
	build := func(opts ...BuildOption) *Store {
		b := NewBuilderSized(2, 1025)
		for i := range 1025 {
			v := float64(i % 200)
			if i == 512 {
				v = 0.125
			}
			b.AddRelationship(1, 2, Attributes{}, map[string]float64{"value": v})
		}
		return b.Build(opts...)
	}
	flat, blocks := build(), build(BlockWeights)
	if blocks.StorageBytes() >= flat.StorageBytes() {
		t.Fatal("blocks did not isolate outlier")
	}
	for i := range 1025 {
		a, _ := flat.Weight(int64(i), "value")
		b, _ := blocks.Weight(int64(i), "value")
		if math.Float64bits(a) != math.Float64bits(b) {
			t.Fatal("optional storage changed values")
		}
	}
}

func TestBuilderValidation(t *testing.T) {
	panics := func(fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Error("expected programmer error to panic")
			}
		}()
		fn()
	}
	b := NewBuilder()
	panics(func() { b.AddRelationship(1, 2, Attributes{}, map[string]float64{"": 1}) })
	s := b.Build()
	if s.NumNodes() != 0 || s.NumRelationships() != 0 {
		t.Fatal("invalid input mutated builder")
	}
	panics(func() { b.AddNode(1) })
	panics(func() { b.AddRelationship(1, 2, Attributes{}, nil) })
	panics(func() { b.Build() })
	var zero Store
	if zero.StorageBytes() != 0 || len(zero.WeightNames()) != 0 {
		t.Fatal("zero store not empty")
	}
	for range zero.Relationships() {
		t.Fatal("zero store has relationship")
	}
}
