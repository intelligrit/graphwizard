// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"errors"
	"math"
	"math/big"
	"path/filepath"
	"testing"

	"github.com/intelligrit/graphwizard/diskgraph"
	"gonum.org/v1/gonum/graph/simple"
)

func bipartiteOverlapGraph() *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	// Five degree-2 bridge nodes and one degree-4 bridge node.
	for _, e := range [][2]int64{
		{0, 10}, {1, 10}, {0, 11}, {1, 11},
		{2, 12}, {3, 12}, {2, 13}, {4, 13}, {3, 14}, {4, 14},
		{0, 15}, {1, 15}, {2, 15}, {3, 15},
	} {
		g.SetEdge(g.NewEdge(simple.Node(e[0]), simple.Node(e[1])))
	}
	return g
}

func TestBipartiteOverlapScorerDegreeBins(t *testing.T) {
	g := bipartiteOverlapGraph()
	scorer, err := NewBipartiteOverlapScorer(context.Background(), g, func(id int64) bool { return id < 10 })
	if err != nil {
		t.Fatal(err)
	}
	got, err := scorer.Score(context.Background(), 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Shared != 3 || math.Abs(got.RarityScore-1.25) > 1e-12 {
		t.Fatalf("got shared=%d rarity=%.12f, want 3 and 1.25", got.Shared, got.RarityScore)
	}
	if len(got.Bins) != 2 {
		t.Fatalf("got bins %v, want 2", got.Bins)
	}
	bin2, bin4 := got.Bins[0], got.Bins[1]
	if bin2.BridgeDegree != 2 || bin2.Population != 5 || bin2.ACount != 2 || bin2.BCount != 2 || bin2.Shared != 2 || math.Abs(bin2.PValue-0.1) > 1e-12 {
		t.Errorf("unexpected degree-2 bin: %+v", bin2)
	}
	if bin4.BridgeDegree != 4 || bin4.Population != 1 || bin4.ACount != 1 || bin4.BCount != 1 || bin4.Shared != 1 || math.Abs(bin4.PValue-1) > 1e-12 {
		t.Errorf("unexpected degree-4 bin: %+v", bin4)
	}
	reversed, err := scorer.Score(context.Background(), 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if reversed.Shared != got.Shared || reversed.RarityScore != got.RarityScore || len(reversed.Bins) != len(got.Bins) {
		t.Fatalf("asymmetric overlap: forward=%+v reverse=%+v", got, reversed)
	}
	for i := range got.Bins {
		if math.Abs(reversed.Bins[i].PValue-got.Bins[i].PValue) > 1e-12 {
			t.Fatalf("asymmetric bin probability: forward=%+v reverse=%+v", got.Bins[i], reversed.Bins[i])
		}
	}

	other, err := scorer.Score(context.Background(), 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if other.Shared != 1 || math.Abs(other.RarityScore-0.25) > 1e-12 {
		t.Fatalf("got other pair %+v, want one common degree-4 bridge", other)
	}
}

func TestBipartiteOverlapScorerRejectsBadPartitionAndNodes(t *testing.T) {
	g := bipartiteOverlapGraph()
	if _, err := NewBipartiteOverlapScorer(context.Background(), g, nil); err == nil {
		t.Fatal("expected nil partition error")
	}
	if _, err := NewBipartiteOverlapScorer(context.Background(), g, func(int64) bool { return true }); err == nil {
		t.Fatal("expected non-bipartite partition error")
	}
	scorer, err := NewBipartiteOverlapScorer(context.Background(), g, func(id int64) bool { return id < 10 })
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]int64{{0, 0}, {0, 10}, {0, 99}} {
		if _, err := scorer.Score(context.Background(), pair[0], pair[1]); err == nil {
			t.Errorf("expected invalid pair error for %v", pair)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := scorer.Score(ctx, 0, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
}

func TestBipartiteOverlapScorerDiskGraph(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bipartite.db")
	b, err := diskgraph.NewUndirectedBuilder(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range [][2]int64{{0, 10}, {1, 10}, {0, 11}, {1, 11}, {2, 12}, {3, 12}} {
		if err := b.AddEdge(e[0], e[1]); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	g, err := diskgraph.OpenUndirected(file)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	scorer, err := NewBipartiteOverlapScorer(context.Background(), g, func(id int64) bool { return id < 10 })
	if err != nil {
		t.Fatal(err)
	}
	got, err := scorer.Score(context.Background(), 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Shared != 2 || got.RarityScore != 1 || len(got.Bins) != 1 || math.Abs(got.Bins[0].PValue-1.0/3.0) > 1e-12 {
		t.Fatalf("unexpected disk graph overlap: %+v", got)
	}
}

func TestHypergeometricOverlapTail(t *testing.T) {
	for _, tc := range []struct {
		N, a, b, observed int
		want              float64
	}{
		{5, 2, 2, 2, 0.1},
		{5, 2, 2, 1, 0.7},
		{5, 2, 2, 0, 1},
		{5, 2, 2, 3, 0},
		{3, 2, 2, 1, 1},
	} {
		got := hypergeometricOverlapTail(tc.N, tc.a, tc.b, tc.observed)
		if math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("P(X >= %d) for N=%d a=%d b=%d: got %.12f, want %.12f", tc.observed, tc.N, tc.a, tc.b, got, tc.want)
		}
	}
}

func TestHypergeometricOverlapTailMatchesCombinatorialCount(t *testing.T) {
	choose := func(n, k int) float64 {
		return float64(new(big.Int).Binomial(int64(n), int64(k)).Int64())
	}
	for N := 1; N <= 12; N++ {
		for a := 0; a <= N; a++ {
			for b := 0; b <= N; b++ {
				for observed := 0; observed <= min(a, b)+1; observed++ {
					want := 0.0
					for x := max(observed, a+b-N, 0); x <= min(a, b); x++ {
						want += choose(a, x) * choose(N-a, b-x) / choose(N, b)
					}
					got := hypergeometricOverlapTail(N, a, b, observed)
					if math.Abs(got-want) > 1e-11 {
						t.Fatalf("N=%d a=%d b=%d observed=%d: got %.12g want %.12g", N, a, b, observed, got, want)
					}
				}
			}
		}
	}
}

func TestHypergeometricOverlapTailLargeSparsePopulation(t *testing.T) {
	const population = 1_000_000_000
	want := 2 / (float64(population) * float64(population-1))
	got := hypergeometricOverlapTail(population, 2, 2, 2)
	if math.Abs(got-want)/want > 1e-10 {
		t.Fatalf("large sparse tail: got %.12g, want %.12g", got, want)
	}
}
