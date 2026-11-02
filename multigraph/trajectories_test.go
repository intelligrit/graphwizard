// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
)

func trajectoryFixture() *Store {
	b := NewBuilder()
	series := map[int64][]float64{
		1: {10, 10, 10, 100, 100, 100},
		2: {10, 10, 10, 10, 10, 10},
		3: {20, 20, 20, 20, 20, 20},
		4: {10, 10, 10, 100, 10, 10},
		5: {10, 10, 10, 10, 10, 10},
	}
	for p, values := range series {
		for i, v := range values {
			b.AddRelationship(p, 100+p, Attributes{Type: "prescribed", Year: int32(2018 + i)}, map[string]float64{"claims": v})
		}
	}
	for i := 0; i < 6; i++ {
		var weights map[string]float64 = map[string]float64{"claims": 0}
		if i == 2 {
			weights = nil
		}
		b.AddRelationship(6, 106, Attributes{Type: "prescribed", Year: int32(2018 + i)}, weights)
		if i != 2 {
			b.AddRelationship(7, 107, Attributes{Type: "prescribed", Year: int32(2018 + i)}, map[string]float64{"claims": 0})
		}
	}
	b.AddRelationship(1, 101, Attributes{Type: "prescribed", Year: 2024}, map[string]float64{"claims": 10000})
	b.AddRelationship(1, 200, Attributes{Type: "affiliated", Year: 0}, nil)
	return b.Build()
}

func TestVisitAnnualTrajectoriesChangePointAndMissing(t *testing.T) {
	s := trajectoryFixture()
	o := TrajectoryOptions{FirstYear: 2018, LastYear: 2023, Type: "prescribed", Weight: "claims", MinSegmentYears: 2, MinPeerCount: 3}
	var got []ProviderTrajectory
	stats, err := s.VisitAnnualTrajectories(context.Background(), o, func(p ProviderTrajectory) error { got = append(got, p); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if stats.Years != 6 || stats.Providers != 7 || stats.CompleteProviders != 5 || stats.ScoredProviders != 5 || len(got) != 7 {
		t.Fatalf("stats=%+v got=%+v", stats, got)
	}
	if got[0].Provider != 1 || !got[0].HasChangePoint || got[0].BreakYear != 2021 || got[0].Score <= got[3].Score || got[0].Years[5].Value != 100 {
		t.Fatalf("persistent change: %+v", got[0])
	}
	if got[5].Provider != 6 || got[5].HasChangePoint || !got[5].Years[0].Present || got[5].Years[0].Value != 0 || got[5].Years[2].Present || got[5].Years[2].Complete || got[5].Years[2].Missing != 1 {
		t.Fatalf("missing versus zero: %+v", got[5])
	}
	if got[6].Provider != 7 || got[6].Years[2].Observed || got[6].HasChangePoint {
		t.Fatalf("absent year: %+v", got[6])
	}
	if math.IsNaN(got[0].Score) || math.IsInf(got[0].Score, 0) {
		t.Fatal("nonfinite score")
	}
	var again []ProviderTrajectory
	stats2, err := s.VisitAnnualTrajectories(context.Background(), o, func(p ProviderTrajectory) error { again = append(again, p); return nil })
	if err != nil || stats != stats2 || !reflect.DeepEqual(got, again) {
		t.Fatalf("nondeterministic: %+v %v", stats2, err)
	}
}

func TestVisitAnnualTrajectoriesHistoricalBoundaryAndErrors(t *testing.T) {
	s := trajectoryFixture()
	for _, o := range []TrajectoryOptions{
		{FirstYear: 2017, LastYear: 2023, Type: "prescribed", Weight: "claims"},
		{FirstYear: 2018, LastYear: 2024, Type: "affiliated"},
		{FirstYear: 2018, LastYear: 2023, Type: "prescribed", TypePrefix: "payment/"},
		{FirstYear: 2018, LastYear: 2023, Type: "prescribed", Weight: "absent"},
		{FirstYear: 2018, LastYear: 2023, Type: "prescribed", MinSegmentYears: 4},
	} {
		if _, err := s.VisitAnnualTrajectories(context.Background(), o, nil); err == nil {
			t.Errorf("expected error for %+v", o)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.VisitAnnualTrajectories(ctx, TrajectoryOptions{FirstYear: 2018, LastYear: 2023, Type: "prescribed", Weight: "claims"}, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
}
