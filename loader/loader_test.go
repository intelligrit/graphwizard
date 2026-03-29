// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package loader

import (
	"fmt"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

// mockRows implements the RowScanner interface for testing without a real DB.
type mockRows struct {
	data    [][]interface{}
	index   int
	scanErr error
	iterErr error
}

func newMockRows(data [][]interface{}) *mockRows {
	return &mockRows{data: data, index: -1}
}

func (m *mockRows) Next() bool {
	m.index++
	return m.index < len(m.data)
}

func (m *mockRows) Scan(dest ...interface{}) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	row := m.data[m.index]
	for i, d := range dest {
		switch ptr := d.(type) {
		case *int64:
			switch v := row[i].(type) {
			case int64:
				*ptr = v
			case int:
				*ptr = int64(v)
			}
		case *float64:
			switch v := row[i].(type) {
			case float64:
				*ptr = v
			case int:
				*ptr = float64(v)
			}
		}
	}
	return nil
}

func (m *mockRows) Err() error {
	return m.iterErr
}

func TestLoadDirectedFromRows(t *testing.T) {
	rows := newMockRows([][]interface{}{
		{int64(1), int64(2)},
		{int64(2), int64(3)},
		{int64(3), int64(1)},
	})

	g, err := loadDirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.HasEdgeFromTo(1, 2) {
		t.Error("missing edge 1->2")
	}
	if !g.HasEdgeFromTo(2, 3) {
		t.Error("missing edge 2->3")
	}
	if !g.HasEdgeFromTo(3, 1) {
		t.Error("missing edge 3->1")
	}
	if g.HasEdgeFromTo(2, 1) {
		t.Error("unexpected edge 2->1 in directed graph")
	}
	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 nodes, got %d", count)
	}
}

func TestLoadDirectedFromRows_Empty(t *testing.T) {
	rows := newMockRows(nil)

	g, err := loadDirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 nodes, got %d", count)
	}
}

func TestLoadDirectedFromRows_ScanError(t *testing.T) {
	rows := &mockRows{
		data:    [][]interface{}{{int64(1), int64(2)}},
		index:   -1,
		scanErr: fmt.Errorf("scan failure"),
	}

	_, err := loadDirectedFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadDirectedFromRows_IterError(t *testing.T) {
	rows := &mockRows{
		data:    nil,
		index:   -1,
		iterErr: fmt.Errorf("iteration failure"),
	}

	_, err := loadDirectedFromRows(rows)
	if err == nil {
		t.Fatal("expected error from rows.Err()")
	}
}

func TestLoadUndirectedFromRows(t *testing.T) {
	rows := newMockRows([][]interface{}{
		{int64(1), int64(2)},
		{int64(2), int64(3)},
	})

	g, err := loadUndirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.HasEdgeBetween(1, 2) {
		t.Error("missing edge 1-2")
	}
	if !g.HasEdgeBetween(2, 3) {
		t.Error("missing edge 2-3")
	}
	if !g.HasEdgeBetween(2, 1) {
		t.Error("undirected edge 2-1 should exist")
	}
}

func TestLoadUndirectedFromRows_Empty(t *testing.T) {
	rows := newMockRows(nil)

	g, err := loadUndirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 nodes, got %d", count)
	}
}

func TestLoadUndirectedFromRows_ScanError(t *testing.T) {
	rows := &mockRows{
		data:    [][]interface{}{{int64(1), int64(2)}},
		index:   -1,
		scanErr: fmt.Errorf("scan failure"),
	}

	_, err := loadUndirectedFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadWeightedDirectedFromRows(t *testing.T) {
	rows := newMockRows([][]interface{}{
		{int64(1), int64(2), float64(1.5)},
		{int64(2), int64(3), float64(2.5)},
	})

	g, err := loadWeightedDirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w, ok := g.Weight(1, 2)
	if !ok || w != 1.5 {
		t.Errorf("expected weight 1.5, got %v (ok=%v)", w, ok)
	}
	w, ok = g.Weight(2, 3)
	if !ok || w != 2.5 {
		t.Errorf("expected weight 2.5, got %v (ok=%v)", w, ok)
	}
}

func TestLoadWeightedDirectedFromRows_Empty(t *testing.T) {
	rows := newMockRows(nil)

	g, err := loadWeightedDirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 nodes, got %d", count)
	}
}

func TestLoadWeightedDirectedFromRows_ScanError(t *testing.T) {
	rows := &mockRows{
		data:    [][]interface{}{{int64(1), int64(2), float64(1.0)}},
		index:   -1,
		scanErr: fmt.Errorf("scan failure"),
	}

	_, err := loadWeightedDirectedFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadWeightedUndirectedFromRows(t *testing.T) {
	rows := newMockRows([][]interface{}{
		{int64(10), int64(20), float64(3.14)},
		{int64(20), int64(30), float64(2.71)},
	})

	g, err := loadWeightedUndirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w, ok := g.Weight(10, 20)
	if !ok || w != 3.14 {
		t.Errorf("expected weight 3.14, got %v (ok=%v)", w, ok)
	}
	w, ok = g.Weight(20, 30)
	if !ok || w != 2.71 {
		t.Errorf("expected weight 2.71, got %v (ok=%v)", w, ok)
	}
	// Undirected: reverse direction should also work.
	w, ok = g.Weight(20, 10)
	if !ok || w != 3.14 {
		t.Errorf("expected weight 3.14 for reverse, got %v (ok=%v)", w, ok)
	}
}

func TestLoadWeightedUndirectedFromRows_Empty(t *testing.T) {
	rows := newMockRows(nil)

	g, err := loadWeightedUndirectedFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 nodes, got %d", count)
	}
}

func TestLoadWeightedUndirectedFromRows_ScanError(t *testing.T) {
	rows := &mockRows{
		data:    [][]interface{}{{int64(1), int64(2), float64(1.0)}},
		index:   -1,
		scanErr: fmt.Errorf("scan failure"),
	}

	_, err := loadWeightedUndirectedFromRows(rows)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnsureNodeDirected(t *testing.T) {
	g := simple.NewDirectedGraph()
	seen := make(map[int64]bool)

	ensureNodeDirected(g, 1, seen)
	ensureNodeDirected(g, 2, seen)
	ensureNodeDirected(g, 1, seen) // duplicate

	if g.Node(1) == nil {
		t.Error("expected node 1")
	}
	if g.Node(2) == nil {
		t.Error("expected node 2")
	}
	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 nodes, got %d", count)
	}
}

func TestEnsureNodeUndirected(t *testing.T) {
	g := simple.NewUndirectedGraph()
	seen := make(map[int64]bool)

	ensureNodeUndirected(g, 10, seen)
	ensureNodeUndirected(g, 20, seen)
	ensureNodeUndirected(g, 10, seen) // duplicate

	if g.Node(10) == nil {
		t.Error("expected node 10")
	}
	if g.Node(20) == nil {
		t.Error("expected node 20")
	}
}

func TestEnsureNodeWeightedDirected(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, 0)
	seen := make(map[int64]bool)

	ensureNodeWeightedDirected(g, 5, seen)
	ensureNodeWeightedDirected(g, 5, seen) // duplicate

	if g.Node(5) == nil {
		t.Error("expected node 5")
	}
}

func TestEnsureNodeWeightedUndirected(t *testing.T) {
	g := simple.NewWeightedUndirectedGraph(0, 0)
	seen := make(map[int64]bool)

	ensureNodeWeightedUndirected(g, 7, seen)
	if g.Node(7) == nil {
		t.Error("expected node 7")
	}
}
