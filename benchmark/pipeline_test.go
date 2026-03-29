// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

//go:build integration

package benchmark

import (
	"database/sql"
	"math/rand"
	"testing"

	"github.com/intelligrit/graphwizard/anomaly"
	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/connectivity"
	"github.com/intelligrit/graphwizard/loader"
	"github.com/intelligrit/graphwizard/structure"
	"github.com/intelligrit/graphwizard/subgraph"
	_ "github.com/marcboeker/go-duckdb/v2"
)

// TestFullPipeline exercises the complete Integrity-style workflow:
// load from DuckDB → run algorithms → write back → verify.
func TestFullPipeline(t *testing.T) {
	// 1. Load graph from DuckDB.
	db, err := sql.Open("duckdb", "../loader/testdata/test.duckdb?access_mode=READ_ONLY")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	g, err := loader.LoadUndirected(db, "SELECT from_id, to_id FROM edges")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("loaded graph: %d nodes", g.Nodes().Len())
	if g.Nodes().Len() != 6 {
		t.Fatalf("expected 6 nodes, got %d", g.Nodes().Len())
	}

	// 2. Community detection (Leiden).
	rng := rand.New(rand.NewSource(42))
	comms := community.Leiden(g, 1.0, rng)
	t.Logf("communities: %v", comms)
	if len(comms) != 6 {
		t.Errorf("expected 6 community assignments, got %d", len(comms))
	}

	// 3. Betweenness centrality.
	scores := centrality.Betweenness(g)
	t.Logf("betweenness: %v", scores)
	if len(scores) < 1 {
		t.Errorf("expected at least 1 betweenness score, got %d", len(scores))
	}

	// 4. Triangle count.
	perNode, total := structure.TriangleCount(g)
	t.Logf("triangles: %d total, per-node: %v", total, perNode)
	if total < 1 {
		t.Error("expected at least 1 triangle")
	}

	// 5. Bridge detection.
	bridges := connectivity.Bridges(g)
	t.Logf("bridges: %d", len(bridges))

	// 6. Anomaly detection.
	isoScores := anomaly.IsolationScore(g)
	t.Logf("isolation scores: %v", isoScores)
	if len(isoScores) != 6 {
		t.Errorf("expected 6 isolation scores, got %d", len(isoScores))
	}

	// 7. Subgraph extraction (2-hop from node 0).
	sub := subgraph.NHopNeighborhoodUndirected(g, 0, 2)
	subCount := sub.Nodes().Len()
	t.Logf("2-hop subgraph from node 0: %d nodes", subCount)
	if subCount < 2 {
		t.Error("subgraph should have at least 2 nodes")
	}

	// 8. Write results back to an in-memory DuckDB.
	writeDB, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer writeDB.Close()

	writeDB.Exec("CREATE TABLE communities (node_id BIGINT, community_id BIGINT)")
	writeDB.Exec("CREATE TABLE scores (node_id BIGINT, score DOUBLE)")

	err = loader.WriteCommunities(writeDB, "communities", comms)
	if err != nil {
		t.Fatalf("WriteCommunities: %v", err)
	}

	err = loader.WriteResults(writeDB, "scores", scores)
	if err != nil {
		t.Fatalf("WriteResults: %v", err)
	}

	// 9. Read back and verify.
	var count int
	writeDB.QueryRow("SELECT COUNT(*) FROM communities").Scan(&count)
	if count != 6 {
		t.Errorf("expected 6 community rows, got %d", count)
	}

	writeDB.QueryRow("SELECT COUNT(*) FROM scores").Scan(&count)
	if count < 1 {
		t.Errorf("expected at least 1 score row, got %d", count)
	}

	t.Log("full pipeline: PASS")
}

func BenchmarkLoadDirected_DuckDB(b *testing.B) {
	db, err := sql.Open("duckdb", "../loader/testdata/test.duckdb?access_mode=READ_ONLY")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader.LoadDirected(db, "SELECT from_id, to_id FROM edges")
	}
}

func BenchmarkWriteResults_DuckDB(b *testing.B) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	scores := make(map[int64]float64, 1000)
	for i := int64(0); i < 1000; i++ {
		scores[i] = float64(i) * 0.01
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Exec("DROP TABLE IF EXISTS bench_results")
		db.Exec("CREATE TABLE bench_results (node_id BIGINT, score DOUBLE)")
		loader.WriteResults(db, "bench_results", scores)
	}
}
