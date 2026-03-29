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
//
// # Link prediction
//
//   - [CommonNeighbors] — count of shared neighbors.
//   - [AdamicAdar] — weighted common neighbors (rare connections count more).
//   - [PreferentialAttachment] — product of degrees.
//   - [PredictLinks] — top-k predicted links using any scoring function.
//
// All functions accept [graph.Undirected] and work on the open neighborhood
// (neighbors of a node, excluding the node itself).
package similarity
