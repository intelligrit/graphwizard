// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package densegraph

import (
	"math"
	"math/rand"
	"reflect"
	"slices"
	"testing"
)

func TestBuilderRemapsInsertionIndexes(t *testing.T) {
	// Insertion order deliberately differs from signed original-ID order.
	ids := []int64{math.MaxInt64, 7, -100, 1 << 40, math.MinInt64, 0, -1, 92, -9999}
	values := []float64{1, 255, 65536, -1.25, math.Copysign(0, -1), math.Inf(1), math.Float64frombits(0x7ff8000000000012)}
	for _, mode := range compactBuildModes() {
		t.Run(mode.name, func(t *testing.T) {
			b := NewUndirectedBuilderSized(100, 1024)
			b.AddNode(ids[1])
			b.AddNode(ids[4])
			want := make(map[[2]int64]uint64)
			rng := rand.New(rand.NewSource(19))
			for k := range 500 {
				u, v := ids[rng.Intn(len(ids)-1)], ids[rng.Intn(len(ids)-1)]
				w := values[k%len(values)]
				if k%3 == 0 {
					b.AddDirectedEntry(u, v, w)
				} else {
					b.AddWeightedEdge(u, v, w)
					want[[2]int64{v, u}] = math.Float64bits(w)
				}
				want[[2]int64{u, v}] = math.Float64bits(w)
			}
			b.AddNode(ids[len(ids)-1]) // isolated node added after all edges
			g := b.Build(mode.opts...)
			sorted := slices.Clone(ids)
			slices.Sort(sorted)
			if !slices.Equal(g.NodeIDs(), sorted) {
				t.Fatalf("node ID order changed: %v", g.NodeIDs())
			}
			got := make(map[[2]int64]uint64)
			g.ScanWeightedEdges(func(u, v int64, w float64) { got[[2]int64{u, v}] = math.Float64bits(w) })
			if !reflect.DeepEqual(got, want) {
				t.Fatal("remapping changed endpoints, duplicate precedence, or weight bits")
			}
			for i, u := range sorted {
				var neighbors []int64
				var dense []int32
				for j, v := range sorted {
					bits, ok := want[[2]int64{u, v}]
					if g.HasEdgeBetween(u, v) != ok {
						t.Fatal("edge lookup changed")
					}
					if ok {
						neighbors = append(neighbors, v)
						dense = append(dense, int32(j))
						if edge := g.WeightedEdge(u, v); edge == nil || math.Float64bits(edge.Weight()) != bits {
							t.Fatal("weighted edge lookup changed")
						}
					}
				}
				if !slices.Equal(collectIDs(g.From(u)), neighbors) || !slices.Equal(g.DenseNeighbors(i), dense) {
					t.Fatal("sorted adjacency changed")
				}
			}
		})
	}
}

func TestBuilderEntryMemory(t *testing.T) {
	if bytes := reflect.TypeOf(directedEntry{}).Size(); bytes != 16 {
		t.Fatalf("temporary directed entries use %d bytes; want 16", bytes)
	}
}
