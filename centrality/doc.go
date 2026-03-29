// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package centrality provides node and edge centrality measures for graphs.
//
// This package unifies custom implementations and gonum wrappers under a
// single import. All functions accept standard gonum graph interfaces and
// return map[int64]float64 keyed by node ID.
//
// # Custom implementations (from academic papers)
//
//   - [Katz] / [KatzUndirected] — Katz centrality via power iteration.
//   - [Degree] / [InDegree] / [OutDegree] — normalized degree centrality.
//   - [PersonalizedPageRank] / [PersonalizedPageRankUndirected] — PPR relative to a seed node.
//   - [Eccentricity] — max shortest-path distance from each node.
//   - [Diameter] / [Radius] — graph diameter and radius.
//
// # Wrapped from gonum/graph/network
//
//   - [PageRank] / [PageRankSparse] — Google PageRank algorithm.
//   - [Betweenness] / [BetweennessWeighted] — shortest-path betweenness.
//   - [EdgeBetweenness] / [EdgeBetweennessWeighted] — edge betweenness.
//   - [Closeness] — closeness centrality.
//   - [Harmonic] — harmonic centrality.
//   - [HITS] — Kleinberg's hub-authority scores.
package centrality
