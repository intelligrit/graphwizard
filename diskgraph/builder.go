// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// UndirectedBuilder builds a disk-backed undirected graph.
// After adding all nodes and edges, call Close to finalize.
//
// For best performance when adding many edges, wrap calls in a Batch:
//
//	b.Batch(func(tx *diskgraph.UndirectedTx) error {
//	    tx.AddEdge(0, 1)
//	    tx.AddEdge(1, 2)
//	    return nil
//	})
type UndirectedBuilder struct {
	db        *bolt.DB
	nodeCount int64
}

// NewUndirectedBuilder creates a new bolt file and returns a builder.
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
		return addNodeTx(tx, id, &b.nodeCount)
	})
}

// AddEdge adds an unweighted edge (weight 1.0) between uid and vid.
func (b *UndirectedBuilder) AddEdge(uid, vid int64) error {
	return b.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted edge between uid and vid.
func (b *UndirectedBuilder) AddWeightedEdge(uid, vid int64, weight float64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return addUndirectedEdgeTx(tx, uid, vid, weight, &b.nodeCount)
	})
}

// UndirectedTx provides batched write operations within a single transaction.
type UndirectedTx struct {
	tx        *bolt.Tx
	nodeCount *int64
}

// AddNode adds a node within the batch transaction.
func (t *UndirectedTx) AddNode(id int64) error {
	return addNodeTx(t.tx, id, t.nodeCount)
}

// AddEdge adds an unweighted edge within the batch transaction.
func (t *UndirectedTx) AddEdge(uid, vid int64) error {
	return addUndirectedEdgeTx(t.tx, uid, vid, 1.0, t.nodeCount)
}

// AddWeightedEdge adds a weighted edge within the batch transaction.
func (t *UndirectedTx) AddWeightedEdge(uid, vid int64, weight float64) error {
	return addUndirectedEdgeTx(t.tx, uid, vid, weight, t.nodeCount)
}

// Batch executes all writes in fn within a single bolt transaction.
// This is much faster than individual AddEdge calls.
func (b *UndirectedBuilder) Batch(fn func(tx *UndirectedTx) error) error {
	return b.db.Update(func(btx *bolt.Tx) error {
		return fn(&UndirectedTx{tx: btx, nodeCount: &b.nodeCount})
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
		return addNodeTx(tx, id, &b.nodeCount)
	})
}

// AddEdge adds an unweighted directed edge from uid to vid.
func (b *DirectedBuilder) AddEdge(uid, vid int64) error {
	return b.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted directed edge from uid to vid.
func (b *DirectedBuilder) AddWeightedEdge(uid, vid int64, weight float64) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return addDirectedEdgeTx(tx, uid, vid, weight, &b.nodeCount)
	})
}

// DirectedTx provides batched write operations within a single transaction.
type DirectedTx struct {
	tx        *bolt.Tx
	nodeCount *int64
}

// AddNode adds a node within the batch transaction.
func (t *DirectedTx) AddNode(id int64) error {
	return addNodeTx(t.tx, id, t.nodeCount)
}

// AddEdge adds an unweighted directed edge within the batch transaction.
func (t *DirectedTx) AddEdge(uid, vid int64) error {
	return addDirectedEdgeTx(t.tx, uid, vid, 1.0, t.nodeCount)
}

// AddWeightedEdge adds a weighted directed edge within the batch transaction.
func (t *DirectedTx) AddWeightedEdge(uid, vid int64, weight float64) error {
	return addDirectedEdgeTx(t.tx, uid, vid, weight, t.nodeCount)
}

// Batch executes all writes in fn within a single bolt transaction.
func (b *DirectedBuilder) Batch(fn func(tx *DirectedTx) error) error {
	return b.db.Update(func(btx *bolt.Tx) error {
		return fn(&DirectedTx{tx: btx, nodeCount: &b.nodeCount})
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

// --- Internal transaction helpers ---

func addNodeTx(tx *bolt.Tx, id int64, count *int64) error {
	bk := tx.Bucket(bucketNodes)
	key := int64ToBytes(id)
	if bk.Get(key) != nil {
		return nil
	}
	*count++
	return bk.Put(key, []byte{1})
}

func addUndirectedEdgeTx(tx *bolt.Tx, uid, vid int64, weight float64, count *int64) error {
	nodes := tx.Bucket(bucketNodes)
	edges := tx.Bucket(bucketEdges)
	adj := tx.Bucket(bucketAdj)
	w := float64ToBytes(weight)

	for _, id := range []int64{uid, vid} {
		key := int64ToBytes(id)
		if nodes.Get(key) == nil {
			if err := nodes.Put(key, []byte{1}); err != nil {
				return err
			}
			*count++
		}
	}

	if err := edges.Put(edgeKey(uid, vid), w); err != nil {
		return err
	}
	if err := edges.Put(edgeKey(vid, uid), w); err != nil {
		return err
	}

	uKey := int64ToBytes(uid)
	vKey := int64ToBytes(vid)
	if err := adj.Put(uKey, appendID(adj.Get(uKey), vid)); err != nil {
		return err
	}
	return adj.Put(vKey, appendID(adj.Get(vKey), uid))
}

func addDirectedEdgeTx(tx *bolt.Tx, uid, vid int64, weight float64, count *int64) error {
	nodes := tx.Bucket(bucketNodes)
	edges := tx.Bucket(bucketEdges)
	adj := tx.Bucket(bucketAdj)
	rev := tx.Bucket(bucketRevAdj)
	w := float64ToBytes(weight)

	for _, id := range []int64{uid, vid} {
		key := int64ToBytes(id)
		if nodes.Get(key) == nil {
			if err := nodes.Put(key, []byte{1}); err != nil {
				return err
			}
			*count++
		}
	}

	if err := edges.Put(edgeKey(uid, vid), w); err != nil {
		return err
	}

	uKey := int64ToBytes(uid)
	if err := adj.Put(uKey, appendID(adj.Get(uKey), vid)); err != nil {
		return err
	}

	vKey := int64ToBytes(vid)
	return rev.Put(vKey, appendID(rev.Get(vKey), uid))
}
