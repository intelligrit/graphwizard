// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph_test

import (
	"runtime"
	"testing"

	"github.com/intelligrit/graphwizard/densegraph"
	"github.com/intelligrit/graphwizard/multigraph"
)

func benchmarkStore() *multigraph.Store {
	const nodes, relationships = 10000, 100000
	b := multigraph.NewBuilderSized(nodes, relationships)
	weights := make(map[string]float64, 3)
	for i := range relationships {
		weights["claims"] = float64(i%200 + 1)
		weights["dollars"] = float64(i%10000) + 0.125
		weights["risk"] = 0.25
		b.AddRelationship(int64(i%nodes), int64((i%nodes+1+i/nodes)%nodes),
			multigraph.Attributes{Type: "prescribed", Year: int32(2014 + i%10)}, weights)
	}
	return b.Build(multigraph.BlockWeights)
}

func BenchmarkRelationships(b *testing.B) {
	b.Run("build", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			s := benchmarkStore()
			b.ReportMetric(float64(s.StorageBytes()), "store-bytes")
			if s.NumRelationships() != 100000 {
				b.Fatal("records lost")
			}
			runtime.KeepAlive(s)
		}
	})
	s := benchmarkStore()
	for _, mode := range []struct {
		name string
		opts []densegraph.BuildOption
	}{
		{"flat", nil}, {"packed24", []densegraph.BuildOption{densegraph.PackNeighbors24}},
		{"delta", []densegraph.BuildOption{densegraph.CompressNeighbors}},
	} {
		b.Run("project/"+mode.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ReportMetric(float64(s.StorageBytes()), "store-bytes")
			for range b.N {
				g, err := s.ProjectUndirected(multigraph.Projection{Weight: "claims", Combine: multigraph.Sum}, mode.opts...)
				if err != nil {
					b.Fatal(err)
				}
				b.ReportMetric(float64(g.StorageBytes()), "projection-bytes")
				runtime.KeepAlive(g)
			}
		})
	}
}
