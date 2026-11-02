// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package similarity provides node similarity measures based on neighbor sets.
//
// # Custom implementations
//
//   - [Jaccard] — Jaccard index: |N(u)∩N(v)| / |N(u)∪N(v)|.
//   - [JaccardAll] — all node pairs above a similarity threshold.
//   - [Overlap] — overlap coefficient: |N(u)∩N(v)| / min(|N(u)|, |N(v)|).
//   - [Cosine] — cosine similarity on neighbor vectors.
//   - [SimRank] — recursive structural similarity (Jeh & Widom, KDD 2002).
//   - [BipartiteOverlapScorer] — degree-aware overlap and per-bin significance.
//   - [VisitBatchBipartiteOverlap] — bounded candidate generation and top-K
//     degree-aware overlap with shared-entity evidence.
//
// # Link prediction
//
//   - [CommonNeighbors] — count of shared neighbors.
//   - [AdamicAdar] — weighted common neighbors (rare connections count more).
//   - [PreferentialAttachment] — product of degrees.
//   - [PredictLinks] — top-k predicted links using any scoring function.
//
// Neighbor-set measures use [graph.Undirected] and work on the open
// neighborhood (neighbors of a node, excluding the node itself).
package similarity
