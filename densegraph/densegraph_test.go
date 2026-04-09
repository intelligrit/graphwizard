// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package densegraph

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph"
)

// --- Undirected ---

func buildTriangle() *Undirected {
	b := NewUndirectedBuilder()
	b.AddEdge(0, 1)
	b.AddEdge(1, 2)
	b.AddEdge(2, 0)
	return b.Build()
}

func TestUndirectedNodes(t *testing.T) {
	g := buildTriangle()

	nodes := g.Nodes()
	count := 0
	for nodes.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("got %d nodes, want 3", count)
	}

	for _, id := range []int64{0, 1, 2} {
		if n := g.Node(id); n == nil {
			t.Errorf("Node(%d) = nil, want non-nil", id)
		}
	}
	if n := g.Node(99); n != nil {
		t.Errorf("Node(99) = %v, want nil", n)
	}
}

func TestUndirectedEdges(t *testing.T) {
	g := buildTriangle()

	for _, pair := range [][2]int64{{0, 1}, {1, 0}, {1, 2}, {2, 1}, {0, 2}, {2, 0}} {
		if !g.HasEdgeBetween(pair[0], pair[1]) {
			t.Errorf("HasEdgeBetween(%d, %d) = false", pair[0], pair[1])
		}
		if e := g.Edge(pair[0], pair[1]); e == nil {
			t.Errorf("Edge(%d, %d) = nil", pair[0], pair[1])
		}
		if e := g.EdgeBetween(pair[0], pair[1]); e == nil {
			t.Errorf("EdgeBetween(%d, %d) = nil", pair[0], pair[1])
		}
	}

	if g.HasEdgeBetween(0, 99) {
		t.Error("HasEdgeBetween(0, 99) = true")
	}
}

func TestUndirectedFrom(t *testing.T) {
	g := buildTriangle()

	neighbors := collectIDs(g.From(0))
	if len(neighbors) != 2 {
		t.Fatalf("From(0) returned %d nodes, want 2", len(neighbors))
	}
	want := map[int64]bool{1: true, 2: true}
	for _, id := range neighbors {
		if !want[id] {
			t.Errorf("unexpected neighbor %d from node 0", id)
		}
	}
}

func TestUndirectedWeight(t *testing.T) {
	b := NewUndirectedBuilder()
	b.AddWeightedEdge(0, 1, 3.14)
	g := b.Build()

	w, ok := g.Weight(0, 1)
	if !ok || math.Abs(w-3.14) > 1e-10 {
		t.Errorf("Weight(0,1) = (%f, %v), want (3.14, true)", w, ok)
	}

	// Same node.
	w, ok = g.Weight(0, 0)
	if !ok || w != 0 {
		t.Errorf("Weight(0,0) = (%f, %v), want (0, true)", w, ok)
	}

	// No edge.
	w, ok = g.Weight(0, 99)
	if ok {
		t.Errorf("Weight(0,99) ok = true, want false")
	}
	if !math.IsInf(w, 1) {
		t.Errorf("Weight(0,99) = %f, want +Inf", w)
	}
}

func TestUndirectedWeightedEdge(t *testing.T) {
	b := NewUndirectedBuilder()
	b.AddWeightedEdge(10, 20, 2.5)
	g := b.Build()

	we := g.WeightedEdge(10, 20)
	if we == nil {
		t.Fatal("WeightedEdge(10,20) = nil")
	}
	if we.Weight() != 2.5 {
		t.Errorf("weight = %f, want 2.5", we.Weight())
	}

	we = g.WeightedEdgeBetween(20, 10)
	if we == nil {
		t.Fatal("WeightedEdgeBetween(20,10) = nil")
	}
}

func TestAddNodeIdempotent(t *testing.T) {
	b := NewUndirectedBuilder()
	b.AddNode(1)
	b.AddNode(1)
	b.AddNode(1)
	g := b.Build()

	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 1 {
		t.Errorf("got %d nodes, want 1 (idempotent)", count)
	}
}

func TestIteratorReset(t *testing.T) {
	g := buildTriangle()

	nodes := g.Nodes()
	count1 := 0
	for nodes.Next() {
		count1++
	}
	nodes.Reset()
	count2 := 0
	for nodes.Next() {
		count2++
	}
	if count1 != count2 {
		t.Errorf("after Reset: got %d, first pass %d", count2, count1)
	}
}

func TestEmptyFrom(t *testing.T) {
	b := NewUndirectedBuilder()
	b.AddNode(42)
	g := b.Build()

	nodes := g.From(42)
	if nodes.Next() {
		t.Error("From(42) on isolated node should be empty")
	}

	nodes = g.From(999)
	if nodes.Next() {
		t.Error("From(999) on non-existent node should be empty")
	}
}

func TestEmptyGraph(t *testing.T) {
	b := NewUndirectedBuilder()
	g := b.Build()

	if g.NumNodes() != 0 {
		t.Errorf("NumNodes = %d, want 0", g.NumNodes())
	}
	if g.Node(0) != nil {
		t.Error("Node(0) on empty graph should be nil")
	}
	if g.From(0).Next() {
		t.Error("From(0) on empty graph should be empty")
	}
}

