// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/intelligrit/graphwizard/anomaly"
	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/connectivity"
	"github.com/intelligrit/graphwizard/diff"
	"github.com/intelligrit/graphwizard/diskgraph"
	"github.com/intelligrit/graphwizard/embedding"
	"github.com/intelligrit/graphwizard/matching"
	"github.com/intelligrit/graphwizard/similarity"
	"github.com/intelligrit/graphwizard/structure"
	"github.com/intelligrit/graphwizard/subgraph"
	"github.com/intelligrit/graphwizard/traverse"
	"gonum.org/v1/gonum/graph/simple"
)

// ============================================================
// Remaining Memory vs Disk benchmarks for full algorithm coverage.
// Organized by package. See diskgraph_test.go for the first batch.
// ============================================================

// --- centrality: ApproximateBetweenness ---

func BenchmarkApproxBetweenness_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.ApproximateBetweenness(g, 100, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkApproxBetweenness_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bab")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.ApproximateBetweenness(g, 100, rand.New(rand.NewSource(int64(i))))
	}
}

// --- centrality: EdgeBetweenness ---

func BenchmarkEdgeBetweenness_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.EdgeBetweenness(g)
	}
}

func BenchmarkEdgeBetweenness_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("beb")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.EdgeBetweenness(g)
	}
}

// --- centrality: EccentricityParallel ---

func BenchmarkEccentricityParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.EccentricityParallel(g)
	}
}

func BenchmarkEccentricityParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("becp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.EccentricityParallel(g)
	}
}

// --- centrality: DiameterParallel ---

func BenchmarkDiameterParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.DiameterParallel(g)
	}
}

func BenchmarkDiameterParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bdp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.DiameterParallel(g)
	}
}

// --- centrality: Radius ---

func BenchmarkRadius_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Radius(g)
	}
}

func BenchmarkRadius_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("brad")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Radius(g)
	}
}

// --- centrality: KatzUndirected ---

func BenchmarkKatzUndirected_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.KatzUndirected(g, 0.01, 1.0, 1e-6, 50)
	}
}

func BenchmarkKatzUndirected_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bkatz")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.KatzUndirected(g, 0.01, 1.0, 1e-6, 50)
	}
}

// --- centrality: KatzUndirectedParallel ---

func BenchmarkKatzUndirectedParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.KatzUndirectedParallel(g, 0.01, 1.0, 1e-6, 50)
	}
}

func BenchmarkKatzUndirectedParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bkatzp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.KatzUndirectedParallel(g, 0.01, 1.0, 1e-6, 50)
	}
}

// --- centrality: PersonalizedPageRankUndirected ---

func BenchmarkPPRUndirected_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.PersonalizedPageRankUndirected(g, 0, 0.85, 1e-6, 100)
	}
}

func BenchmarkPPRUndirected_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bppr")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.PersonalizedPageRankUndirected(g, 0, 0.85, 1e-6, 100)
	}
}

// --- centrality: InfluenceMaximization ---

func BenchmarkInfluence_Memory_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(50, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.InfluenceMaximization(g, 3, 0.1, 100, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkInfluence_Disk_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("binf")
	defer cleanup()
	g, err := DiskBarabasiAlbert(50, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.InfluenceMaximization(g, 3, 0.1, 100, rand.New(rand.NewSource(int64(i))))
	}
}

// --- community: LeidenParallel ---

func BenchmarkLeidenParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(50, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.LeidenParallel(g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkLeidenParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bleip")
	defer cleanup()
	g, err := DiskTwoClusterGraph(50, 0.3, 0.01, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.LeidenParallel(g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

// --- community: SpectralClustering ---

func BenchmarkSpectral_Memory_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(25, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.SpectralClustering(g, 2)
	}
}

func BenchmarkSpectral_Disk_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bspec")
	defer cleanup()
	g, err := DiskTwoClusterGraph(25, 0.3, 0.01, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.SpectralClustering(g, 2)
	}
}

// --- connectivity: BridgesParallel ---

func BenchmarkBridgesParallel_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.BridgesParallel(g)
	}
}

func BenchmarkBridgesParallel_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bbrp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.BridgesParallel(g)
	}
}

// --- connectivity: ArticulationPoints ---

func BenchmarkArticulationPoints_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.ArticulationPoints(g)
	}
}

func BenchmarkArticulationPoints_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bap")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.ArticulationPoints(g)
	}
}

// --- connectivity: BiconnectedComponents ---

func BenchmarkBiconnected_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.BiconnectedComponents(g)
	}
}

func BenchmarkBiconnected_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bbc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.BiconnectedComponents(g)
	}
}

// --- connectivity: DegeneracyOrdering ---

func BenchmarkDegeneracy_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.DegeneracyOrdering(g)
	}
}

func BenchmarkDegeneracy_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bdeg2")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.DegeneracyOrdering(g)
	}
}

// --- connectivity: UndirectedCycles ---

func BenchmarkUndirectedCycles_Memory_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(50, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.UndirectedCycles(g)
	}
}

