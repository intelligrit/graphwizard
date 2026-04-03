// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"database/sql"
	"math"

	"gonum.org/v1/gonum/graph"
)

// Undirected is a read-only, disk-backed undirected graph that implements
// graph.Undirected and graph.WeightedUndirected. It reads directly from
// a SQLite database using prepared statements.
//
// By default, the graph reads adjacency data via SQL queries. Call
// PreloadAdjacency() or pass Preload to cache adjacency in Go memory
// for additional speed on algorithms that call HasEdgeBetween in tight
// loops (e.g., ClusteringCoefficient).
type Undirected struct {
	db *sql.DB

	// Prepared statements for hot-path queries.
	stmtFrom     *sql.Stmt
	stmtHasEdge  *sql.Stmt
	stmtWeight   *sql.Stmt

	// nodeIDs is populated at open time.
	nodeIDs []int64
	nodeSet map[int64]struct{}

	// adjCache, when non-nil, holds preloaded adjacency lists.
	adjCache map[int64][]int64
	// adjSet mirrors adjCache as sets for O(1) HasEdgeBetween lookups.
	adjSet map[int64]map[int64]struct{}
}

// OpenUndirected opens a SQLite file previously created by UndirectedBuilder.
// Pass Preload to cache adjacency data in memory for maximum speed, or call
// PreloadAdjacency() after opening.
func OpenUndirected(path string, opts ...Option) (*Undirected, error) {
	var cfg openConfig
	for _, o := range opts {
		o(&cfg)
	}

	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}

	g := &Undirected{db: db}

	// Prepare hot-path statements.
	g.stmtFrom, err = db.Prepare("SELECT dst FROM edges WHERE src = ?")
	if err != nil {
		db.Close()
		return nil, err
	}
	g.stmtHasEdge, err = db.Prepare("SELECT 1 FROM edges WHERE src = ? AND dst = ? LIMIT 1")
	if err != nil {
		g.stmtFrom.Close()
		db.Close()
		return nil, err
	}
	g.stmtWeight, err = db.Prepare("SELECT weight FROM edges WHERE src = ? AND dst = ? LIMIT 1")
	if err != nil {
		g.stmtFrom.Close()
		g.stmtHasEdge.Close()
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
	g.nodeSet = make(map[int64]struct{}, len(g.nodeIDs))
	for _, id := range g.nodeIDs {
		g.nodeSet[id] = struct{}{}
	}

	// Preload adjacency if requested.
	if cfg.preload {
		tryPreload(g, cfg)
	}

	return g, nil
}

// PreloadAdjacency loads all adjacency lists into memory.
func (g *Undirected) PreloadAdjacency() {
	g.preloadAdj()
}

func (g *Undirected) preloadAdj() {
	n := len(g.nodeIDs)
	g.adjCache = make(map[int64][]int64, n)
	g.adjSet = make(map[int64]map[int64]struct{}, n)

	rows, err := g.db.Query("SELECT src, dst FROM edges")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var src, dst int64
		rows.Scan(&src, &dst)
		g.adjCache[src] = append(g.adjCache[src], dst)
	}
	for id, neighbors := range g.adjCache {
		s := make(map[int64]struct{}, len(neighbors))
		for _, nid := range neighbors {
			s[nid] = struct{}{}
		}
		g.adjSet[id] = s
	}
}

// adjBucketSize estimates adjacency data size for memory checking.
func (g *Undirected) adjBucketSize() int64 {
	var count int64
	g.db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&count)
	return count * 16 // 2 int64s per row
}

func (g *Undirected) closeStmts() {
	if g.stmtFrom != nil {
		g.stmtFrom.Close()
	}
	if g.stmtHasEdge != nil {
		g.stmtHasEdge.Close()
	}
	if g.stmtWeight != nil {
		g.stmtWeight.Close()
	}
}

// Close releases the database.
func (g *Undirected) Close() error {
	g.nodeIDs = nil
	g.nodeSet = nil
	g.adjCache = nil
	g.adjSet = nil
	g.closeStmts()
	return g.db.Close()
}

// Node returns the node with the given ID, or nil if it doesn't exist.
func (g *Undirected) Node(id int64) graph.Node {
	if _, ok := g.nodeSet[id]; !ok {
		return nil
	}
	return boltNode{id: id}
}

// Nodes returns an iterator over all nodes.
func (g *Undirected) Nodes() graph.Nodes {
	return newSliceNodes(g.nodeIDs)
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
	rows, err := g.stmtFrom.Query(id)
	if err != nil {
		return emptyNodes{}
	}
	defer rows.Close()
	var neighbors []int64
	for rows.Next() {
		var dst int64
		rows.Scan(&dst)
		neighbors = append(neighbors, dst)
	}
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
	var one int
	err := g.stmtHasEdge.QueryRow(xid, yid).Scan(&one)
	return err == nil
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
	err := g.stmtWeight.QueryRow(uid, vid).Scan(&w)
	if err != nil {
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
