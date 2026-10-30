// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"maps"
	"math/rand"
	"testing"

	"github.com/intelligrit/graphwizard"
	"github.com/intelligrit/graphwizard/densegraph"
	"gonum.org/v1/gonum/graph"
)

// Hides optional adjacency interfaces to exercise the existing gonum fallback.
type genericWeightedGraph struct{ graph.WeightedUndirected }

type scannedWeightedGraph struct {
	genericWeightedGraph
	graphwizard.EdgeScanner
}

// Keeps the original dense-copy path visible while hiding shared traversal.
type copiedDenseGraph struct {
	genericWeightedGraph
	graphwizard.DenseAdjacency
}

func (g genericWeightedGraph) EdgeBetween(x, y int64) graph.Edge { return g.WeightedEdgeBetween(x, y) }

func sharedFixture(n int, opts ...densegraph.BuildOption) *densegraph.Undirected {
	b := densegraph.NewUndirectedBuilder()
	for i := range n {
		b.AddNode(int64(i)*101 - 3000)
	}
	for i := range n {
		for j := i + 1; j < min(n, i+24); j++ {
			w := float64(1 + (i+j)%17)
			if i/64 != j/64 {
				w = 0.125
			}
			b.AddWeightedEdge(int64(i)*101-3000, int64(j)*101-3000, w)
		}
	}
	return b.Build(opts...)
}

func TestSharedAdjacencyMatchesFallback(t *testing.T) {
	flat := sharedFixture(256)
	packed := sharedFixture(256, densegraph.CompressNeighbors)
	if packed.StorageBytes() >= flat.StorageBytes() {
		t.Fatal("fixture did not compress")
	}
	ctx := context.Background()
	for _, g := range []*densegraph.Undirected{flat, packed,
		sharedFixture(256, densegraph.PackNeighbors24),
		sharedFixture(256, densegraph.BlockWeights),
		sharedFixture(256, densegraph.PackNeighbors24, densegraph.BlockWeights),
		sharedFixture(256, densegraph.CompressNeighbors, densegraph.BlockWeights),
	} {
		ids, adj, degree, weight := buildWeightedAdj(g)
		if adj.shared != g || adj.rows != nil {
			t.Fatal("graph adjacency was copied")
		}
		wantIDs, wantAdj, wantDegree, wantWeight := buildWeightedAdj(genericWeightedGraph{g})
		if len(ids) != len(wantIDs) || weight != wantWeight {
			t.Fatal("build totals differ")
		}
		for i := range ids {
			if ids[i] != wantIDs[i] || degree[i] != wantDegree[i] {
				t.Fatal("node degree differs")
			}
			got, want := make(map[int]float64), make(map[int]float64)
			c := adj.cursor()
			c.reset(i)
			for c.next() {
				j, w := c.value()
				got[j] = w
			}
			c = wantAdj.cursor()
			c.reset(i)
			for c.next() {
				j, w := c.value()
				want[j] = w
			}
			if !maps.Equal(got, want) {
				t.Fatal("weighted adjacency differs")
			}
		}
		for seed := range int64(4) {
			want := Leiden(ctx, genericWeightedGraph{g}, 1, rand.New(rand.NewSource(seed)))
			serial := Leiden(ctx, g, 1, rand.New(rand.NewSource(seed)))
			parallel := LeidenParallel(ctx, g, 1, rand.New(rand.NewSource(seed)))
			if !maps.Equal(want, serial) || !maps.Equal(want, parallel) {
				t.Fatalf("partition differs for seed %d", seed)
			}
		}
	}
}

func TestSharedSelfEdgeWeight(t *testing.T) {
	b := densegraph.NewUndirectedBuilder()
	b.AddWeightedEdge(1, 1, 9)
	b.AddWeightedEdge(1, 2, 2)
	g := b.Build(densegraph.CompressNeighbors)
	_, _, wantDegree, wantTotal := buildWeightedAdj(genericWeightedGraph{g})
	for _, input := range []graph.Undirected{g, scannedWeightedGraph{genericWeightedGraph{g}, g}} {
		_, _, degree, total := buildWeightedAdj(input)
		if degree[0] != wantDegree[0] || total != wantTotal {
			t.Fatal("self-edge semantics changed")
		}
	}
}

func BenchmarkWeightedAdjacencyBuild(b *testing.B) {
	for _, compressed := range []bool{false, true} {
		name := "flat"
		var opts []densegraph.BuildOption
		if compressed {
			name = "compressed"
			opts = []densegraph.BuildOption{densegraph.CompressNeighbors}
		}
		g := sharedFixture(10000, opts...)
		for _, shared := range []bool{false, true} {
			mode := "copy"
			var input graph.Undirected = copiedDenseGraph{genericWeightedGraph{g}, g}
			if shared {
				mode = "shared"
				input = g
			}
			b.Run(name+"/"+mode, func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					_, adj, degree, _ := buildWeightedAdj(input)
					if len(degree) != g.NumNodes() || adj.shared == nil && adj.rows == nil {
						b.Fatal("missing adjacency")
					}
				}
			})
		}
	}
}

func BenchmarkSharedLeiden(b *testing.B) {
	for _, compressed := range []bool{false, true} {
		name := "flat"
		var opts []densegraph.BuildOption
		if compressed {
			name = "compressed"
			opts = []densegraph.BuildOption{densegraph.CompressNeighbors}
		}
		g := sharedFixture(1024, opts...)
		for _, shared := range []bool{false, true} {
			mode := "copy"
			var input graph.Undirected = copiedDenseGraph{genericWeightedGraph{g}, g}
			if shared {
				mode = "shared"
				input = g
			}
			b.Run(name+"/"+mode, func(b *testing.B) {
				b.ReportAllocs()
				for range b.N {
					result := Leiden(context.Background(), input, 1, rand.New(rand.NewSource(42)))
					if len(result) != 1024 {
						b.Fatal("missing nodes")
					}
				}
			})
		}
	}
}