func BenchmarkUndirectedCycles_Disk_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("buc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(50, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.UndirectedCycles(g)
	}
}

// --- structure: TriangleCountParallel ---

func BenchmarkTriangleCountParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 5, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCountParallel(g)
	}
}

func BenchmarkTriangleCountParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("btrip")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 5, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCountParallel(g)
	}
}

// --- structure: ClusteringCoefficientParallel ---

func BenchmarkClusteringCoeffParallel_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.ClusteringCoefficientParallel(g)
	}
}

func BenchmarkClusteringCoeffParallel_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bccp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.ClusteringCoefficientParallel(g)
	}
}

// --- structure: MaximalCliques ---

func BenchmarkMaximalCliques_Memory_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(50, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.MaximalCliques(g)
	}
}

func BenchmarkMaximalCliques_Disk_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bmc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(50, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.MaximalCliques(g)
	}
}

// --- structure: GraphColoring ---

func BenchmarkGraphColoring_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.GraphColoring(g)
	}
}

func BenchmarkGraphColoring_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bgc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.GraphColoring(g)
	}
}

// --- structure: Kruskal ---

func BenchmarkKruskal_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := WeightedErdosRenyi(1000, 0.01, 10.0, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.Kruskal(g)
	}
}

func BenchmarkKruskal_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bkru")
	defer cleanup()
	g, err := DiskWeightedErdosRenyi(1000, 0.01, 10.0, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.Kruskal(g)
	}
}

// --- structure: Prim ---

func BenchmarkPrim_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := WeightedErdosRenyi(1000, 0.01, 10.0, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.Prim(g, 0)
	}
}

func BenchmarkPrim_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bprim")
	defer cleanup()
	g, err := DiskWeightedErdosRenyi(1000, 0.01, 10.0, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.Prim(g, 0)
	}
}

// --- structure: TSP ---

func BenchmarkTSP_Memory_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := WeightedErdosRenyi(50, 0.5, 10.0, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TSP(g)
	}
}

func BenchmarkTSP_Disk_50(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("btsp")
	defer cleanup()
	g, err := DiskWeightedErdosRenyi(50, 0.5, 10.0, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TSP(g)
	}
}

// --- similarity: Overlap ---

func BenchmarkOverlap_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.Overlap(g, 0, 1)
	}
}

func BenchmarkOverlap_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bov")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.Overlap(g, 0, 1)
	}
}

// --- similarity: CommonNeighbors ---

func BenchmarkCommonNeighbors_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.CommonNeighbors(g, 0, 1)
	}
}

func BenchmarkCommonNeighbors_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bcn")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.CommonNeighbors(g, 0, 1)
	}
}

// --- similarity: AdamicAdar ---

func BenchmarkAdamicAdar_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.AdamicAdar(g, 0, 1)
	}
}

func BenchmarkAdamicAdar_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("baa")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.AdamicAdar(g, 0, 1)
	}
}

// --- similarity: PreferentialAttachment ---

func BenchmarkPrefAttach_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.PreferentialAttachment(g, 0, 1)
	}
}

func BenchmarkPrefAttach_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bpa")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.PreferentialAttachment(g, 0, 1)
	}
}

// --- similarity: Cosine ---

func BenchmarkCosine_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.Cosine(g, 0, 1)
	}
}

func BenchmarkCosine_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bcos")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.Cosine(g, 0, 1)
	}
}

// --- similarity: JaccardAllParallel ---

func BenchmarkJaccardAllParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := ErdosRenyi(100, 0.1, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.JaccardAllParallel(g, 0.1)
	}
}

func BenchmarkJaccardAllParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bjap")
	defer cleanup()
	g, err := DiskErdosRenyi(100, 0.1, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.JaccardAllParallel(g, 0.1)
	}
}

// --- similarity: PredictLinks ---

func BenchmarkPredictLinks_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.PredictLinks(g, 5, similarity.Jaccard)
	}
}

func BenchmarkPredictLinks_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bpl")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.PredictLinks(g, 5, similarity.Jaccard)
	}
}

// --- similarity: PredictLinksParallel ---

func BenchmarkPredictLinksParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.PredictLinksParallel(g, 5, similarity.Jaccard)
	}
}

func BenchmarkPredictLinksParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bplp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.PredictLinksParallel(g, 5, similarity.Jaccard)
	}
}

// --- traverse: BFSPath ---

func BenchmarkBFSPath_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		traverse.BFSPath(g, 0, 999)
	}
}

func BenchmarkBFSPath_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bbfsp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		traverse.BFSPath(g, 0, 999)
	}
}

// --- anomaly: IsolationScore ---

func BenchmarkIsolationScore_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		anomaly.IsolationScore(g)
	}
}

func BenchmarkIsolationScore_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("biso")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		anomaly.IsolationScore(g)
	}
}

// --- anomaly: StructuralOutliers ---

func BenchmarkStructuralOutliers_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		anomaly.StructuralOutliers(g, 5)
	}
}

