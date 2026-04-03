// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"math"

	"gonum.org/v1/gonum/graph"

	bolt "go.etcd.io/bbolt"
)

// Directed is a read-only, disk-backed directed graph that implements
// graph.Directed and graph.WeightedDirected. It reads directly from
// a memory-mapped bolt file.
type Directed struct {
	db *bolt.DB
	n  int64
}

// OpenDirected opens a bolt file previously created by DirectedBuilder.
func OpenDirected(path string) (*Directed, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}
	return &Directed{db: db, n: nodeCount(db)}, nil
}

// Close releases the bolt database.
func (g *Directed) Close() error { return g.db.Close() }

// Node returns the node with the given ID, or nil if it doesn't exist.
func (g *Directed) Node(id int64) graph.Node {
	var exists bool
	_ = g.db.View(func(tx *bolt.Tx) error {
		exists = tx.Bucket(bucketNodes).Get(int64ToBytes(id)) != nil
		return nil
	})
	if !exists {
		return nil
	}
	return boltNode{id: id}
}

// Nodes returns an iterator over all nodes.
func (g *Directed) Nodes() graph.Nodes {
	return newAllNodes(g)
}

// allNodeIDs returns all node IDs from the nodes bucket.
func (g *Directed) allNodeIDs() []int64 {
	var ids []int64
	_ = g.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketNodes).ForEach(func(k, _ []byte) error {
			ids = append(ids, bytesToInt64(k))
			return nil
		})
	})
	return ids
}

// From returns all nodes reachable from the node with the given ID
// (outgoing edges).
func (g *Directed) From(id int64) graph.Nodes {
	var neighbors []int64
	_ = g.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketAdj).Get(int64ToBytes(id))
		neighbors = decodeIDs(v)
		return nil
	})
	if neighbors == nil {
		return emptyNodes{}
	}
	return newSliceNodes(neighbors)
}

// To returns all nodes that have an edge to the node with the given ID
// (incoming edges).
func (g *Directed) To(id int64) graph.Nodes {
	var sources []int64
	_ = g.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketRevAdj).Get(int64ToBytes(id))
		sources = decodeIDs(v)
		return nil
	})
	if sources == nil {
		return emptyNodes{}
	}
	return newSliceNodes(sources)
}

// HasEdgeBetween returns whether an edge exists between xid and yid
// in either direction.
func (g *Directed) HasEdgeBetween(xid, yid int64) bool {
	var exists bool
	_ = g.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEdges)
		exists = b.Get(edgeKey(xid, yid)) != nil || b.Get(edgeKey(yid, xid)) != nil
		return nil
	})
	return exists
}

// HasEdgeFromTo returns whether an edge exists from uid to vid.
func (g *Directed) HasEdgeFromTo(uid, vid int64) bool {
	var exists bool
	_ = g.db.View(func(tx *bolt.Tx) error {
		exists = tx.Bucket(bucketEdges).Get(edgeKey(uid, vid)) != nil
		return nil
	})
	return exists
}

// Edge returns the edge from xid to yid, or nil.
func (g *Directed) Edge(xid, yid int64) graph.Edge {
	if !g.HasEdgeFromTo(xid, yid) {
		return nil
	}
	return boltEdge{from: boltNode{id: xid}, to: boltNode{id: yid}}
}

// WeightedEdge returns the weighted edge from uid to vid, or nil.
func (g *Directed) WeightedEdge(uid, vid int64) graph.WeightedEdge {
	var w float64
	var found bool
	_ = g.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketEdges).Get(edgeKey(uid, vid))
		if v != nil {
			w = bytesToFloat64(v)
			found = true
		}
		return nil
	})
	if !found {
		return nil
	}
	return boltWeightedEdge{
		boltEdge: boltEdge{from: boltNode{id: uid}, to: boltNode{id: vid}},
		w:        w,
	}
}

// Weight returns the weight of the edge from xid to yid.
func (g *Directed) Weight(xid, yid int64) (w float64, ok bool) {
	if xid == yid {
		return 0, true
	}
	e := g.WeightedEdge(xid, yid)
	if e != nil {
		return e.Weight(), true
	}
	return math.Inf(1), false
}

// Compile-time interface checks.
var (
	_ graph.Directed         = (*Directed)(nil)
	_ graph.WeightedDirected = (*Directed)(nil)
)
