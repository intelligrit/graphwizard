// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"gonum.org/v1/gonum/graph"
)

// Edge represents an undirected edge between two nodes.
type Edge struct {
	From, To int64
}

// BiconnectedComponents returns the biconnected components of an undirected
// graph. Each component is a set of edges that form a maximal biconnected
// subgraph.
//
// A biconnected component is a maximal set of edges such that any two edges
// in the set lie on a common simple cycle. Articulation points (cut vertices)
// appear in multiple components.
//
// Reference: J. Hopcroft and R. Tarjan, "Algorithm 447: efficient algorithms
// for graph manipulation", Communications of the ACM, 1973.
func BiconnectedComponents(g graph.Undirected) [][]Edge {
	disc := make(map[int64]int)
	low := make(map[int64]int)
	visited := make(map[int64]bool)
	timer := 0
	var stack []Edge
	var components [][]Edge

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		if !visited[n.ID()] {
			biconnDFS(g, n.ID(), -1, visited, disc, low, &timer, &stack, &components)
			// Remaining edges on the stack form a component.
			if len(stack) > 0 {
				comp := make([]Edge, len(stack))
				copy(comp, stack)
				components = append(components, comp)
				stack = stack[:0]
			}
		}
	}

	return components
}

func biconnDFS(g graph.Undirected, u, parent int64, visited map[int64]bool, disc, low map[int64]int, timer *int, stack *[]Edge, components *[][]Edge) {
	visited[u] = true
	disc[u] = *timer
	low[u] = *timer
	*timer++
	children := 0

	neighbors := g.From(u)
	for neighbors.Next() {
		v := neighbors.Node().ID()
		if !visited[v] {
			children++
			*stack = append(*stack, Edge{From: u, To: v})
			biconnDFS(g, v, u, visited, disc, low, timer, stack, components)

			if low[v] < low[u] {
				low[u] = low[v]
			}

			// u is an articulation point: pop a component.
			if (parent == -1 && children > 1) || (parent != -1 && low[v] >= disc[u]) {
				var comp []Edge
				for {
					if len(*stack) == 0 {
						break
					}
					e := (*stack)[len(*stack)-1]
					*stack = (*stack)[:len(*stack)-1]
					comp = append(comp, e)
					if e.From == u && e.To == v {
						break
					}
				}
				if len(comp) > 0 {
					*components = append(*components, comp)
				}
			}
		} else if v != parent && disc[v] < disc[u] {
			*stack = append(*stack, Edge{From: u, To: v})
			if disc[v] < low[u] {
				low[u] = disc[v]
			}
		}
	}
}

// ArticulationPoints returns all articulation points (cut vertices) in an
// undirected graph. An articulation point is a node whose removal disconnects
// the graph.
func ArticulationPoints(g graph.Undirected) []int64 {
	disc := make(map[int64]int)
	low := make(map[int64]int)
	visited := make(map[int64]bool)
	isAP := make(map[int64]bool)
	timer := 0

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		if !visited[n.ID()] {
			apDFS(g, n.ID(), -1, visited, disc, low, isAP, &timer)
		}
	}

	var result []int64
	for id := range isAP {
		result = append(result, id)
	}
	return result
}

func apDFS(g graph.Undirected, u, parent int64, visited map[int64]bool, disc, low map[int64]int, isAP map[int64]bool, timer *int) {
	visited[u] = true
	disc[u] = *timer
	low[u] = *timer
	*timer++
	children := 0

	neighbors := g.From(u)
	for neighbors.Next() {
		v := neighbors.Node().ID()
		if !visited[v] {
			children++
			apDFS(g, v, u, visited, disc, low, isAP, timer)
			if low[v] < low[u] {
				low[u] = low[v]
			}
			if parent == -1 && children > 1 {
				isAP[u] = true
			}
			if parent != -1 && low[v] >= disc[u] {
				isAP[u] = true
			}
		} else if v != parent {
			if disc[v] < low[u] {
				low[u] = disc[v]
			}
		}
	}
}
