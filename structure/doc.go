// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package structure provides structural analysis and combinatorial algorithms.
//
// # Custom implementations (from academic papers)
//
//   - [ClusteringCoefficient] — local clustering coefficient per node
//     (Watts-Strogatz 1998).
//   - [AverageClusteringCoefficient] — mean of all local coefficients.
//   - [SetCover] — greedy minimum set cover, O(ln n) approximation (Chvatal 1979).
//   - [TSP] — travelling salesman heuristic: nearest-neighbor + 2-opt (Croes 1958).
//   - [TriangleCount] — count triangles per node and total.
//   - [Kruskal] — minimum spanning tree/forest via Kruskal's algorithm.
//   - [Prim] — minimum spanning tree via Prim's algorithm.
//
// # Wrapped from gonum
//
//   - [MaximalCliques] — Bron-Kerbosch algorithm for all maximal cliques.
//   - [GraphColoring] — exact vertex coloring via DSatur.
package structure
