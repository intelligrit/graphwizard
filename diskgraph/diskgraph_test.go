// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"gonum.org/v1/gonum/graph"
)

func tempPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

// --- Undirected ---

func buildTriangle(t *testing.T) string {
	t.Helper()
	path := tempPath(t, "tri.db")
	b, err := NewUndirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	// Triangle: 0-1, 1-2, 2-0
	if err := b.AddEdge(0, 1); err != nil {
		t.Fatal(err)
	}
	if err := b.AddEdge(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := b.AddEdge(2, 0); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestUndirectedNodes(t *testing.T) {
	path := buildTriangle(t)
	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	// Check node count.
	nodes := g.Nodes()
	count := 0
	for nodes.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("got %d nodes, want 3", count)
	}

	// Check specific nodes.
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
	path := buildTriangle(t)
	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	// All edges exist in both directions.
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

	// Non-existent edge.
	if g.HasEdgeBetween(0, 99) {
		t.Error("HasEdgeBetween(0, 99) = true")
	}
}

func TestUndirectedFrom(t *testing.T) {
	path := buildTriangle(t)
	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	// Node 0 should reach nodes 1 and 2.
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
	path := tempPath(t, "weighted.db")
	b, err := NewUndirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.AddWeightedEdge(0, 1, 3.14); err != nil {
		t.Fatal(err)
	}
	b.Close()

	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

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
	path := tempPath(t, "we.db")
	b, err := NewUndirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.AddWeightedEdge(10, 20, 2.5); err != nil {
		t.Fatal(err)
	}
	b.Close()

	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	we := g.WeightedEdge(10, 20)
	if we == nil {
		t.Fatal("WeightedEdge(10,20) = nil")
	}
	if we.Weight() != 2.5 {
		t.Errorf("weight = %f, want 2.5", we.Weight())
	}

	// Reverse direction should also work for undirected.
	we = g.WeightedEdgeBetween(20, 10)
	if we == nil {
		t.Fatal("WeightedEdgeBetween(20,10) = nil")
	}
}

// --- Directed ---

func buildDAG(t *testing.T) string {
	t.Helper()
	path := tempPath(t, "dag.db")
	b, err := NewDirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	// 0->1, 0->2, 1->2
	if err := b.AddEdge(0, 1); err != nil {
		t.Fatal(err)
	}
	if err := b.AddEdge(0, 2); err != nil {
		t.Fatal(err)
	}
	if err := b.AddEdge(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDirectedFrom(t *testing.T) {
	path := buildDAG(t)
	g, err := OpenDirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	// Node 0 -> {1, 2}.
	ids := collectIDs(g.From(0))
	if len(ids) != 2 {
		t.Fatalf("From(0) = %v, want 2 nodes", ids)
	}

	// Node 2 has no outgoing edges.
	ids = collectIDs(g.From(2))
	if len(ids) != 0 {
		t.Errorf("From(2) = %v, want empty", ids)
	}
}

func TestDirectedTo(t *testing.T) {
	path := buildDAG(t)
	g, err := OpenDirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	// Node 2 <- {0, 1}.
	ids := collectIDs(g.To(2))
	if len(ids) != 2 {
		t.Fatalf("To(2) = %v, want 2 nodes", ids)
	}

	// Node 0 has no incoming edges.
	ids = collectIDs(g.To(0))
	if len(ids) != 0 {
		t.Errorf("To(0) = %v, want empty", ids)
	}
}

func TestDirectedHasEdge(t *testing.T) {
	path := buildDAG(t)
	g, err := OpenDirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if !g.HasEdgeFromTo(0, 1) {
		t.Error("HasEdgeFromTo(0,1) = false")
	}
	if g.HasEdgeFromTo(1, 0) {
		t.Error("HasEdgeFromTo(1,0) = true, want false")
	}

	// HasEdgeBetween is direction-agnostic.
	if !g.HasEdgeBetween(1, 0) {
		t.Error("HasEdgeBetween(1,0) = false")
	}
}

func TestDirectedWeight(t *testing.T) {
	path := tempPath(t, "dw.db")
	b, err := NewDirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.AddWeightedEdge(0, 1, 7.77); err != nil {
		t.Fatal(err)
	}
	b.Close()

	g, err := OpenDirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	w, ok := g.Weight(0, 1)
	if !ok || math.Abs(w-7.77) > 1e-10 {
		t.Errorf("Weight(0,1) = (%f, %v), want (7.77, true)", w, ok)
	}

	// Reverse should not exist.
	_, ok = g.Weight(1, 0)
	if ok {
		t.Error("Weight(1,0) ok = true, want false")
	}
}

func TestBuilderAddNodeIdempotent(t *testing.T) {
	path := tempPath(t, "idem.db")
	b, err := NewUndirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	b.AddNode(1)
	b.AddNode(1)
	b.AddNode(1)
	b.Close()

	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

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
	path := buildTriangle(t)
	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

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

func TestOpenNonexistent(t *testing.T) {
	_, err := OpenUndirected("/nonexistent/path.db")
	if err == nil {
		t.Error("expected error opening nonexistent file")
	}
	_, err = OpenDirected("/nonexistent/path.db")
	if err == nil {
		t.Error("expected error opening nonexistent file")
	}
}

func TestEmptyFrom(t *testing.T) {
	path := tempPath(t, "iso.db")
	b, _ := NewUndirectedBuilder(path)
	b.AddNode(42)
	b.Close()

	g, _ := OpenUndirected(path)
	defer g.Close()

	nodes := g.From(42)
	if nodes.Next() {
		t.Error("From(42) on isolated node should be empty")
	}

	// Non-existent node.
	nodes = g.From(999)
	if nodes.Next() {
		t.Error("From(999) on non-existent node should be empty")
	}
}

func TestFileExistsAfterClose(t *testing.T) {
	path := buildTriangle(t)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("bolt file does not exist after builder.Close()")
	}
}

func TestBatchUndirected(t *testing.T) {
	path := tempPath(t, "batch.db")
	b, err := NewUndirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Batch(func(tx *UndirectedTx) error {
		for i := int64(0); i < 100; i++ {
			for j := i + 1; j < 100; j++ {
				if (i+j)%7 == 0 {
					if err := tx.AddEdge(i, j); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	b.Close()

	g, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 100 {
		t.Errorf("got %d nodes, want 100", count)
	}

	// Spot check an edge.
	if !g.HasEdgeBetween(0, 7) {
		t.Error("expected edge 0-7")
	}
}

func TestBatchDirected(t *testing.T) {
	path := tempPath(t, "dbatch.db")
	b, err := NewDirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Batch(func(tx *DirectedTx) error {
		tx.AddEdge(0, 1)
		tx.AddEdge(1, 2)
		tx.AddEdge(2, 0)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	b.Close()

	g, err := OpenDirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()

	if !g.HasEdgeFromTo(0, 1) {
		t.Error("expected 0->1")
	}
	if g.HasEdgeFromTo(1, 0) {
		t.Error("unexpected 1->0")
	}
	ids := collectIDs(g.To(0))
	if len(ids) != 1 || ids[0] != 2 {
		t.Errorf("To(0) = %v, want [2]", ids)
	}
}

// collectIDs drains a graph.Nodes iterator into a slice of IDs.
func collectIDs(it graph.Nodes) []int64 {
	var ids []int64
	for it.Next() {
		ids = append(ids, it.Node().ID())
	}
	return ids
}
