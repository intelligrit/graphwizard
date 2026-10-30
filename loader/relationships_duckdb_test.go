// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

//go:build integration

package loader

import (
	"database/sql"
	"testing"

	"github.com/intelligrit/graphwizard/multigraph"
	_ "github.com/marcboeker/go-duckdb/v2"
)

func TestLoadRelationshipsDuckDB(t *testing.T) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s, err := LoadRelationships(db, `
SELECT src,dst,kind,year,claims,dollars,NULL::DOUBLE AS risk
FROM (VALUES
 (1::BIGINT,2::BIGINT,'prescribed',2022,10::DOUBLE,100::DOUBLE),
 (1,2,'prescribed',2023,15,180),
 (1,2,'paid',2023,NULL,0)
) AS r(src,dst,kind,year,claims,dollars)
ORDER BY year,kind`, []string{"claims", "dollars", "risk"})
	if err != nil {
		t.Fatal(err)
	}
	if s.NumRelationships() != 3 || s.NumNodes() != 2 {
		t.Fatal("DuckDB relationships collapsed")
	}
	for r := range s.Relationships() {
		if _, ok := s.Weight(r.ID, "risk"); ok {
			t.Fatal("NULL risk became a value")
		}
		if r.Type == "paid" {
			if _, ok := s.Weight(r.ID, "claims"); ok {
				t.Fatal("NULL claims became a value")
			}
			if w, ok := s.Weight(r.ID, "dollars"); !ok || w != 0 {
				t.Fatal("zero dollars became missing")
			}
		}
	}
	g, err := s.ProjectUndirected(multigraph.Projection{
		Weight: "claims", Filter: func(r multigraph.Relationship) bool { return r.Type == "prescribed" && r.Year == 2023 },
	})
	if err != nil {
		t.Fatal(err)
	}
	if w, ok := g.Weight(1, 2); !ok || w != 15 {
		t.Fatal("year-specific claims changed")
	}
}
