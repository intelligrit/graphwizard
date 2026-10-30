// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package compact

import (
	"math"
	"math/rand"
	"slices"
	"strconv"
	"testing"
)

func TestNeighborRoundTrip(t *testing.T) {
	// Rows cross restart boundaries, include empty rows, duplicates, large gaps,
	// and resets from large to small dense indices between rows.
	rng := rand.New(rand.NewSource(42))
	offsets := []int{0}
	var targets []int32
	for i := range 80 {
		n := rng.Intn(100)
		if i%7 == 0 {
			n = 0
		}
		row := make([]int32, n)
		for k := range row {
			row[k] = int32(rng.Intn(500))
		}
		if n > 1 {
			row[n-1] = 0xffffff
		}
		slices.Sort(row)
		targets = append(targets, row...)
		offsets = append(offsets, len(targets))
	}
	flat := NewNeighbors(offsets, targets, FlatNeighbors)
	packed := NewNeighbors(offsets, targets, DeltaNeighbors)
	if packed.flat != nil {
		t.Fatal("fixture should compress")
	}
	if packed.StorageBytes() >= flat.StorageBytes() {
		t.Fatal("compression did not save memory")
	}
	for _, a := range []*Neighbors{flat, packed, NewNeighbors(offsets, targets, Packed24Neighbors)} {
		for i := range len(offsets) - 1 {
			want := targets[offsets[i]:offsets[i+1]]
			if got := a.Row(i); !slices.Equal(got, want) {
				t.Fatalf("row %d differs", i)
			}
			c := a.Cursor(i)
			for k, v := range want {
				if !c.Next() || c.Value() != v || c.Position() != offsets[i]+k {
					t.Fatalf("cursor row=%d k=%d", i, k)
				}
			}
			if c.Next() || c.Next() {
				t.Fatal("cursor not exhausted")
			}
			for _, v := range append(slices.Clone(want), -1, 501, math.MaxInt32) {
				p, ok := slices.BinarySearch(want, v)
				got := a.Find(i, v)
				if ok && got != offsets[i]+p || !ok && got != -1 {
					t.Fatalf("Find(%d,%d)=%d", i, v, got)
				}
			}
		}
		for _, i := range []int{-1, len(offsets) - 1, math.MaxInt} {
			if a.Row(i) != nil || a.Cursor(i).end != 0 || a.Find(i, 0) != -1 {
				t.Fatalf("invalid row %d", i)
			}
		}
	}
}

func TestCompressionFallback(t *testing.T) {
	for _, values := range [][]int32{nil, {7}, {0, math.MaxInt32, 0, math.MaxInt32}} {
		a := NewNeighbors([]int{0, len(values)}, values, DeltaNeighbors)
		if len(a.data) != 0 || len(a.bases) != 0 {
			t.Fatal("expanded payload should remain flat")
		}
		if !slices.Equal(a.Row(0), values) {
			t.Fatal("fallback changed neighbors")
		}
	}
}

func TestPacked24BoundariesAndFallback(t *testing.T) {
	for _, values := range [][]int32{
		nil, {0}, {0, 255, 256, 65535, 65536, 0xffffff, 0xffffff},
		{0, 1 << 24}, {0, math.MaxInt32}, {-1, 0},
	} {
		a := NewNeighbors([]int{0, 0, len(values), len(values)}, values, Packed24Neighbors)
		if !slices.Equal(a.Row(1), values) || a.Row(0) != nil || a.Row(2) != nil {
			t.Fatalf("round trip changed %v", values)
		}
		fits := len(values) != 0 && values[0] >= 0 && values[len(values)-1] <= 0xffffff
		if fits != (a.packed24 != nil) {
			t.Fatalf("wrong format for %v", values)
		}
		width := 4
		if fits {
			width = 3
			if a.flat != nil || a.data != nil || a.bases != nil || a.positions != nil {
				t.Fatal("packed storage retained another neighbor representation")
			}
		}
		if got, want := a.StorageBytes(), int64(4*(strconv.IntSize/8)+width*len(values)); got != want {
			t.Fatalf("bytes=%d want %d", got, want)
		}
		for _, v := range append(slices.Clone(values), -2, 257, 1<<24) {
			p, ok := slices.BinarySearch(values, v)
			got := a.Find(1, v)
			if ok && got != p || !ok && got != -1 {
				t.Fatalf("Find(%d)=%d in %v", v, got, values)
			}
		}
	}
}

