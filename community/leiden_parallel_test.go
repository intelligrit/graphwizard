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

// samePartition returns true when a and b induce the same equivalence classes
// on node IDs, ignoring community label numbering. This is the correct
// comparison for determinism: the algorithm is allowed to assign different
// integer labels to the same community across runs, as long as the SET of
// nodes in each community is identical.
func samePartition(a, b map[int64]int64) bool {
	if len(a) != len(b) {
		return false
	}
	// Build canonical form: for each community, its representative is the
	// smallest node ID in the community. Two maps are the same partition iff
	// they produce identical canonical maps.
	canon := func(m map[int64]int64) map[int64]int64 {
		rep := make(map[int64]int64, len(m))
		for nid, cid := range m {
			if r, ok := rep[cid]; !ok || nid < r {
				rep[cid] = nid
			}
		}
		out := make(map[int64]int64, len(m))
		for nid, cid := range m {
			out[nid] = rep[cid]
		}
		return out
	}
	return maps.Equal(canon(a), canon(b))
}

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

// TestLeidenParallelLargeScaleDeterministic is a scaled-up regression test for
// non-determinism in mega-communities. It uses a mixed SBM with 5 large blocks
// (200 nodes each) and 40 small blocks (20 nodes each) = 1,800 nodes total.
// The large blocks produce communities of 100-200+ nodes, which stress the
// parallel refinement path in ways that the smaller 1,000-node SBM test does
// not. Running with GOMAXPROCS=4 exercises concurrent goroutine scheduling.
func TestLeidenParallelLargeScaleDeterministic(t *testing.T) {
	g := buildMixedSBMGraph(t)

	old := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(old)

	const runs = 20
	var first map[int64]int64
	for i := 0; i < runs; i++ {
		rng := rand.New(rand.NewSource(42))
		got := LeidenParallel(context.Background(), g, 1.0, rng)
		if i == 0 {
			first = got
			continue
		}
		if !samePartition(got, first) {
			// Also report by maps.Equal to distinguish label drift from true splits.
			if maps.Equal(got, first) {
				t.Fatalf("run %d: community labels differ (same partition, different IDs)", i)
			}
			diff := 0
			for k, v := range got {
				if first[k] != v {
					diff++
				}
			}
			t.Fatalf("run %d: partition differs from run 0 (%d/%d nodes in different communities)",
				i, diff, len(got))
		}
	}
}

// TestLeidenParallelLargeScaleGOMAXPROCS verifies GOMAXPROCS-independence on
// the larger mixed-SBM graph. This is the targeted regression for the
// mega-community non-determinism reported against the Medicare bipartite graph.
func TestLeidenParallelLargeScaleGOMAXPROCS(t *testing.T) {
	g := buildMixedSBMGraph(t)

	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)

	rng0 := rand.New(rand.NewSource(42))
	reference := LeidenParallel(context.Background(), g, 1.0, rng0)

	for _, procs := range []int{2, 4, 8} {
		runtime.GOMAXPROCS(procs)
		rng := rand.New(rand.NewSource(42))
		got := LeidenParallel(context.Background(), g, 1.0, rng)
		if !samePartition(got, reference) {
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

// TestLeidenParallelLargeScaleMatchesSerial verifies that the parallel and
// serial implementations agree on the larger mixed-SBM graph.
func TestLeidenParallelLargeScaleMatchesSerial(t *testing.T) {
	g := buildMixedSBMGraph(t)

	old := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(old)

	rng1 := rand.New(rand.NewSource(42))
	serial := Leiden(context.Background(), g, 1.0, rng1)

	rng2 := rand.New(rand.NewSource(42))
	parallel := LeidenParallel(context.Background(), g, 1.0, rng2)

	if !samePartition(serial, parallel) {
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

// buildMixedSBMGraph builds a stochastic block model with mixed block sizes:
// 5 large blocks of 200 nodes and 40 small blocks of 20 nodes = 1,800 total.
// The large blocks produce communities large enough to stress the parallel
// refinement code paths that the 1,000-node uniform SBM does not exercise.
func buildMixedSBMGraph(t *testing.T) *simple.UndirectedGraph {
	t.Helper()
	rng := rand.New(rand.NewSource(7))
	g := simple.NewUndirectedGraph()

	// Build block structure: 5 blocks of 200 + 40 blocks of 20.
	type block struct{ start, size int }
	var blocks []block
	nodeID := 0
	for i := 0; i < 5; i++ {
		blocks = append(blocks, block{nodeID, 200})
		nodeID += 200
	}
	for i := 0; i < 40; i++ {
		blocks = append(blocks, block{nodeID, 20})
		nodeID += 20
	}
	for i := 0; i < nodeID; i++ {
		g.AddNode(simple.Node(i))
	}

	const pIntra = 0.10
	const pInter = 0.005

	for bi, b := range blocks {
		// Intra-block edges.
		for u := b.start; u < b.start+b.size; u++ {
			for v := u + 1; v < b.start+b.size; v++ {
				if rng.Float64() < pIntra {
					g.SetEdge(g.NewEdge(simple.Node(u), simple.Node(v)))
				}
			}
		}
		// Inter-block edges (only to later blocks to avoid duplicates).
		for bj := bi + 1; bj < len(blocks); bj++ {
			c := blocks[bj]
			for u := b.start; u < b.start+b.size; u++ {
				for v := c.start; v < c.start+c.size; v++ {
					if rng.Float64() < pInter {
						g.SetEdge(g.NewEdge(simple.Node(u), simple.Node(v)))
					}
				}
			}
		}
	}
	return g
}

