// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package densegraph

import (
	"math"
	"runtime"
	"slices"
	"testing"
)

func compactFixture(n int, opts ...BuildOption) *Undirected {
	return compactWeightedFixture(n, false, opts...)
}

func compactWeightedFixture(n int, outlier bool, opts ...BuildOption) *Undirected {
	b := NewUndirectedBuilderSized(n, n*32)
	for i := range n {
		b.AddNode(compactID(i))
	}
	for i := range n {
		for j := i + 1; j < min(n, i+17); j++ {
			w := float64(1 + (i+j)%200)
			if outlier && i == n/2 && j == i+1 {
				w = 0.125
			}
			b.AddWeightedEdge(compactID(i), compactID(j), w)
		}
	}
	return b.Build(opts...)
}

func compactID(i int) int64 { return -1<<40 + int64(i)*1000003 }

func TestCompactGraphExact(t *testing.T) {
	for _, tc := range compactBuildModes() {
		t.Run(tc.name, func(t *testing.T) {
			flat := compactWeightedFixture(256, true)
			packed := compactWeightedFixture(256, true, tc.opts...)
			if packed.StorageBytes() > flat.StorageBytes() {
				t.Fatal("optional encoding increased storage")
			}
			checkCompactGraphExact(t, flat, packed)
		})
	}
}

func checkCompactGraphExact(t *testing.T, flat, packed *Undirected) {
	t.Helper()
	for i, id := range flat.NodeIDs() {
		want := collectIDs(flat.From(id))
		if got := collectIDs(packed.From(id)); !slices.Equal(got, want) {
			t.Fatalf("From(%d) differs", id)
		}
		if !slices.Equal(flat.DenseNeighbors(i), packed.DenseNeighbors(i)) {
			t.Fatalf("DenseNeighbors(%d) differs", i)
		}
		var streamed []int32
		for j := range packed.Neighbors(i) {
			streamed = append(streamed, j)
		}
		if !slices.Equal(streamed, flat.DenseNeighbors(i)) {
			t.Fatal("streamed neighbors differ")
		}
		for j, w := range packed.WeightedNeighbors(i) {
			want, ok := flat.Weight(id, flat.NodeIDs()[j])
			if !ok || w != want {
				t.Fatal("streamed weight differs")
			}
		}
		for _, dst := range flat.NodeIDs() {
			if flat.HasEdgeBetween(id, dst) != packed.HasEdgeBetween(id, dst) {
				t.Fatal("edge existence differs")
			}
			w1, ok1 := flat.Weight(id, dst)
			w2, ok2 := packed.Weight(id, dst)
			if ok1 != ok2 || math.Float64bits(w1) != math.Float64bits(w2) {
				t.Fatal("weight differs")
			}
		}
	}
	for _, g := range []*Undirected{flat, packed} {
		cursor, other := g.NewNeighborCursor(), g.NewNeighborCursor()
		weighted := g.NewWeightedCursor()
		for i := range g.NumNodes() {
			cursor.Reset(i)
			weighted.Reset(i)
			other.Reset(i)
			for _, want := range g.DenseNeighbors(i) {
				if !cursor.Next() || !other.Next() || !weighted.Next() || cursor.Neighbor() != want || other.Neighbor() != want || weighted.Neighbor() != want {
					t.Fatal("cursor differs")
				}
				w, _ := g.Weight(g.NodeIDs()[i], g.NodeIDs()[want])
				if weighted.Weight() != w {
					t.Fatal("cursor weight differs")
				}
			}
			if cursor.Next() || other.Next() || weighted.Next() {
				t.Fatal("cursor not exhausted")
			}
		}
		cursor.Reset(math.MaxInt)
		if cursor.Next() {
			t.Fatal("invalid cursor row")
		}
		weighted.Reset(-1)
		if weighted.Next() {
			t.Fatal("invalid weighted cursor row")
		}
		if g.From(math.MaxInt64).Next() || g.HasEdgeBetween(compactID(0), math.MaxInt64) {
			t.Fatal("unknown ID found")
		}
		seq := g.Neighbors(0)
		calls := 0
		for range seq {
			calls++
			break
		}
		if calls != 1 {
			t.Fatal("early stop failed")
		}
		calls = 0
		for range seq {
			calls++
		}
		if calls != len(g.DenseNeighbors(0)) {
			t.Fatal("sequence not reusable")
		}
		calls = 0
		for range g.WeightedNeighbors(0) {
			calls++
			break
		}
		if calls != 1 {
			t.Fatal("weighted early stop failed")
		}
		for range g.Neighbors(math.MaxInt) {
			t.Fatal("invalid row yielded neighbor")
		}
	}
}

func TestWeightLookupDoesNotAllocate(t *testing.T) {
	for _, tc := range compactBuildModes() {
		g := compactWeightedFixture(128, true, tc.opts...)
		if n := testing.AllocsPerRun(100, func() {
			w, ok := g.Weight(compactID(0), compactID(1))
			if !ok || w != 2 {
				panic("incorrect weight")
			}
		}); n != 0 {
			t.Fatalf("%s Weight allocated %g objects", tc.name, n)
		}
	}
}