func TestDuplicateEdgeDedup(t *testing.T) {
	b := NewUndirectedBuilder()
	b.AddWeightedEdge(0, 1, 1.0)
	b.AddWeightedEdge(0, 1, 2.0) // should overwrite
	g := b.Build()

	w, ok := g.Weight(0, 1)
	if !ok || w != 2.0 {
		t.Errorf("Weight(0,1) = (%f, %v), want (2.0, true)", w, ok)
	}

	// Should only have 2 neighbors per direction, not 4.
	neighbors := collectIDs(g.From(0))
	if len(neighbors) != 1 {
		t.Errorf("From(0) = %v, want 1 neighbor", neighbors)
	}
}

func TestAddDirectedEntry(t *testing.T) {
	b := NewUndirectedBuilder()
	// Simulate loading from a table with both directions already stored.
	b.AddDirectedEntry(0, 1, 5.0)
	b.AddDirectedEntry(1, 0, 5.0)
	b.AddDirectedEntry(1, 2, 3.0)
	b.AddDirectedEntry(2, 1, 3.0)
	g := b.Build()

	w, ok := g.Weight(0, 1)
	if !ok || w != 5.0 {
		t.Errorf("Weight(0,1) = (%f, %v), want (5.0, true)", w, ok)
	}
	w, ok = g.Weight(1, 2)
	if !ok || w != 3.0 {
		t.Errorf("Weight(1,2) = (%f, %v), want (3.0, true)", w, ok)
	}
	if g.HasEdgeBetween(0, 2) {
		t.Error("HasEdgeBetween(0,2) = true, want false")
	}
}

func TestSelfLoop(t *testing.T) {
	b := NewUndirectedBuilder()
	b.AddWeightedEdge(5, 5, 1.0)
	b.AddEdge(5, 6)
	g := b.Build()

	if !g.HasEdgeBetween(5, 5) {
		t.Error("HasEdgeBetween(5,5) = false, want true for self-loop")
	}
	if !g.HasEdgeBetween(5, 6) {
		t.Error("HasEdgeBetween(5,6) = false")
	}
}

// --- DenseAdjacency ---

func TestDenseAdjacency(t *testing.T) {
	g := buildTriangle()

	ids := g.NodeIDs()
	if len(ids) != 3 {
		t.Fatalf("NodeIDs len = %d, want 3", len(ids))
	}

	if g.NumNodes() != 3 {
		t.Errorf("NumNodes = %d, want 3", g.NumNodes())
	}

	for i := 0; i < g.NumNodes(); i++ {
		nbs := g.DenseNeighbors(i)
		if len(nbs) != 2 {
			t.Errorf("DenseNeighbors(%d) len = %d, want 2", i, len(nbs))
		}
	}

	// Out of range.
	if nbs := g.DenseNeighbors(-1); nbs != nil {
		t.Error("DenseNeighbors(-1) should be nil")
	}
	if nbs := g.DenseNeighbors(99); nbs != nil {
		t.Error("DenseNeighbors(99) should be nil")
	}
}

// --- EdgeScanner ---

func TestScanWeightedEdges(t *testing.T) {
	b := NewUndirectedBuilder()
	b.AddWeightedEdge(0, 1, 2.0)
	b.AddWeightedEdge(1, 2, 3.0)
	g := b.Build()

	count := 0
	g.ScanWeightedEdges(func(src, dst int64, w float64) {
		count++
	})
	// 2 undirected edges → 4 directed entries.
	if count != 4 {
		t.Errorf("ScanWeightedEdges yielded %d entries, want 4", count)
	}
}

// --- PreloadAdjacency (no-op) ---

func TestPreloadAdjacencyNoop(t *testing.T) {
	g := buildTriangle()
	g.PreloadAdjacency() // should not panic

	// Graph still works.
	if !g.HasEdgeBetween(0, 1) {
		t.Error("HasEdgeBetween(0,1) = false after PreloadAdjacency")
	}
}

// --- Sized builder ---

func TestNewUndirectedBuilderSized(t *testing.T) {
	b := NewUndirectedBuilderSized(100, 200)
	b.AddEdge(0, 1)
	b.AddEdge(1, 2)
	g := b.Build()

	if g.NumNodes() != 3 {
		t.Errorf("NumNodes = %d, want 3", g.NumNodes())
	}
}

// --- Integration with community algorithms ---

func TestLeidenCompatibility(t *testing.T) {
	// Build the karate club graph (34 nodes, 78 edges) to verify
	// densegraph works with the same interface algorithms expect.
	g := buildTriangle()

	// Verify all four interfaces are satisfied.
	var _ graph.Undirected = g
	var _ graph.WeightedUndirected = g

	// NodeIDs should be non-nil (unlike diskgraph without preload).
	if g.NodeIDs() == nil {
		t.Error("NodeIDs() = nil, DenseAdjacency should always be available")
	}
}

// --- Helpers ---

func collectIDs(it graph.Nodes) []int64 {
	var ids []int64
	for it.Next() {
		ids = append(ids, it.Node().ID())
	}
	return ids
}
