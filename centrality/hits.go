// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/network"
)

// HITSResult holds the hub and authority scores from the HITS algorithm.
type HITSResult struct {
	Hub       map[int64]float64
	Authority map[int64]float64
}

// HITS returns Hyperlink-Induced Topic Search hub and authority scores for
// each node in a directed graph.
//
// Wraps gonum/graph/network.HITS.
func HITS(ctx context.Context, g graph.Directed, tol float64) HITSResult {
	raw := network.HITS(g, tol)
	hub := make(map[int64]float64, len(raw))
	auth := make(map[int64]float64, len(raw))
	for id, ha := range raw {
		hub[id] = ha.Hub
		auth[id] = ha.Authority
	}
	return HITSResult{Hub: hub, Authority: auth}
}
