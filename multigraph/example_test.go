// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package multigraph_test

import (
	"context"
	"fmt"

	"github.com/intelligrit/graphwizard/densegraph"
	"github.com/intelligrit/graphwizard/multigraph"
)

func ExampleStore_ProjectUndirected() {
	b := multigraph.NewBuilder()
	b.AddRelationship(1, 2, multigraph.Attributes{Type: "prescribed", Year: 2022},
		map[string]float64{"claims": 10, "dollars": 100})
	b.AddRelationship(1, 2, multigraph.Attributes{Type: "prescribed", Year: 2023},
		map[string]float64{"claims": 15, "dollars": 180})
	store := b.Build(multigraph.BlockWeights)

	annual, err := store.ProjectUndirected(multigraph.Projection{
		Weight: "claims",
		Filter: func(r multigraph.Relationship) bool { return r.Type == "prescribed" && r.Year == 2023 },
	}, densegraph.PackNeighbors24)
	if err != nil {
		panic(err)
	}
	w, _ := annual.Weight(1, 2)
	fmt.Println("2023 claims:", w)

	allYears, err := store.ProjectUndirected(multigraph.Projection{
		Weight: "dollars", Combine: multigraph.Sum,
		Filter: func(r multigraph.Relationship) bool { return r.Type == "prescribed" },
	})
	if err != nil {
		panic(err)
	}
	w, _ = allYears.Weight(1, 2)
	fmt.Println("All-year dollars:", w)
	fmt.Println("Preserved relationships:", store.NumRelationships())
	// Output:
	// 2023 claims: 15
	// All-year dollars: 280
	// Preserved relationships: 2
}

func ExampleStore_VisitAnnualTrajectories() {
	b := multigraph.NewBuilder()
	for _, provider := range []int64{1, 2, 3} {
		for year := int32(2020); year <= 2023; year++ {
			claims := 10.0
			if provider == 1 && year >= 2022 {
				claims = 100
			}
			b.AddRelationship(provider, 100+provider,
				multigraph.Attributes{Type: "prescribed", Year: year},
				map[string]float64{"claims": claims})
		}
	}
	store := b.Build()
	stats, err := store.VisitAnnualTrajectories(context.Background(),
		multigraph.TrajectoryOptions{FirstYear: 2020, LastYear: 2023,
			Type: "prescribed", Weight: "claims", MinPeerCount: 2},
		func(p multigraph.ProviderTrajectory) error {
			if p.Provider == 1 {
				fmt.Println("break year:", p.BreakYear)
			}
			return nil
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("complete providers:", stats.CompleteProviders)
	// Output:
	// break year: 2022
	// complete providers: 3
}
