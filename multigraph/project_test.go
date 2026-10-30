// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"
	"sync"
	"testing"

	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/densegraph"
	"github.com/intelligrit/graphwizard/multigraph"
	"github.com/intelligrit/graphwizard/paths"
	"gonum.org/v1/gonum/graph"
)

func exampleStore() *multigraph.Store {
	b := multigraph.NewBuilder()
	b.AddNode(99)
	b.AddRelationship(1, 2, multigraph.Attributes{Type: "prescribed", Year: 2022}, map[string]float64{"claims": 2, "dollars": 7})
	b.AddRelationship(2, 1, multigraph.Attributes{Type: "prescribed", Year: 2023}, map[string]float64{"claims": 3, "dollars": 1})
	b.AddRelationship(1, 2, multigraph.Attributes{Type: "referred", Year: 2023}, map[string]float64{"claims": 10})
	b.AddRelationship(2, 3, multigraph.Attributes{Type: "prescribed", Year: 2023}, map[string]float64{"claims": 4, "dollars": 20})
	b.AddRelationship(1, 3, multigraph.Attributes{Type: "prescribed", Year: 2023}, map[string]float64{"claims": 15, "dollars": 2})
	return b.Build(multigraph.BlockWeights)
}

func currentPrescriptions(r multigraph.Relationship) bool {
	return r.Type == "prescribed" && r.Year == 2023
}

func TestSelectedMetricChangesWeightedPaths(t *testing.T) {
	s := exampleStore()
	for _, tc := range []struct {
		metric string
		path   []int64
		weight float64
	}{
		{"claims", []int64{1, 2, 3}, 7},
		{"dollars", []int64{1, 3}, 2},
	} {
		g, err := s.ProjectUndirected(multigraph.Projection{Weight: tc.metric, Filter: currentPrescriptions}, densegraph.PackNeighbors24)
		if err != nil {
			t.Fatal(err)
		}
		var _ graph.WeightedUndirected = g
		nodes, weight := paths.ShortestPath(context.Background(), g, 1, 3)
		var ids []int64
		for _, n := range nodes {
			ids = append(ids, n.ID())
		}
		if !slices.Equal(ids, tc.path) || weight != tc.weight {
			t.Fatalf("%s: %v, %g", tc.metric, ids, weight)
		}
		if g.Node(99) == nil || g.From(99).Next() {
			t.Fatal("isolated node changed")
		}
	}
	if s.NumRelationships() != 5 {
		t.Fatal("projection modified original records")
	}
}

func TestParallelRelationshipsRequireExplicitReduction(t *testing.T) {
	s := exampleStore()
	if g, err := s.ProjectUndirected(multigraph.Projection{Weight: "claims"}); !errors.Is(err, multigraph.ErrParallelRelationships) || g != nil {
		t.Fatal("parallel relationships silently overwritten")
	}
	for _, tc := range []struct {
		combine multigraph.Combine
		want    float64
	}{
		{multigraph.Sum, 15}, {multigraph.Minimum, 2}, {multigraph.Maximum, 10},
	} {
		g, err := s.ProjectUndirected(multigraph.Projection{Weight: "claims", Combine: tc.combine})
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range [][2]int64{{1, 2}, {2, 1}} {
			if w, ok := g.Weight(pair[0], pair[1]); !ok || w != tc.want {
				t.Fatalf("weight=%g want %g", w, tc.want)
			}
		}
	}
	counts, err := s.ProjectUndirected(multigraph.Projection{Combine: multigraph.Sum})
	if err != nil {
		t.Fatal(err)
	}
	if w, _ := counts.Weight(1, 2); w != 3 {
		t.Fatal("unit sum should count relationships")
	}
	selected, err := s.ProjectUndirected(multigraph.Projection{
		Weight: "claims", Filter: func(r multigraph.Relationship) bool { return r.Type == "prescribed" && r.Year == 2022 },
	})
	if err != nil {
		t.Fatal(err)
	}
	if w, _ := selected.Weight(1, 2); w != 2 || selected.HasEdgeBetween(1, 3) {
		t.Fatal("year selection changed")
	}
}

