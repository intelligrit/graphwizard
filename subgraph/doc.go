// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package subgraph provides utilities for extracting induced subgraphs from
// larger graphs. It supports N-hop neighborhood extraction (BFS from a center
// node) and predicate-based node filtering.
//
// All functions accept gonum/graph interfaces and return new simple graphs
// containing copies of the selected nodes and edges.
package subgraph
