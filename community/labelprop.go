// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"math/rand"

	"gonum.org/v1/gonum/graph"
)

// LabelPropagation performs community detection using the label propagation
// algorithm, returning a map from node ID to community ID.
//
// Each node starts with a unique label and iteratively adopts the most
// frequent label among its neighbors. The algorithm converges when no node
// changes its label. Ties are broken randomly using the provided RNG.
//
// Label propagation is fast (near-linear time) but non-deterministic: results
// depend on the random iteration order and tie-breaking.
//
// The maxIter parameter caps the number of iterations (use 100 as a default).
//
// Reference: U. Raghavan, R. Albert, S. Kumara, "Near linear time algorithm
// to detect community structures in large-scale networks", Physical Review E,
// 2007.
func LabelPropagation(g graph.Undirected, maxIter int, rng *rand.Rand) map[int64]int64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return make(map[int64]int64)
	}

	// Initialize: each node gets its own label.
	label := make(map[int64]int64, n)
	for _, id := range ids {
		label[id] = id
	}

	for iter := 0; iter < maxIter; iter++ {
		changed := false
		order := rng.Perm(n)

		for _, idx := range order {
			id := ids[idx]
			// Count label frequencies among neighbors.
			freq := make(map[int64]int)
			neighbors := g.From(id)
			hasNeighbors := false
			for neighbors.Next() {
				freq[label[neighbors.Node().ID()]]++
				hasNeighbors = true
			}
			if !hasNeighbors {
				continue
			}

			// Find the maximum frequency.
			maxFreq := 0
			for _, f := range freq {
				if f > maxFreq {
					maxFreq = f
				}
			}

			// Collect all labels with max frequency (ties).
			var candidates []int64
			for l, f := range freq {
				if f == maxFreq {
					candidates = append(candidates, l)
				}
			}

			// Pick one at random.
			newLabel := candidates[rng.Intn(len(candidates))]
			if newLabel != label[id] {
				label[id] = newLabel
				changed = true
			}
		}

		if !changed {
			break
		}
	}

	return label
}
