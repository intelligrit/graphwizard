// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package traverse

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestBFS_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	visited := BFS(g, 0)
	if len(visited) != 4 {
		t.Fatalf("expected 4 visited nodes, got %d", len(visited))
	}
	if visited[0] != 0 {
		t.Errorf("BFS should start at source, got %d", visited[0])
	}
}

func TestBFS_Disconnected(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.AddNode(simple.Node(2))

	visited := BFS(g, 0)
	if len(visited) != 2 {
		t.Fatalf("expected 2 reachable nodes, got %d", len(visited))
	}
}

func TestDFS_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	visited := DFS(g, 0)
	if len(visited) != 3 {
		t.Fatalf("expected 3 visited nodes, got %d", len(visited))
	}
	if visited[0] != 0 {
		t.Errorf("DFS should start at source, got %d", visited[0])
	}
}

func TestDFS_SingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	visited := DFS(g, 0)
	if len(visited) != 1 {
		t.Fatalf("expected 1 visited node, got %d", len(visited))
	}
}

func TestBFSPath_Chain(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	path := BFSPath(g, 0, 3)
	if len(path) != 4 {
		t.Fatalf("expected path of 4 nodes, got %d: %v", len(path), path)
	}
	if path[0] != 0 || path[3] != 3 {
		t.Errorf("expected path [0..3], got %v", path)
	}
}

func TestBFSPath_NoPath(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	path := BFSPath(g, 0, 1)
	if path != nil {
		t.Errorf("expected nil for disconnected nodes, got %v", path)
	}
}

func TestBFSPath_SameNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))

	path := BFSPath(g, 0, 0)
	if len(path) != 1 || path[0] != 0 {
		t.Errorf("expected [0] for same source/target, got %v", path)
	}
}

func TestBFSPath_ShortestInGrid(t *testing.T) {
	// Diamond: 0-1, 0-2, 1-3, 2-3. BFS path 0->3 should be length 3.
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	path := BFSPath(g, 0, 3)
	if len(path) != 3 {
		t.Errorf("expected shortest path of length 3, got %d: %v", len(path), path)
	}
}
