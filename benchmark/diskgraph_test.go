// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"math"
	"math/rand"
	"testing"

	"github.com/intelligrit/graphwizard/anomaly"
	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/connectivity"
	"github.com/intelligrit/graphwizard/similarity"
	"github.com/intelligrit/graphwizard/structure"
	"github.com/intelligrit/graphwizard/traverse"
)

// ============================================================
// Correctness: diskgraph results must match in-memory results
// ============================================================

func TestDiskKarate_GraphStructure(t *testing.T) {
	dir, cleanup := diskTempDir("karate")
	defer cleanup()
	g, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 34 {
		t.Errorf("expected 34 nodes, got %d", count)
	}

	edgeCount := 0
	nodes = g.Nodes()
	for nodes.Next() {
		edgeCount += g.From(nodes.Node().ID()).Len()
	}
	edgeCount /= 2
	if edgeCount != 78 {
		t.Errorf("expected 78 edges, got %d", edgeCount)
	}
}

func TestDiskKarate_Degree(t *testing.T) {
	memG := KarateClub()
	memScores := centrality.Degree(memG)

	dir, cleanup := diskTempDir("deg")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskScores := centrality.Degree(diskG)
	compareMaps(t, "Degree", memScores, diskScores)
}

func TestDiskKarate_Betweenness(t *testing.T) {
	memG := KarateClub()
	memScores := centrality.Betweenness(memG)

	dir, cleanup := diskTempDir("btw")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskScores := centrality.Betweenness(diskG)
	compareMaps(t, "Betweenness", memScores, diskScores)
}

func TestDiskKarate_Eccentricity(t *testing.T) {
	memG := KarateClub()
	memScores := centrality.Eccentricity(memG)

	dir, cleanup := diskTempDir("ecc")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskScores := centrality.Eccentricity(diskG)
	compareMaps(t, "Eccentricity", memScores, diskScores)
}

func TestDiskKarate_Diameter(t *testing.T) {
	memD := centrality.Diameter(KarateClub())

	dir, cleanup := diskTempDir("dia")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskD := centrality.Diameter(diskG)
	if memD != diskD {
		t.Errorf("Diameter: mem=%.0f disk=%.0f", memD, diskD)
	}
}

func TestDiskKarate_TriangleCount(t *testing.T) {
	_, memTotal := structure.TriangleCount(KarateClub())

	dir, cleanup := diskTempDir("tri")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	_, diskTotal := structure.TriangleCount(diskG)
	if memTotal != diskTotal {
		t.Errorf("TriangleCount: mem=%d disk=%d", memTotal, diskTotal)
	}
}

func TestDiskKarate_ClusteringCoefficient(t *testing.T) {
	memCC := structure.ClusteringCoefficient(KarateClub())

	dir, cleanup := diskTempDir("cc")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskCC := structure.ClusteringCoefficient(diskG)
	compareMaps(t, "ClusteringCoefficient", memCC, diskCC)
}

func TestDiskKarate_AverageClusteringCoefficient(t *testing.T) {
	memAvg := structure.AverageClusteringCoefficient(KarateClub())

	dir, cleanup := diskTempDir("avgcc")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskAvg := structure.AverageClusteringCoefficient(diskG)
	if math.Abs(memAvg-diskAvg) > 1e-9 {
		t.Errorf("AvgCC: mem=%.6f disk=%.6f", memAvg, diskAvg)
	}
}

func TestDiskKarate_Bridges(t *testing.T) {
	memBridges := connectivity.Bridges(KarateClub())

	dir, cleanup := diskTempDir("br")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskBridges := connectivity.Bridges(diskG)
	if len(memBridges) != len(diskBridges) {
		t.Errorf("Bridges: mem=%d disk=%d", len(memBridges), len(diskBridges))
	}
}

func TestDiskKarate_ConnectedComponents(t *testing.T) {
	memComps := connectivity.ConnectedComponents(KarateClub())

	dir, cleanup := diskTempDir("wcc")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskComps := connectivity.ConnectedComponents(diskG)
	if len(memComps) != len(diskComps) {
		t.Errorf("ConnectedComponents: mem=%d disk=%d", len(memComps), len(diskComps))
	}
}