func TestMissingWeightsAndZeroEdges(t *testing.T) {
	b := multigraph.NewBuilder()
	b.AddRelationship(1, 2, multigraph.Attributes{}, map[string]float64{"value": 0})
	b.AddRelationship(2, 3, multigraph.Attributes{}, nil)
	b.AddRelationship(3, 4, multigraph.Attributes{}, map[string]float64{"value": 2})
	s := b.Build()
	if g, err := s.ProjectUndirected(multigraph.Projection{Weight: "value"}); !errors.Is(err, multigraph.ErrMissingWeight) || g != nil {
		t.Fatal("missing weight accepted implicitly")
	}
	for _, policy := range []multigraph.Missing{multigraph.SkipMissing, multigraph.ZeroMissing} {
		g, err := s.ProjectUndirected(multigraph.Projection{Weight: "value", Missing: policy})
		if err != nil {
			t.Fatal(err)
		}
		if w, ok := g.Weight(1, 2); !ok || w != 0 || !g.HasEdgeBetween(1, 2) {
			t.Fatal("stored zero edge lost")
		}
		if g.HasEdgeBetween(2, 3) != (policy == multigraph.ZeroMissing) {
			t.Fatal("missing edge policy ignored")
		}
		if w, ok := g.Weight(2, 3); policy == multigraph.ZeroMissing && (!ok || w != 0) {
			t.Fatal("missing weight not set to zero")
		}
		if g.NumNodes() != 4 {
			t.Fatal("missing weight policy removed nodes")
		}
	}
	if _, err := s.ProjectUndirected(multigraph.Projection{Weight: "value", Filter: func(r multigraph.Relationship) bool { return r.ID != 1 }}); err != nil {
		t.Fatal("weight validated before filtering:", err)
	}
	for _, p := range []multigraph.Projection{
		{Weight: "typo", Missing: multigraph.SkipMissing},
		{Weight: "typo", Filter: func(multigraph.Relationship) bool { return false }},
		{Combine: multigraph.Combine(255)}, {Missing: multigraph.Missing(255)},
	} {
		if g, err := s.ProjectUndirected(p); err == nil || g != nil {
			t.Fatal("invalid projection accepted")
		}
	}
	g, err := s.ProjectUndirected(multigraph.Projection{Weight: "value", Filter: func(multigraph.Relationship) bool { return false }})
	if err != nil || g.NumNodes() != 4 || g.From(1).Next() {
		t.Fatal("empty selection lost nodes")
	}
	var empty multigraph.Store
	g, err = empty.ProjectUndirected(multigraph.Projection{})
	if err != nil || g.NumNodes() != 0 {
		t.Fatal("empty store projection failed")
	}
}

func TestStableReductionAndExactSelectedBits(t *testing.T) {
	b := multigraph.NewBuilder()
	for _, v := range []float64{1e16, -1e16, 1} {
		b.AddRelationship(2, 1, multigraph.Attributes{}, map[string]float64{"v": v})
		b.AddRelationship(3, 4, multigraph.Attributes{}, map[string]float64{"v": 1})
	}
	s := b.Build()
	for range 5 {
		g, err := s.ProjectUndirected(multigraph.Projection{Weight: "v", Combine: multigraph.Sum})
		if err != nil {
			t.Fatal(err)
		}
		if w, _ := g.Weight(1, 2); w != 1 {
			t.Fatal("floating point reduction order changed")
		}
	}
	for _, v := range []float64{math.Copysign(0, -1), math.Inf(1), math.Float64frombits(0x7ff8000000000012)} {
		b := multigraph.NewBuilder()
		b.AddRelationship(-1, 7, multigraph.Attributes{}, map[string]float64{"v": v})
		g, err := b.Build().ProjectUndirected(multigraph.Projection{Weight: "v"}, densegraph.CompressNeighbors, densegraph.BlockWeights)
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := g.Weight(-1, 7); !ok || math.Float64bits(got) != math.Float64bits(v) {
			t.Fatal("uncombined weight bits changed")
		}
	}
}

func TestSelfRelationshipsCountOnce(t *testing.T) {
	b := multigraph.NewBuilder()
	b.AddRelationship(1, 1, multigraph.Attributes{}, map[string]float64{"v": 3})
	b.AddRelationship(1, 1, multigraph.Attributes{}, map[string]float64{"v": 4})
	s := b.Build()
	g, err := s.ProjectUndirected(multigraph.Projection{Weight: "v", Combine: multigraph.Sum})
	if err != nil {
		t.Fatal(err)
	}
	if e := g.WeightedEdge(1, 1); e == nil || e.Weight() != 7 {
		t.Fatal("self relationships double counted")
	}
	if w, ok := g.Weight(1, 1); !ok || w != 0 {
		t.Fatal("gonum self-weight convention changed")
	}
	count := 0
	g.ScanWeightedEdges(func(from, to int64, w float64) {
		count++
		if from != 1 || to != 1 || w != 7 {
			t.Fatal("stored self edge changed")
		}
	})
	if count != 1 {
		t.Fatal("self edge duplicated")
	}
}

