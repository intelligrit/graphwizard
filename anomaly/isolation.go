// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly

import (
	"math"

	"gonum.org/v1/gonum/graph"
)

// IsolationScore computes an anomaly score for each node in an undirected
// graph. Higher scores indicate more structurally unusual nodes.
//
// The score combines three signals:
//   - Degree z-score: how unusual the node's degree is relative to the graph.
//   - Clustering coefficient deviation: how different the node's local
//     clustering coefficient is from its neighbors' mean.
//   - Neighbor degree variance: how much the degrees of the node's neighbors
//     vary (high variance suggests an unusual bridging role).
//
// Each component is normalized to [0,1] and the final score is their mean.
func IsolationScore(g graph.Undirected) map[int64]float64 {
	zscores := DegreeZScore(g)
	cc := localClusteringCoefficients(g)
	ndv := neighborDegreeVariance(g)

	// Normalize each component.
	normZ := normalizeAbsolute(zscores)
	normCC := normalizeClusteringDeviation(g, cc)
	normNDV := normalizeMap(ndv)

	result := make(map[int64]float64, len(zscores))
	for id := range zscores {
		result[id] = (normZ[id] + normCC[id] + normNDV[id]) / 3.0
	}
	return result
}

// localClusteringCoefficients computes the local clustering coefficient for
// each node. Reimplemented here to avoid circular imports with structure/.
func localClusteringCoefficients(g graph.Undirected) map[int64]float64 {
	result := make(map[int64]float64)
	nodes := g.Nodes()
	for nodes.Next() {
		id := nodes.Node().ID()
		neighbors := neighborIDs(g, id)
		k := len(neighbors)
		if k < 2 {
			result[id] = 0
			continue
		}
		edges := 0
		for i := 0; i < len(neighbors); i++ {
			for j := i + 1; j < len(neighbors); j++ {
				if g.HasEdgeBetween(neighbors[i], neighbors[j]) {
					edges++
				}
			}
		}
		result[id] = 2.0 * float64(edges) / float64(k*(k-1))
	}
	return result
}

// neighborDegreeVariance computes the variance of neighbor degrees for each
// node.
func neighborDegreeVariance(g graph.Undirected) map[int64]float64 {
	result := make(map[int64]float64)
	nodes := g.Nodes()
	for nodes.Next() {
		id := nodes.Node().ID()
		neighbors := neighborIDs(g, id)
		if len(neighbors) == 0 {
			result[id] = 0
			continue
		}
		sum := 0.0
		degs := make([]float64, len(neighbors))
		for i, nid := range neighbors {
			d := float64(nodeDegree(g, nid))
			degs[i] = d
			sum += d
		}
		mean := sum / float64(len(neighbors))
		varSum := 0.0
		for _, d := range degs {
			diff := d - mean
			varSum += diff * diff
		}
		result[id] = varSum / float64(len(neighbors))
	}
	return result
}

// normalizeAbsolute normalizes a map by dividing absolute values by the max.
func normalizeAbsolute(m map[int64]float64) map[int64]float64 {
	maxVal := 0.0
	for _, v := range m {
		if math.Abs(v) > maxVal {
			maxVal = math.Abs(v)
		}
	}
	result := make(map[int64]float64, len(m))
	if maxVal == 0 {
		for id := range m {
			result[id] = 0
		}
		return result
	}
	for id, v := range m {
		result[id] = math.Abs(v) / maxVal
	}
	return result
}

// normalizeClusteringDeviation computes how much each node's clustering
// coefficient deviates from its neighborhood mean, then normalizes to [0,1].
func normalizeClusteringDeviation(g graph.Undirected, cc map[int64]float64) map[int64]float64 {
	dev := make(map[int64]float64, len(cc))
	for id, c := range cc {
		neighbors := neighborIDs(g, id)
		if len(neighbors) == 0 {
			dev[id] = 0
			continue
		}
		sum := 0.0
		for _, nid := range neighbors {
			sum += cc[nid]
		}
		neighborMean := sum / float64(len(neighbors))
		dev[id] = math.Abs(c - neighborMean)
	}
	return normalizeMap(dev)
}

// normalizeMap normalizes values to [0,1] by dividing by max.
func normalizeMap(m map[int64]float64) map[int64]float64 {
	maxVal := 0.0
	for _, v := range m {
		if v > maxVal {
			maxVal = v
		}
	}
	result := make(map[int64]float64, len(m))
	if maxVal == 0 {
		for id := range m {
			result[id] = 0
		}
		return result
	}
	for id, v := range m {
		result[id] = v / maxVal
	}
	return result
}

// neighborIDs returns the IDs of all neighbors of a node.
func neighborIDs(g graph.Undirected, id int64) []int64 {
	var result []int64
	it := g.From(id)
	for it.Next() {
		result = append(result, it.Node().ID())
	}
	return result
}
