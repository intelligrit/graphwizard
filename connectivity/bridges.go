// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"

	"gonum.org/v1/gonum/graph"
)

// Bridge represents an edge whose removal disconnects the graph.
type Bridge struct {
	From, To graph.Node
}

// Bridges returns all bridge edges in an undirected graph using Tarjan's
// bridge-finding algorithm.
//
// A bridge is an edge whose removal increases the number of connected
// components. The algorithm runs in O(V + E) time using a single DFS pass,
// tracking discovery times and low-link values.
//
// Reference: R. Tarjan, "A Note on Finding the Bridges of a Graph",
// Information Processing Letters, 1974.
func Bridges(ctx context.Context, g graph.Undirected) []Bridge {
	disc := make(map[int64]int)
	low := make(map[int64]int)
	visited := make(map[int64]bool)
	var bridges []Bridge
	timer := 0

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		if !visited[n.ID()] {
			bridgeDFS(g, n, nil, visited, disc, low, &timer, &bridges)
		}
	}

	return bridges
}

// bridgeFrame is a stack frame for the iterative bridge-finding DFS.
type bridgeFrame struct {
	u        graph.Node
	parent   graph.Node
	neighbors []graph.Node
	idx      int // index into neighbors slice
}

func bridgeDFS(g graph.Undirected, root graph.Node, parent graph.Node, visited map[int64]bool, disc, low map[int64]int, timer *int, bridges *[]Bridge) {
	// Iterative DFS using an explicit stack to avoid stack overflow on deep graphs.
	stack := []bridgeFrame{{u: root, parent: parent}}
	visited[root.ID()] = true
	disc[root.ID()] = *timer
	low[root.ID()] = *timer
	*timer++

	// Pre-collect neighbors for root.
	stack[0].neighbors = collectNeighbors(g, root.ID())

	for len(stack) > 0 {
		top := &stack[len(stack)-1]

		if top.idx < len(top.neighbors) {
			v := top.neighbors[top.idx]
			top.idx++

			if top.parent != nil && v.ID() == top.parent.ID() {
				continue
			}

			if !visited[v.ID()] {
				visited[v.ID()] = true
				disc[v.ID()] = *timer
				low[v.ID()] = *timer
				*timer++

				frame := bridgeFrame{
					u:         v,
					parent:    top.u,
					neighbors: collectNeighbors(g, v.ID()),
				}
				stack = append(stack, frame)
			} else {
				if disc[v.ID()] < low[top.u.ID()] {
					low[top.u.ID()] = disc[v.ID()]
				}
			}
		} else {
			// Done with this node; pop and update parent.
			finished := *top
			stack = stack[:len(stack)-1]

			if len(stack) > 0 {
				parentFrame := &stack[len(stack)-1]
				if low[finished.u.ID()] < low[parentFrame.u.ID()] {
					low[parentFrame.u.ID()] = low[finished.u.ID()]
				}
				if low[finished.u.ID()] > disc[parentFrame.u.ID()] {
					*bridges = append(*bridges, Bridge{From: parentFrame.u, To: finished.u})
				}
			}
		}
	}
}

func collectNeighbors(g graph.Undirected, id int64) []graph.Node {
	var result []graph.Node
	it := g.From(id)
	for it.Next() {
		result = append(result, it.Node())
	}
	return result
}
