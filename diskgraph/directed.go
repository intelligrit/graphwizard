// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"database/sql"
	"math"

	"gonum.org/v1/gonum/graph"
)

// Directed is a read-only, disk-backed directed graph that implements
// graph.Directed and graph.WeightedDirected.
type Directed struct {
	db *sql.DB

	stmtFrom       *sql.Stmt
	stmtTo         *sql.Stmt
	stmtHasEdge    *sql.Stmt
	stmtHasEdgeDir *sql.Stmt
	stmtWeight     *sql.Stmt

	// nodeIDs is populated at open time (sorted ascending).
	nodeIDs []int64
}

// OpenDirected opens a SQLite file previously created by DirectedBuilder.
func OpenDirected(path string, opts ...Option) (*Directed, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}

	g := &Directed{db: db}

	g.stmtFrom, err = db.Prepare("SELECT dst FROM edges WHERE src = ?")
	if err != nil {
		db.Close()
		return nil, err
	}
	g.stmtTo, err = db.Prepare("SELECT src FROM edges WHERE dst = ?")
	if err != nil {
		g.stmtFrom.Close()
		db.Close()
		return nil, err
	}
	g.stmtHasEdgeDir, err = db.Prepare("SELECT 1 FROM edges WHERE src = ? AND dst = ? LIMIT 1")
	if err != nil {
		g.stmtFrom.Close()
		g.stmtTo.Close()
		db.Close()
		return nil, err
	}
	g.stmtWeight, err = db.Prepare("SELECT weight FROM edges WHERE src = ? AND dst = ? LIMIT 1")
	if err != nil {
		g.stmtFrom.Close()
		g.stmtTo.Close()
		g.stmtHasEdgeDir.Close()
		db.Close()
		return nil, err
	}

	// Preload node IDs.
	rows, err := db.Query("SELECT id FROM nodes ORDER BY id")
	if err != nil {
		g.closeStmts()
		db.Close()
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		g.nodeIDs = append(g.nodeIDs, id)
	}
	return g, nil
}

func (g *Directed) closeStmts() {
	if g.stmtFrom != nil {
		g.stmtFrom.Close()
	}
	if g.stmtTo != nil {
		g.stmtTo.Close()
	}
	if g.stmtHasEdgeDir != nil {
		g.stmtHasEdgeDir.Close()
	}
	if g.stmtWeight != nil {
		g.stmtWeight.Close()
	}
}

// Close releases the database.
func (g *Directed) Close() error {
	g.nodeIDs = nil
	g.closeStmts()
	return g.db.Close()
}

// Node returns the node with the given ID, or nil if it doesn't exist.
func (g *Directed) Node(id int64) graph.Node {
	if !nodeExists(g.nodeIDs, id) {
		return nil
	}
	return boltNode{id: id}
}

// Nodes returns an iterator over all nodes.
func (g *Directed) Nodes() graph.Nodes {
	return newSliceNodes(g.nodeIDs)
}

func queryIDs(stmt *sql.Stmt, id int64) graph.Nodes {
	rows, err := stmt.Query(id)
	if err != nil {
		return emptyNodes{}
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var v int64
		rows.Scan(&v)
		ids = append(ids, v)
	}
	if ids == nil {
		return emptyNodes{}
	}
	return newSliceNodes(ids)
}

// From returns all nodes reachable from the node with the given ID.
func (g *Directed) From(id int64) graph.Nodes {
	return queryIDs(g.stmtFrom, id)
}

// To returns all nodes that have an edge to the node with the given ID.
func (g *Directed) To(id int64) graph.Nodes {
	return queryIDs(g.stmtTo, id)
}

// HasEdgeBetween returns whether an edge exists between xid and yid
// in either direction.
func (g *Directed) HasEdgeBetween(xid, yid int64) bool {
	return g.HasEdgeFromTo(xid, yid) || g.HasEdgeFromTo(yid, xid)
}

// HasEdgeFromTo returns whether an edge exists from uid to vid.
func (g *Directed) HasEdgeFromTo(uid, vid int64) bool {
	var one int
	return g.stmtHasEdgeDir.QueryRow(uid, vid).Scan(&one) == nil
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
	err := g.stmtWeight.QueryRow(uid, vid).Scan(&w)
	if err != nil {
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