func TestOptionalStorageSelection(t *testing.T) {
	flat := compactWeightedFixture(256, true)
	packed := compactWeightedFixture(256, true, PackNeighbors24)
	blocks := compactWeightedFixture(256, true, BlockWeights)
	combined := compactWeightedFixture(256, true, PackNeighbors24, BlockWeights)
	delta := compactWeightedFixture(256, true, CompressNeighbors)
	entries := 0
	flat.ScanWeightedEdges(func(_, _ int64, _ float64) { entries++ })
	if flat.weights.StorageBytes() != int64(entries)*8 {
		t.Fatal("default should use graph-wide float64 weights for this fixture")
	}
	if flat.StorageBytes()-packed.StorageBytes() != int64(entries) {
		t.Fatal("24-bit packing should save exactly one byte per directed entry")
	}
	if blocks.weights.StorageBytes() >= flat.weights.StorageBytes() {
		t.Fatal("block weights should isolate the fractional outlier")
	}
	if combined.StorageBytes() != blocks.StorageBytes()-int64(entries) {
		t.Fatal("options did not compose independently")
	}
	for _, tc := range []struct {
		opts []BuildOption
		want int64
	}{
		{[]BuildOption{CompressNeighbors, PackNeighbors24}, packed.StorageBytes()},
		{[]BuildOption{PackNeighbors24, CompressNeighbors}, delta.StorageBytes()},
	} {
		g := compactWeightedFixture(256, true, tc.opts...)
		if g.StorageBytes() != tc.want {
			t.Fatal("last neighbor option should win")
		}
	}
	for _, tc := range compactBuildModes() {
		empty := NewUndirectedBuilder().Build(tc.opts...)
		if empty.NumNodes() != 0 || empty.From(1).Next() {
			t.Fatal("nonempty graph")
		}
		unit := NewUndirectedBuilder()
		for i := range 256 {
			unit.AddEdge(int64(i), int64(i+1))
		}
		if unit.Build(tc.opts...).weights.StorageBytes() != 0 {
			t.Fatal("optional encoding should preserve implicit unit weights")
		}
	}
}

func TestCompactWeightsAndDuplicateOrder(t *testing.T) {
	b := NewUndirectedBuilder()
	for i := range 200 {
		b.AddWeightedEdge(-7, 999, float64(i))
	}
	b.AddWeightedEdge(-7, -7, 4.5)
	b.AddNode(math.MaxInt64)
	g := b.Build(CompressNeighbors)
	if w, _ := g.Weight(-7, 999); w != 199 {
		t.Fatalf("last weight lost: %g", w)
	}
	if e := g.WeightedEdge(-7, -7); e == nil || e.Weight() != 4.5 {
		t.Fatal("stored self weight lost")
	}
	for j, w := range g.WeightedNeighbors(0) {
		if j == 0 && w != 0 {
			t.Fatal("self weight inconsistent with Weight")
		}
	}
	if g.From(math.MaxInt64).Next() {
		t.Fatal("isolated node has neighbors")
	}
	var self float64
	g.ScanWeightedEdges(func(src, dst int64, w float64) {
		if src == dst {
			self = w
		}
	})
	if self != 4.5 {
		t.Fatal("scan should preserve stored self weight")
	}
}

func TestUnitWeightsHaveNoEdgeArray(t *testing.T) {
	b := NewUndirectedBuilder()
	for i := range 100 {
		b.AddEdge(int64(i), int64(i+1))
	}
	g := b.Build()
	if g.weights.StorageBytes() != 0 {
		t.Fatal("unit weights should be implicit")
	}
	if g.StorageBytes() >= 20*200+4*102+8*101 {
		t.Fatal("no memory improvement over original layout")
	}
}

func compactBuildModes() []struct {
	name string
	opts []BuildOption
} {
	return []struct {
		name string
		opts []BuildOption
	}{
		{"flat", nil},
		{"compressed", []BuildOption{CompressNeighbors}},
		{"packed24", []BuildOption{PackNeighbors24}},
		{"flat-blocks", []BuildOption{BlockWeights}},
		{"compressed-blocks", []BuildOption{CompressNeighbors, BlockWeights}},
		{"packed24-blocks", []BuildOption{PackNeighbors24, BlockWeights}},
	}
}

func BenchmarkCompactGraph(b *testing.B) {
	benchmarkCompactGraph(b, false)
}

// A single fractional edge otherwise makes the whole graph use float64 weights.
func BenchmarkCompactGraphOutlier(b *testing.B) {
	benchmarkCompactGraph(b, true)
}

func benchmarkCompactGraph(b *testing.B, outlier bool) {
	for _, tc := range compactBuildModes() {
		name, opts := tc.name, tc.opts
		g := compactWeightedFixture(10000, outlier, opts...)
		b.Run(name+"/scan", func(b *testing.B) {
			b.ReportAllocs()
			b.ReportMetric(float64(g.StorageBytes()), "storage-bytes")
			var sum float64
			for range b.N {
				for i := range g.NumNodes() {
					for _, w := range g.WeightedNeighbors(i) {
						sum += w
					}
				}
			}
			if sum == 0 {
				b.Fatal("empty scan")
			}
		})
		b.Run(name+"/build", func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				result := compactWeightedFixture(10000, outlier, opts...)
				if result.NumNodes() != 10000 {
					b.Fatal("wrong count")
				}
				runtime.KeepAlive(result)
			}
		})
	}
}
