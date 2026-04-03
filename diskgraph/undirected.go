// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"math"

	"gonum.org/v1/gonum/graph"

	bolt "go.etcd.io/bbolt"
)

// Undirected is a read-only, disk-backed undirected graph that implements
// graph.Undirected and graph.WeightedUndirected. It reads directly from
// a memory-mapped bolt file using a single long-lived read transaction.
type Undirected struct {
	db    *bolt.DB
	tx    *bolt.Tx
	nodes *bolt.Bucket
	edges *bolt.Bucket
	adj   *bolt.Bucket
	n     int64

	// adjCache, when non-nil, holds preloaded adjacency lists keyed by
	// node ID. This accelerates From() and HasEdgeBetween() at the cost
	// of holding the adjacency structure in memory. Enable via
	// PreloadAdjacency().
	adjCache map[int64][]int64
	// adjSet mirrors adjCache as sets for O(1) HasEdgeBetween lookups.
	adjSet map[int64]map[int64]struct{}
}

// OpenUndirected opens a bolt file previously created by UndirectedBuilder.
func OpenUndirected(path string) (*Undirected, error) {
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

	return &Undirected{
		db:    db,
		tx:    tx,
		nodes: tx.Bucket(bucketNodes),
		edges: tx.Bucket(bucketEdges),
		adj:   tx.Bucket(bucketAdj),
		n:     n,
	}, nil
}

// PreloadAdjacency loads all adjacency lists into memory. This trades memory
// for speed — HasEdgeBetween becomes an O(1) set lookup instead of a B-tree
// traversal, which dramatically improves algorithms like ClusteringCoefficient
// that call HasEdgeBetween in tight loops.
//
// For a graph with E edges, this uses roughly O(E * 16) bytes of memory.
func (g *Undirected) PreloadAdjacency() {
	g.adjCache = make(map[int64][]int64, g.n)
	g.adjSet = make(map[int64]map[int64]struct{}, g.n)

	g.adj.ForEach(func(k, v []byte) error {
		id := bytesToInt64(k)
		neighbors := decodeIDs(v)
		g.adjCache[id] = neighbors
		s := make(map[int64]struct{}, len(neighbors))
		for _, nid := range neighbors {
			s[nid] = struct{}{}
		}
		g.adjSet[id] = s
		return nil
	})
}

// Close releases the read transaction and bolt database.
func (g *Undirected) Close() error {
	g.adjCache = nil
	g.adjSet = nil
	g.tx.Rollback()
	return g.db.Close()
}

// Node returns the node with the given ID, or nil if it doesn't exist.
func (g *Undirected) Node(id int64) graph.Node {
	if g.nodes.Get(int64ToBytes(id)) == nil {
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
	ids := make([]int64, 0, g.n)
	g.nodes.ForEach(func(k, _ []byte) error {
		ids = append(ids, bytesToInt64(k))
		return nil
	})
	return ids
}

// From returns all nodes reachable from the node with the given ID.
func (g *Undirected) From(id int64) graph.Nodes {
	if g.adjCache != nil {
		neighbors := g.adjCache[id]
		if neighbors == nil {
			return emptyNodes{}
		}
		return newSliceNodes(neighbors)
	}
	v := g.adj.Get(int64ToBytes(id))
	neighbors := decodeIDs(v)
	if neighbors == nil {
		return emptyNodes{}
	}
	return newSliceNodes(neighbors)
}

// HasEdgeBetween returns whether an edge exists between xid and yid.
func (g *Undirected) HasEdgeBetween(xid, yid int64) bool {
	if g.adjSet != nil {
		s := g.adjSet[xid]
		if s == nil {
			return false
		}
		_, ok := s[yid]
		return ok
	}
	return g.edges.Get(edgeKey(xid, yid)) != nil
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
	v := g.edges.Get(edgeKey(uid, vid))
	if v == nil {
		return nil
	}
	return boltWeightedEdge{
		boltEdge: boltEdge{from: boltNode{id: uid}, to: boltNode{id: vid}},
		w:        bytesToFloat64(v),
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