func TestProjectionStorageModesAndLeiden(t *testing.T) {
	s := exampleStore()
	manual := densegraph.NewUndirectedBuilder()
	manual.AddNode(99)
	manual.AddWeightedEdge(1, 2, 3)
	manual.AddWeightedEdge(2, 3, 4)
	manual.AddWeightedEdge(1, 3, 15)
	want := manual.Build()
	for _, opts := range [][]densegraph.BuildOption{
		nil, {densegraph.PackNeighbors24}, {densegraph.CompressNeighbors},
		{densegraph.BlockWeights}, {densegraph.PackNeighbors24, densegraph.BlockWeights},
		{densegraph.CompressNeighbors, densegraph.BlockWeights},
	} {
		g, err := s.ProjectUndirected(multigraph.Projection{Weight: "claims", Filter: currentPrescriptions}, opts...)
		if err != nil {
			t.Fatal(err)
		}
		for _, u := range want.NodeIDs() {
			for _, v := range want.NodeIDs() {
				a, oka := g.Weight(u, v)
				b, okb := want.Weight(u, v)
				if oka != okb || a != b || g.HasEdgeBetween(u, v) != want.HasEdgeBetween(u, v) {
					t.Fatal("projected graph differs")
				}
			}
		}
		for seed := range int64(3) {
			expected := community.Leiden(context.Background(), want, 1, rand.New(rand.NewSource(seed)))
			for _, got := range []map[int64]int64{
				community.Leiden(context.Background(), g, 1, rand.New(rand.NewSource(seed))),
				community.LeidenParallel(context.Background(), g, 1, rand.New(rand.NewSource(seed))),
			} {
				if !maps.Equal(got, expected) {
					t.Fatal("Leiden partition differs from manually selected graph")
				}
			}
		}
	}
}

func TestConcurrentProjectionsAndReads(t *testing.T) {
	s := exampleStore()
	var wg sync.WaitGroup
	errors := make(chan error, 8)
	for worker := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metric, want := "claims", 3.0
			if worker%2 == 0 {
				metric, want = "dollars", 1
			}
			for range 10 {
				g, err := s.ProjectUndirected(multigraph.Projection{Weight: metric, Filter: currentPrescriptions}, densegraph.PackNeighbors24)
				if err != nil {
					errors <- err
					return
				}
				if w, ok := g.Weight(1, 2); !ok || w != want {
					errors <- fmt.Errorf("concurrent metric changed")
					return
				}
				for r := range s.Relationships() {
					if _, ok := s.Weight(r.ID, "claims"); !ok {
						errors <- fmt.Errorf("stored value lost")
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

func TestBulkProjectionMatchesIndependentAggregation(t *testing.T) {
	b := multigraph.NewBuilderSized(64, 5000)
	for i := range 64 {
		b.AddNode(int64(i)*1009 - 90000)
	}
	rng := rand.New(rand.NewSource(73))
	expected := make(map[[2]int64]float64)
	for i := range 5000 {
		u, v := int64(rng.Intn(63))*1009-90000, int64(rng.Intn(63))*1009-90000
		attrs := multigraph.Attributes{Type: "prescribed", Year: int32(2020 + i%4)}
		if i%5 == 0 {
			attrs.Type = "referred"
		}
		w := float64(i % 251)
		values := map[string]float64{"other": float64(i) + 0.5}
		if i%7 != 0 {
			values["claims"] = w
		}
		b.AddRelationship(u, v, attrs, values)
		if attrs.Type == "prescribed" && attrs.Year >= 2022 && i%7 != 0 {
			expected[[2]int64{min(u, v), max(u, v)}] += w
		}
	}
	s := b.Build(multigraph.BlockWeights)
	for _, opts := range [][]densegraph.BuildOption{
		nil, {densegraph.PackNeighbors24}, {densegraph.CompressNeighbors},
		{densegraph.BlockWeights}, {densegraph.PackNeighbors24, densegraph.BlockWeights},
		{densegraph.CompressNeighbors, densegraph.BlockWeights},
	} {
		g, err := s.ProjectUndirected(multigraph.Projection{
			Weight: "claims", Combine: multigraph.Sum, Missing: multigraph.SkipMissing,
			Filter: func(r multigraph.Relationship) bool { return r.Type == "prescribed" && r.Year >= 2022 },
		}, opts...)
		if err != nil {
			t.Fatal(err)
		}
		actual := make(map[[2]int64]float64)
		g.ScanWeightedEdges(func(u, v int64, w float64) {
			if u <= v {
				actual[[2]int64{u, v}] = w
			}
		})
		if !maps.Equal(expected, actual) {
			t.Fatal("bulk aggregation changed edges or values")
		}
		if g.NumNodes() != 64 || g.From(63*1009-90000).Next() {
			t.Fatal("bulk projection lost isolated node")
		}
	}
}