func TestBlockWeightsMixedRepresentations(t *testing.T) {
	values := make([]float64, 10*weightBlockSize+3)
	nan := math.Float64frombits(0x7ff8000000000012)
	exceptions := []float64{math.Copysign(0, -1), 0, -2, 0.125, math.Inf(1), math.Inf(-1), nan, math.Float64frombits(0x7ff0000000000001)}
	for i := range values {
		switch i / weightBlockSize {
		case 0:
			values[i] = 1.25
		case 1, 7, 10:
			values[i] = float64(i % 256)
		case 2:
			values[i] = float64(65535 - i%weightBlockSize)
		case 3:
			values[i] = float64(math.MaxUint32 - i%weightBlockSize)
		case 4:
			values[i] = exceptions[i%len(exceptions)]
		case 5:
			values[i] = nan
		case 6:
			values[i] = math.Copysign(0, -1)
		case 8:
			values[i] = float64(math.MaxUint32 + 1 + i)
		case 9:
			values[i] = 1
		}
	}
	w := NewBlockWeights(len(values), func(i int) float64 { return values[i] })
	if !slices.Equal(w.blockKinds, []byte{0, 1, 2, 4, 8, 0, 0, 1, 8, 0, 1}) {
		t.Fatalf("unexpected block representations: %v", w.blockKinds)
	}
	if w.u8 != nil || w.u16 != nil || w.u32 != nil || w.f64 != nil {
		t.Fatal("block weights retained a global weight array")
	}
	wantBytes := int64(11*(strconv.IntSize/8+1) + 4*8 + weightBlockSize*(1+2+4+8+1+8) + 3)
	if w.StorageBytes() != wantBytes {
		t.Fatalf("bytes=%d want %d", w.StorageBytes(), wantBytes)
	}
	// Reverse access crosses boundaries without relying on decoder history.
	for i := len(values) - 1; i >= 0; i-- {
		if math.Float64bits(w.At(i)) != math.Float64bits(values[i]) {
			t.Fatalf("weight %d changed bits", i)
		}
	}
}

func TestBlockWeightsSelectionAndExactBits(t *testing.T) {
	rng := rand.New(rand.NewSource(17))
	for _, n := range []int{0, 1, 127, 128, 129, 255, 256, 257, 1025} {
		for _, pattern := range []string{"constant", "byte", "short", "int", "float", "outlier", "constant blocks", "random bits"} {
			t.Run(pattern+"/"+strconv.Itoa(n), func(t *testing.T) {
				values := make([]float64, n)
				for i := range values {
					switch pattern {
					case "constant":
						values[i] = 1
					case "byte", "outlier":
						values[i] = float64(i % 255)
					case "short":
						values[i] = float64(256 + i%256)
					case "int":
						values[i] = float64(65536 + i%256)
					case "float":
						values[i] = float64(i) + 0.5
					case "constant blocks":
						values[i] = float64(i/weightBlockSize) + 0.25
					case "random bits":
						values[i] = math.Float64frombits(rng.Uint64())
					}
				}
				if pattern == "outlier" && n != 0 {
					values[n-1] = -0.25
				}
				get := func(i int) float64 { return values[i] }
				flat, block := NewWeights(n, get), NewBlockWeights(n, get)
				if block.StorageBytes() > flat.StorageBytes() {
					t.Fatal("optional block storage increased retained payload")
				}
				if pattern != "outlier" && pattern != "constant blocks" && block.blockKinds != nil {
					t.Fatal("homogeneous representation should fall back to global encoding")
				}
				for i, v := range values {
					if math.Float64bits(block.At(i)) != math.Float64bits(v) {
						t.Fatalf("weight %d changed bits", i)
					}
				}
			})
		}
	}
}

