// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package loader

import (
	"database/sql"
	"fmt"
	"sort"
)

// WriteResults writes centrality scores to a two-column table (node_id, score).
// The table is created if it does not exist. Existing data is replaced.
func WriteResults(db *sql.DB, table string, results map[int64]float64) error {
	createSQL := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (node_id BIGINT PRIMARY KEY, score DOUBLE PRECISION)",
		table,
	)
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("loader: create table: %w", err)
	}

	if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
		return fmt.Errorf("loader: truncate table: %w", err)
	}

	// Sort keys for deterministic insertion order.
	ids := make([]int64, 0, len(results))
	for id := range results {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	insertSQL := fmt.Sprintf("INSERT INTO %s (node_id, score) VALUES ($1, $2)", table)
	for _, id := range ids {
		if _, err := db.Exec(insertSQL, id, results[id]); err != nil {
			return fmt.Errorf("loader: insert node %d: %w", id, err)
		}
	}
	return nil
}

// WriteCommunities writes community assignments to a two-column table
// (node_id, community_id). The table is created if it does not exist.
func WriteCommunities(db *sql.DB, table string, comms map[int64]int64) error {
	createSQL := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (node_id BIGINT PRIMARY KEY, community_id BIGINT)",
		table,
	)
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("loader: create table: %w", err)
	}

	if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
		return fmt.Errorf("loader: truncate table: %w", err)
	}

	ids := make([]int64, 0, len(comms))
	for id := range comms {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	insertSQL := fmt.Sprintf("INSERT INTO %s (node_id, community_id) VALUES ($1, $2)", table)
	for _, id := range ids {
		if _, err := db.Exec(insertSQL, id, comms[id]); err != nil {
			return fmt.Errorf("loader: insert node %d: %w", id, err)
		}
	}
	return nil
}

// WriteRows executes the given parameterized query once per row. Each row
// element is passed as a positional parameter ($1, $2, ...).
//
// Example:
//
//	rows := [][]interface{}{{1, 2, 3.14}, {4, 5, 2.71}}
//	err := WriteRows(db, "INSERT INTO edges (from_id, to_id, weight) VALUES ($1, $2, $3)", rows)
func WriteRows(db *sql.DB, query string, rows [][]interface{}) error {
	// Validate that the query has the expected number of placeholders.
	for i, row := range rows {
		if _, err := db.Exec(query, row...); err != nil {
			return fmt.Errorf("loader: write row %d: %w", i, err)
		}
	}
	return nil
}

