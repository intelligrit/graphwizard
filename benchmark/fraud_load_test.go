// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/densegraph"
	"github.com/intelligrit/graphwizard/similarity"
)

// TestFraudAlgorithmsLargeBipartite is an opt-in load test. Its generated
// topology has many providers connected through a smaller set of bridge nodes,
// as in a provider-drug graph. The real Medicare fixture is not in this repo.
func TestFraudAlgorithmsLargeBipartite(t *testing.T) {
	if os.Getenv("GRAPHWIZARD_LARGE_TEST") != "1" {
		t.Skip("set GRAPHWIZARD_LARGE_TEST=1 to run the synthetic load test")
	}
	left := loadTestSize(t, "GRAPHWIZARD_LEFT", 200_000)
	right := loadTestSize(t, "GRAPHWIZARD_RIGHT", 2_000)
	degree := loadTestSize(t, "GRAPHWIZARD_DEGREE", 20)
	if left <= right || degree > right || left > math.MaxInt/degree {
		t.Fatal("invalid synthetic graph dimensions")
	}
	edges := left * degree
	t.Logf("synthetic graph: %d left nodes, %d bridge nodes, %d undirected edges", left, right, edges)

	var g *densegraph.Undirected
	measureFraudLoadPhase(t, "build", 0, func() error {
		builder := densegraph.NewUndirectedBuilderSized(left+right, 2*edges)
		for id := 0; id < left; id++ {
			for j := 0; j < degree; j++ {
				bridge := left + (id%right+j)%right
				builder.AddEdge(int64(id), int64(bridge))
			}
		}
		g = builder.Build()
		return nil
	})

	ctx := context.Background()
	isLeft := func(id int64) bool { return id < int64(left) }
	var scorer *similarity.BipartiteOverlapScorer
	measureFraudLoadPhase(t, "overlap constructor", 1<<30, func() (err error) {
		scorer, err = similarity.NewBipartiteOverlapScorer(ctx, g, isLeft)
		return err
	})
	measureFraudLoadPhase(t, "overlap pair", 1<<30, func() error {
		got, err := scorer.Score(ctx, 0, int64(right))
		if err != nil {
			return err
		}
		if got.Shared != degree || len(got.Bins) == 0 || math.IsNaN(got.RarityScore) {
			return fmt.Errorf("unexpected overlap: shared=%d rarity=%g bins=%d", got.Shared, got.RarityScore, len(got.Bins))
		}
		t.Logf("overlap: shared=%d rarity=%.8g bins=%d", got.Shared, got.RarityScore, len(got.Bins))
		return nil
	})
	measureFraudLoadPhase(t, "local PPR", 1<<30, func() error {
		scores, err := centrality.LocalPersonalizedPageRank(ctx, g,
			map[int64]float64{0: 2, int64(right): 1}, 0.85, 1e-7)
		if err != nil {
			return err
		}
		if scores[0] <= 0 || scores[int64(right)] <= 0 {
			return fmt.Errorf("missing seed scores: %v", scores)
		}
		t.Logf("local PPR: %d scored nodes, seed scores %.8g and %.8g",
			len(scores), scores[0], scores[int64(right)])
		return nil
	})
	runtime.KeepAlive(g)
	runtime.KeepAlive(scorer)
}

func loadTestSize(t *testing.T, name string, fallback int) int {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		t.Fatalf("%s must be a positive integer", name)
	}
	return n
}

func measureFraudLoadPhase(t *testing.T, name string, maxExtraHeap uint64, run func() error) {
	t.Helper()
	runtime.GC()
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)
	peak := baseline.HeapAlloc
	stop := make(chan struct{})
	finished := make(chan uint64, 1)
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var current runtime.MemStats
				runtime.ReadMemStats(&current)
				peak = max(peak, current.HeapAlloc)
			case <-stop:
				finished <- peak
				return
			}
		}
	}()
	start := time.Now()
	err := run()
	elapsed := time.Since(start)
	close(stop)
	peak = <-finished
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	peak = max(peak, after.HeapAlloc)
	runtime.GC()
	var retained runtime.MemStats
	runtime.ReadMemStats(&retained)
	t.Logf("%s: time=%s baseline_heap=%.1f MiB peak_heap=%.1f MiB retained_heap=%.1f MiB extra_peak=%.1f MiB",
		name, elapsed.Round(time.Millisecond), float64(baseline.HeapAlloc)/1048576,
		float64(peak)/1048576, float64(retained.HeapAlloc)/1048576,
		float64(peak-baseline.HeapAlloc)/1048576)
	if err != nil {
		t.Fatal(err)
	}
	if maxExtraHeap > 0 && peak-baseline.HeapAlloc > maxExtraHeap {
		t.Fatalf("%s added %.1f MiB peak heap, limit %.1f MiB", name,
			float64(peak-baseline.HeapAlloc)/1048576, float64(maxExtraHeap)/1048576)
	}
}
