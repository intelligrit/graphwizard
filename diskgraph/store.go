// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Schema used by all graph types.
const schemaNodes = `CREATE TABLE IF NOT EXISTS nodes (id INTEGER PRIMARY KEY)`
const schemaEdges = `CREATE TABLE IF NOT EXISTS edges (
	src INTEGER NOT NULL,
	dst INTEGER NOT NULL,
	weight REAL NOT NULL DEFAULT 1.0,
	PRIMARY KEY (src, dst)
)`
const schemaEdgesIdx = `CREATE INDEX IF NOT EXISTS idx_edges_dst ON edges (dst, src)`
const schemaMeta = `CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value INTEGER)`

// openReadOnly opens a SQLite database in read-only mode with performance tunings.
func openReadOnly(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?mode=ro&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("diskgraph: open %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	pragmas(db)
	return db, nil
}

// openReadWrite opens a SQLite database for writing.
func openReadWrite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("diskgraph: create %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	pragmas(db)
	return db, nil
}

func pragmas(db *sql.DB) {
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA cache_size=-64000") // 64MB cache
	db.Exec("PRAGMA mmap_size=268435456") // 256MB mmap
	db.Exec("PRAGMA temp_store=MEMORY")
}

// nodeCount reads the cached node count from the meta table.
func nodeCount(db *sql.DB) int64 {
	var count int64
	db.QueryRow("SELECT value FROM meta WHERE key='node_count'").Scan(&count)
	return count
}
