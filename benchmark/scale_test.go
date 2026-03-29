// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"math/rand"
	"testing"

	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/connectivity"
	"github.com/intelligrit/graphwizard/embedding"
	"github.com/intelligrit/graphwizard/similarity"
	"github.com/intelligrit/graphwizard/structure"
	"github.com/intelligrit/graphwizard/traverse"
)

// --- Correctness at scale: verify algorithms don't crash or produce
//     degenerate results on graphs with 1K-10K nodes ---

func TestScale_Leiden_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(500, 0.05, 0.001, rng)

	comms := community.Leiden(g, 1.0, rng)
	labels := make(map[int64]bool)
	for _, c := range comms {
		labels[c] = true
	}
	t.Logf("1K nodes, %d communities", len(labels))
	if len(labels) < 2 {
		t.Error("expected at least 2 communities in two-cluster graph")
	}
}

func TestScale_Louvain_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(500, 0.05, 0.001, rng)

	comms := community.Louvain(g, 1.0, nil)
	labels := make(map[int64]bool)
	for _, c := range comms {
		labels[c] = true
	}
	t.Logf("1K nodes, %d communities", len(labels))
	if len(labels) < 2 {
		t.Error("expected at least 2 communities")
	}
}

func TestScale_LabelProp_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(500, 0.05, 0.001, rng)

	comms := community.LabelPropagation(g, 50, rng)
	labels := make(map[int64]bool)
	for _, c := range comms {
		labels[c] = true
	}
	t.Logf("1K nodes, %d communities", len(labels))
	if len(labels) < 2 {
		t.Error("expected at least 2 communities")
	}
}

func TestScale_Bridges_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)

	bridges := connectivity.Bridges(g)
	t.Logf("1K BA graph: %d bridges", len(bridges))
}

func TestScale_WCC_10K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 2, rng)

	comps := connectivity.ConnectedComponents(g)
	t.Logf("10K BA graph: %d components", len(comps))
	if len(comps) != 1 {
		t.Error("BA graph should be connected")
	}
}

func TestScale_Triangles_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 5, rng)

	_, total := structure.TriangleCount(g)
	t.Logf("1K BA(m=5) graph: %d triangles", total)
	if total == 0 {
		t.Error("BA graph with m=5 should have triangles")
	}
}

func TestScale_ClusteringCoefficient_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)

	avg := structure.AverageClusteringCoefficient(g)
	t.Logf("1K BA(m=3) avg CC: %.4f", avg)
	if avg == 0 {
		t.Error("BA graph should have non-zero clustering")
	}
}

func TestScale_Degree_10K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 3, rng)

	scores := centrality.Degree(g)
	if len(scores) != 10000 {
		t.Errorf("expected 10000 scores, got %d", len(scores))
	}

	// Early nodes should have higher degree (preferential attachment).
	if scores[0] < scores[9999] {
		t.Error("early nodes in BA graph should have higher degree centrality")
	}
}

func TestScale_BFS_10K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 2, rng)

	visited := traverse.BFS(g, 0)
	if len(visited) != 10000 {
		t.Errorf("BFS should reach all 10K nodes, reached %d", len(visited))
	}
}

func TestScale_Jaccard_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)

	// Spot check: nodes 0 and 1 (both early, high degree).
	j := similarity.Jaccard(g, 0, 1)
	t.Logf("J(0,1) in 1K BA: %.4f", j)
}

func TestScale_Node2Vec_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)

	walks := embedding.Node2VecWalks(g, embedding.WalkParams{
		WalkLength:   20,
		WalksPerNode: 5,
		P:            1.0,
		Q:            1.0,
	}, rng)

	if len(walks) != 5000 {
		t.Errorf("expected 5000 walks, got %d", len(walks))
	}
	for i, walk := range walks {
		if len(walk) != 20 {
			t.Errorf("walk %d: expected length 20, got %d", i, len(walk))
			break
		}
	}
}

func TestScale_Kruskal_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := WeightedErdosRenyi(1000, 0.01, 10.0, rng)

	result := structure.Kruskal(g)
	t.Logf("1K ER MST: %d edges, weight %.2f", len(result.Edges), result.Weight)
}

func TestScale_KCore_10K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 3, rng)

	core := connectivity.KCore(3, g)
	t.Logf("10K BA 3-core: %d nodes", len(core))
	if len(core) == 0 {
		t.Error("BA(m=3) should have a non-empty 3-core")
	}
}

// --- Benchmarks ---

func BenchmarkLeiden_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(50, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.Leiden(g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkLouvain_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(50, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.Louvain(g, 1.0, nil)
	}
}

func BenchmarkLabelProp_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(500, 0.05, 0.001, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.LabelPropagation(g, 20, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkBridges_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.Bridges(g)
	}
}

func BenchmarkBetweenness_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Betweenness(g)
	}
}

func BenchmarkBetweenness_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Betweenness(g)
	}
}

func BenchmarkDegree_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Degree(g)
	}
}

func BenchmarkPageRank_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	ug := BarabasiAlbert(1000, 3, rng)
	dg := toDirected(ug)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.PageRank(dg, 0.85, 1e-6)
	}
}

func BenchmarkTriangleCount_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 5, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCount(g)
	}
}

func BenchmarkBFS_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 2, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		traverse.BFS(g, 0)
	}
}

func BenchmarkKruskal_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := WeightedErdosRenyi(1000, 0.01, 10.0, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.Kruskal(g)
	}
}

func BenchmarkNode2Vec_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	params := embedding.WalkParams{WalkLength: 20, WalksPerNode: 5, P: 1, Q: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.Node2VecWalks(g, params, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkWCC_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 2, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.ConnectedComponents(g)
	}
}

func BenchmarkClusteringCoefficient_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.ClusteringCoefficient(g)
	}
}
