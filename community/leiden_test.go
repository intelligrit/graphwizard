// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"maps"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestLeiden_TwoClusters(t *testing.T) {
	// Two dense clusters connected by a single weak link.
	// Cluster A: 0-1-2-0 (triangle)
	// Cluster B: 3-4-5-3 (triangle)
	// Bridge: 2-3
	g := simple.NewUndirectedGraph()
	// Cluster A
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	// Cluster B
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))
	// Bridge
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)

	// Nodes in the same cluster should be in the same community.
	if comms[0] != comms[1] || comms[1] != comms[2] {
		t.Errorf("cluster A nodes should be in same community: %d, %d, %d", comms[0], comms[1], comms[2])
	}
	if comms[3] != comms[4] || comms[4] != comms[5] {
		t.Errorf("cluster B nodes should be in same community: %d, %d, %d", comms[3], comms[4], comms[5])
	}
	// The two clusters should be in different communities.
	if comms[0] == comms[3] {
		t.Error("clusters A and B should be in different communities")
	}
}

func TestLeiden_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)
	if len(comms) != 1 {
		t.Errorf("expected 1 community assignment, got %d", len(comms))
	}
}

func TestLeiden_CompleteGraph(t *testing.T) {
	// K4: all nodes should be in the same community at resolution 1.0
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)

	first := comms[0]
	for id, c := range comms {
		if c != first {
			t.Errorf("K4: all nodes should be in same community; node %d in %d, expected %d", id, c, first)
		}
	}
}

func TestLeiden_EmptyGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(42))
	comms := Leiden(context.Background(), g, 1.0, rng)
	if len(comms) != 0 {
		t.Errorf("expected empty result for empty graph, got %v", comms)
	}
}

// TestLeidenDeterministic verifies that Leiden produces the same partition
// across 20 runs with the same seed. Regression test for non-determinism
// caused by Go map iteration order leaking into refine's rng.Perm call order.
func TestLeidenDeterministic(t *testing.T) {
	g := buildSBMGraph(t)

	const runs = 20
	var first map[int64]int64
	for i := 0; i < runs; i++ {
		rng := rand.New(rand.NewSource(42))
		got := Leiden(context.Background(), g, 1.0, rng)
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

// buildSBMGraph builds a stochastic block model graph: 10 blocks of 100 nodes,
// intra-block edge probability 0.10, inter-block 0.01. At 1000 nodes the graph
// is large enough that Go's map iteration randomization reliably fires across
// runs, making it a reliable regression target for determinism bugs.
func buildSBMGraph(t *testing.T) *simple.UndirectedGraph {
	t.Helper()
	const (
		blocks   = 10
		perBlock = 100
		pIntra   = 0.10
		pInter   = 0.01
	)
	rng := rand.New(rand.NewSource(1))
	g := simple.NewUndirectedGraph()
	for i := 0; i < blocks*perBlock; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := 0; i < blocks*perBlock; i++ {
		for j := i + 1; j < blocks*perBlock; j++ {
			p := pInter
			if i/perBlock == j/perBlock {
				p = pIntra
			}
			if rng.Float64() < p {
				g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
			}
		}
	}
	return g
}
