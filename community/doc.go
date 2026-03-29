// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package community provides community detection algorithms for graphs.
//
// # Custom implementations (from academic papers)
//
//   - [Leiden] — Leiden community detection (Traag et al. 2019). Guarantees
//     well-connected communities via a refinement phase that improves on Louvain.
//   - [LabelPropagation] — fast near-linear community detection via iterative
//     label adoption (Raghavan et al. 2007).
//
// # Wrapped from gonum/graph/community
//
//   - [Louvain] — Louvain modularity optimization (Blondel et al. 2008).
//   - [LouvainQ] — compute modularity Q score for a given partition.
//
// Both algorithms accept a resolution parameter: higher values produce more,
// smaller communities. Use 1.0 for standard modularity optimization.
package community