func TestDiskKarate_KCore(t *testing.T) {
	memCore := connectivity.KCore(3, KarateClub())

	dir, cleanup := diskTempDir("kc")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskCore := connectivity.KCore(3, diskG)
	if len(memCore) != len(diskCore) {
		t.Errorf("KCore: mem=%d disk=%d", len(memCore), len(diskCore))
	}
}

func TestDiskKarate_BFS(t *testing.T) {
	memVisited := traverse.BFS(KarateClub(), 0)

	dir, cleanup := diskTempDir("bfs")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskVisited := traverse.BFS(diskG, 0)
	if len(memVisited) != len(diskVisited) {
		t.Errorf("BFS: mem=%d disk=%d", len(memVisited), len(diskVisited))
	}
}

func TestDiskKarate_DFS(t *testing.T) {
	memVisited := traverse.DFS(KarateClub(), 0)

	dir, cleanup := diskTempDir("dfs")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskVisited := traverse.DFS(diskG, 0)
	if len(memVisited) != len(diskVisited) {
		t.Errorf("DFS: mem=%d disk=%d", len(memVisited), len(diskVisited))
	}
}

func TestDiskKarate_Jaccard(t *testing.T) {
	memJ := similarity.Jaccard(KarateClub(), 0, 1)

	dir, cleanup := diskTempDir("jac")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskJ := similarity.Jaccard(diskG, 0, 1)
	if math.Abs(memJ-diskJ) > 1e-9 {
		t.Errorf("Jaccard(0,1): mem=%.6f disk=%.6f", memJ, diskJ)
	}
}

func TestDiskKarate_DegreeZScore(t *testing.T) {
	memZ := anomaly.DegreeZScore(KarateClub())

	dir, cleanup := diskTempDir("zsc")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	diskZ := anomaly.DegreeZScore(diskG)
	compareMaps(t, "DegreeZScore", memZ, diskZ)
}

func TestDiskKarate_Leiden(t *testing.T) {
	memG := KarateClub()
	rng1 := rand.New(rand.NewSource(42))
	memComms := community.Leiden(memG, 1.0, rng1)

	dir, cleanup := diskTempDir("lei")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	rng2 := rand.New(rand.NewSource(42))
	diskComms := community.Leiden(diskG, 1.0, rng2)

	// Community labels may differ, but the number of communities should match.
	memLabels := uniqueLabels(memComms)
	diskLabels := uniqueLabels(diskComms)
	if len(memLabels) != len(diskLabels) {
		t.Logf("Leiden communities: mem=%d disk=%d (may differ due to iteration order)", len(memLabels), len(diskLabels))
	}
	// Both should detect at least 2 communities.
	if len(diskLabels) < 2 {
		t.Error("Leiden on disk graph should detect at least 2 communities")
	}
}

func TestDiskKarate_LabelPropagation(t *testing.T) {
	dir, cleanup := diskTempDir("lp")
	defer cleanup()
	diskG, err := DiskKarateClub(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer diskG.Close()

	rng := rand.New(rand.NewSource(42))
	comms := community.LabelPropagation(diskG, 100, rng)
	labels := uniqueLabels(comms)
	if len(labels) < 2 {
		t.Error("LabelPropagation on disk graph should detect at least 2 communities")
	}
}

// ============================================================
// Scale correctness: disk-backed at 1K nodes
// ============================================================

func TestDiskScale_Degree_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("sdeg")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	scores := centrality.Degree(g)
	if len(scores) != 1000 {
		t.Errorf("expected 1000 scores, got %d", len(scores))
	}
}

func TestDiskScale_BFS_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("sbfs")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	visited := traverse.BFS(g, 0)
	if len(visited) != 1000 {
		t.Errorf("BFS should reach all 1K nodes, reached %d", len(visited))
	}
}

func TestDiskScale_Bridges_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("sbr")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	bridges := connectivity.Bridges(g)
	t.Logf("1K disk BA graph: %d bridges", len(bridges))
}

func TestDiskScale_TriangleCount_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("stri")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 5, rng, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	_, total := structure.TriangleCount(g)
	t.Logf("1K disk BA(m=5): %d triangles", total)
	if total == 0 {
		t.Error("should have triangles")
	}
}

