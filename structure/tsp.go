// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"
	"math"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
)

// TSPResult holds the tour and its total weight.
type TSPResult struct {
	Tour   []graph.Node
	Weight float64
}

// TSP finds an approximate solution to the Travelling Salesman Problem on a
// weighted undirected graph using nearest-neighbor heuristic followed by 2-opt
// local search improvement.
//
// The graph should be complete (or at least have edges between all node pairs
// to be visited). If an edge is missing, its weight is treated as +Inf.
//
// This is a heuristic — it does not guarantee an optimal solution but
// typically produces tours within 5-20% of optimal.
//
// Reference: G. Croes, "A Method for Solving Traveling-Salesman Problems",
// Operations Research, 1958. (2-opt)
func TSP(ctx context.Context, g graph.WeightedUndirected) TSPResult {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n <= 1 {
		var tour []graph.Node
		for _, id := range ids {
			tour = append(tour, g.Node(id))
		}
		return TSPResult{Tour: tour, Weight: 0}
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	// Build distance matrix.
	dist := make([][]float64, n)
	for i := range dist {
		dist[i] = make([]float64, n)
		for j := range dist[i] {
			if i == j {
				dist[i][j] = 0
				continue
			}
			w, ok := g.Weight(ids[i], ids[j])
			if !ok {
				dist[i][j] = math.Inf(1)
			} else {
				dist[i][j] = w
			}
		}
	}

	// Nearest-neighbor heuristic starting from each node; keep best.
	progress.Report(ctx, progress.Progress{Phase: "nn-heuristic", Step: 0, Total: 2})
	bestTour := nearestNeighbor(dist, n, 0)
	bestCost := tourCost(dist, bestTour)
	for start := 1; start < n; start++ {
		tour := nearestNeighbor(dist, n, start)
		cost := tourCost(dist, tour)
		if cost < bestCost {
			bestTour = tour
			bestCost = cost
		}
	}

	// 2-opt improvement.
	progress.Report(ctx, progress.Progress{Phase: "2opt", Step: 1, Total: 2})
	bestTour, bestCost = twoOpt(dist, bestTour, bestCost)

	// Convert to nodes.
	result := make([]graph.Node, len(bestTour))
	for i, idx := range bestTour {
		result[i] = g.Node(ids[idx])
	}
	return TSPResult{Tour: result, Weight: bestCost}
}

func nearestNeighbor(dist [][]float64, n, start int) []int {
	visited := make([]bool, n)
	tour := make([]int, 0, n)
	cur := start
	visited[cur] = true
	tour = append(tour, cur)

	for len(tour) < n {
		best := -1
		bestDist := math.Inf(1)
		for j := 0; j < n; j++ {
			if !visited[j] && dist[cur][j] < bestDist {
				best = j
				bestDist = dist[cur][j]
			}
		}
		if best == -1 {
			break
		}
		visited[best] = true
		tour = append(tour, best)
		cur = best
	}

	return tour
}

func tourCost(dist [][]float64, tour []int) float64 {
	cost := 0.0
	for i := 0; i < len(tour)-1; i++ {
		cost += dist[tour[i]][tour[i+1]]
	}
	// Return to start.
	if len(tour) > 1 {
		cost += dist[tour[len(tour)-1]][tour[0]]
	}
	return cost
}

func twoOpt(dist [][]float64, tour []int, bestCost float64) ([]int, float64) {
	n := len(tour)
	improved := true
	for improved {
		improved = false
		for i := 0; i < n-1; i++ {
			for j := i + 2; j < n; j++ {
				// Cost of removing edges (i, i+1) and (j, j+1 mod n)
				// and adding (i, j) and (i+1, j+1 mod n).
				jNext := (j + 1) % n
				oldDist := dist[tour[i]][tour[i+1]] + dist[tour[j]][tour[jNext]]
				newDist := dist[tour[i]][tour[j]] + dist[tour[i+1]][tour[jNext]]
				if newDist < oldDist-1e-10 {
					// Reverse segment [i+1 .. j].
					for l, r := i+1, j; l < r; l, r = l+1, r-1 {
						tour[l], tour[r] = tour[r], tour[l]
					}
					bestCost += newDist - oldDist
					improved = true
				}
			}
		}
	}
	return tour, bestCost
}
