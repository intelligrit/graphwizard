// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math"
	"testing"

	"github.com/intelligrit/graphwizard"
	"github.com/intelligrit/graphwizard/densegraph"
	"gonum.org/v1/gonum/graph"
)

type genericUndirected struct{ graph.Undirected }

type denseWithoutCursor struct {
	graph.Undirected
	graphwizard.DenseAdjacency
}

func TestKatzCompressedMatchesGeneric(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts []densegraph.BuildOption
	}{
		{"delta", []densegraph.BuildOption{densegraph.CompressNeighbors}},
		{"packed24", []densegraph.BuildOption{densegraph.PackNeighbors24}},
		{"packed24-blocks", []densegraph.BuildOption{densegraph.PackNeighbors24, densegraph.BlockWeights}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := densegraph.NewUndirectedBuilder()
			for i := range 128 {
				b.AddNode(int64(i)*1009 - 77)
				for j := i + 1; j < min(128, i+9); j++ {
					b.AddEdge(int64(i)*1009-77, int64(j)*1009-77)
				}
			}
			g := b.Build(tc.opts...)
			ctx := context.Background()
			want := KatzUndirected(ctx, genericUndirected{g}, 0.01, 1, 1e-10, 100)
			serial := KatzUndirected(ctx, g, 0.01, 1, 1e-10, 100)
			parallel := KatzUndirectedParallel(ctx, g, 0.01, 1, 1e-10, 100)
			fallback := KatzUndirected(ctx, denseWithoutCursor{g, g}, 0.01, 1, 1e-10, 100)
			fallbackParallel := KatzUndirectedParallel(ctx, denseWithoutCursor{g, g}, 0.01, 1, 1e-10, 100)
			for id, w := range want {
				if math.Abs(serial[id]-w) > 1e-12 || math.Abs(parallel[id]-w) > 1e-12 || math.Abs(fallback[id]-w) > 1e-12 || math.Abs(fallbackParallel[id]-w) > 1e-12 {
					t.Fatalf("node %d differs", id)
				}
			}
		})
	}
}
