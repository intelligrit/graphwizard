// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package graphwizard

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
