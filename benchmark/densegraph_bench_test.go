// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"context"
	"math/rand"
	"runtime"
	"testing"

	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/connectivity"
	"github.com/intelligrit/graphwizard/structure"
)

// ============================================================
// Dense vs Memory (gonum simple) benchmarks: speed and memory.
//
// Naming convention: Benchmark<Algo>_<Backend>_<Size>
// Backends: Memory (gonum simple), Dense (densegraph CSR)
// ============================================================

// --- Leiden ---

func BenchmarkLeiden_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := TwoClusterGraph(500, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.Leiden(context.Background(), g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

func BenchmarkLeiden_Dense_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := DenseTwoClusterGraph(500, 0.3, 0.01, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		community.Leiden(context.Background(), g, 1.0, rand.New(rand.NewSource(int64(i))))
	}
}

// --- Degree ---

func BenchmarkDegree_Memory_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Degree(context.Background(), g)
	}
}

func BenchmarkDegree_Dense_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := DenseBarabasiAlbert(10000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		centrality.Degree(context.Background(), g)
	}
}

// --- ConnectedComponents ---

func BenchmarkConnComp_Memory_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.ConnectedComponents(context.Background(), g)
	}
}

func BenchmarkConnComp_Dense_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := DenseBarabasiAlbert(10000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.ConnectedComponents(context.Background(), g)
	}
}

// --- KCore ---

func BenchmarkKCore_Memory_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(10000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.KCore(context.Background(), 3, g)
	}
}

func BenchmarkKCore_Dense_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := DenseBarabasiAlbert(10000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connectivity.KCore(context.Background(), 3, g)
	}
}

// --- TriangleCount ---

func BenchmarkTriangleCount_Memory_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := BarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCount(context.Background(), g)
	}
}

func BenchmarkTriangleCount_Dense_1K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	g := DenseBarabasiAlbert(1000, 3, rng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structure.TriangleCount(context.Background(), g)
	}
}

// ============================================================
// Memory allocation benchmarks.
//
// These measure the memory footprint of just the graph structure
// (not the algorithm working set).
// ============================================================

func BenchmarkGraphMemory_Memory_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	var before, after runtime.MemStats
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng.Seed(42)
		runtime.GC()
		runtime.ReadMemStats(&before)
		g := BarabasiAlbert(10000, 3, rng)
		runtime.GC()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc), "graph-bytes")
		_ = g
	}
}

func BenchmarkGraphMemory_Dense_10K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	var before, after runtime.MemStats
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng.Seed(42)
		runtime.GC()
		runtime.ReadMemStats(&before)
		g := DenseBarabasiAlbert(10000, 3, rng)
		runtime.GC()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc), "graph-bytes")
		_ = g
	}
}

func BenchmarkGraphMemory_Memory_50K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	var before, after runtime.MemStats
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng.Seed(42)
		runtime.GC()
		runtime.ReadMemStats(&before)
		g := BarabasiAlbert(50000, 3, rng)
		runtime.GC()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc), "graph-bytes")
		_ = g
	}
}

func BenchmarkGraphMemory_Dense_50K(b *testing.B) {
	rng := rand.New(rand.NewSource(42))
	var before, after runtime.MemStats
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng.Seed(42)
		runtime.GC()
		runtime.ReadMemStats(&before)
		g := DenseBarabasiAlbert(50000, 3, rng)
		runtime.GC()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc), "graph-bytes")
		_ = g
	}
}

// ============================================================
// Build time benchmarks: how long to construct the graph.
// ============================================================

func BenchmarkBuildGraph_Memory_10K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(42))
		_ = BarabasiAlbert(10000, 3, rng)
	}
}

func BenchmarkBuildGraph_Dense_10K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(42))
		_ = DenseBarabasiAlbert(10000, 3, rng)
	}
}
