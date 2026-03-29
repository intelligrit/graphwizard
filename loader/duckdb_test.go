// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

//go:build integration

package loader

import (
	"database/sql"
	"math"
	"testing"

	_ "github.com/marcboeker/go-duckdb/v2"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("duckdb", "testdata/test.duckdb?access_mode=READ_ONLY")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func openWritableDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatal(err)
	}
	// Create tables in memory.
	db.Exec("CREATE TABLE results (node_id BIGINT, score DOUBLE)")
	db.Exec("CREATE TABLE communities (node_id BIGINT, community_id BIGINT)")
	return db
}

func TestDuckDB_LoadDirected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	g, err := LoadDirected(db, "SELECT from_id, to_id FROM edges")
	if err != nil {
		t.Fatal(err)
	}

	if g.Nodes().Len() != 6 {
		t.Errorf("expected 6 nodes, got %d", g.Nodes().Len())
	}
	if !g.HasEdgeFromTo(0, 1) {
		t.Error("expected edge 0->1")
	}
	if !g.HasEdgeFromTo(2, 3) {
		t.Error("expected edge 2->3")
	}
}

func TestDuckDB_LoadUndirected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	g, err := LoadUndirected(db, "SELECT from_id, to_id FROM edges")
	if err != nil {
		t.Fatal(err)
	}

	if g.Nodes().Len() != 6 {
		t.Errorf("expected 6 nodes, got %d", g.Nodes().Len())
	}
	if !g.HasEdgeBetween(0, 1) {
		t.Error("expected edge 0-1")
	}
}

func TestDuckDB_LoadWeightedDirected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	g, err := LoadWeightedDirected(db, "SELECT from_id, to_id, weight FROM weighted_edges")
	if err != nil {
		t.Fatal(err)
	}

	if g.Nodes().Len() != 6 {
		t.Errorf("expected 6 nodes, got %d", g.Nodes().Len())
	}
	w, ok := g.Weight(0, 1)
	if !ok || math.Abs(w-1.0) > 1e-9 {
		t.Errorf("expected weight 1.0 for 0->1, got %f (ok=%v)", w, ok)
	}
}

func TestDuckDB_LoadWeightedUndirected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	g, err := LoadWeightedUndirected(db, "SELECT from_id, to_id, weight FROM weighted_edges")
	if err != nil {
		t.Fatal(err)
	}

	if g.Nodes().Len() != 6 {
		t.Errorf("expected 6 nodes, got %d", g.Nodes().Len())
	}
	w, ok := g.Weight(2, 3)
	if !ok || math.Abs(w-0.5) > 1e-9 {
		t.Errorf("expected weight 0.5 for 2-3, got %f (ok=%v)", w, ok)
	}
}

func TestDuckDB_WriteResults(t *testing.T) {
	db := openWritableDB(t)
	defer db.Close()

	scores := map[int64]float64{0: 1.5, 1: 2.5, 2: 0.5}
	err := WriteResults(db, "results", scores)
	if err != nil {
		t.Fatal(err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM results").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}

	var score float64
	db.QueryRow("SELECT score FROM results WHERE node_id = 1").Scan(&score)
	if math.Abs(score-2.5) > 1e-9 {
		t.Errorf("expected 2.5, got %f", score)
	}
}

func TestDuckDB_WriteCommunities(t *testing.T) {
	db := openWritableDB(t)
	defer db.Close()

	comms := map[int64]int64{0: 1, 1: 1, 2: 2, 3: 2}
	err := WriteCommunities(db, "communities", comms)
	if err != nil {
		t.Fatal(err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM communities").Scan(&count)
	if count != 4 {
		t.Errorf("expected 4 rows, got %d", count)
	}
}

func TestDuckDB_LoadDirected_EmptyTable(t *testing.T) {
	db := openWritableDB(t)
	defer db.Close()
	db.Exec("CREATE TABLE empty_edges (from_id BIGINT, to_id BIGINT)")

	g, err := LoadDirected(db, "SELECT from_id, to_id FROM empty_edges")
	if err != nil {
		t.Fatal(err)
	}
	if g.Nodes().Len() != 0 {
		t.Errorf("expected 0 nodes, got %d", g.Nodes().Len())
	}
}

func TestDuckDB_LoadDirected_BadQuery(t *testing.T) {
	db := openWritableDB(t)
	defer db.Close()

	_, err := LoadDirected(db, "SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("expected error for bad query")
	}
}