func TestNodesIterator(t *testing.T) {
	ids := []int64{-100, 10, math.MaxInt64}
	a := NewNeighbors([]int{0, 2}, []int32{0, 2}, DeltaNeighbors)
	it := NewNodes(ids, a.Cursor(0))
	for range 2 {
		if it.Len() != 2 || it.Node() != nil {
			t.Fatal("iterator initial state")
		}
		for k, id := range []int64{-100, math.MaxInt64} {
			if !it.Next() || it.Node().ID() != id || it.Len() != 1-k {
				t.Fatal("iterator value")
			}
		}
		if it.Next() || it.Node() != nil || it.Len() != 0 {
			t.Fatal("iterator end state")
		}
		it.Reset()
	}
}

func TestExactWeights(t *testing.T) {
	for _, tc := range []struct {
		name   string
		values []float64
		bytes  int64
	}{
		{"empty", nil, 0},
		{"unit", []float64{1, 1, 1}, 0},
		{"constant fractional", []float64{1.25, 1.25}, 0},
		{"byte", []float64{0, 1, 255}, 3},
		{"uint16", []float64{0, 256, 65535}, 6},
		{"uint32", []float64{0, 65536, math.MaxUint32}, 12},
		{"large", []float64{0, math.MaxUint32 + 1}, 16},
		{"fractional", []float64{0, 1.25}, 16},
		{"negative", []float64{-1, 1}, 16},
		{"signed zero", []float64{0, math.Copysign(0, -1)}, 16},
		{"nonfinite", []float64{math.Inf(1), math.Inf(-1), math.Float64frombits(0x7ff8000000000012)}, 24},
		{"constant nan", []float64{math.Float64frombits(0x7ff8000000000012), math.Float64frombits(0x7ff8000000000012)}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, build := range []func(int, func(int) float64) Weights{NewWeights, NewBlockWeights} {
				w := build(len(tc.values), func(i int) float64 { return tc.values[i] })
				if w.StorageBytes() != tc.bytes {
					t.Fatalf("bytes=%d want %d", w.StorageBytes(), tc.bytes)
				}
				for i, v := range tc.values {
					if math.Float64bits(w.At(i)) != math.Float64bits(v) {
						t.Fatalf("weight %d changed bits", i)
					}
				}
			}
		})
	}
}

func BenchmarkNeighbors(b *testing.B) {
	const n, degree = 10000, 32
	offsets := make([]int, n+1)
	targets := make([]int32, n*degree)
	for i := range n {
		offsets[i+1] = (i + 1) * degree
		for j := range degree {
			targets[i*degree+j] = int32(j*200 + i%100)
		}
	}
	for _, tc := range []struct {
		name     string
		encoding NeighborEncoding
	}{{"flat", FlatNeighbors}, {"compressed", DeltaNeighbors}, {"packed24", Packed24Neighbors}} {
		name := tc.name
		a := NewNeighbors(offsets, targets, tc.encoding)
		b.Run(name+"/scan", func(b *testing.B) {
			b.ReportAllocs()
			b.ReportMetric(float64(a.StorageBytes()), "storage-bytes")
			var sum int64
			for range b.N {
				for i := range n {
					c := a.Cursor(i)
					for c.Next() {
						sum += int64(c.Value())
					}
				}
			}
			if sum == 0 {
				b.Fatal("empty scan")
			}
		})
		b.Run(name+"/lookup", func(b *testing.B) {
			b.ReportAllocs()
			for k := range b.N {
				i := k % n
				if a.Find(i, int32((k%degree)*200+i%100)) < 0 {
					b.Fatal("missing neighbor")
				}
			}
		})
	}
}
