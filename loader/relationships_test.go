// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package loader

import (
	"database/sql"
	"errors"
	"slices"
	"testing"

	"github.com/intelligrit/graphwizard/multigraph"
	_ "modernc.org/sqlite"
)

func relationshipDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`CREATE TABLE relations (src INTEGER, dst INTEGER, kind TEXT, year INTEGER, claims REAL, dollars REAL, missing REAL);
INSERT INTO relations VALUES
(1,2,'prescribed',2022,10,100,NULL),
(1,2,'prescribed',2023,15,180,NULL),
(1,2,'referred',2023,0,NULL,NULL),
(2,3,NULL,NULL,NULL,5,NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestLoadRelationshipsPreservesRowsAndNulls(t *testing.T) {
	db := relationshipDB(t)
	s, err := LoadRelationships(db, "SELECT * FROM relations ORDER BY rowid", []string{"claims", "dollars", "missing"}, multigraph.BlockWeights)
	if err != nil {
		t.Fatal(err)
	}
	if s.NumRelationships() != 4 || s.NumNodes() != 3 {
		t.Fatal("relationships collapsed")
	}
	if !slices.Equal(s.WeightNames(), []string{"claims", "dollars", "missing"}) {
		t.Fatal("all-null weight name lost")
	}
	for id, want := range []float64{10, 15, 0} {
		if w, ok := s.Weight(int64(id), "claims"); !ok || w != want {
			t.Fatal("claims changed")
		}
	}
	if _, ok := s.Weight(2, "dollars"); ok {
		t.Fatal("NULL became a number")
	}
	if _, ok := s.Weight(3, "claims"); ok {
		t.Fatal("previous row value leaked")
	}
	if r, _ := s.Relationship(3); r.Type != "" || r.Year != 0 {
		t.Fatal("NULL attributes changed")
	}
	if _, err := s.ProjectUndirected(multigraph.Projection{Weight: "missing"}); !errors.Is(err, multigraph.ErrMissingWeight) {
		t.Fatal("all-null metric not recognized")
	}
	g, err := s.ProjectUndirected(multigraph.Projection{
		Weight: "claims", Filter: func(r multigraph.Relationship) bool { return r.Type == "prescribed" && r.Year == 2023 },
	})
	if err != nil {
		t.Fatal(err)
	}
	if w, ok := g.Weight(1, 2); !ok || w != 15 {
		t.Fatal("SQL year selection failed")
	}
}

func TestScanRelationshipsRetainsIsolatesAndEmptySchema(t *testing.T) {
	db := relationshipDB(t)
	b := multigraph.NewBuilderSized(4, 4)
	b.AddNode(99)
	rows, err := db.Query("SELECT src,dst,kind,year,claims FROM relations ORDER BY rowid")
	if err != nil {
		t.Fatal(err)
	}
	if err := ScanRelationships(rows, b, []string{"claims"}); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	s := b.Build()
	if s.NumNodes() != 4 || s.NumRelationships() != 4 {
		t.Fatal("isolated node lost")
	}
	empty, err := LoadRelationships(db, "SELECT src,dst,kind,year,claims FROM relations WHERE 0", []string{"claims"})
	if err != nil {
		t.Fatal(err)
	}
	if empty.NumRelationships() != 0 || !slices.Equal(empty.WeightNames(), []string{"claims"}) {
		t.Fatal("empty input lost weight schema")
	}
	if _, err := empty.ProjectUndirected(multigraph.Projection{Weight: "claims"}); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRelationshipsErrors(t *testing.T) {
	db := relationshipDB(t)
	for _, tc := range []struct {
		query string
		names []string
	}{
		{"SELECT * FROM relations", []string{""}},
		{"SELECT * FROM relations", []string{"same", "same"}},
		{"SELECT bad_column FROM relations", nil},
		{"SELECT src,dst FROM relations", nil},
		{"SELECT NULL,2,'kind',2023", nil},
		{"SELECT 1,2,'kind',2147483648", nil},
		{"SELECT 1,2,'kind',-2147483649", nil},
		{"SELECT 1,2,'kind',2023,'not-a-number'", []string{"v"}},
	} {
		if s, err := LoadRelationships(db, tc.query, tc.names); err == nil || s != nil {
			t.Fatalf("invalid input accepted: %s", tc.query)
		}
	}
	for _, policy := range []multigraph.Missing{multigraph.SkipMissing, multigraph.ZeroMissing} {
		s, err := LoadRelationships(db, "SELECT src,dst,kind,year,missing FROM relations", []string{"v"})
		if err != nil {
			t.Fatal(err)
		}
		g, err := s.ProjectUndirected(multigraph.Projection{Weight: "v", Combine: multigraph.Sum, Missing: policy})
		if err != nil {
			t.Fatal(err)
		}
		if g.HasEdgeBetween(1, 2) != (policy == multigraph.ZeroMissing) {
			t.Fatal("all-null projection policy failed")
		}
	}
}

type failedRelationshipRows struct{ err error }

func (r failedRelationshipRows) Next() bool                { return false }
func (r failedRelationshipRows) Scan(...interface{}) error { panic("unexpected scan") }
func (r failedRelationshipRows) Err() error                { return r.err }

func TestScanRelationshipsPropagatesReadError(t *testing.T) {
	failure := errors.New("read failed")
	if err := ScanRelationships(failedRelationshipRows{failure}, multigraph.NewBuilder(), nil); !errors.Is(err, failure) {
		t.Fatal("read error lost")
	}
}
