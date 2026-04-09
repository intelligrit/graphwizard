// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package densegraph

import (
	"math"
	"slices"

	"github.com/intelligrit/graphwizard"
	"gonum.org/v1/gonum/graph"
)

// Undirected is a read-only, memory-efficient undirected graph stored in
// Compressed Sparse Row (CSR) format with inline weights. It provides the
// same interfaces as diskgraph.Undirected but with no SQLite dependency —
// all data lives in flat Go slices.
//
// Memory cost: 20*E + 12*(N+1) + 8*N bytes, where E is the number of
// directed edge entries (2× undirected edges) and N is the number of nodes.
// For 81M undirected edges and 3M nodes this is roughly 3.3 GB, compared
// to ~200 GB for gonum simple.UndirectedGraph.
//
// Construct with NewUndirectedBuilder followed by Build.
type Undirected struct {
	// nodeIDs is the sorted slice of original node IDs. The dense
	// index of a node is its position in this slice.
	nodeIDs []int64

	// offsets has length len(nodeIDs)+1. The neighbors of the node
	// at dense index i are targets[offsets[i]:offsets[i+1]].
	offsets []int32

	// targets stores all neighbor node IDs (original IDs) in a single
	// contiguous slice. Each per-node segment is sorted for binary search.
	targets []int64

	// weights stores edge weights parallel to targets.
	weights []float64

	// denseTargets mirrors targets but stores dense indices (int32).
	denseTargets []int32
}

// Node returns the node with the given ID, or nil if it doesn't exist.
func (g *Undirected) Node(id int64) graph.Node {
	if searchID(g.nodeIDs, id) < 0 {
		return nil
	}
	return denseNode{id: id}
}

// Nodes returns an iterator over all nodes.
func (g *Undirected) Nodes() graph.Nodes {
	return newSliceNodes(g.nodeIDs)
}

// From returns all nodes reachable from the node with the given ID.
func (g *Undirected) From(id int64) graph.Nodes {
	idx := searchID(g.nodeIDs, id)
	if idx < 0 {
		return emptyNodes{}
	}
	lo := g.offsets[idx]
	hi := g.offsets[idx+1]
	if lo == hi {
		return emptyNodes{}
	}
	return newSliceNodes(g.targets[lo:hi])
}

// HasEdgeBetween returns whether an edge exists between xid and yid.
func (g *Undirected) HasEdgeBetween(xid, yid int64) bool {
	idx := searchID(g.nodeIDs, xid)
	if idx < 0 {
		return false
	}
	lo := int(g.offsets[idx])
	hi := int(g.offsets[idx+1])
	if lo == hi {
		return false
	}
	_, found := slices.BinarySearch(g.targets[lo:hi], yid)
	return found
}

// Edge returns the edge between xid and yid, or nil.
func (g *Undirected) Edge(xid, yid int64) graph.Edge {
	if !g.HasEdgeBetween(xid, yid) {
		return nil
	}
	return denseEdge{from: denseNode{id: xid}, to: denseNode{id: yid}}
}

// EdgeBetween returns the edge between xid and yid, or nil.
func (g *Undirected) EdgeBetween(xid, yid int64) graph.Edge {
	return g.Edge(xid, yid)
}

// WeightedEdge returns the weighted edge from uid to vid, or nil.
func (g *Undirected) WeightedEdge(uid, vid int64) graph.WeightedEdge {
	idx := searchID(g.nodeIDs, uid)
	if idx < 0 {
		return nil
	}
	lo := int(g.offsets[idx])
	hi := int(g.offsets[idx+1])
	if lo == hi {
		return nil
	}
	pos, found := slices.BinarySearch(g.targets[lo:hi], vid)
	if !found {
		return nil
	}
	return denseWeightedEdge{
		denseEdge: denseEdge{from: denseNode{id: uid}, to: denseNode{id: vid}},
		w:         g.weights[lo+pos],
	}
}

// WeightedEdgeBetween returns the weighted edge between xid and yid, or nil.
func (g *Undirected) WeightedEdgeBetween(xid, yid int64) graph.WeightedEdge {
	return g.WeightedEdge(xid, yid)
}

// Weight returns the weight of the edge between xid and yid.
func (g *Undirected) Weight(xid, yid int64) (w float64, ok bool) {
	if xid == yid {
		return 0, true
	}
	e := g.WeightedEdge(xid, yid)
	if e != nil {
		return e.Weight(), true
	}
	return math.Inf(1), false
}

// --- DenseAdjacency interface ---

// NodeIDs returns the original node IDs in dense-index order.
func (g *Undirected) NodeIDs() []int64 { return g.nodeIDs }

// DenseNeighbors returns the dense indices of all neighbors of the node
// at dense index i.
func (g *Undirected) DenseNeighbors(i int) []int32 {
	if i < 0 || i >= len(g.nodeIDs) {
		return nil
	}
	lo := g.offsets[i]
	hi := g.offsets[i+1]
	if lo == hi {
		return nil
	}
	return g.denseTargets[lo:hi]
}

// NumNodes returns the number of nodes.
func (g *Undirected) NumNodes() int { return len(g.nodeIDs) }

// --- EdgeScanner interface ---

// ScanWeightedEdges calls yield for every directed edge entry.
func (g *Undirected) ScanWeightedEdges(yield func(src, dst int64, weight float64)) {
	for i, id := range g.nodeIDs {
		lo := g.offsets[i]
		hi := g.offsets[i+1]
		for j := lo; j < hi; j++ {
			yield(id, g.targets[j], g.weights[j])
		}
	}
}

// PreloadAdjacency is a no-op (adjacency is always in memory).
func (g *Undirected) PreloadAdjacency() {}

// --- Helpers ---

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

// --- Node / Edge types ---

type denseNode struct{ id int64 }

func (n denseNode) ID() int64 { return n.id }

type denseEdge struct{ from, to denseNode }

func (e denseEdge) From() graph.Node         { return e.from }
func (e denseEdge) To() graph.Node           { return e.to }
func (e denseEdge) ReversedEdge() graph.Edge { return denseEdge{from: e.to, to: e.from} }

type denseWeightedEdge struct {
	denseEdge
	w float64
}

func (e denseWeightedEdge) Weight() float64 { return e.w }
func (e denseWeightedEdge) ReversedEdge() graph.Edge {
	return denseWeightedEdge{denseEdge: denseEdge{from: e.to, to: e.from}, w: e.w}
}

// --- Iterator types ---

type sliceNodes struct {
	ids []int64
	pos int
}

func newSliceNodes(ids []int64) *sliceNodes { return &sliceNodes{ids: ids, pos: -1} }

func (it *sliceNodes) Next() bool {
	it.pos++
	return it.pos < len(it.ids)
}

func (it *sliceNodes) Len() int {
	remaining := len(it.ids) - it.pos - 1
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (it *sliceNodes) Reset()          { it.pos = -1 }
func (it *sliceNodes) Node() graph.Node {
	if it.pos < 0 || it.pos >= len(it.ids) {
		return nil
	}
	return denseNode{id: it.ids[it.pos]}
}

type emptyNodes struct{}

func (emptyNodes) Next() bool       { return false }
func (emptyNodes) Len() int         { return 0 }
func (emptyNodes) Reset()           {}
func (emptyNodes) Node() graph.Node { return nil }

// Compile-time interface checks.
var (
	_ graph.Undirected         = (*Undirected)(nil)
	_ graph.WeightedUndirected = (*Undirected)(nil)
	_ graphwizard.DenseAdjacency = (*Undirected)(nil)
	_ graphwizard.EdgeScanner    = (*Undirected)(nil)
)
