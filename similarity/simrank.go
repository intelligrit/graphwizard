// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"sort"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
)

// SimRank computes the SimRank similarity between all pairs of nodes in a
// directed graph. Two nodes are similar if they are referenced by similar
// nodes. The decay factor C (typically 0.6-0.8) controls how quickly
// similarity decays with distance.
//
// Returns a map from node ID pairs [a, b] (where a <= b) to similarity scores.
// SimRank(a, a) = 1 for all nodes.
//
// Time complexity: O(K * V^2 * d^2) where K = iterations, d = avg in-degree.
//
// Reference: G. Jeh and J. Widom, "SimRank: A Measure of Structural-Context
// Similarity", KDD 2002.
func SimRank(ctx context.Context, g graph.Directed, decay float64, maxIter int) map[[2]int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	n := len(ids)
	if n == 0 {
		return make(map[[2]int64]float64)
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	// Build in-neighbor lists.
	inNeighbors := make([][]int, n)
	for i, id := range ids {
		to := g.To(id)
		for to.Next() {
			j := idx[to.Node().ID()]
			inNeighbors[i] = append(inNeighbors[i], j)
		}
	}

	// Initialize: sim(a,a) = 1, sim(a,b) = 0 for a != b.
	sim := make([][]float64, n)
	for i := range sim {
		sim[i] = make([]float64, n)
		sim[i][i] = 1.0
	}

	for iter := 0; iter < maxIter; iter++ {
		progress.Report(ctx, progress.Progress{Phase: "iterate", Step: iter, Total: maxIter})
		newSim := make([][]float64, n)
		for i := range newSim {
			newSim[i] = make([]float64, n)
			newSim[i][i] = 1.0
		}

		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				inI := inNeighbors[i]
				inJ := inNeighbors[j]
				if len(inI) == 0 || len(inJ) == 0 {
					continue
				}
				sum := 0.0
				for _, a := range inI {
					for _, b := range inJ {
						sum += sim[a][b]
					}
				}
				s := decay * sum / float64(len(inI)*len(inJ))
				newSim[i][j] = s
				newSim[j][i] = s
			}
		}

		sim = newSim
	}

	// Convert to map.
	result := make(map[[2]int64]float64)
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			if sim[i][j] > 0 {
				result[[2]int64{ids[i], ids[j]}] = sim[i][j]
			}
		}
	}
	return result
}
