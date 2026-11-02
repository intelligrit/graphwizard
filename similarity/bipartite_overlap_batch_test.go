// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/intelligrit/graphwizard/multigraph"
	"gonum.org/v1/gonum/graph/simple"
)

func TestVisitBatchBipartiteOverlap(t *testing.T) {
	g := bipartiteOverlapGraph()
	var got []BatchOverlapMatch
	o := BatchOverlapOptions{MaxBridgeDegree: 2, MaxCandidatesPerLeft: 8, TopK: 2, MinSeedShared: 1, LeftIDs: []int64{1, 0, 0}}
	stats, err := VisitBatchBipartiteOverlap(context.Background(), g, func(id int64) bool { return id < 10 }, o, func(m BatchOverlapMatch) error { got = append(got, m); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if stats.LeftNodes != 2 || stats.MatchedLeftNodes != 2 || stats.Matches != 2 || got[0].A != 0 || got[0].B != 1 || got[1].A != 1 || got[1].B != 0 {
		t.Fatalf("stats=%+v matches=%+v", stats, got)
	}
	m := got[0]
	if m.SeedShared != 2 || m.Shared != 3 || math.Abs(m.SeedRarity-1) > 1e-12 || math.Abs(m.RarityScore-1.25) > 1e-12 || len(m.Evidence) != 3 {
		t.Fatalf("unexpected match %+v", m)
	}
	if m.Evidence[0].ID != 10 || m.Evidence[1].ID != 11 || m.Evidence[2].ID != 15 || m.Evidence[2].Eligible || m.Evidence[2].Degree != 4 {
		t.Fatalf("evidence %+v", m.Evidence)
	}
	var again []BatchOverlapMatch
	_, err = VisitBatchBipartiteOverlap(context.Background(), g, func(id int64) bool { return id < 10 }, o, func(m BatchOverlapMatch) error { again = append(again, m); return nil })
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatalf("nondeterministic output: %v %+v", err, again)
	}
}

func TestVisitBatchBipartiteOverlapBoundsAndErrors(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for _, pair := range [][2]int64{{0, 10}, {1, 10}, {2, 10}, {3, 10}, {0, 11}, {1, 11}} {
		g.SetEdge(g.NewEdge(simple.Node(pair[0]), simple.Node(pair[1])))
	}
	o := BatchOverlapOptions{MaxBridgeDegree: 3, MaxCandidatesPerLeft: 1, TopK: 1, MinSeedShared: 1, LeftIDs: []int64{0}}
	var got []BatchOverlapMatch
	stats, err := VisitBatchBipartiteOverlap(context.Background(), g, func(id int64) bool { return id < 10 }, o, func(m BatchOverlapMatch) error { got = append(got, m); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if stats.Candidates != 1 || stats.DroppedPairEncounters != 0 || len(got) != 1 || got[0].B != 1 {
		t.Fatalf("stats=%+v got=%+v", stats, got)
	}
	o.MaxBridgeDegree = 4
	got = nil
	stats, err = VisitBatchBipartiteOverlap(context.Background(), g, func(id int64) bool { return id < 10 }, o, func(m BatchOverlapMatch) error { got = append(got, m); return nil })
	if err != nil || stats.DroppedPairEncounters == 0 || stats.TruncatedLeftNodes != 1 || len(got) != 1 || !got[0].CandidateSearchTruncated {
		t.Fatalf("expected truncated candidates: %+v %v", stats, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = VisitBatchBipartiteOverlap(ctx, g, func(id int64) bool { return id < 10 }, o, func(BatchOverlapMatch) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	o.LeftIDs = []int64{10}
	_, err = VisitBatchBipartiteOverlap(context.Background(), g, func(id int64) bool { return id < 10 }, o, func(BatchOverlapMatch) error { return nil })
	if err == nil {
		t.Fatal("expected invalid left ID")
	}
}

func TestVisitBatchBipartiteOverlapAnnualProjectionBoundary(t *testing.T) {
	b := multigraph.NewBuilder()
	for _, e := range [][3]int64{{0, 10, 2023}, {1, 10, 2023}, {0, 11, 2023}, {1, 11, 2023}, {0, 12, 2024}, {2, 12, 2024}, {0, 13, 2024}, {2, 13, 2024}} {
		b.AddRelationship(e[0], e[1], multigraph.Attributes{Type: "prescribed", Year: int32(e[2])}, nil)
	}
	b.AddRelationship(0, 20, multigraph.Attributes{Type: "affiliated", Year: 0}, nil)
	b.AddRelationship(3, 20, multigraph.Attributes{Type: "affiliated", Year: 0}, nil)
	s := b.Build()
	for year, want := range map[int32]int64{2023: 1, 2024: 2} {
		g, err := s.ProjectUndirected(multigraph.Projection{Filter: func(r multigraph.Relationship) bool { return r.Type == "prescribed" && r.Year == year }})
		if err != nil {
			t.Fatal(err)
		}
		var matches []BatchOverlapMatch
		_, err = VisitBatchBipartiteOverlap(context.Background(), g, func(id int64) bool { return id < 10 }, BatchOverlapOptions{MaxBridgeDegree: 2, MaxCandidatesPerLeft: 8, TopK: 2, MinSeedShared: 2, LeftIDs: []int64{0}}, func(m BatchOverlapMatch) error { matches = append(matches, m); return nil })
		if err != nil || len(matches) != 1 || matches[0].B != want || matches[0].Shared != 2 {
			t.Fatalf("year %d: %+v %v", year, matches, err)
		}
	}
}
