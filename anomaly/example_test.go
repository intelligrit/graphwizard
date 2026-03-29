// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly_test

import (
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
