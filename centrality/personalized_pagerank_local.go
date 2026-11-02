// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"errors"
	"math"
	"slices"

	"gonum.org/v1/gonum/graph"
)

// LocalPersonalizedPageRank approximates Personalized PageRank from a weighted
// set of seed nodes on an undirected graph. Seed weights are normalized to sum
// to one. Damping has the same meaning as in PersonalizedPageRankUndirected.
// Epsilon controls the stopping condition: every non-isolated node's remaining
// residual divided by its degree is less than epsilon on return. Smaller
// epsilon gives a closer approximation and explores more of the graph.
//
// The implementation uses the local push method of Andersen, Chung, and Lang,
// with a lazy-walk restart probability of (1-damping)/(1+damping). This
// transformation makes its limiting scores equal to standard PPR with the
// requested damping. Scores are lower bounds until all residual is pushed;
// they are not renormalized. One minus the sum of returned scores estimates
// unpushed probability mass. On highly connected graphs, a small epsilon may
// visit most nodes. The method stores scores, residuals, and degrees only for
// visited nodes, and never constructs whole-graph adjacency.
//
// Damping must be in [0,1), epsilon must be finite and positive, and seed
// weights must be finite and nonnegative with a positive total. Missing seed
// nodes return an error. The graph must not change during the computation.
//
// Reference: R. Andersen, F. Chung, K. Lang, "Local Graph Partitioning using
// PageRank Vectors", FOCS 2006.
func LocalPersonalizedPageRank(ctx context.Context, g graph.Undirected, seeds map[int64]float64, damping, epsilon float64) (map[int64]float64, error) {
	if math.IsNaN(damping) || damping < 0 || damping >= 1 {
		return nil, errors.New("centrality: damping must be in [0,1)")
	}
	if math.IsNaN(epsilon) || math.IsInf(epsilon, 0) || epsilon <= 0 {
		return nil, errors.New("centrality: epsilon must be finite and positive")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(seeds))
	total := 0.0
	for id, weight := range seeds {
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return nil, errors.New("centrality: seed weights must be finite and nonnegative")
		}
		if weight == 0 {
			continue
		}
		if g.Node(id) == nil {
			return nil, errors.New("centrality: seed node is missing from graph")
		}
		ids = append(ids, id)
		total += weight
	}
	if total <= 0 || math.IsInf(total, 0) {
		return nil, errors.New("centrality: seed weights must have a finite positive total")
	}
	slices.Sort(ids)

	alpha := (1 - damping) / (1 + damping)
	residual := make(map[int64]float64, len(ids))
	scores := make(map[int64]float64, len(ids))
	degrees := make(map[int64]int, len(ids))
	queued := make(map[int64]bool, len(ids))
	queue := make([]int64, 0, len(ids))

	degree := func(id int64) (int, error) {
		if d, ok := degrees[id]; ok {
			return d, nil
		}
		d := 0
		neighbors := g.From(id)
		for neighbors.Next() {
			d++
			if d&1023 == 0 {
				if err := ctx.Err(); err != nil {
					return 0, err
				}
			}
		}
		degrees[id] = d
		return d, nil
	}
	addResidual := func(id int64, amount float64) error {
		if amount == 0 && residual[id] == 0 {
			return nil
		}
		residual[id] += amount
		if queued[id] {
			return nil
		}
		d, err := degree(id)
		if err != nil {
			return err
		}
		if d == 0 {
			// An isolated node cannot diffuse its mass to another node.
			scores[id] += residual[id]
			delete(residual, id)
			return nil
		}
		if residual[id]/float64(d) >= epsilon {
			queued[id] = true
			queue = append(queue, id)
		}
		return nil
	}
	for _, id := range ids {
		if err := addResidual(id, seeds[id]/total); err != nil {
			return nil, err
		}
	}

	for head := 0; head < len(queue); head++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		id := queue[head]
		delete(queued, id)
		d := degrees[id]
		mass := residual[id]
		if mass/float64(d) < epsilon {
			continue
		}
		scores[id] += alpha * mass
		residual[id] = (1 - alpha) * mass / 2
		share := residual[id] / float64(d)
		if share > 0 {
			neighbors := g.From(id)
			for seen := 0; neighbors.Next(); seen++ {
				if seen&1023 == 0 {
					if err := ctx.Err(); err != nil {
						return nil, err
					}
				}
				if err := addResidual(neighbors.Node().ID(), share); err != nil {
					return nil, err
				}
			}
		}
		if err := addResidual(id, 0); err != nil {
			return nil, err
		}
		// Discard consumed queue entries so repeated pushes do not retain a
		// history proportional to total work.
		if head > 1024 && head*2 >= len(queue) {
			queue = append([]int64(nil), queue[head+1:]...)
			head = -1
		}
	}
	return scores, nil
}
