// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"encoding/binary"
	"fmt"
	"math"

	bolt "go.etcd.io/bbolt"
)

// Bucket names used in the bolt database.
var (
	bucketNodes   = []byte("nodes")
	bucketEdges   = []byte("edges")   // key: src|dst, value: weight (8 bytes float64)
	bucketAdj     = []byte("adj")     // key: src, value: concatenated int64 neighbor IDs
	bucketRevAdj  = []byte("rev_adj") // key: dst, value: concatenated int64 source IDs (directed only)
	bucketMeta    = []byte("meta")
	metaNodeCount = []byte("node_count")
)

// int64ToBytes encodes an int64 as big-endian bytes for bolt key ordering.
func int64ToBytes(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// bytesToInt64 decodes big-endian bytes back to int64.
func bytesToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

// edgeKey builds a 16-byte key from two node IDs.
func edgeKey(uid, vid int64) []byte {
	k := make([]byte, 16)
	binary.BigEndian.PutUint64(k[:8], uint64(uid))
	binary.BigEndian.PutUint64(k[8:], uint64(vid))
	return k
}

// float64ToBytes encodes a float64 as 8 bytes.
func float64ToBytes(f float64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(f))
	return b
}

// bytesToFloat64 decodes 8 bytes to a float64.
func bytesToFloat64(b []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

// appendID appends an int64 to a packed neighbor list.
func appendID(existing []byte, id int64) []byte {
	return append(existing, int64ToBytes(id)...)
}

// decodeIDs unpacks a concatenated list of int64 IDs.
func decodeIDs(b []byte) []int64 {
	if len(b) == 0 {
		return nil
	}
	n := len(b) / 8
	ids := make([]int64, n)
	for i := range n {
		ids[i] = bytesToInt64(b[i*8 : (i+1)*8])
	}
	return ids
}

// openReadOnly opens a bolt database in read-only mode.
func openReadOnly(path string) (*bolt.DB, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("diskgraph: open %s: %w", path, err)
	}
	return db, nil
}

// nodeCount reads the cached node count from the meta bucket.
func nodeCount(db *bolt.DB) int64 {
	var count int64
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMeta)
		if b == nil {
			return nil
		}
		v := b.Get(metaNodeCount)
		if v != nil {
			count = bytesToInt64(v)
		}
		return nil
	})
	return count
}
