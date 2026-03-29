// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package flow provides network flow algorithms.
//
// # Wrapped from gonum/graph/network
//
//   - [MaxFlow] — maximum flow via Dinic's algorithm, O(V²E).
//
// # Custom implementations
//
//   - [MinCut] — minimum s-t cut via Edmonds-Karp (Ford-Fulkerson with BFS).
//     Returns the cut partition and weight (equals max flow).
//
// The graph must implement [graph.WeightedDirected]; edge weights are treated
// as capacities.
package flow
