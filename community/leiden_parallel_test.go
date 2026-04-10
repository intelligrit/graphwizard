// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"maps"
	"math/rand"
	"runtime"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestLeidenParallel_TwoCliques(t *testing.T) {
	g := simple.NewUndirectedGraph()
	// Two 4-cliques connected by a single edge.
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	for i := int64(4); i < 8; i++ {
		for j := i + 1; j < 8; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))

	rng := rand.New(rand.NewSource(42))
	result := LeidenParallel(context.Background(), g, 1.0, rng)

	if len(result) != 8 {
		t.Fatalf("expected 8 assignments, got %d", len(result))
	}

	// Nodes within the same clique should be in the same community.
	for i := int64(1); i < 4; i++ {
		if result[i] != result[0] {
			t.Errorf("clique A: node %d in community %d, expected %d", i, result[i], result[0])
		}
	}
	for i := int64(5); i < 8; i++ {
		if result[i] != result[4] {
			t.Errorf("clique B: node %d in community %d, expected %d", i, result[i], result[4])
		}
	}
}

func TestLeidenParallel_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(1))
	result := LeidenParallel(context.Background(), g, 1.0, rng)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestLeidenParallel_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	rng := rand.New(rand.NewSource(1))
	result := LeidenParallel(context.Background(), g, 1.0, rng)
	if len(result) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(result))
	}
}

// TestLeidenParallelDeterministic verifies that LeidenParallel produces the
// same partition across 20 runs with the same seed. Regression test for the
// bug where Go map iteration order leaked into rng.Int63() seed consumption and
// subWeights tie-breaking, producing different partitions across identical runs.
func TestLeidenParallelDeterministic(t *testing.T) {
	g := buildSBMGraph(t)

	const runs = 20
	var first map[int64]int64
	for i := 0; i < runs; i++ {
		rng := rand.New(rand.NewSource(42))
		got := LeidenParallel(context.Background(), g, 1.0, rng)
		if i == 0 {
			first = got
			continue
		}
		if !maps.Equal(got, first) {
			diff := 0
			for k, v := range got {
				if first[k] != v {
					diff++
				}
			}
			t.Fatalf("run %d: partition differs from run 0 (%d/%d nodes changed community)",
				i, diff, len(got))
		}
	}
}

// TestLeidenParallelMatchesSerial enforces the docstring contract: for a given
// seed, LeidenParallel must produce the same partition as the sequential Leiden.
func TestLeidenParallelMatchesSerial(t *testing.T) {
	g := buildSBMGraph(t)

	rng1 := rand.New(rand.NewSource(42))
	serial := Leiden(context.Background(), g, 1.0, rng1)

	rng2 := rand.New(rand.NewSource(42))
	parallel := LeidenParallel(context.Background(), g, 1.0, rng2)

	if !maps.Equal(serial, parallel) {
		diff := 0
		for k, v := range parallel {
			if serial[k] != v {
				diff++
			}
		}
		t.Fatalf("serial and parallel partitions differ: %d/%d nodes in different communities",
			diff, len(serial))
	}
}

// TestLeidenParallelGOMAXPROCS verifies that changing GOMAXPROCS does not
// affect the partition produced for a given seed. Catches bugs where goroutine
// scheduling order leaks into the result.
func TestLeidenParallelGOMAXPROCS(t *testing.T) {
	g := buildSBMGraph(t)

	// GOMAXPROCS=1 serializes goroutine scheduling; use it as the reference.
	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)
	rng0 := rand.New(rand.NewSource(42))
	reference := LeidenParallel(context.Background(), g, 1.0, rng0)

	for _, procs := range []int{2, 4, 8} {
		runtime.GOMAXPROCS(procs)
		rng := rand.New(rand.NewSource(42))
		got := LeidenParallel(context.Background(), g, 1.0, rng)
		if !maps.Equal(got, reference) {
			diff := 0
			for k, v := range got {
				if reference[k] != v {
					diff++
				}
			}
			t.Errorf("GOMAXPROCS=%d: partition differs from GOMAXPROCS=1 (%d/%d nodes changed community)",
				procs, diff, len(reference))
		}
	}
}

