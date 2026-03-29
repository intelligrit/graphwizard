// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package loader_test

import (
	"database/sql"
	"log"

	"github.com/intelligrit/graphwizard/loader"
)

// openTestDB is a placeholder that would open a real database connection.
func openTestDB() *sql.DB {
	// In production: sql.Open("postgres", connStr)
	return nil
}

// ExampleLoadDirected demonstrates loading a directed graph from a SQL query.
// The query must return rows of (from_id INT, to_id INT).
func ExampleLoadDirected() {
	db := openTestDB()
	if db == nil {
		return
	}
	g, err := loader.LoadDirected(db, "SELECT from_id, to_id FROM edges")
	if err != nil {
		log.Fatal(err)
	}
	_ = g // use g with graphwizard algorithms
}

// ExampleLoadWeightedDirected demonstrates loading a weighted directed graph.
// The query must return rows of (from_id INT, to_id INT, weight FLOAT).
func ExampleLoadWeightedDirected() {
	db := openTestDB()
	if db == nil {
		return
	}
	g, err := loader.LoadWeightedDirected(db, "SELECT from_id, to_id, weight FROM edges")
	if err != nil {
		log.Fatal(err)
	}
	_ = g
}

// ExampleLoadUndirected demonstrates loading an undirected graph from a SQL query.
func ExampleLoadUndirected() {
	db := openTestDB()
	if db == nil {
		return
	}
	g, err := loader.LoadUndirected(db, "SELECT from_id, to_id FROM edges")
	if err != nil {
		log.Fatal(err)
	}
	_ = g
}

// ExampleLoadWeightedUndirected demonstrates loading a weighted undirected graph.
func ExampleLoadWeightedUndirected() {
	db := openTestDB()
	if db == nil {
		return
	}
	g, err := loader.LoadWeightedUndirected(db, "SELECT from_id, to_id, weight FROM edges")
	if err != nil {
		log.Fatal(err)
	}
	_ = g
}

// ExampleWriteResults demonstrates writing centrality scores to a database table.
func ExampleWriteResults() {
	db := openTestDB()
	if db == nil {
		return
	}
	scores := map[int64]float64{1: 0.85, 2: 0.42, 3: 0.73}
	err := loader.WriteResults(db, "centrality_scores", scores)
	if err != nil {
		log.Fatal(err)
	}
}

// ExampleWriteCommunities demonstrates writing community assignments to a database table.
func ExampleWriteCommunities() {
	db := openTestDB()
	if db == nil {
		return
	}
	comms := map[int64]int64{1: 0, 2: 0, 3: 1, 4: 1}
	err := loader.WriteCommunities(db, "community_assignments", comms)
	if err != nil {
		log.Fatal(err)
	}
}

// ExampleWriteRows demonstrates writing arbitrary rows to a database table.
func ExampleWriteRows() {
	db := openTestDB()
	if db == nil {
		return
	}
	rows := [][]interface{}{
		{1, 2, 3.14},
		{4, 5, 2.71},
	}
	err := loader.WriteRows(db, "INSERT INTO edges (from_id, to_id, weight) VALUES ($1, $2, $3)", rows)
	if err != nil {
		log.Fatal(err)
	}
}
