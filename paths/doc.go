// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package paths provides shortest-path and k-shortest-path algorithms.
//
// # Custom implementations (from academic papers)
//
//   - [YenKShortest] — K shortest loopless paths in a weighted directed graph
//     (Yen 1971). Uses Dijkstra as a subroutine.
//
// # Wrapped from gonum/graph/path
//
//   - [ShortestPath] — single-source shortest path via Dijkstra.
//   - [AllShortestPaths] — all-pairs shortest paths via Dijkstra. The result
//     can be passed to [centrality.BetweennessWeighted], [centrality.Closeness], etc.
//   - [BellmanFord] — single-source shortest path supporting negative weights.
//   - [FloydWarshall] — all-pairs shortest paths supporting negative weights.
//   - [AStar] — A* search with a user-supplied admissible heuristic.
package paths
