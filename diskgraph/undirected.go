// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"math"

	"gonum.org/v1/gonum/graph"

	bolt "go.etcd.io/bbolt"
)

// Undirected is a read-only, disk-backed undirected graph that implements
// graph.Undirected and graph.WeightedUndirected. It reads directly from
// a memory-mapped bolt file.
type Undirected struct {
	db *bolt.DB
	n  int64 // cached node count
}

// OpenUndirected opens a bolt file previously created by UndirectedBuilder.
func OpenUndirected(path string) (*Undirected, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}
	return &Undirected{db: db, n: nodeCount(db)}, nil
}

// Close releases the bolt database.
func (g *Undirected) Close() error { return g.db.Close() }

// Node returns the node with the given ID, or nil if it doesn't exist.
func (g *Undirected) Node(id int64) graph.Node {
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
func (g *Undirected) Nodes() graph.Nodes {
	return newAllNodes(g)
}

// allNodeIDs returns all node IDs by scanning the nodes bucket.
func (g *Undirected) allNodeIDs() []int64 {
	var ids []int64
	_ = g.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNodes)
		return b.ForEach(func(k, _ []byte) error {
			ids = append(ids, bytesToInt64(k))
			return nil
		})
	})
	return ids
}

// From returns all nodes reachable from the node with the given ID.
func (g *Undirected) From(id int64) graph.Nodes {
	var neighbors []int64
	_ = g.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAdj)
		v := b.Get(int64ToBytes(id))
		neighbors = decodeIDs(v)
		return nil
	})
	if neighbors == nil {
		return emptyNodes{}
	}
	return newSliceNodes(neighbors)
}

// HasEdgeBetween returns whether an edge exists between xid and yid.
func (g *Undirected) HasEdgeBetween(xid, yid int64) bool {
	var exists bool
	_ = g.db.View(func(tx *bolt.Tx) error {
		exists = tx.Bucket(bucketEdges).Get(edgeKey(xid, yid)) != nil
		return nil
	})
	return exists
}

// Edge returns the edge between xid and yid, or nil.
func (g *Undirected) Edge(xid, yid int64) graph.Edge {
	if !g.HasEdgeBetween(xid, yid) {
		return nil
	}
	return boltEdge{from: boltNode{id: xid}, to: boltNode{id: yid}}
}

// EdgeBetween returns the edge between xid and yid, or nil.
func (g *Undirected) EdgeBetween(xid, yid int64) graph.Edge {
	return g.Edge(xid, yid)
}

// WeightedEdge returns the weighted edge from uid to vid, or nil.
func (g *Undirected) WeightedEdge(uid, vid int64) graph.WeightedEdge {
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

// WeightedEdgeBetween returns the weighted edge between xid and yid, or nil.
func (g *Undirected) WeightedEdgeBetween(xid, yid int64) graph.WeightedEdge {
	return g.WeightedEdge(xid, yid)
}

// Weight returns the weight of the edge between xid and yid.
// If the edge exists, it returns the weight and true.
// If xid == yid, it returns 0 and true.
// Otherwise it returns infinity and false.
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

// Compile-time interface checks.
var (
	_ graph.Undirected         = (*Undirected)(nil)
	_ graph.WeightedUndirected = (*Undirected)(nil)
)
