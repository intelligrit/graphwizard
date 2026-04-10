// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math/rand"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
)

// InfluenceMaximization returns the k most influential seed nodes in an
// undirected graph using the CELF (Cost-Effective Lazy Forward) greedy
// algorithm with Independent Cascade simulation.
//
// The probability parameter is the cascade probability along each edge
// (typically 0.01-0.1). The simulations parameter controls the number of
// Monte Carlo runs per candidate evaluation (higher = more accurate but
// slower; 100-1000 is typical).
//
// Returns the seed node IDs in order of selection and the estimated total
// influence (expected number of activated nodes).
//
// Reference: J. Leskovec et al., "Cost-effective Outbreak Detection in
// Networks", KDD 2007 (CELF optimization). D. Kempe, J. Kleinberg,
// E. Tardos, "Maximizing the Spread of Influence through a Social Network",
// KDD 2003 (influence maximization).
func InfluenceMaximization(ctx context.Context, g graph.Undirected, k int, probability float64, simulations int, rng *rand.Rand) (seeds []int64, influence float64) {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 || k <= 0 {
		return nil, 0
	}
	if k > n {
		k = n
	}

	adj := make(map[int64][]int64)
	for _, id := range ids {
		it := g.From(id)
		for it.Next() {
			adj[id] = append(adj[id], it.Node().ID())
		}
	}

	seedSet := make(map[int64]bool)
	totalSpread := 0.0

	// CELF: evaluate all nodes first, then lazily re-evaluate.

	// Initial evaluation.
	cands := make([]candidate, n)
	for i, id := range ids {
		spread := simulateSpread(adj, map[int64]bool{id: true}, probability, simulations, rng)
		cands[i] = candidate{id: id, marginal: spread, round: 0}
	}

	for round := 0; round < k; round++ {
		progress.Report(ctx, progress.Progress{Phase: "select", Step: round, Total: k})
		// Sort by marginal gain descending.
		sortCandidates(cands)

		// CELF: re-evaluate top candidate if stale.
		for {
			top := &cands[0]
			if top.round == round {
				// Fresh evaluation — select this node.
				break
			}
			// Re-evaluate marginal gain.
			seedSet[top.id] = true
			spreadWith := simulateSpread(adj, seedSet, probability, simulations, rng)
			delete(seedSet, top.id)
			top.marginal = spreadWith - totalSpread
			top.round = round
			sortCandidates(cands)
		}

		// Select top node.
		selected := cands[0]
		seedSet[selected.id] = true
		seeds = append(seeds, selected.id)
		totalSpread += selected.marginal
		cands = cands[1:]
	}

	return seeds, totalSpread
}

func simulateSpread(adj map[int64][]int64, seeds map[int64]bool, prob float64, sims int, rng *rand.Rand) float64 {
	total := 0
	for s := 0; s < sims; s++ {
		activated := make(map[int64]bool)
		var queue []int64
		for id := range seeds {
			activated[id] = true
			queue = append(queue, id)
		}
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]
			for _, neighbor := range adj[node] {
				if !activated[neighbor] && rng.Float64() < prob {
					activated[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
		total += len(activated)
	}
	return float64(total) / float64(sims)
}

type candidate struct {
	id       int64
	marginal float64
	round    int
}

func sortCandidates(cands []candidate) {
	// Simple insertion sort — usually nearly sorted after first round.
	for i := 1; i < len(cands); i++ {
		j := i
		for j > 0 && cands[j].marginal > cands[j-1].marginal {
			cands[j], cands[j-1] = cands[j-1], cands[j]
			j--
		}
	}
}
