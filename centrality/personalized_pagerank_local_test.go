// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"testing"

	"github.com/intelligrit/graphwizard/diskgraph"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type noNodesUndirected struct{ graph.Undirected }

func (noNodesUndirected) Nodes() graph.Nodes { panic("whole-graph scan") }

func TestLocalPersonalizedPageRankMatchesFullPPR(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for _, e := range [][2]int64{{0, 1}, {1, 2}, {2, 3}, {1, 3}} {
		g.SetEdge(g.NewEdge(simple.Node(e[0]), simple.Node(e[1])))
	}
	g.AddNode(simple.Node(4))
	seeds := map[int64]float64{0: 2, 3: 1}
	got, err := LocalPersonalizedPageRank(context.Background(), noNodesUndirected{g}, seeds, 0.85, 1e-9)
	if err != nil {
		t.Fatal(err)
	}
	from0 := PersonalizedPageRankUndirected(context.Background(), g, 0, 0.85, 1e-14, 1000)
	from3 := PersonalizedPageRankUndirected(context.Background(), g, 3, 0.85, 1e-14, 1000)
	for _, id := range []int64{0, 1, 2, 3} {
		want := (2*from0[id] + from3[id]) / 3
		if math.Abs(got[id]-want) > 1e-6 {
			t.Errorf("node %d: got %.9f, want %.9f", id, got[id], want)
		}
	}
	if _, ok := got[4]; ok {
		t.Fatal("unvisited isolated node appeared in sparse result")
	}
}

func TestLocalPersonalizedPageRankIsolatedSeeds(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(1))
	g.AddNode(simple.Node(2))
	got, err := LocalPersonalizedPageRank(context.Background(), g, map[int64]float64{1: 3, 2: 1}, 0.85, 1e-6)
	if err != nil {
		t.Fatal(err)
	}
	if got[1] != 0.75 || got[2] != 0.25 {
		t.Fatalf("got isolated seed scores %v", got)
	}
}

func TestLocalPersonalizedPageRankDiskGraph(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ppr.db")
	b, err := diskgraph.NewUndirectedBuilder(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range [][2]int64{{0, 1}, {1, 2}, {2, 3}} {
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
	got, err := LocalPersonalizedPageRank(
		context.Background(), noNodesUndirected{g}, map[int64]float64{0: 1, 3: 1}, 0.85, 1e-8,
	)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got[0]-got[3]) > 1e-7 || math.Abs(got[1]-got[2]) > 1e-7 {
		t.Fatalf("symmetric seeds on a path produced asymmetric scores: %v", got)
	}
	inMemory := simple.NewUndirectedGraph()
	for _, e := range [][2]int64{{0, 1}, {1, 2}, {2, 3}} {
		inMemory.SetEdge(inMemory.NewEdge(simple.Node(e[0]), simple.Node(e[1])))
	}
	from0 := PersonalizedPageRankUndirected(context.Background(), inMemory, 0, 0.85, 1e-14, 1000)
	from3 := PersonalizedPageRankUndirected(context.Background(), inMemory, 3, 0.85, 1e-14, 1000)
	for _, id := range []int64{0, 1, 2, 3} {
		want := (from0[id] + from3[id]) / 2
		if math.Abs(got[id]-want) > 1e-6 {
			t.Errorf("node %d: got %.9f, want %.9f", id, got[id], want)
		}
	}
}

func TestLocalPersonalizedPageRankValidationAndCancellation(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(1))
	for _, tc := range []struct {
		seeds   map[int64]float64
		damping float64
		epsilon float64
	}{
		{map[int64]float64{1: 1}, -0.1, 1e-3},
		{map[int64]float64{1: 1}, 1, 1e-3},
		{map[int64]float64{1: 1}, 0.85, 0},
		{map[int64]float64{1: math.NaN()}, 0.85, 1e-3},
		{map[int64]float64{1: -1}, 0.85, 1e-3},
		{map[int64]float64{2: 1}, 0.85, 1e-3},
		{map[int64]float64{}, 0.85, 1e-3},
	} {
		if _, err := LocalPersonalizedPageRank(context.Background(), g, tc.seeds, tc.damping, tc.epsilon); err == nil {
			t.Errorf("expected error for %+v", tc)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := LocalPersonalizedPageRank(ctx, g, map[int64]float64{1: 1}, 0.85, 1e-3); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context cancellation", err)
	}
}
