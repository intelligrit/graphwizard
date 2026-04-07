// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import "slices"

// csr is a Compressed Sparse Row representation of graph adjacency.
// It stores all neighbor lists in a single flat slice (targets) with
// an offset table indexed by dense node index. Neighbor lists are
// sorted to allow binary-search HasEdgeBetween in O(log degree).
//
// Memory cost: 8*E bytes (targets) + 4*(N+1) bytes (offsets), where
// E = number of directed edge entries and N = number of nodes.
// For 162M edges and 6M nodes this is ~1.3 GB, compared to ~20 GB+
// for the previous map-of-maps representation.
type csr struct {
	// nodeIDs is the sorted slice of original node IDs. The dense
	// index of a node is its position in this slice.
	nodeIDs []int64

	// offsets has length len(nodeIDs)+1. The neighbors of the node
	// at dense index i are targets[offsets[i]:offsets[i+1]].
	offsets []int32

	// targets stores all neighbor node IDs (original IDs, not dense
	// indices) in a single contiguous slice. Each per-node segment
	// is sorted for binary search.
	targets []int64

	// denseTargets mirrors targets but stores dense indices (int32)
	// instead of original node IDs. Built once at preload time so
	// algorithms can read dense-indexed adjacency without copying.
	// Memory cost: 4*E bytes (half of targets).
	denseTargets []int32
}

// buildCSR constructs a CSR from sorted nodeIDs and an edge stream.
// The edges function is called twice (degree counting, then filling)
// and must yield the same (src, dst) pairs each time.
// nodeIDs must be sorted in ascending order.
func buildCSR(nodeIDs []int64, edges func(yield func(src, dst int64))) *csr {
	n := len(nodeIDs)

	// First pass: count degree of each node.
	degree := make([]int32, n)
	edges(func(src, _ int64) {
		idx := searchID(nodeIDs, src)
		if idx >= 0 {
			degree[idx]++
		}
	})

	// Build offset table from degree counts.
	offsets := make([]int32, n+1)
	for i := range n {
		offsets[i+1] = offsets[i] + degree[i]
	}
	totalEdges := int(offsets[n])

	// Second pass: fill targets using write cursors.
	targets := make([]int64, totalEdges)
	cursor := make([]int32, n) // current write position per node
	copy(cursor, offsets[:n])
	edges(func(src, dst int64) {
		idx := searchID(nodeIDs, src)
		if idx >= 0 {
			targets[cursor[idx]] = dst
			cursor[idx]++
		}
	})

	// Sort each neighbor list for binary search.
	for i := range n {
		seg := targets[offsets[i]:offsets[i+1]]
		slices.Sort(seg)
	}

	// Build dense-index version of targets for DenseAdjacency consumers.
	denseTargets := make([]int32, totalEdges)
	for i, tid := range targets {
		denseTargets[i] = int32(searchID(nodeIDs, tid))
	}

	return &csr{
		nodeIDs:      nodeIDs,
		offsets:      offsets,
		targets:      targets,
		denseTargets: denseTargets,
	}
}

// neighbors returns the sorted neighbor IDs for the given original node ID.
// Returns nil if the node is not found.
func (c *csr) neighbors(id int64) []int64 {
	idx := searchID(c.nodeIDs, id)
	if idx < 0 {
		return nil
	}
	lo := c.offsets[idx]
	hi := c.offsets[idx+1]
	if lo == hi {
		return nil
	}
	return c.targets[lo:hi]
}

// denseNeighbors returns the dense indices of neighbors for the node at dense
// index i. Returns nil if i is out of range or the node has no neighbors.
func (c *csr) denseNeighbors(i int) []int32 {
	if i < 0 || i >= len(c.nodeIDs) {
		return nil
	}
	lo := c.offsets[i]
	hi := c.offsets[i+1]
	if lo == hi {
		return nil
	}
	return c.denseTargets[lo:hi]
}

// hasEdge returns true if there is an edge from xid to yid.
// Uses binary search on the sorted neighbor list of xid.
func (c *csr) hasEdge(xid, yid int64) bool {
	idx := searchID(c.nodeIDs, xid)
	if idx < 0 {
		return false
	}
	lo := int(c.offsets[idx])
	hi := int(c.offsets[idx+1])
	if lo == hi {
		return false
	}
	seg := c.targets[lo:hi]
	_, found := slices.BinarySearch(seg, yid)
	return found
}

// searchID performs binary search on a sorted slice of int64s.
// Returns the index if found, or -1.
func searchID(sorted []int64, id int64) int {
	lo, hi := 0, len(sorted)
	for lo < hi {
		mid := lo + (hi-lo)/2
		v := sorted[mid]
		if v == id {
			return mid
		}
		if v < id {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return -1
}

// nodeExists checks whether id exists in a sorted slice of int64s.
func nodeExists(sorted []int64, id int64) bool {
	return searchID(sorted, id) >= 0
}
