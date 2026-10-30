// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package densegraph

import (
	"iter"
	"math"

	"github.com/intelligrit/graphwizard"
	"github.com/intelligrit/graphwizard/internal/compact"
	"gonum.org/v1/gonum/graph"
)

// Undirected is a read-only, memory-efficient undirected graph stored in
// Compressed Sparse Row (CSR) format with inline weights. It provides the
// same interfaces as diskgraph.Undirected but with no SQLite dependency —
// all data lives in flat Go slices.
//
// Flat adjacency uses 4 bytes per directed entry, plus machine-sized CSR
// offsets and 8 bytes per node ID. Weights use a constant, uint8, uint16,
// uint32, or float64 representation, selected automatically without loss.
// Build(CompressNeighbors) additionally enables lossless block delta encoding.
//
// Construct with NewUndirectedBuilder followed by Build.
type Undirected struct {
	// nodeIDs is the sorted slice of original node IDs. The dense
	// index of a node is its position in this slice.
	nodeIDs []int64

	adj     *compact.Neighbors
	weights compact.Weights
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
	return compact.NewNodes(g.nodeIDs, g.adj.Cursor(idx))
}

// HasEdgeBetween returns whether an edge exists between xid and yid.
func (g *Undirected) HasEdgeBetween(xid, yid int64) bool {
	return g.edgePosition(xid, yid) >= 0
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
	pos := g.edgePosition(uid, vid)
	if pos < 0 {
		return nil
	}
	return denseWeightedEdge{
		denseEdge: denseEdge{from: denseNode{id: uid}, to: denseNode{id: vid}},
		w:         g.weights.At(pos),
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
	if pos := g.edgePosition(xid, yid); pos >= 0 {
		return g.weights.At(pos), true
	}
	return math.Inf(1), false
}

// --- DenseAdjacency interface ---

// NodeIDs returns the original node IDs in dense-index order.
func (g *Undirected) NodeIDs() []int64 { return g.nodeIDs }

// DenseNeighbors returns the dense indices of all neighbors of the node
// at dense index i. In compressed mode the row is decoded into a fresh slice.
// Use Neighbors for allocation-free traversal.
func (g *Undirected) DenseNeighbors(i int) []int32 {
	if g.adj == nil {
		return nil
	}
	return g.adj.Row(i)
}

// NumNodes returns the number of nodes.
func (g *Undirected) NumNodes() int { return len(g.nodeIDs) }

// --- EdgeScanner interface ---

// ScanWeightedEdges calls yield for every directed edge entry.
func (g *Undirected) ScanWeightedEdges(yield func(src, dst int64, weight float64)) {
	for i, id := range g.nodeIDs {
		c := g.adj.Cursor(i)
		for c.Next() {
			yield(id, g.nodeIDs[c.Value()], g.weights.At(c.Position()))
		}
	}
}

// PreloadAdjacency is a no-op (adjacency is always in memory).
func (g *Undirected) PreloadAdjacency() {}

// Neighbors traverses dense neighbor indexes without allocating a decoded row.
func (g *Undirected) Neighbors(i int) iter.Seq[int32] {
	return func(yield func(int32) bool) {
		if g.adj == nil {
			return
		}
		c := g.adj.Cursor(i)
		for c.Next() {
			if !yield(c.Value()) {
				return
			}
		}
	}
}

// NewNeighborCursor creates a reusable decoder for allocation-free traversal
// through interfaces. Each concurrent reader must create its own cursor.
func (g *Undirected) NewNeighborCursor() graphwizard.NeighborCursor {
	if g.adj == nil {
		return &compact.Cursor{}
	}
	c := g.adj.Cursor(-1)
	return &c
}

type weightedCursor struct {
	g   *Undirected
	c   compact.Cursor
	row int
}

// NewWeightedCursor creates a reusable decoder over shared weighted adjacency.
func (g *Undirected) NewWeightedCursor() graphwizard.WeightedNeighborCursor {
	c := &weightedCursor{g: g}
	if g.adj != nil {
		c.c = g.adj.Cursor(-1)
	}
	return c
}

func (c *weightedCursor) Reset(i int)     { c.row = i; c.c.Reset(i) }
func (c *weightedCursor) Next() bool      { return c.c.Next() }
func (c *weightedCursor) Neighbor() int32 { return c.c.Value() }
func (c *weightedCursor) Weight() float64 {
	if int(c.c.Value()) == c.row {
		return 0
	}
	return c.g.weights.At(c.c.Position())
}

// WeightedNeighbors traverses sorted neighbors and exact weights without copying
// adjacency. Self-edge weights are zero, consistent with Weight.
func (g *Undirected) WeightedNeighbors(i int) iter.Seq2[int32, float64] {
	return func(yield func(int32, float64) bool) {
		if g.adj == nil {
			return
		}
		c := g.adj.Cursor(i)
		for c.Next() {
			w := g.weights.At(c.Position())
			if int(c.Value()) == i {
				w = 0
			}
			if !yield(c.Value(), w) {
				return
			}
		}
	}
}

// StorageBytes reports retained slice payloads, including IDs, offsets,
// compressed neighbors and weights. It excludes headers and builder memory.
func (g *Undirected) StorageBytes() int64 {
	size := int64(cap(g.nodeIDs))*8 + g.weights.StorageBytes()
	if g.adj != nil {
		size += g.adj.StorageBytes()
	}
	return size
}

func (g *Undirected) edgePosition(uid, vid int64) int {
	i, j := searchID(g.nodeIDs, uid), searchID(g.nodeIDs, vid)
	if i < 0 || j < 0 || g.adj == nil {
		return -1
	}
	return g.adj.Find(i, int32(j))
}

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

func (it *sliceNodes) Reset() { it.pos = -1 }
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
	_ graph.Undirected                   = (*Undirected)(nil)
	_ graph.WeightedUndirected           = (*Undirected)(nil)
	_ graphwizard.DenseAdjacency         = (*Undirected)(nil)
	_ graphwizard.EdgeScanner            = (*Undirected)(nil)
	_ graphwizard.DenseWeightedAdjacency = (*Undirected)(nil)
)
