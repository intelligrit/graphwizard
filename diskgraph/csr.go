// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"math"
	"slices"

	"github.com/intelligrit/graphwizard/internal/compact"
)

// csr stores only dense neighbor indices; original IDs are mapped lazily.
// Flat storage costs 4 bytes per edge entry plus machine-sized row offsets.
// Optional 24-bit packing or block delta encoding trades decoding time for memory.
type csr struct {
	nodeIDs []int64
	adj     *compact.Neighbors
}

// buildCSR takes two passes over a repeatable edge stream. All endpoints must
// occur in sorted nodeIDs. Edge counts use int; node indices use int32.
func buildCSR(nodeIDs []int64, edges func(yield func(src, dst int64)), encoding compact.NeighborEncoding) *csr {
	n := len(nodeIDs)
	if n > math.MaxInt32 {
		panic("diskgraph: more than MaxInt32 nodes")
	}
	offsets := make([]int, n+1)
	edges(func(src, _ int64) {
		if i := searchID(nodeIDs, src); i >= 0 {
			offsets[i+1]++
		}
	})
	for i := range n {
		offsets[i+1] += offsets[i]
	}
	targets := make([]int32, offsets[n])
	cursor := slices.Clone(offsets[:n])
	edges(func(src, dst int64) {
		i, j := searchID(nodeIDs, src), searchID(nodeIDs, dst)
		if i >= 0 {
			if j < 0 {
				panic("diskgraph: edge target absent from node table")
			}
			targets[cursor[i]] = int32(j)
			cursor[i]++
		}
	})
	for i := range n {
		slices.Sort(targets[offsets[i]:offsets[i+1]])
	}
	return &csr{nodeIDs: nodeIDs, adj: compact.NewNeighbors(offsets, targets, encoding)}
}

func (c *csr) denseNeighbors(i int) []int32 { return c.adj.Row(i) }

func (c *csr) hasEdge(xid, yid int64) bool {
	i, j := searchID(c.nodeIDs, xid), searchID(c.nodeIDs, yid)
	return i >= 0 && j >= 0 && c.adj.Find(i, int32(j)) >= 0
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
