// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package centrality

import (
	"context"
	"math"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// Eccentricity returns the eccentricity of each node in a graph, keyed by
// node ID. The eccentricity of a node is the maximum shortest-path distance
// from that node to any other reachable node.
//
// Nodes with no outgoing paths have eccentricity 0.
func Eccentricity(ctx context.Context, g graph.Graph) map[int64]float64 {
	allPaths := path.DijkstraAllPaths(g)
	nodes := g.Nodes()
	result := make(map[int64]float64)

	// Collect node IDs to know the total for progress reporting.
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)

	for i, uid := range ids {
		progress.Report(ctx, progress.Progress{Phase: "nodes", Step: i, Total: n})
		u := g.Node(uid)
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
		result[uid] = maxDist
	}

	return result
}

// Diameter returns the diameter of a graph: the maximum eccentricity across
// all nodes. This is the longest shortest path between any two nodes.
//
// Returns 0 for empty graphs or graphs with no edges.
func Diameter(ctx context.Context, g graph.Graph) float64 {
	ecc := Eccentricity(ctx, g)
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
func Radius(ctx context.Context, g graph.Graph) float64 {
	ecc := Eccentricity(ctx, g)
	r := math.Inf(1)
	for _, e := range ecc {
		if e < r {
			r = e
		}
	}
	return r
}