func BenchmarkStructuralOutliers_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bso")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		anomaly.StructuralOutliers(g, 5)
	}
}

// --- embedding: Node2VecWalksParallel ---

func BenchmarkNode2VecParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.Node2VecWalksParallel(g, embedding.WalkParams{WalkLength: 10, WalksPerNode: 3, P: 1, Q: 1}, int64(i))
	}
}

func BenchmarkNode2VecParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bn2vp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.Node2VecWalksParallel(g, embedding.WalkParams{WalkLength: 10, WalksPerNode: 3, P: 1, Q: 1}, int64(i))
	}
}

// --- embedding: DeepWalkWalks ---

func BenchmarkDeepWalk_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.DeepWalkWalks(g, 10, 3, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkDeepWalk_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bdw")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.DeepWalkWalks(g, 10, 3, rand.New(rand.NewSource(int64(i))))
	}
}

// --- embedding: DeepWalkWalksParallel ---

func BenchmarkDeepWalkParallel_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.DeepWalkWalksParallel(g, 10, 3, int64(i))
	}
}

func BenchmarkDeepWalkParallel_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bdwp")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		embedding.DeepWalkWalksParallel(g, 10, 3, int64(i))
	}
}

// --- subgraph: NHopNeighborhoodUndirected ---

func BenchmarkNHop_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subgraph.NHopNeighborhoodUndirected(g, 0, 2)
	}
}

func BenchmarkNHop_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bnhop")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subgraph.NHopNeighborhoodUndirected(g, 0, 2)
	}
}

// --- subgraph: FilterNodes ---

func BenchmarkFilterNodes_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	keep := func(id int64) bool { return id%2 == 0 }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subgraph.FilterNodes(g, keep)
	}
}

func BenchmarkFilterNodes_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bfilt")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	keep := func(id int64) bool { return id%2 == 0 }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subgraph.FilterNodes(g, keep)
	}
}

// --- diff: Compare ---

func BenchmarkCompare_Memory_100(b *testing.B) {
	rng1 := rand.New(rand.NewSource(42))
	rng2 := rand.New(rand.NewSource(43))
	g1 := BarabasiAlbert(100, 3, rng1)
	g2 := BarabasiAlbert(100, 3, rng2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff.Compare(g1, g2)
	}
}

func BenchmarkCompare_Disk_100(b *testing.B) {
	rng1 := rand.New(rand.NewSource(42))
	rng2 := rand.New(rand.NewSource(43))
	dir1, cleanup1 := diskTempDir("bcmp1")
	defer cleanup1()
	g1, err := DiskBarabasiAlbert(100, 3, rng1, dir1)
	if err != nil {
		b.Fatal(err)
	}
	defer g1.Close()
	dir2, cleanup2 := diskTempDir("bcmp2")
	defer cleanup2()
	g2, err := DiskBarabasiAlbert(100, 3, rng2, dir2)
	if err != nil {
		b.Fatal(err)
	}
	defer g2.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff.Compare(g1, g2)
	}
}

// --- matching: HopcroftKarp ---

func BenchmarkHopcroftKarp_Memory_100(b *testing.B) {
	// Build a bipartite graph: left=0..49, right=50..99
	rng := rand.New(rand.NewSource(42))
	g := buildBipartiteMemory(50, 50, 0.2, rng)
	left := make([]int64, 50)
	for i := range left {
		left[i] = int64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matching.HopcroftKarp(g, left)
	}
}

func BenchmarkHopcroftKarp_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bhk")
	defer cleanup()
	g, err := buildBipartiteDisk(50, 50, 0.2, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	left := make([]int64, 50)
	for i := range left {
		left[i] = int64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matching.HopcroftKarp(g, left)
	}
}

// --- helpers for bipartite graphs ---

func buildBipartiteMemory(nLeft, nRight int, p float64, rng *rand.Rand) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < int64(nLeft); i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(nLeft); i < int64(nLeft+nRight); i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < int64(nLeft); i++ {
		for j := int64(nLeft); j < int64(nLeft+nRight); j++ {
			if rng.Float64() < p {
				g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
			}
		}
	}
	return g
}

func buildBipartiteDisk(nLeft, nRight int, p float64, rng *rand.Rand, dir string) (*diskgraph.Undirected, error) {
	path := filepath.Join(dir, "bip.db")
	b, err := diskgraph.NewUndirectedBuilder(path)
	if err != nil {
		return nil, err
	}
	err = b.Batch(func(tx *diskgraph.UndirectedTx) error {
		for i := int64(0); i < int64(nLeft+nRight); i++ {
			if err := tx.AddNode(i); err != nil {
				return err
			}
		}
		for i := int64(0); i < int64(nLeft); i++ {
			for j := int64(nLeft); j < int64(nLeft+nRight); j++ {
				if rng.Float64() < p {
					if err := tx.AddEdge(i, j); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		b.Close()
		return nil, err
	}
	if err := b.Close(); err != nil {
		return nil, err
	}
	return diskgraph.OpenUndirected(path)
}
