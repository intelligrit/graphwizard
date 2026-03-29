// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package traverse

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
)

// BFS performs a breadth-first search from the given source node, returning
// all reachable node IDs in BFS order.
//
// Wraps gonum/graph/traverse.BreadthFirst.
func BFS(g graph.Graph, source int64) []int64 {
	var visited []int64
	w := traverse.BreadthFirst{
		Visit: func(n graph.Node) {
			visited = append(visited, n.ID())
		},
	}
	w.Walk(g, g.Node(source), nil)
	return visited
}

// DFS performs a depth-first search from the given source node, returning
// all reachable node IDs in DFS order.
//
// Wraps gonum/graph/traverse.DepthFirst.
func DFS(g graph.Graph, source int64) []int64 {
	var visited []int64
	w := traverse.DepthFirst{
		Visit: func(n graph.Node) {
			visited = append(visited, n.ID())
		},
	}
	w.Walk(g, g.Node(source), nil)
	return visited
}

// BFSPath returns the shortest unweighted path from source to target as a
// slice of node IDs. Returns nil if no path exists.
func BFSPath(g graph.Graph, source, target int64) []int64 {
	parent := make(map[int64]int64)
	found := false

	w := traverse.BreadthFirst{
		Visit: func(n graph.Node) {},
		Traverse: func(e graph.Edge) bool {
			if _, seen := parent[e.To().ID()]; seen {
				return false
			}
			parent[e.To().ID()] = e.From().ID()
			if e.To().ID() == target {
				found = true
			}
			return !found
		},
	}
	parent[source] = -1
	w.Walk(g, g.Node(source), func(n graph.Node, d int) bool {
		return found
	})

	if !found && source != target {
		return nil
	}

	// Reconstruct path.
	var path []int64
	for at := target; at != -1; at = parent[at] {
		path = append(path, at)
	}
	// Reverse.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
