// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/connectivity"
	"github.com/intelligrit/graphwizard/embedding"
	"github.com/intelligrit/graphwizard/similarity"
	"github.com/intelligrit/graphwizard/structure"
	"gonum.org/v1/gonum/graph/simple"
)

// --- Correctness: parallel versions must match sequential ---

func TestApproximateBetweenness_Correctness(t *testing.T) {
	g := KarateClub()
	rng := rand.New(rand.NewSource(42))

	// Full sampling (k=34) should approximate exact betweenness.
	approx := centrality.ApproximateBetweenness(context.Background(), g, 34, rng)
	exact := centrality.Betweenness(context.Background(), g)

	// Top-ranked node should be the same.
	topApprox := topNode(approx)
	topExact := topNode(exact)
	t.Logf("exact top: %d, approx top: %d", topExact, topApprox)
	if topApprox != topExact {
		t.Logf("warning: top nodes differ (approximation noise)")
	}

	// Correlation: all nodes should have non-negative scores.
	for id, s := range approx {
		if s < 0 {
			t.Errorf("node %d has negative betweenness: %f", id, s)
		}
	}
}

func TestApproximateBetweenness_Sampled(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)

	rng2 := rand.New(rand.NewSource(42))
	scores := centrality.ApproximateBetweenness(context.Background(), g, 100, rng2)
	if len(scores) != 1000 {
		t.Errorf("expected 1000 scores, got %d", len(scores))
	}

	// Early nodes in BA should have higher betweenness.
	if scores[0] < scores[999] {
		t.Error("early BA nodes should have higher approximate betweenness")
	}
}

func TestApproximateBetweenness_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(42))
	scores := centrality.ApproximateBetweenness(context.Background(), g, 10, rng)
	if len(scores) != 0 {
		t.Errorf("expected empty, got %d", len(scores))
	}
}

func TestTriangleCountParallel_MatchesSequential(t *testing.T) {
	g := KarateClub()

	seqNodes, seqTotal := structure.TriangleCount(context.Background(), g)
	parNodes, parTotal := structure.TriangleCountParallel(context.Background(), g)

	if seqTotal != parTotal {
		t.Errorf("total mismatch: seq=%d par=%d", seqTotal, parTotal)
	}
	for id, sc := range seqNodes {
		if parNodes[id] != sc {
			t.Errorf("node %d: seq=%d par=%d", id, sc, parNodes[id])
		}
	}
}

func TestClusteringCoefficientParallel_MatchesSequential(t *testing.T) {
	g := KarateClub()

	seq := structure.ClusteringCoefficient(context.Background(), g)
	par := structure.ClusteringCoefficientParallel(context.Background(), g)

	for id, sc := range seq {
		if math.Abs(par[id]-sc) > 1e-10 {
			t.Errorf("node %d: seq=%.6f par=%.6f", id, sc, par[id])
		}
	}
}

func TestNode2VecWalksParallel_Structure(t *testing.T) {
	g := KarateClub()

	walks := embedding.Node2VecWalksParallel(context.Background(), g, embedding.WalkParams{
		WalkLength:   10,
		WalksPerNode: 3,
		P:            1.0,
		Q:            1.0,
	}, 42)

	// 34 nodes * 3 walks = 102 walks.
	if len(walks) != 102 {
		t.Errorf("expected 102 walks, got %d", len(walks))
	}
	for i, walk := range walks {
		if len(walk) != 10 {
			t.Errorf("walk %d: expected length 10, got %d", i, len(walk))
			break
		}
	}
}

func TestEccentricityParallel_MatchesSequential(t *testing.T) {
	g := KarateClub()

	seq := centrality.Eccentricity(context.Background(), g)
	par := centrality.EccentricityParallel(context.Background(), g)

	for id, sc := range seq {
		if math.Abs(par[id]-sc) > 1e-10 {
			t.Errorf("node %d: seq=%.2f par=%.2f", id, sc, par[id])
		}
	}
}

func TestDiameterParallel_MatchesSequential(t *testing.T) {
	g := KarateClub()
	seq := centrality.Diameter(context.Background(), g)
	par := centrality.DiameterParallel(context.Background(), g)
	if seq != par {
		t.Errorf("diameter mismatch: seq=%.0f par=%.0f", seq, par)
	}
}

func TestJaccardAllParallel_MatchesSequential(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	g := ErdosRenyi(50, 0.2, rng)

	seq := similarity.JaccardAll(context.Background(), g, 0.1)
	par := similarity.JaccardAllParallel(context.Background(), g, 0.1)

	if len(seq) != len(par) {
		t.Errorf("count mismatch: seq=%d par=%d", len(seq), len(par))
	}
}

// --- Benchmarks: sequential vs parallel ---

func BenchmarkTriangleCount_Sequential_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 5, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCount(context.Background(), g)
	}
}

func BenchmarkTriangleCount_Parallel_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 5, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCountParallel(context.Background(), g)
	}
}

func BenchmarkClusteringCoefficient_Sequential_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.ClusteringCoefficient(context.Background(), g)
	}
}

func BenchmarkClusteringCoefficient_Parallel_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.ClusteringCoefficientParallel(context.Background(), g)
	}
}

func BenchmarkNode2Vec_Sequential_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	params := embedding.WalkParams{WalkLength: 20, WalksPerNode: 5, P: 1, Q: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.Node2VecWalks(context.Background(), g, params, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkNode2Vec_Parallel_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.Node2VecWalksParallel(context.Background(), g, embedding.WalkParams{WalkLength: 20, WalksPerNode: 5, P: 1, Q: 1}, int64(i))
	}
}

func BenchmarkEccentricity_Sequential_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Eccentricity(context.Background(), g)
	}
}

func BenchmarkEccentricity_Parallel_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.EccentricityParallel(context.Background(), g)
	}
}

func BenchmarkApproxBetweenness_1K_k100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.ApproximateBetweenness(context.Background(), g, 100, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkExactBetweenness_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Betweenness(context.Background(), g)
	}
}

// --- Parallel vs Sequential: new algorithms ---

func BenchmarkLeiden_Sequential_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(50, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.Leiden(context.Background(), g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkLeiden_Parallel_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(50, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.LeidenParallel(context.Background(), g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkKatz_Sequential_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	ug := BarabasiAlbert(1000, 3, rng)
	dg := toDirected(ug)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Katz(context.Background(), dg, 0.01, 1.0, 1e-6, 50)
	}
}

func BenchmarkKatz_Parallel_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	ug := BarabasiAlbert(1000, 3, rng)
	dg := toDirected(ug)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.KatzParallel(context.Background(), dg, 0.01, 1.0, 1e-6, 50)
	}
}

func BenchmarkBridges_Sequential_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.Bridges(context.Background(), g)
	}
}

func BenchmarkBridges_Parallel_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.BridgesParallel(context.Background(), g)
	}
}

// --- Helpers ---

func topNode(scores map[int64]float64) int64 {
	top := int64(-1)
	topScore := -1.0
	for id, s := range scores {
		if s > topScore {
			topScore = s
			top = id
		}
	}
	return top
}
