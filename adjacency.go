// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package graphwizard

import "iter"

// DenseAdjacency is an optional interface that graph implementations may
// provide to give algorithms direct access to a precomputed, dense-indexed
// adjacency structure. When a graph.Undirected or graph.Directed also
// implements DenseAdjacency, algorithms can skip the expensive pattern of
// calling From() for every node and building their own adjacency slices.
//
// The dense index assigns each node a contiguous integer in [0, N), where
// N is the number of nodes. NodeIDs returns the original IDs in dense-index
// order, so NodeIDs()[i] is the original ID for dense index i.
//
// Implementations should build the dense adjacency lazily or at preload
// time and cache it for reuse across multiple algorithm calls.
type DenseAdjacency interface {
	// NodeIDs returns the original node IDs in dense-index order.
	// The returned slice must not be modified by the caller.
	NodeIDs() []int64

	// DenseNeighbors returns the dense indices of all neighbors of the
	// node at dense index i. The returned slice must not be modified by
	// the caller.
	DenseNeighbors(i int) []int32

	// NumNodes returns the number of nodes (the length of NodeIDs).
	NumNodes() int
}

// DenseNeighborIterator provides allocation-free traversal of dense-indexed
// adjacency, including compressed storage. NodeIDs and NumNodes have the same
// meaning as in DenseAdjacency. A nil NodeIDs slice means adjacency is unavailable.
// Neighbor sequences must support repeated and concurrent read-only traversal.
type DenseNeighborIterator interface {
	NodeIDs() []int64
	NumNodes() int
	Neighbors(i int) iter.Seq[int32]
	NewNeighborCursor() NeighborCursor
}

// NeighborCursor is an independent reusable decoder. Reset positions it before
// the first neighbor of a dense node index. Neighbor is valid after Next returns
// true. A cursor belongs to one reader; separate cursors may read concurrently.
type NeighborCursor interface {
	Reset(i int)
	Next() bool
	Neighbor() int32
}

// WeightedNeighborCursor additionally exposes the weight of the current edge,
// consistent with graph.Weight. Weight is valid after Next returns true.
type WeightedNeighborCursor interface {
	NeighborCursor
	Weight() float64
}

// DenseWeightedAdjacency lets algorithms share a graph's weighted adjacency
// without materializing an edge-sized copy. WeightedNeighbors yields neighbors
// in increasing dense-index order and weights consistent with graph.Weight
// (including a zero weight for self edges). Sequences must support repeated and
// concurrent read-only traversal. The graph must remain unchanged during use.
type DenseWeightedAdjacency interface {
	DenseNeighborIterator
	WeightedNeighbors(i int) iter.Seq2[int32, float64]
	NewWeightedCursor() WeightedNeighborCursor
}

// EdgeScanner is an optional interface for graphs that can stream all
// edges efficiently (e.g. a single sequential table scan). This avoids
// the N+E individual queries that the From()+Weight() pattern requires.
//
// Each directed entry (src, dst, weight) is yielded exactly once.
// For undirected graphs stored with both directions, both (u,v) and
// (v,u) are yielded.
type EdgeScanner interface {
	// ScanWeightedEdges calls yield for every directed edge entry.
	ScanWeightedEdges(yield func(src, dst int64, weight float64))
}
