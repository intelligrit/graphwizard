// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package stream

import (
	"sync"
	"testing"
)

func TestStreamGraph_AddNode(t *testing.T) {
	sg := New()
	sg.AddNode(1)
	sg.AddNode(2)
	sg.AddNode(1) // duplicate, should be no-op

	changes := sg.Changes()
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes (duplicate ignored), got %d", len(changes))
	}
	if changes[0].Kind != AddNodeChange || changes[0].From != 1 {
		t.Errorf("unexpected change 0: %+v", changes[0])
	}
	if changes[1].Kind != AddNodeChange || changes[1].From != 2 {
		t.Errorf("unexpected change 1: %+v", changes[1])
	}

	g := sg.Graph()
	if g.Node(1) == nil || g.Node(2) == nil {
		t.Error("nodes should exist in graph")
	}
}

func TestStreamGraph_RemoveNode(t *testing.T) {
	sg := New()
	sg.AddNode(1)
	sg.AddNode(2)
	sg.AddEdge(1, 2, 1.0)
	sg.RemoveNode(1)

	g := sg.Graph()
	if g.Node(1) != nil {
		t.Error("node 1 should be removed")
	}
	if g.HasEdgeBetween(1, 2) {
		t.Error("edge 1-2 should be removed with node")
	}
}

func TestStreamGraph_RemoveNonexistent(t *testing.T) {
	sg := New()
	sg.RemoveNode(99) // no-op
	sg.RemoveEdge(1, 2) // no-op

	if len(sg.Changes()) != 0 {
		t.Errorf("expected 0 changes for no-ops, got %d", len(sg.Changes()))
	}
}

func TestStreamGraph_AddEdge(t *testing.T) {
	sg := New()
	sg.AddEdge(1, 2, 3.14)

	changes := sg.Changes()
	// Should have: AddNode(1), AddNode(2), AddEdge(1,2)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d: %+v", len(changes), changes)
	}

	g := sg.Graph()
	w, ok := g.Weight(1, 2)
	if !ok || w != 3.14 {
		t.Errorf("expected weight 3.14, got %f (ok=%v)", w, ok)
	}
}

func TestStreamGraph_RemoveEdge(t *testing.T) {
	sg := New()
	sg.AddEdge(1, 2, 1.0)
	sg.RemoveEdge(1, 2)

	g := sg.Graph()
	if g.HasEdgeBetween(1, 2) {
		t.Error("edge should be removed")
	}

	changes := sg.Changes()
	last := changes[len(changes)-1]
	if last.Kind != RemoveEdgeChange || last.From != 1 || last.To != 2 {
		t.Errorf("unexpected last change: %+v", last)
	}
}

func TestStreamGraph_Flush(t *testing.T) {
	sg := New()
	sg.AddNode(1)
	sg.AddNode(2)

	if len(sg.Changes()) != 2 {
		t.Fatalf("expected 2 changes before flush")
	}

	sg.Flush()

	if len(sg.Changes()) != 0 {
		t.Errorf("expected 0 changes after flush, got %d", len(sg.Changes()))
	}

	// Graph should still have the nodes.
	g := sg.Graph()
	if g.Node(1) == nil || g.Node(2) == nil {
		t.Error("flush should not affect graph state")
	}
}

func TestStreamGraph_ChangesIsCopy(t *testing.T) {
	sg := New()
	sg.AddNode(1)

	c := sg.Changes()
	c[0].From = 999 // mutate the copy

	original := sg.Changes()
	if original[0].From == 999 {
		t.Error("Changes should return a copy, not the internal slice")
	}
}

func TestStreamGraph_ConcurrentAccess(t *testing.T) {
	sg := New()
	var wg sync.WaitGroup

	// Concurrent writers.
	for i := int64(0); i < 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			sg.AddNode(id)
		}(i)
	}
	wg.Wait()

	// Concurrent readers.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sg.Changes()
			_ = sg.Graph()
		}()
	}
	wg.Wait()

	g := sg.Graph()
	count := 0
	nodes := g.Nodes()
	for nodes.Next() {
		count++
	}
	if count != 100 {
		t.Errorf("expected 100 nodes, got %d", count)
	}
}

func TestStreamGraph_UpdateEdgeWeight(t *testing.T) {
	sg := New()
	sg.AddEdge(1, 2, 1.0)
	sg.AddEdge(1, 2, 5.0) // update weight

	g := sg.Graph()
	w, ok := g.Weight(1, 2)
	if !ok || w != 5.0 {
		t.Errorf("expected updated weight 5.0, got %f", w)
	}
}
