// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
)

func annualChangesFixture() *Store {
	b := NewBuilder()
	add := func(p, e int64, kind string, year int32, weights map[string]float64) {
		b.AddRelationship(p, e, Attributes{Type: kind, Year: year}, weights)
	}
	// Deliberately unsorted, with a current affiliation and a future record.
	add(1, 10, "prescribed", 2023, map[string]float64{"claims": 30})
	add(1, 11, "prescribed", 2022, map[string]float64{"claims": 0})
	add(2, 13, "prescribed", 2023, map[string]float64{"claims": 0})
	add(1, 10, "prescribed", 2022, map[string]float64{"claims": 10})
	add(1, 11, "prescribed", 2023, nil)
	add(1, 12, "prescribed", 2023, map[string]float64{"claims": 5})
	add(2, 10, "prescribed", 2022, map[string]float64{"claims": 2})
	add(2, 10, "prescribed", 2023, map[string]float64{"claims": 4})
	add(2, 13, "prescribed", 2022, map[string]float64{"claims": 8})
	add(3, 14, "prescribed", 2022, nil)
	add(3, 15, "prescribed", 2023, map[string]float64{"claims": 0})
	add(1, 10, "prescribed", 2024, map[string]float64{"claims": 1000})
	add(1, 99, "affiliated", 0, nil)
	add(1, 20, "payment/Cash", 2023, map[string]float64{"dollars": 2})
	add(1, 20, "payment/Cash", 2024, map[string]float64{"dollars": 12})
	return b.Build()
}

func TestVisitAnnualChangesHistoryAndMissing(t *testing.T) {
	s := annualChangesFixture()
	o := AnnualChangeOptions{Earlier: 2022, Later: 2023, Type: "prescribed", Weight: "claims"}
	var providers []AnnualProviderChange
	var edges []AnnualEdgeChange
	stats, err := s.VisitAnnualChanges(context.Background(), o, func(p AnnualProviderChange) error { providers = append(providers, p); return nil }, func(e AnnualEdgeChange) error { edges = append(edges, e); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if stats.SelectedRelationships != 11 || stats.Edges != 7 || stats.Providers != 3 || stats.NewEdges != 2 || stats.LostEdges != 1 || stats.SortedInput {
		t.Fatalf("stats %+v", stats)
	}
	if providers[0].Provider != 1 || providers[0].EarlierDegree != 2 || providers[0].LaterDegree != 3 || providers[0].NewConnections != 1 || providers[0].HasMetricChange || providers[0].Later.Missing != 1 {
		t.Fatalf("provider 1: %+v", providers[0])
	}
	if math.Abs(providers[0].TurnoverRate-0.2) > 1e-12 || !providers[0].HasPeerTurnoverZ || providers[0].PeerTurnoverCount != 2 {
		t.Fatalf("provider 1 turnover: %+v", providers[0])
	}
	if providers[1].Provider != 2 || !providers[1].HasMetricChange || !providers[1].HasConcentration || math.Abs(providers[1].EarlierHHI-0.68) > 1e-12 || math.Abs(providers[1].LaterHHI-1) > 1e-12 {
		t.Fatalf("provider 2: %+v", providers[1])
	}
	if providers[2].Provider != 3 || providers[2].HasMetricChange || providers[2].NewConnections != 1 || providers[2].LostConnections != 1 {
		t.Fatalf("provider 3: %+v", providers[2])
	}
	var zero, missing, future bool
	for _, e := range edges {
		if e.Type != "prescribed" || e.Entity == 99 || e.Entity == 20 {
			t.Fatalf("historical/type leak: %+v", e)
		}
		if e.Provider == 2 && e.Entity == 13 {
			zero = e.Later.Observed && e.Later.Present && e.Later.Complete && e.Later.Value == 0 && e.HasAdjustedChange
		}
		if e.Provider == 1 && e.Entity == 11 {
			missing = e.Later.Observed && !e.Later.Present && !e.Later.Complete && !e.HasAdjustedChange
		}
		if e.Later.Value == 1000 {
			future = true
		}
	}
	if !zero || !missing || future {
		t.Fatalf("zero=%v missing=%v future=%v", zero, missing, future)
	}
	var again []AnnualEdgeChange
	var againProviders []AnnualProviderChange
	_, err = s.VisitAnnualChanges(context.Background(), o, func(p AnnualProviderChange) error { againProviders = append(againProviders, p); return nil }, func(e AnnualEdgeChange) error { again = append(again, e); return nil })
	if err != nil || !reflect.DeepEqual(edges, again) || !reflect.DeepEqual(providers, againProviders) {
		t.Fatalf("nondeterministic output: %v", err)
	}
}

func TestVisitAnnualChangesPaymentAndErrors(t *testing.T) {
	s := annualChangesFixture()
	o := AnnualChangeOptions{Earlier: 2023, Later: 2024, TypePrefix: "payment/", Weight: "dollars"}
	var got []AnnualEdgeChange
	stats, err := s.VisitAnnualChanges(context.Background(), o, nil, func(e AnnualEdgeChange) error { got = append(got, e); return nil })
	if err != nil || stats.SelectedRelationships != 2 || len(got) != 1 || got[0].Type != "payment/Cash" || got[0].Earlier.Value != 2 || got[0].Later.Value != 12 {
		t.Fatalf("payment: %+v %+v %v", stats, got, err)
	}
	for _, invalid := range []AnnualChangeOptions{
		{Earlier: 0, Later: 2023, Type: "affiliated"},
		{Earlier: 2024, Later: 2023, Type: "prescribed"},
		{Earlier: 2024, Later: 2025, Type: "prescribed", Weight: "claims"},
		{Earlier: 2022, Later: 2023, Type: "prescribed", TypePrefix: "payment/"},
		{Earlier: 2022, Later: 2023, Type: "prescribed", Weight: "unknown"},
	} {
		if _, err := s.VisitAnnualChanges(context.Background(), invalid, nil, nil); err == nil {
			t.Errorf("expected error for %+v", invalid)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.VisitAnnualChanges(ctx, o, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}

func TestVisitAnnualChangesSortedByTypeName(t *testing.T) {
	b := NewBuilder()
	// Type codes are assigned by first use, so numeric code order differs
	// from the fixture's lexicographic type order.
	b.AddRelationship(1, 10, Attributes{Type: "payment/B", Year: 2022}, map[string]float64{"dollars": 1})
	b.AddRelationship(2, 20, Attributes{Type: "payment/A", Year: 2022}, map[string]float64{"dollars": 2})
	b.AddRelationship(2, 20, Attributes{Type: "payment/B", Year: 2023}, map[string]float64{"dollars": 3})
	s := b.Build()
	stats, err := s.VisitAnnualChanges(context.Background(), AnnualChangeOptions{Earlier: 2022, Later: 2023, TypePrefix: "payment/", Weight: "dollars"}, nil, nil)
	if err != nil || !stats.SortedInput {
		t.Fatalf("expected type-name-sorted input: %+v %v", stats, err)
	}
}
