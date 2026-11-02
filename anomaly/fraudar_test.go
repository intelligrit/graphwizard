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

func fraudarFixture() *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph()
	for p := int64(0); p < 3; p++ {
		for b := int64(100); b < 103; b++ {
			g.SetEdge(g.NewEdge(simple.Node(p), simple.Node(b)))
		}
	}
	for p := int64(10); p < 20; p++ {
		g.SetEdge(g.NewEdge(simple.Node(p), simple.Node(110)))
	}
	g.AddNode(simple.Node(200))
	return g
}

func TestFRAUDARDenseBlockAndHubControl(t *testing.T) {
	g := fraudarFixture()
	isProvider := func(id int64) bool { return id < 100 }
	got, err := FRAUDAR(context.Background(), g, isProvider, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.ProviderIDs, []int64{0, 1, 2}) || !reflect.DeepEqual(got.BridgeIDs, []int64{100, 101, 102}) || got.Edges != 9 || got.Density != 1 || got.IDsTruncated {
		t.Fatalf("unexpected block %+v", got)
	}
	want := 9 / math.Log(8) / 6
	if math.Abs(got.WeightedDensity-want) > 1e-12 {
		t.Fatalf("weighted density %.12g want %.12g", got.WeightedDensity, want)
	}
	again, err := FRAUDAR(context.Background(), g, isProvider, 10)
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatalf("nondeterministic: %+v %v", again, err)
	}
	capped, err := FRAUDAR(context.Background(), g, isProvider, 2)
	if err != nil || !capped.IDsTruncated || capped.Providers != 3 || capped.Bridges != 3 || len(capped.ProviderIDs)+len(capped.BridgeIDs) != 2 {
		t.Fatalf("cap: %+v %v", capped, err)
	}
}

func TestFRAUDARErrors(t *testing.T) {
	g := fraudarFixture()
	if _, err := FRAUDAR(context.Background(), g, nil, 10); err == nil {
		t.Fatal("expected partition error")
	}
	if _, err := FRAUDAR(context.Background(), g, func(int64) bool { return true }, 10); err == nil {
		t.Fatal("expected bipartition error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := FRAUDAR(ctx, g, func(id int64) bool { return id < 100 }, 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
