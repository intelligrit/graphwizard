// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diff

import (
	"reflect"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestCompare_AddedNodes(t *testing.T) {
	before := simple.NewDirectedGraph()
	before.AddNode(simple.Node(0))
	before.AddNode(simple.Node(1))

	after := simple.NewDirectedGraph()
	after.AddNode(simple.Node(0))
	after.AddNode(simple.Node(1))
	after.AddNode(simple.Node(2))

	d := Compare(before, after)
	if !reflect.DeepEqual(d.AddedNodes, []int64{2}) {
		t.Errorf("expected added nodes [2], got %v", d.AddedNodes)
	}
	if len(d.RemovedNodes) != 0 {
		t.Errorf("expected no removed nodes, got %v", d.RemovedNodes)
	}
}

func TestCompare_RemovedNodes(t *testing.T) {
	before := simple.NewDirectedGraph()
	before.AddNode(simple.Node(0))
	before.AddNode(simple.Node(1))
	before.AddNode(simple.Node(2))

	after := simple.NewDirectedGraph()
	after.AddNode(simple.Node(0))

	d := Compare(before, after)
	if !reflect.DeepEqual(d.RemovedNodes, []int64{1, 2}) {
		t.Errorf("expected removed nodes [1 2], got %v", d.RemovedNodes)
	}
}

func TestCompare_AddedEdges(t *testing.T) {
	before := simple.NewDirectedGraph()
	before.AddNode(simple.Node(0))
	before.AddNode(simple.Node(1))
	before.SetEdge(before.NewEdge(simple.Node(0), simple.Node(1)))

	after := simple.NewDirectedGraph()
	after.AddNode(simple.Node(0))
	after.AddNode(simple.Node(1))
	after.SetEdge(after.NewEdge(simple.Node(0), simple.Node(1)))
	after.SetEdge(after.NewEdge(simple.Node(1), simple.Node(0)))

	d := Compare(before, after)
	if !reflect.DeepEqual(d.AddedEdges, [][2]int64{{1, 0}}) {
		t.Errorf("expected added edge [1,0], got %v", d.AddedEdges)
	}
	if len(d.RemovedEdges) != 0 {
		t.Errorf("expected no removed edges, got %v", d.RemovedEdges)
	}
}

func TestCompare_RemovedEdges(t *testing.T) {
	before := simple.NewDirectedGraph()
	before.AddNode(simple.Node(0))
	before.AddNode(simple.Node(1))
	before.SetEdge(before.NewEdge(simple.Node(0), simple.Node(1)))

	after := simple.NewDirectedGraph()
	after.AddNode(simple.Node(0))
	after.AddNode(simple.Node(1))

	d := Compare(before, after)
	if !reflect.DeepEqual(d.RemovedEdges, [][2]int64{{0, 1}}) {
		t.Errorf("expected removed edge [0,1], got %v", d.RemovedEdges)
	}
}

func TestCompare_NoChanges(t *testing.T) {
	before := simple.NewDirectedGraph()
	before.AddNode(simple.Node(0))
	before.AddNode(simple.Node(1))
	before.SetEdge(before.NewEdge(simple.Node(0), simple.Node(1)))

	after := simple.NewDirectedGraph()
	after.AddNode(simple.Node(0))
	after.AddNode(simple.Node(1))
	after.SetEdge(after.NewEdge(simple.Node(0), simple.Node(1)))

	d := Compare(before, after)
	if len(d.AddedNodes) != 0 || len(d.RemovedNodes) != 0 ||
		len(d.AddedEdges) != 0 || len(d.RemovedEdges) != 0 {
		t.Errorf("expected no changes, got %+v", d)
	}
}

func TestCompare_Undirected(t *testing.T) {
	before := simple.NewUndirectedGraph()
	before.SetEdge(before.NewEdge(simple.Node(0), simple.Node(1)))

	after := simple.NewUndirectedGraph()
	after.SetEdge(after.NewEdge(simple.Node(0), simple.Node(1)))
	after.SetEdge(after.NewEdge(simple.Node(1), simple.Node(2)))

	d := Compare(before, after)
	if !reflect.DeepEqual(d.AddedNodes, []int64{2}) {
		t.Errorf("expected added node [2], got %v", d.AddedNodes)
	}
	// Undirected edges show up as both directions in graph.From.
	if len(d.AddedEdges) != 2 {
		t.Errorf("expected 2 added edge directions for undirected, got %v", d.AddedEdges)
	}
}

func TestCompare_EmptyGraphs(t *testing.T) {
	before := simple.NewDirectedGraph()
	after := simple.NewDirectedGraph()

	d := Compare(before, after)
	if len(d.AddedNodes) != 0 || len(d.RemovedNodes) != 0 ||
		len(d.AddedEdges) != 0 || len(d.RemovedEdges) != 0 {
		t.Errorf("expected no changes for empty graphs, got %+v", d)
	}
}
