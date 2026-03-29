// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
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
func Bridges(g graph.Undirected) []Bridge {
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

func bridgeDFS(g graph.Undirected, u graph.Node, parent graph.Node, visited map[int64]bool, disc, low map[int64]int, timer *int, bridges *[]Bridge) {
	visited[u.ID()] = true
	disc[u.ID()] = *timer
	low[u.ID()] = *timer
	*timer++

	neighbors := g.From(u.ID())
	for neighbors.Next() {
		v := neighbors.Node()
		if parent != nil && v.ID() == parent.ID() {
			continue
		}
		if !visited[v.ID()] {
			bridgeDFS(g, v, u, visited, disc, low, timer, bridges)
			if low[v.ID()] < low[u.ID()] {
				low[u.ID()] = low[v.ID()]
			}
			if low[v.ID()] > disc[u.ID()] {
				*bridges = append(*bridges, Bridge{From: u, To: v})
			}
		} else {
			if disc[v.ID()] < low[u.ID()] {
				low[u.ID()] = disc[v.ID()]
			}
		}
	}
}
