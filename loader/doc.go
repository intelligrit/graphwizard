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
//	s, err := loader.LoadRelationships(db, "SELECT from_id, to_id, kind, year, claims, dollars FROM relations", []string{"claims", "dollars"})
//
// LoadRelationships preserves repeated endpoints and named weights in a
// multigraph.Store. Select type/year, weight, and reduction policy with
// Store.ProjectUndirected before passing its graph to an algorithm.
//
// Writing:
//
//	err := loader.WriteResults(db, "centrality_scores", scores)
//	err := loader.WriteCommunities(db, "community_assignments", communities)
package loader
