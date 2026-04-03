// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"math"

	"gonum.org/v1/gonum/graph"

	bolt "go.etcd.io/bbolt"
)

// Directed is a read-only, disk-backed directed graph that implements
// graph.Directed and graph.WeightedDirected. It reads directly from
// a memory-mapped bolt file using a single long-lived read transaction.
type Directed struct {
	db     *bolt.DB
	tx     *bolt.Tx
	nodes  *bolt.Bucket
	edges  *bolt.Bucket
	adj    *bolt.Bucket
	revAdj *bolt.Bucket
	n      int64
}

// OpenDirected opens a bolt file previously created by DirectedBuilder.
func OpenDirected(path string) (*Directed, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}
	n := nodeCount(db)

	tx, err := db.Begin(false)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Directed{
		db:     db,
		tx:     tx,
		nodes:  tx.Bucket(bucketNodes),
		edges:  tx.Bucket(bucketEdges),
		adj:    tx.Bucket(bucketAdj),
		revAdj: tx.Bucket(bucketRevAdj),
		n:      n,
	}, nil
}

// Close releases the read transaction and bolt database.
func (g *Directed) Close() error {
	g.tx.Rollback()
	return g.db.Close()
}

// Node returns the node with the given ID, or nil if it doesn't exist.
func (g *Directed) Node(id int64) graph.Node {
	if g.nodes.Get(int64ToBytes(id)) == nil {
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
	ids := make([]int64, 0, g.n)
	g.nodes.ForEach(func(k, _ []byte) error {
		ids = append(ids, bytesToInt64(k))
		return nil
	})
	return ids
}

// From returns all nodes reachable from the node with the given ID
// (outgoing edges).
func (g *Directed) From(id int64) graph.Nodes {
	v := g.adj.Get(int64ToBytes(id))
	neighbors := decodeIDs(v)
	if neighbors == nil {
		return emptyNodes{}
	}
	return newSliceNodes(neighbors)
}

// To returns all nodes that have an edge to the node with the given ID
// (incoming edges).
func (g *Directed) To(id int64) graph.Nodes {
	v := g.revAdj.Get(int64ToBytes(id))
	sources := decodeIDs(v)
	if sources == nil {
		return emptyNodes{}
	}
	return newSliceNodes(sources)
}

// HasEdgeBetween returns whether an edge exists between xid and yid
// in either direction.
func (g *Directed) HasEdgeBetween(xid, yid int64) bool {
	return g.edges.Get(edgeKey(xid, yid)) != nil || g.edges.Get(edgeKey(yid, xid)) != nil
}

// HasEdgeFromTo returns whether an edge exists from uid to vid.
func (g *Directed) HasEdgeFromTo(uid, vid int64) bool {
	return g.edges.Get(edgeKey(uid, vid)) != nil
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
	v := g.edges.Get(edgeKey(uid, vid))
	if v == nil {
		return nil
	}
	return boltWeightedEdge{
		boltEdge: boltEdge{from: boltNode{id: uid}, to: boltNode{id: vid}},
		w:        bytesToFloat64(v),
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
