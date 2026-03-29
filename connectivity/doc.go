// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package connectivity provides algorithms for analyzing graph connectivity.
//
// # Custom implementations (from academic papers)
//
//   - [Bridges] — find bridge edges whose removal disconnects the graph (Tarjan 1974).
//   - [BiconnectedComponents] — find maximal biconnected subgraphs (Hopcroft-Tarjan 1973).
//   - [ArticulationPoints] — find cut vertices whose removal disconnects the graph.
//   - [UnionFind] — disjoint-set data structure with path compression and union by rank.
//
// # Wrapped from gonum/graph/topo
//
//   - [ConnectedComponents] — weakly connected components of an undirected graph.
//   - [StronglyConnectedComponents] — Tarjan's SCC for directed graphs.
//   - [DirectedCycles] — all elementary cycles in a directed graph (Johnson's algorithm).
//   - [UndirectedCycles] — cycle basis for undirected graphs (Paton's algorithm).
//   - [KCore] — k-core subgraph extraction.
//   - [DegeneracyOrdering] — degeneracy ordering and core layers.
//
// # DAG condensation
//
//   - [Condensation] — collapse SCCs into a DAG.
package connectivity
