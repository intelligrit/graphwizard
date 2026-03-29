// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package loader provides utilities for loading graphs from SQL databases
// and writing algorithm results back to database tables.
//
// All functions use the standard database/sql interfaces, so any SQL driver
// (PostgreSQL, MySQL, SQLite, DuckDB, etc.) works without additional imports.
//
// Loading:
//
//	db, _ := sql.Open("postgres", connStr)
//	g, err := loader.LoadDirected(db, "SELECT from_id, to_id FROM edges")
//	g, err := loader.LoadWeightedDirected(db, "SELECT from_id, to_id, weight FROM edges")
//
// Writing:
//
//	err := loader.WriteResults(db, "centrality_scores", scores)
//	err := loader.WriteCommunities(db, "community_assignments", communities)
package loader
