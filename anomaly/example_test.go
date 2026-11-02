// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly_test

import (
	"context"
	"fmt"

	"github.com/intelligrit/graphwizard/anomaly"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleDegreeZScore() {
	g := simple.NewUndirectedGraph()
	// Star: center 0 has degree 3, leaves have degree 1.
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))

	scores := anomaly.DegreeZScore(g)
	fmt.Printf("center z-score > 0: %v\n", scores[0] > 0)
	// Output: center z-score > 0: true
}

func ExampleFindNearBicliqueBlocks() {
	g := simple.NewUndirectedGraph()
	for _, provider := range []int64{1, 2, 3} {
		for _, drug := range []int64{10, 11} {
			g.SetEdge(g.NewEdge(simple.Node(provider), simple.Node(drug)))
		}
	}
	blocks, _, err := anomaly.FindNearBicliqueBlocks(context.Background(), g,
		func(id int64) bool { return id < 10 },
		[]anomaly.NearBicliqueSeed{{A: 1, B: 2}},
		anomaly.DefaultNearBicliqueOptions())
	if err != nil {
		panic(err)
	}
	fmt.Printf("providers=%d drugs=%d density=%.1f\n",
		len(blocks[0].ProviderIDs), len(blocks[0].Bridges), blocks[0].Density)
	// Output: providers=3 drugs=2 density=1.0
}

func ExampleIsolationScore() {
	g := simple.NewUndirectedGraph()
	// Star: center 0 connects to 1,2,3; leaf 4 hangs off 3.
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(4)))

	scores := anomaly.IsolationScore(g)
	fmt.Printf("scores computed: %d\n", len(scores))
	// Output: scores computed: 5
}

func ExampleStructuralOutliers() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(4)))

	outliers := anomaly.StructuralOutliers(g, 1)
	fmt.Printf("top outlier count: %d\n", len(outliers))
	// Output: top outlier count: 1
}
