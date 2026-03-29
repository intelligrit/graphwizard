// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package flow

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/network"
)

// MaxFlow computes the maximum flow from source to target in a weighted
// directed graph using Dinic's algorithm. The eps parameter is the tolerance
// for considering flow as zero.
//
// Wraps gonum/graph/network.MaxFlowDinic.
func MaxFlow(g graph.WeightedDirected, source, target int64, eps float64) float64 {
	return network.MaxFlowDinic(g, g.Node(source), g.Node(target), eps)
}