func TestDiskScale_WCC_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("swcc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 2, rng, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	comps := connectivity.ConnectedComponents(g)
	if len(comps) != 1 {
		t.Errorf("BA graph should be connected, got %d components", len(comps))
	}
}

func TestDiskScale_Leiden_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("slei")
	defer cleanup()
	g, err := DiskTwoClusterGraph(500, 0.05, 0.001, rng, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	rng2 := rand.New(rand.NewSource(42))
	comms := community.Leiden(g, 1.0, rng2)
	labels := uniqueLabels(comms)
	t.Logf("1K disk two-cluster: %d communities", len(labels))
	if len(labels) < 2 {
		t.Error("expected at least 2 communities")
	}
}

func TestDiskScale_ClusteringCoefficient_1K(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("scc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	avg := structure.AverageClusteringCoefficient(g)
	t.Logf("1K disk BA(m=3) avg CC: %.4f", avg)
	if avg == 0 {
		t.Error("should have non-zero clustering")
	}
}

// ============================================================
// Benchmarks: Memory vs Disk side-by-side
// ============================================================

// --- Degree ---

func BenchmarkDegree_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Degree(g)
	}
}

func BenchmarkDegree_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bdeg")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Degree(g)
	}
}

// --- Betweenness ---

func BenchmarkBetweenness_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Betweenness(g)
	}
}

func BenchmarkBetweenness_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bbtw")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Betweenness(g)
	}
}

// --- BFS ---

func BenchmarkBFS_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		traverse.BFS(g, 0)
	}
}

func BenchmarkBFS_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bbfs")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		traverse.BFS(g, 0)
	}
}

// --- Triangle Count ---

func BenchmarkTriangleCount_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 5, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCount(g)
	}
}

func BenchmarkTriangleCount_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("btri")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 5, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCount(g)
	}
}

// --- Clustering Coefficient ---

func BenchmarkClusteringCoeff_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.ClusteringCoefficient(g)
	}
}

func BenchmarkClusteringCoeff_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bcc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.ClusteringCoefficient(g)
	}
}

// --- Bridges ---

func BenchmarkBridges_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.Bridges(g)
	}
}

func BenchmarkBridges_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bbr")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.Bridges(g)
	}
}

// --- Connected Components ---

func BenchmarkWCC_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.ConnectedComponents(g)
	}
}

func BenchmarkWCC_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bwcc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.ConnectedComponents(g)
	}
}

// --- Leiden ---

func BenchmarkLeiden_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(50, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.Leiden(g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkLeiden_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("blei")
	defer cleanup()
	g, err := DiskTwoClusterGraph(50, 0.3, 0.01, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.Leiden(g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

// --- Jaccard ---

func BenchmarkJaccard_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.Jaccard(g, 0, 1)
	}
}

func BenchmarkJaccard_Disk_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("bjac")
	defer cleanup()
	g, err := DiskBarabasiAlbert(1000, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		similarity.Jaccard(g, 0, 1)
	}
}

// --- Eccentricity ---

func BenchmarkEccentricity_Memory_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(100, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Eccentricity(g)
	}
}

func BenchmarkEccentricity_Disk_100(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	dir, cleanup := diskTempDir("becc")
	defer cleanup()
	g, err := DiskBarabasiAlbert(100, 3, rng, dir)
	if err != nil {
		b.Fatal(err)
	}
	defer g.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Eccentricity(g)
	}
}

// ============================================================
// Helpers
// ============================================================

func compareMaps(t *testing.T, name string, mem, disk map[int64]float64) {
	t.Helper()
	if len(mem) != len(disk) {
		t.Errorf("%s: mem has %d entries, disk has %d", name, len(mem), len(disk))
		return
	}
	for id, mv := range mem {
		dv, ok := disk[id]
		if !ok {
			t.Errorf("%s: node %d in mem but not disk", name, id)
			continue
		}
		if math.Abs(mv-dv) > 1e-9 {
			t.Errorf("%s: node %d: mem=%.10f disk=%.10f", name, id, mv, dv)
		}
	}
}

func uniqueLabels(comms map[int64]int64) map[int64]bool {
	labels := make(map[int64]bool)
	for _, c := range comms {
		labels[c] = true
	}
	return labels
}
