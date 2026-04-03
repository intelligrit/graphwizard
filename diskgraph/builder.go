// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"database/sql"
	"fmt"
)

// UndirectedBuilder builds a disk-backed undirected graph.
// After adding all nodes and edges, call Close to finalize.
//
// For best performance, wrap calls in a Batch:
//
//	b.Batch(func(tx *UndirectedTx) error {
//	    tx.AddEdge(0, 1)
//	    tx.AddEdge(1, 2)
//	    return nil
//	})
type UndirectedBuilder struct {
	db        *sql.DB
	nodeCount int64
}

// NewUndirectedBuilder creates a new SQLite file and returns a builder.
func NewUndirectedBuilder(path string) (*UndirectedBuilder, error) {
	db, err := openReadWrite(path)
	if err != nil {
		return nil, err
	}
	for _, stmt := range []string{schemaNodes, schemaEdges, schemaEdgesIdx, schemaMeta} {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("diskgraph: schema: %w", err)
		}
	}
	return &UndirectedBuilder{db: db}, nil
}

// AddNode adds a node with the given ID.
func (b *UndirectedBuilder) AddNode(id int64) error {
	res, err := b.db.Exec("INSERT OR IGNORE INTO nodes (id) VALUES (?)", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	b.nodeCount += n
	return nil
}

// AddEdge adds an unweighted edge (weight 1.0) between uid and vid.
func (b *UndirectedBuilder) AddEdge(uid, vid int64) error {
	return b.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted edge between uid and vid.
// Both nodes are created if they don't exist.
func (b *UndirectedBuilder) AddWeightedEdge(uid, vid int64, weight float64) error {
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	if err := addUndirectedEdgeTx(tx, uid, vid, weight, &b.nodeCount); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// UndirectedTx provides batched write operations within a single transaction.
type UndirectedTx struct {
	tx           *sql.Tx
	nodeCount    *int64
	stmtNode     *sql.Stmt
	stmtEdge     *sql.Stmt
}

// AddNode adds a node within the batch transaction.
func (t *UndirectedTx) AddNode(id int64) error {
	res, err := t.stmtNode.Exec(id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	*t.nodeCount += n
	return nil
}

// AddEdge adds an unweighted edge within the batch transaction.
func (t *UndirectedTx) AddEdge(uid, vid int64) error {
	return t.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted edge within the batch transaction.
func (t *UndirectedTx) AddWeightedEdge(uid, vid int64, weight float64) error {
	// Ensure nodes exist.
	for _, id := range []int64{uid, vid} {
		res, err := t.stmtNode.Exec(id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		*t.nodeCount += n
	}
	// Insert both directions for undirected.
	if _, err := t.stmtEdge.Exec(uid, vid, weight); err != nil {
		return err
	}
	_, err := t.stmtEdge.Exec(vid, uid, weight)
	return err
}

// Batch executes all writes in fn within a single SQLite transaction.
// This is much faster than individual AddEdge calls.
func (b *UndirectedBuilder) Batch(fn func(tx *UndirectedTx) error) error {
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	stmtNode, err := tx.Prepare("INSERT OR IGNORE INTO nodes (id) VALUES (?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	stmtEdge, err := tx.Prepare("INSERT OR REPLACE INTO edges (src, dst, weight) VALUES (?, ?, ?)")
	if err != nil {
		stmtNode.Close()
		tx.Rollback()
		return err
	}
	utx := &UndirectedTx{
		tx:       tx,
		nodeCount: &b.nodeCount,
		stmtNode: stmtNode,
		stmtEdge: stmtEdge,
	}
	if err := fn(utx); err != nil {
		stmtNode.Close()
		stmtEdge.Close()
		tx.Rollback()
		return err
	}
	stmtNode.Close()
	stmtEdge.Close()
	return tx.Commit()
}

// Close finalizes the database, writing the node count to metadata.
func (b *UndirectedBuilder) Close() error {
	_, err := b.db.Exec("INSERT OR REPLACE INTO meta (key, value) VALUES ('node_count', ?)", b.nodeCount)
	if err != nil {
		b.db.Close()
		return err
	}
	return b.db.Close()
}

// DirectedBuilder builds a disk-backed directed graph.
type DirectedBuilder struct {
	db        *sql.DB
	nodeCount int64
}

// NewDirectedBuilder creates a new SQLite file for a directed graph.
func NewDirectedBuilder(path string) (*DirectedBuilder, error) {
	db, err := openReadWrite(path)
	if err != nil {
		return nil, err
	}
	for _, stmt := range []string{schemaNodes, schemaEdges, schemaEdgesIdx, schemaMeta} {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("diskgraph: schema: %w", err)
		}
	}
	return &DirectedBuilder{db: db}, nil
}

// AddNode adds a node with the given ID.
func (b *DirectedBuilder) AddNode(id int64) error {
	res, err := b.db.Exec("INSERT OR IGNORE INTO nodes (id) VALUES (?)", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	b.nodeCount += n
	return nil
}

// AddEdge adds an unweighted directed edge from uid to vid.
func (b *DirectedBuilder) AddEdge(uid, vid int64) error {
	return b.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted directed edge from uid to vid.
func (b *DirectedBuilder) AddWeightedEdge(uid, vid int64, weight float64) error {
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	if err := addDirectedEdgeTx(tx, uid, vid, weight, &b.nodeCount); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// DirectedTx provides batched write operations within a single transaction.
type DirectedTx struct {
	tx       *sql.Tx
	nodeCount *int64
	stmtNode *sql.Stmt
	stmtEdge *sql.Stmt
}

// AddNode adds a node within the batch transaction.
func (t *DirectedTx) AddNode(id int64) error {
	res, err := t.stmtNode.Exec(id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	*t.nodeCount += n
	return nil
}

// AddEdge adds an unweighted directed edge within the batch transaction.
func (t *DirectedTx) AddEdge(uid, vid int64) error {
	return t.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted directed edge within the batch transaction.
func (t *DirectedTx) AddWeightedEdge(uid, vid int64, weight float64) error {
	for _, id := range []int64{uid, vid} {
		res, err := t.stmtNode.Exec(id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		*t.nodeCount += n
	}
	_, err := t.stmtEdge.Exec(uid, vid, weight)
	return err
}

// Batch executes all writes in fn within a single SQLite transaction.
func (b *DirectedBuilder) Batch(fn func(tx *DirectedTx) error) error {
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	stmtNode, err := tx.Prepare("INSERT OR IGNORE INTO nodes (id) VALUES (?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	stmtEdge, err := tx.Prepare("INSERT OR REPLACE INTO edges (src, dst, weight) VALUES (?, ?, ?)")
	if err != nil {
		stmtNode.Close()
		tx.Rollback()
		return err
	}
	dtx := &DirectedTx{
		tx:        tx,
		nodeCount: &b.nodeCount,
		stmtNode:  stmtNode,
		stmtEdge:  stmtEdge,
	}
	if err := fn(dtx); err != nil {
		stmtNode.Close()
		stmtEdge.Close()
		tx.Rollback()
		return err
	}
	stmtNode.Close()
	stmtEdge.Close()
	return tx.Commit()
}

// Close finalizes the directed graph database.
func (b *DirectedBuilder) Close() error {
	_, err := b.db.Exec("INSERT OR REPLACE INTO meta (key, value) VALUES ('node_count', ?)", b.nodeCount)
	if err != nil {
		b.db.Close()
		return err
	}
	return b.db.Close()
}

// --- Internal transaction helpers (for single-edge writes) ---

func addUndirectedEdgeTx(tx *sql.Tx, uid, vid int64, weight float64, count *int64) error {
	for _, id := range []int64{uid, vid} {
		res, err := tx.Exec("INSERT OR IGNORE INTO nodes (id) VALUES (?)", id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		*count += n
	}
	if _, err := tx.Exec("INSERT OR REPLACE INTO edges (src, dst, weight) VALUES (?, ?, ?)", uid, vid, weight); err != nil {
		return err
	}
	_, err := tx.Exec("INSERT OR REPLACE INTO edges (src, dst, weight) VALUES (?, ?, ?)", vid, uid, weight)
	return err
}

func addDirectedEdgeTx(tx *sql.Tx, uid, vid int64, weight float64, count *int64) error {
	for _, id := range []int64{uid, vid} {
		res, err := tx.Exec("INSERT OR IGNORE INTO nodes (id) VALUES (?)", id)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		*count += n
	}
	_, err := tx.Exec("INSERT OR REPLACE INTO edges (src, dst, weight) VALUES (?, ?, ?)", uid, vid, weight)
	return err
}
