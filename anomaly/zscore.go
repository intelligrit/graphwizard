// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly

import (
	"math"

	"gonum.org/v1/gonum/graph"
)

// DegreeZScore returns the z-score of each node's degree relative to the
// graph-wide mean and standard deviation. Nodes with unusually high or low
// degree will have large absolute z-scores.
//
// For graphs with fewer than 2 nodes or zero degree variance, all z-scores
// are 0.
func DegreeZScore(g graph.Undirected) map[int64]float64 {
	degrees := make(map[int64]int)
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		id := nodes.Node().ID()
		ids = append(ids, id)
		deg := 0
		it := g.From(id)
		for it.Next() {
			deg++
		}
		degrees[id] = deg
	}

	n := len(ids)
	result := make(map[int64]float64, n)
	if n < 2 {
		for _, id := range ids {
			result[id] = 0
		}
		return result
	}

	// Compute mean and standard deviation.
	sum := 0.0
	for _, id := range ids {
		sum += float64(degrees[id])
	}
	mean := sum / float64(n)

	varSum := 0.0
	for _, id := range ids {
		d := float64(degrees[id]) - mean
		varSum += d * d
	}
	stddev := math.Sqrt(varSum / float64(n))

	if stddev == 0 {
		for _, id := range ids {
			result[id] = 0
		}
		return result
	}

	for _, id := range ids {
		result[id] = (float64(degrees[id]) - mean) / stddev
	}
	return result
}

// nodeDegree returns the degree of a node in an undirected graph.
func nodeDegree(g graph.Undirected, id int64) int {
	deg := 0
	it := g.From(id)
	for it.Next() {
		deg++
	}
	return deg
}
