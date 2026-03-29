// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package flow provides network flow algorithms.
//
// # Wrapped from gonum/graph/network
//
//   - [MaxFlow] — maximum flow via Dinic's algorithm, O(V²E).
//
// The graph must implement [graph.WeightedDirected]; edge weights are treated
// as capacities.
package flow
