// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"
	"sort"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestBridgesParallel_MatchesSequential(t *testing.T) {
	g := simple.NewUndirectedGraph()
	// Two triangles connected by a bridge.
	// Triangle 1: 0-1-2
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	// Triangle 2: 3-4-5
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))
	// Bridge: 2-3
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	seqBridges := Bridges(context.Background(), g)
	parBridges := BridgesParallel(context.Background(), g)

	// Normalize both for comparison.
	normalize := func(bridges []Bridge) [][2]int64 {
		var result [][2]int64
		for _, b := range bridges {
			a, z := b.From.ID(), b.To.ID()
			if a > z {
				a, z = z, a
			}
			result = append(result, [2]int64{a, z})
		}
		sort.Slice(result, func(i, j int) bool {
			if result[i][0] != result[j][0] {
				return result[i][0] < result[j][0]
			}
			return result[i][1] < result[j][1]
		})
		return result
	}

	seqNorm := normalize(seqBridges)
	parNorm := normalize(parBridges)

	if len(seqNorm) != len(parNorm) {
		t.Fatalf("bridge count mismatch: seq=%d par=%d", len(seqNorm), len(parNorm))
	}
	for i := range seqNorm {
		if seqNorm[i] != parNorm[i] {
			t.Errorf("bridge %d: seq=%v par=%v", i, seqNorm[i], parNorm[i])
		}
	}
}

func TestBridgesParallel_SingleComponent(t *testing.T) {
	// Chain: 0-1-2 (both edges are bridges).
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	bridges := BridgesParallel(context.Background(), g)
	if len(bridges) != 2 {
		t.Errorf("expected 2 bridges in chain, got %d", len(bridges))
	}
}

func TestBridgesParallel_DisconnectedNoBridges(t *testing.T) {
	// Two separate triangles (no bridges).
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(3)))

	bridges := BridgesParallel(context.Background(), g)
	if len(bridges) != 0 {
		t.Errorf("expected 0 bridges in disconnected triangles, got %d", len(bridges))
	}
}
