// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"gonum.org/v1/gonum/graph"
)

// boltNode implements graph.Node backed by an ID.
type boltNode struct {
	id int64
}

func (n boltNode) ID() int64 { return n.id }

// sliceNodes is a graph.Nodes iterator over a slice of node IDs.
type sliceNodes struct {
	ids []int64
	pos int
}

func newSliceNodes(ids []int64) *sliceNodes {
	return &sliceNodes{ids: ids, pos: -1}
}

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
	return boltNode{id: it.ids[it.pos]}
}

// emptyNodes is a graph.Nodes iterator that is always exhausted.
type emptyNodes struct{}

func (emptyNodes) Next() bool      { return false }
func (emptyNodes) Len() int        { return 0 }
func (emptyNodes) Reset()          {}
func (emptyNodes) Node() graph.Node { return nil }

// boltEdge implements graph.Edge.
type boltEdge struct {
	from, to boltNode
}

func (e boltEdge) From() graph.Node         { return e.from }
func (e boltEdge) To() graph.Node           { return e.to }
func (e boltEdge) ReversedEdge() graph.Edge { return boltEdge{from: e.to, to: e.from} }

// boltWeightedEdge implements graph.WeightedEdge.
type boltWeightedEdge struct {
	boltEdge
	w float64
}

func (e boltWeightedEdge) Weight() float64            { return e.w }
func (e boltWeightedEdge) ReversedEdge() graph.Edge {
	return boltWeightedEdge{boltEdge: boltEdge{from: e.to, to: e.from}, w: e.w}
}
