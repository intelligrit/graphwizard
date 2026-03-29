// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package subgraph

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// NHopNeighborhood extracts a directed subgraph containing all nodes reachable
// from center within the given number of hops, plus all edges between those
// nodes that exist in the original graph.
//
// The graph g must implement graph.Directed. If center does not exist in g,
// an empty graph is returned.
func NHopNeighborhood(g graph.Directed, center int64, hops int) *simple.DirectedGraph {
	result := simple.NewDirectedGraph()

	if g.Node(center) == nil {
		return result
	}

	// BFS to collect all reachable nodes within hops.
	visited := map[int64]bool{center: true}
	frontier := []int64{center}

	for h := 0; h < hops && len(frontier) > 0; h++ {
		var next []int64
		for _, id := range frontier {
			// Follow outgoing edges.
			it := g.From(id)
			for it.Next() {
				nid := it.Node().ID()
				if !visited[nid] {
					visited[nid] = true
					next = append(next, nid)
				}
			}
			// Follow incoming edges.
			it = g.To(id)
			for it.Next() {
				nid := it.Node().ID()
				if !visited[nid] {
					visited[nid] = true
					next = append(next, nid)
				}
			}
		}
		frontier = next
	}

	// Add all visited nodes.
	for id := range visited {
		result.AddNode(simple.Node(id))
	}

	// Add all edges between visited nodes.
	for id := range visited {
		it := g.From(id)
		for it.Next() {
			nid := it.Node().ID()
			if visited[nid] {
				result.SetEdge(result.NewEdge(simple.Node(id), simple.Node(nid)))
			}
		}
	}

	return result
}

// NHopNeighborhoodUndirected extracts an undirected subgraph containing all
// nodes reachable from center within the given number of hops, plus all edges
// between those nodes that exist in the original graph.
//
// If center does not exist in g, an empty graph is returned.
func NHopNeighborhoodUndirected(g graph.Undirected, center int64, hops int) *simple.UndirectedGraph {
	result := simple.NewUndirectedGraph()

	if g.Node(center) == nil {
		return result
	}

	// BFS to collect all reachable nodes within hops.
	visited := map[int64]bool{center: true}
	frontier := []int64{center}

	for h := 0; h < hops && len(frontier) > 0; h++ {
		var next []int64
		for _, id := range frontier {
			it := g.From(id)
			for it.Next() {
				nid := it.Node().ID()
				if !visited[nid] {
					visited[nid] = true
					next = append(next, nid)
				}
			}
		}
		frontier = next
	}

	// Add all visited nodes.
	for id := range visited {
		result.AddNode(simple.Node(id))
	}

	// Add all edges between visited nodes (avoid duplicates by using u < v).
	for id := range visited {
		it := g.From(id)
		for it.Next() {
			nid := it.Node().ID()
			if visited[nid] && id < nid {
				result.SetEdge(result.NewEdge(simple.Node(id), simple.Node(nid)))
			}
		}
	}

	return result
}
