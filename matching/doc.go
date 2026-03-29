// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package matching provides graph matching algorithms.
//
// # Custom implementations (from academic papers)
//
//   - [HopcroftKarp] — maximum cardinality bipartite matching in O(E√V) time
//     (Hopcroft-Karp 1973). The caller supplies the left partition; all other
//     nodes are assumed to be in the right partition.
//
// The result is a [Matching] (map from left node IDs to matched right node IDs).
package matching
