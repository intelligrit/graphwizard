// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package loader

import (
	"database/sql"
	"fmt"

	"gonum.org/v1/gonum/graph/simple"
)

// RowScanner is an interface for iterating over query result rows.
// *sql.Rows satisfies this interface.
type RowScanner interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

// LoadDirected executes query against db and builds a directed graph.
// The query must return rows of (from_id INT, to_id INT).
func LoadDirected(db *sql.DB, query string) (*simple.DirectedGraph, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("loader: query failed: %w", err)
	}
	defer rows.Close()

	return loadDirectedFromRows(rows)
}

// loadDirectedFromRows builds a directed graph from a RowScanner.
func loadDirectedFromRows(rows RowScanner) (*simple.DirectedGraph, error) {
	g := simple.NewDirectedGraph()
	if err := scanEdges(rows, g, false); err != nil {
		return nil, err
	}
	return g, nil
}

// LoadWeightedDirected executes query against db and builds a weighted
// directed graph. The query must return rows of (from_id INT, to_id INT,
// weight FLOAT).
func LoadWeightedDirected(db *sql.DB, query string) (*simple.WeightedDirectedGraph, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("loader: query failed: %w", err)
	}
	defer rows.Close()

	return loadWeightedDirectedFromRows(rows)
}

// loadWeightedDirectedFromRows builds a weighted directed graph from a RowScanner.
func loadWeightedDirectedFromRows(rows RowScanner) (*simple.WeightedDirectedGraph, error) {
	g := simple.NewWeightedDirectedGraph(0, 0)
	if err := scanWeightedEdges(rows, g); err != nil {
		return nil, err
	}
	return g, nil
}

// LoadUndirected executes query against db and builds an undirected graph.
// The query must return rows of (from_id INT, to_id INT).
func LoadUndirected(db *sql.DB, query string) (*simple.UndirectedGraph, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("loader: query failed: %w", err)
	}
	defer rows.Close()

	return loadUndirectedFromRows(rows)
}

// loadUndirectedFromRows builds an undirected graph from a RowScanner.
func loadUndirectedFromRows(rows RowScanner) (*simple.UndirectedGraph, error) {
	g := simple.NewUndirectedGraph()
	if err := scanEdges(rows, g, true); err != nil {
		return nil, err
	}
	return g, nil
}

// LoadWeightedUndirected executes query against db and builds a weighted
// undirected graph. The query must return rows of (from_id INT, to_id INT,
// weight FLOAT).
func LoadWeightedUndirected(db *sql.DB, query string) (*simple.WeightedUndirectedGraph, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("loader: query failed: %w", err)
	}
	defer rows.Close()

	return loadWeightedUndirectedFromRows(rows)
}

// loadWeightedUndirectedFromRows builds a weighted undirected graph from a RowScanner.
func loadWeightedUndirectedFromRows(rows RowScanner) (*simple.WeightedUndirectedGraph, error) {
	g := simple.NewWeightedUndirectedGraph(0, 0)
	if err := scanWeightedEdges(rows, g); err != nil {
		return nil, err
	}
	return g, nil
}

func scanEdges(rows RowScanner, g interface{}, undirected bool) error {
	seen := make(map[int64]bool)
	for rows.Next() {
		var from, to int64
		if err := rows.Scan(&from, &to); err != nil {
			return fmt.Errorf("loader: scan failed: %w", err)
		}
		switch gr := g.(type) {
		case *simple.DirectedGraph:
			ensureNodeDirected(gr, from, seen)
			ensureNodeDirected(gr, to, seen)
			gr.SetEdge(gr.NewEdge(simple.Node(from), simple.Node(to)))
		case *simple.UndirectedGraph:
			ensureNodeUndirected(gr, from, seen)
			ensureNodeUndirected(gr, to, seen)
			gr.SetEdge(gr.NewEdge(simple.Node(from), simple.Node(to)))
		}
	}
	return rows.Err()
}

func scanWeightedEdges(rows RowScanner, g interface{}) error {
	seen := make(map[int64]bool)
	for rows.Next() {
		var from, to int64
		var weight float64
		if err := rows.Scan(&from, &to, &weight); err != nil {
			return fmt.Errorf("loader: scan failed: %w", err)
		}
		switch gr := g.(type) {
		case *simple.WeightedDirectedGraph:
			ensureNodeWeightedDirected(gr, from, seen)
			ensureNodeWeightedDirected(gr, to, seen)
			gr.SetWeightedEdge(gr.NewWeightedEdge(simple.Node(from), simple.Node(to), weight))
		case *simple.WeightedUndirectedGraph:
			ensureNodeWeightedUndirected(gr, from, seen)
			ensureNodeWeightedUndirected(gr, to, seen)
			gr.SetWeightedEdge(gr.NewWeightedEdge(simple.Node(from), simple.Node(to), weight))
		}
	}
	return rows.Err()
}

func ensureNodeDirected(g *simple.DirectedGraph, id int64, seen map[int64]bool) {
	if !seen[id] {
		g.AddNode(simple.Node(id))
		seen[id] = true
	}
}

func ensureNodeUndirected(g *simple.UndirectedGraph, id int64, seen map[int64]bool) {
	if !seen[id] {
		g.AddNode(simple.Node(id))
		seen[id] = true
	}
}

func ensureNodeWeightedDirected(g *simple.WeightedDirectedGraph, id int64, seen map[int64]bool) {
	if !seen[id] {
		g.AddNode(simple.Node(id))
		seen[id] = true
	}
}

func ensureNodeWeightedUndirected(g *simple.WeightedUndirectedGraph, id int64, seen map[int64]bool) {
	if !seen[id] {
		g.AddNode(simple.Node(id))
		seen[id] = true
	}
}
