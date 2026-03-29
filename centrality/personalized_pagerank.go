// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"

	"gonum.org/v1/gonum/graph"
)

// PersonalizedPageRank returns the Personalized PageRank (PPR) scores relative
// to a seed node, keyed by node ID.
//
// PPR measures the relevance of every node to the seed: the random walker
// restarts at the seed node with probability (1 - damping) at each step,
// rather than jumping to a uniformly random node as in standard PageRank.
//
// The damping factor (typically 0.85) controls how far the walk explores.
// Lower values concentrate scores near the seed.
//
// Reference: T. Haveliwala, "Topic-Sensitive PageRank", WWW 2002.
func PersonalizedPageRank(g graph.Directed, seed int64, damping, tol float64, maxIter int) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return nil
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	seedIdx, ok := idx[seed]
	if !ok {
		return nil
	}

	// Build out-degree and adjacency.
	outDeg := make([]int, n)
	adj := make([][]int, n) // adj[i] = list of nodes i points to
	for i, id := range ids {
		to := g.From(id)
		for to.Next() {
			j, ok := idx[to.Node().ID()]
			if ok {
				adj[i] = append(adj[i], j)
				outDeg[i]++
			}
		}
	}

	// Power iteration.
	score := make([]float64, n)
	score[seedIdx] = 1.0

	for iter := 0; iter < maxIter; iter++ {
		newScore := make([]float64, n)
		// Teleport component: restart at seed.
		newScore[seedIdx] = 1.0 - damping

		// Diffusion component. Dangling node mass is redirected to the seed.
		danglingMass := 0.0
		for i := 0; i < n; i++ {
			if score[i] == 0 {
				continue
			}
			if outDeg[i] == 0 {
				danglingMass += score[i]
				continue
			}
			share := damping * score[i] / float64(outDeg[i])
			for _, j := range adj[i] {
				newScore[j] += share
			}
		}
		newScore[seedIdx] += damping * danglingMass

		// Check convergence.
		diff := 0.0
		for i := 0; i < n; i++ {
			diff += math.Abs(newScore[i] - score[i])
		}
		score = newScore
		if diff < tol {
			break
		}
	}

	result := make(map[int64]float64, n)
	for i, id := range ids {
		result[id] = score[i]
	}
	return result
}

// PersonalizedPageRankUndirected runs PPR on an undirected graph by treating
// each undirected edge as two directed edges.
func PersonalizedPageRankUndirected(g graph.Undirected, seed int64, damping, tol float64, maxIter int) map[int64]float64 {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 {
		return nil
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	seedIdx, ok := idx[seed]
	if !ok {
		return nil
	}

	adj := make([][]int, n)
	deg := make([]int, n)
	for i, id := range ids {
		neighbors := g.From(id)
		for neighbors.Next() {
			j, ok := idx[neighbors.Node().ID()]
			if ok {
				adj[i] = append(adj[i], j)
				deg[i]++
			}
		}
	}

	score := make([]float64, n)
	score[seedIdx] = 1.0

	for iter := 0; iter < maxIter; iter++ {
		newScore := make([]float64, n)
		newScore[seedIdx] = 1.0 - damping

		danglingMass := 0.0
		for i := 0; i < n; i++ {
			if score[i] == 0 {
				continue
			}
			if deg[i] == 0 {
				danglingMass += score[i]
				continue
			}
			share := damping * score[i] / float64(deg[i])
			for _, j := range adj[i] {
				newScore[j] += share
			}
		}
		newScore[seedIdx] += damping * danglingMass

		diff := 0.0
		for i := 0; i < n; i++ {
			diff += math.Abs(newScore[i] - score[i])
		}
		score = newScore
		if diff < tol {
			break
		}
	}

	result := make(map[int64]float64, n)
	for i, id := range ids {
		result[id] = score[i]
	}
	return result
}
