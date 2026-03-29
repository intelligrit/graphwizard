// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"sort"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestBridges_SingleBridge(t *testing.T) {
	// Graph: 0--1--2--3  with 1--2 as a bridge
	//        |     |
	//        +--1  2--+
	// Actually: two triangles connected by a bridge
	// 0-1, 1-0 (triangle left: 0-1, 1-3, 3-0)
	// bridge: 1-2
	// triangle right: 2-4, 4-5, 5-2

	g := simple.NewUndirectedGraph()
	// Left triangle: 0-1-3
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))
	// Bridge
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	// Right triangle: 2-4-5
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(2)))

	bridges := Bridges(g)
	if len(bridges) != 1 {
		t.Fatalf("expected 1 bridge, got %d", len(bridges))
	}

	b := bridges[0]
	ids := []int64{b.From.ID(), b.To.ID()}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	if ids[0] != 1 || ids[1] != 2 {
		t.Errorf("expected bridge 1-2, got %d-%d", ids[0], ids[1])
	}
}

func TestBridges_NoBridges(t *testing.T) {
	// Complete graph K4: no bridges
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}

	bridges := Bridges(g)
	if len(bridges) != 0 {
		t.Fatalf("expected 0 bridges in K4, got %d", len(bridges))
	}
}

func TestBridges_AllBridges(t *testing.T) {
	// Path graph: 0--1--2--3 (all edges are bridges)
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	bridges := Bridges(g)
	if len(bridges) != 3 {
		t.Fatalf("expected 3 bridges in path graph, got %d", len(bridges))
	}
}

func TestBridges_DisconnectedGraph(t *testing.T) {
	// Two disconnected edges: 0--1  2--3
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	bridges := Bridges(g)
	if len(bridges) != 2 {
		t.Fatalf("expected 2 bridges, got %d", len(bridges))
	}
}

func TestBridges_EmptyGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	bridges := Bridges(g)
	if len(bridges) != 0 {
		t.Fatalf("expected 0 bridges in empty graph, got %d", len(bridges))
	}
}
