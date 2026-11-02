// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func nearBicliqueFixture() *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	for p := int64(0); p < 4; p++ {
		for e := int64(10); e < 13; e++ {
			if p == 3 && e == 12 {
				continue
			}
			g.SetEdge(g.NewEdge(simple.Node(p), simple.Node(e)))
		}
	}
	for p := int64(4); p < 7; p++ {
		for e := int64(20); e < 23; e++ {
			g.SetEdge(g.NewEdge(simple.Node(p), simple.Node(e)))
		}
	}
	for p := int64(0); p < 9; p++ {
		g.SetEdge(g.NewEdge(simple.Node(p), simple.Node(99)))
	}
	return g
}

func TestFindNearBicliqueBlocksTwoGroupsAndHub(t *testing.T) {
	g := nearBicliqueFixture()
	o := DefaultNearBicliqueOptions()
	o.MaxBridgeDegree, o.MaxBridges, o.MaxProviders = 4, 6, 6
	o.MinCoverage, o.MinDensity, o.TopK = 0.65, 0.8, 2
	seeds := []NearBicliqueSeed{{5, 4}, {1, 0}, {0, 1}}
	blocks, stats, err := FindNearBicliqueBlocks(context.Background(), g, func(id int64) bool { return id < 10 }, seeds, o)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Seeds != 2 || stats.Returned != 2 || len(blocks) != 2 {
		t.Fatalf("stats=%+v blocks=%+v", stats, blocks)
	}
	var first, second *NearBicliqueBlock
	for i := range blocks {
		if blocks[i].Seed == (NearBicliqueSeed{0, 1}) {
			first = &blocks[i]
		}
		if blocks[i].Seed == (NearBicliqueSeed{4, 5}) {
			second = &blocks[i]
		}
		for _, b := range blocks[i].Bridges {
			if b.ID == 99 {
				t.Fatalf("hub leaked: %+v", blocks[i])
			}
		}
	}
	if first == nil || second == nil {
		t.Fatalf("missing groups: %+v", blocks)
	}
	if !reflect.DeepEqual(first.ProviderIDs, []int64{0, 1, 2, 3}) || first.Edges != 11 || math.Abs(first.Density-11.0/12) > 1e-12 {
		t.Fatalf("first=%+v", first)
	}
	if !reflect.DeepEqual(second.ProviderIDs, []int64{4, 5, 6}) || second.Edges != 9 || second.Density != 1 {
		t.Fatalf("second=%+v", second)
	}
	again, againStats, err := FindNearBicliqueBlocks(context.Background(), g, func(id int64) bool { return id < 10 }, seeds, o)
	if err != nil || !reflect.DeepEqual(blocks, again) || stats != againStats {
		t.Fatalf("nondeterministic: %+v %+v %v", again, againStats, err)
	}
}

func TestFindNearBicliqueBlocksErrorsAndCaps(t *testing.T) {
	g := nearBicliqueFixture()
	left := func(id int64) bool { return id < 10 }
	o := DefaultNearBicliqueOptions()
	o.MaxBridgeDegree, o.MaxBridges, o.MaxProviders = 4, 6, 6
	o.MaxCandidateProviders = 2
	blocks, stats, err := FindNearBicliqueBlocks(context.Background(), g, left, []NearBicliqueSeed{{0, 1}}, o)
	if err != nil || len(blocks) != 0 || stats.Truncated != 1 {
		t.Fatalf("cap should reject group and report truncation: %+v %+v %v", blocks, stats, err)
	}
	if _, _, err := FindNearBicliqueBlocks(context.Background(), g, left, []NearBicliqueSeed{{0, 0}}, DefaultNearBicliqueOptions()); err == nil {
		t.Fatal("expected invalid seed")
	}
	if _, _, err := FindNearBicliqueBlocks(context.Background(), g, func(int64) bool { return true }, []NearBicliqueSeed{{0, 1}}, DefaultNearBicliqueOptions()); err == nil {
		t.Fatal("expected invalid bipartition")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := FindNearBicliqueBlocks(ctx, g, left, []NearBicliqueSeed{{0, 1}}, DefaultNearBicliqueOptions()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
