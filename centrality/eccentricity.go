// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"math"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// Eccentricity returns the eccentricity of each node in a graph, keyed by
// node ID. The eccentricity of a node is the maximum shortest-path distance
// from that node to any other reachable node.
//
// Nodes with no outgoing paths have eccentricity 0.
func Eccentricity(g graph.Graph) map[int64]float64 {
	allPaths := path.DijkstraAllPaths(g)
	nodes := g.Nodes()
	result := make(map[int64]float64)

	for nodes.Next() {
		u := nodes.Node()
		maxDist := 0.0
		inner := g.Nodes()
		for inner.Next() {
			v := inner.Node()
			if u.ID() == v.ID() {
				continue
			}
			w := allPaths.Weight(u.ID(), v.ID())
			if !math.IsInf(w, 1) && w > maxDist {
				maxDist = w
			}
		}
		result[u.ID()] = maxDist
	}

	return result
}

// Diameter returns the diameter of a graph: the maximum eccentricity across
// all nodes. This is the longest shortest path between any two nodes.
//
// Returns 0 for empty graphs or graphs with no edges.
func Diameter(g graph.Graph) float64 {
	ecc := Eccentricity(g)
	d := 0.0
	for _, e := range ecc {
		if e > d {
			d = e
		}
	}
	return d
}

// Radius returns the radius of a graph: the minimum eccentricity across all
// nodes. Returns +Inf for empty graphs.
func Radius(g graph.Graph) float64 {
	ecc := Eccentricity(g)
	r := math.Inf(1)
	for _, e := range ecc {
		if e < r {
			r = e
		}
	}
	return r
}
