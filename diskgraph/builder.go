// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// UndirectedBuilder builds a disk-backed undirected graph.
// After adding all nodes and edges, call Close to finalize.
type UndirectedBuilder struct {
	db        *bolt.DB
	nodeCount int64
}

// NewUndirectedBuilder creates a new bolt file and returns a builder.
// The file is created or truncated if it already exists.
func NewUndirectedBuilder(path string) (*UndirectedBuilder, error) {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("diskgraph: create %s: %w", path, err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{bucketNodes, bucketEdges, bucketAdj, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &UndirectedBuilder{db: db}, nil
}

// AddNode adds a node with the given ID.
func (b *UndirectedBuilder) AddNode(id int64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bk := tx.Bucket(bucketNodes)
		key := int64ToBytes(id)
		if bk.Get(key) != nil {
			return nil // already exists
		}
		b.nodeCount++
		return bk.Put(key, []byte{1})
	})
}

// AddEdge adds an unweighted edge (weight 1.0) between uid and vid.
// Both nodes are created if they don't exist.
func (b *UndirectedBuilder) AddEdge(uid, vid int64) error {
	return b.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted edge between uid and vid.
// Both nodes are created if they don't exist.
func (b *UndirectedBuilder) AddWeightedEdge(uid, vid int64, weight float64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		nodes := tx.Bucket(bucketNodes)
		edges := tx.Bucket(bucketEdges)
		adj := tx.Bucket(bucketAdj)
		w := float64ToBytes(weight)

		// Ensure both nodes exist.
		for _, id := range []int64{uid, vid} {
			key := int64ToBytes(id)
			if nodes.Get(key) == nil {
				if err := nodes.Put(key, []byte{1}); err != nil {
					return err
				}
				b.nodeCount++
			}
		}

		// Store edge in both directions for undirected.
		if err := edges.Put(edgeKey(uid, vid), w); err != nil {
			return err
		}
		if err := edges.Put(edgeKey(vid, uid), w); err != nil {
			return err
		}

		// Update adjacency lists.
		uKey := int64ToBytes(uid)
		vKey := int64ToBytes(vid)
		if err := adj.Put(uKey, appendID(adj.Get(uKey), vid)); err != nil {
			return err
		}
		return adj.Put(vKey, appendID(adj.Get(vKey), uid))
	})
}

// Close finalizes the database, writing the node count to metadata.
func (b *UndirectedBuilder) Close() error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMeta).Put(metaNodeCount, int64ToBytes(b.nodeCount))
	})
	if err != nil {
		b.db.Close()
		return err
	}
	return b.db.Close()
}

// DirectedBuilder builds a disk-backed directed graph.
type DirectedBuilder struct {
	db        *bolt.DB
	nodeCount int64
}

// NewDirectedBuilder creates a new bolt file for a directed graph.
func NewDirectedBuilder(path string) (*DirectedBuilder, error) {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("diskgraph: create %s: %w", path, err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{bucketNodes, bucketEdges, bucketAdj, bucketRevAdj, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &DirectedBuilder{db: db}, nil
}

// AddNode adds a node with the given ID.
func (b *DirectedBuilder) AddNode(id int64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bk := tx.Bucket(bucketNodes)
		key := int64ToBytes(id)
		if bk.Get(key) != nil {
			return nil
		}
		b.nodeCount++
		return bk.Put(key, []byte{1})
	})
}

// AddEdge adds an unweighted directed edge from uid to vid.
func (b *DirectedBuilder) AddEdge(uid, vid int64) error {
	return b.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted directed edge from uid to vid.
func (b *DirectedBuilder) AddWeightedEdge(uid, vid int64, weight float64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		nodes := tx.Bucket(bucketNodes)
		edges := tx.Bucket(bucketEdges)
		adj := tx.Bucket(bucketAdj)
		rev := tx.Bucket(bucketRevAdj)
		w := float64ToBytes(weight)

		// Ensure both nodes exist.
		for _, id := range []int64{uid, vid} {
			key := int64ToBytes(id)
			if nodes.Get(key) == nil {
				if err := nodes.Put(key, []byte{1}); err != nil {
					return err
				}
				b.nodeCount++
			}
		}

		// Store edge (one direction only for directed).
		if err := edges.Put(edgeKey(uid, vid), w); err != nil {
			return err
		}

		// Forward adjacency: uid -> vid.
		uKey := int64ToBytes(uid)
		if err := adj.Put(uKey, appendID(adj.Get(uKey), vid)); err != nil {
			return err
		}

		// Reverse adjacency: vid <- uid.
		vKey := int64ToBytes(vid)
		return rev.Put(vKey, appendID(rev.Get(vKey), uid))
	})
}

// Close finalizes the directed graph database.
func (b *DirectedBuilder) Close() error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMeta).Put(metaNodeCount, int64ToBytes(b.nodeCount))
	})
	if err != nil {
		b.db.Close()
		return err
	}
	return b.db.Close()
}
