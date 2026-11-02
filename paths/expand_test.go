// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/intelligrit/graphwizard/diskgraph"
	"github.com/intelligrit/graphwizard/paths"
	"gonum.org/v1/gonum/graph/simple"
)

func TestExpandPathsBetween(t *testing.T) {
	g := simple.NewDirectedGraph()
	for _, edge := range [][2]int64{{0, 1}, {0, 2}, {1, 3}, {2, 3}, {3, 0}} {
		g.SetEdge(g.NewEdge(simple.Node(edge[0]), simple.Node(edge[1])))
	}
	target := int64(3)
	var got [][]int64
	n, err := paths.ExpandPaths(context.Background(), g, 0, paths.ExpandOptions{
		MinHops: 1, MaxHops: 3, MaxPaths: 3, Target: &target,
	}, func(path []int64) bool {
		got = append(got, path)
		return true
	})
	if err != nil || n != 2 {
		t.Fatalf("got count=%d err=%v, want count=2", n, err)
	}
	slices.SortFunc(got, func(a, b []int64) int { return slices.Compare(a, b) })
	want := [][]int64{{0, 1, 3}, {0, 2, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got paths %v, want %v", got, want)
	}
	// The back edge from 3 to 0 must not produce a three-hop path.
	n, err = paths.ExpandPaths(context.Background(), g, 0, paths.ExpandOptions{
		MinHops: 3, MaxHops: 3, MaxPaths: 10,
	}, func(path []int64) bool {
		t.Errorf("unexpected cyclic path %v", path)
		return true
	})
	if err != nil || n != 0 {
		t.Fatalf("got count=%d err=%v, want no cyclic paths", n, err)
	}
}

func TestExpandPathsLimitsAndFilter(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for _, edge := range [][2]int64{{0, 1}, {0, 2}, {1, 3}, {2, 3}} {
		g.SetEdge(g.NewEdge(simple.Node(edge[0]), simple.Node(edge[1])))
	}
	var got [][]int64
	n, err := paths.ExpandPaths(context.Background(), g, 0, paths.ExpandOptions{
		MinHops: 1, MaxHops: 2, MaxPaths: 2,
		AllowStep: func(from, to int64) bool { return to != 2 },
	}, func(path []int64) bool {
		got = append(got, path)
		return true
	})
	if err != nil || n != 2 {
		t.Fatalf("got count=%d err=%v, want count=2", n, err)
	}
	want := [][]int64{{0, 1}, {0, 1, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got paths %v, want %v", got, want)
	}
	// Returned paths are independent even after traversal has continued.
	got[0][1] = 99
	if got[1][1] != 1 {
		t.Fatal("visitor paths shared their backing array")
	}
}

func TestExpandPathsVisitStopAndZeroHop(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(7))
	called := 0
	n, err := paths.ExpandPaths(context.Background(), g, 7, paths.ExpandOptions{
		MaxHops: 0, MaxPaths: 1,
	}, func(path []int64) bool {
		called++
		if !reflect.DeepEqual(path, []int64{7}) {
			t.Errorf("got path %v, want [7]", path)
		}
		return false
	})
	if err != nil || n != 1 || called != 1 {
		t.Fatalf("got count=%d calls=%d err=%v", n, called, err)
	}
	if n, err := paths.ExpandPaths(context.Background(), g, 8, paths.ExpandOptions{MaxPaths: 1}, func([]int64) bool { return true }); n != 0 || err != nil {
		t.Fatalf("missing start: got count=%d err=%v", n, err)
	}

	g.SetEdge(g.NewEdge(simple.Node(7), simple.Node(8)))
	called = 0
	n, err = paths.ExpandPaths(context.Background(), g, 7, paths.ExpandOptions{
		MaxHops: 1, MaxPaths: 10,
	}, func([]int64) bool {
		called++
		return false
	})
	if err != nil || n != 1 || called != 1 {
		t.Fatalf("visitor stop: got count=%d calls=%d err=%v", n, called, err)
	}
}

func TestExpandPathsValidationAndCancellation(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	visit := func([]int64) bool { return true }
	for _, opts := range []paths.ExpandOptions{
		{MinHops: -1, MaxHops: 1, MaxPaths: 1},
		{MinHops: 2, MaxHops: 1, MaxPaths: 1},
		{MaxHops: 1, MaxPaths: 0},
	} {
		if _, err := paths.ExpandPaths(context.Background(), g, 0, opts, visit); err == nil {
			t.Errorf("expected validation error for %+v", opts)
		}
	}
	if _, err := paths.ExpandPaths(context.Background(), g, 0, paths.ExpandOptions{MaxPaths: 1}, nil); err == nil {
		t.Fatal("expected validation error for nil visitor")
	}
	ctx, cancel := context.WithCancel(context.Background())
	count, err := paths.ExpandPaths(ctx, g, 0, paths.ExpandOptions{MaxHops: 1, MaxPaths: 2}, func([]int64) bool {
		cancel()
		return true
	})
	if count != 1 || !errors.Is(err, context.Canceled) {
		t.Fatalf("got count=%d err=%v, want count=1 and cancellation", count, err)
	}
}

func TestExpandPathsDiskGraph(t *testing.T) {
	file := filepath.Join(t.TempDir(), "graph.db")
	b, err := diskgraph.NewUndirectedBuilder(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range [][2]int64{{0, 1}, {1, 2}, {0, 2}} {
		if err := b.AddEdge(edge[0], edge[1]); err != nil {
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
	target := int64(2)
	var got [][]int64
	n, err := paths.ExpandPaths(context.Background(), g, 0, paths.ExpandOptions{
		MinHops: 1, MaxHops: 2, MaxPaths: 3, Target: &target,
	}, func(path []int64) bool {
		got = append(got, path)
		return true
	})
	if err != nil || n != 2 {
		t.Fatalf("got count=%d err=%v, want count=2", n, err)
	}
	slices.SortFunc(got, func(a, b []int64) int { return slices.Compare(a, b) })
	want := [][]int64{{0, 1, 2}, {0, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got paths %v, want %v", got, want)
	}
}
