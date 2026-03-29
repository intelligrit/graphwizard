// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

//go:build integration

package loader

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/marcboeker/go-duckdb/v2"
)

// TestCreateTestDB creates the DuckDB test fixture. Run with:
//
//	go test -tags integration -run TestCreateTestDB ./loader/
func TestCreateTestDB(t *testing.T) {
	path := "testdata/test.duckdb"
	os.MkdirAll("testdata", 0o755)
	os.Remove(path)

	db, err := sql.Open("duckdb", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Unweighted edges table.
	_, err = db.Exec(`
		CREATE TABLE edges (from_id BIGINT, to_id BIGINT);
		INSERT INTO edges VALUES (0, 1), (1, 2), (2, 0), (2, 3), (3, 4), (4, 5), (5, 3);
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Weighted edges table.
	_, err = db.Exec(`
		CREATE TABLE weighted_edges (from_id BIGINT, to_id BIGINT, weight DOUBLE);
		INSERT INTO weighted_edges VALUES
			(0, 1, 1.0), (1, 2, 2.0), (2, 0, 3.0),
			(2, 3, 0.5), (3, 4, 1.5), (4, 5, 2.5), (5, 3, 1.0);
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Results table (empty, for write tests).
	_, err = db.Exec(`
		CREATE TABLE results (node_id BIGINT, score DOUBLE);
		CREATE TABLE communities (node_id BIGINT, community_id BIGINT);
	`)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("created %s", path)
}
