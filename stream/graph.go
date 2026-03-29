// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package stream

import (
	"sync"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// ChangeKind describes the type of graph mutation.
type ChangeKind int

const (
	// AddNodeChange indicates a node was added.
	AddNodeChange ChangeKind = iota
	// RemoveNodeChange indicates a node was removed.
	RemoveNodeChange
	// AddEdgeChange indicates an edge was added.
	AddEdgeChange
	// RemoveEdgeChange indicates an edge was removed.
	RemoveEdgeChange
)

// Change records a single graph mutation.
type Change struct {
	Kind   ChangeKind
	From   int64
	To     int64
	Weight float64
}

// StreamGraph is a thread-safe mutable graph that records all changes.
type StreamGraph struct {
	mu      sync.RWMutex
	g       *simple.WeightedUndirectedGraph
	changes []Change
}

// New creates a new empty StreamGraph.
func New() *StreamGraph {
	return &StreamGraph{
		g: simple.NewWeightedUndirectedGraph(0, 0),
	}
}

// AddNode adds a node to the graph. If the node already exists, this is a
// no-op and no change is recorded.
func (sg *StreamGraph) AddNode(id int64) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	if sg.g.Node(id) != nil {
		return
	}
	sg.g.AddNode(simple.Node(id))
	sg.changes = append(sg.changes, Change{Kind: AddNodeChange, From: id})
}

// RemoveNode removes a node and all its incident edges from the graph. If the
// node does not exist, this is a no-op.
func (sg *StreamGraph) RemoveNode(id int64) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	if sg.g.Node(id) == nil {
		return
	}
	sg.g.RemoveNode(id)
	sg.changes = append(sg.changes, Change{Kind: RemoveNodeChange, From: id})
}

// AddEdge adds a weighted edge between two nodes. If either node does not
// exist, it is created automatically (and recorded as a change). If the edge
// already exists, its weight is updated.
func (sg *StreamGraph) AddEdge(from, to int64, weight float64) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	if sg.g.Node(from) == nil {
		sg.g.AddNode(simple.Node(from))
		sg.changes = append(sg.changes, Change{Kind: AddNodeChange, From: from})
	}
	if sg.g.Node(to) == nil {
		sg.g.AddNode(simple.Node(to))
		sg.changes = append(sg.changes, Change{Kind: AddNodeChange, From: to})
	}
	sg.g.SetWeightedEdge(sg.g.NewWeightedEdge(simple.Node(from), simple.Node(to), weight))
	sg.changes = append(sg.changes, Change{Kind: AddEdgeChange, From: from, To: to, Weight: weight})
}

// RemoveEdge removes the edge between two nodes. If the edge does not exist,
// this is a no-op.
func (sg *StreamGraph) RemoveEdge(from, to int64) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	if !sg.g.HasEdgeBetween(from, to) {
		return
	}
	sg.g.RemoveEdge(from, to)
	sg.changes = append(sg.changes, Change{Kind: RemoveEdgeChange, From: from, To: to})
}

// Graph returns the underlying weighted undirected graph for use with
// graphwizard algorithms. The returned graph should only be read while
// the StreamGraph is not being mutated, or within the read lock.
func (sg *StreamGraph) Graph() graph.WeightedUndirected {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.g
}

// Changes returns a copy of all mutations since the last call to Flush.
func (sg *StreamGraph) Changes() []Change {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	result := make([]Change, len(sg.changes))
	copy(result, sg.changes)
	return result
}

// Flush clears the change log.
func (sg *StreamGraph) Flush() {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.changes = sg.changes[:0]
}
