// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"context"

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
func BiconnectedComponents(ctx context.Context, g graph.Undirected) [][]Edge {
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

// biconnFrame is a stack frame for iterative biconnected-component DFS.
type biconnFrame struct {
	u         int64
	parent    int64
	neighbors []int64
	idx       int
	children  int
}

func biconnDFS(g graph.Undirected, root, parent int64, visited map[int64]bool, disc, low map[int64]int, timer *int, stack *[]Edge, components *[][]Edge) {
	// Iterative DFS using explicit stack to avoid stack overflow.
	callStack := []biconnFrame{{u: root, parent: parent, neighbors: collectNeighborIDs(g, root)}}
	visited[root] = true
	disc[root] = *timer
	low[root] = *timer
	*timer++

	for len(callStack) > 0 {
		top := &callStack[len(callStack)-1]

		if top.idx < len(top.neighbors) {
			v := top.neighbors[top.idx]
			top.idx++

			if !visited[v] {
				top.children++
				*stack = append(*stack, Edge{From: top.u, To: v})

				visited[v] = true
				disc[v] = *timer
				low[v] = *timer
				*timer++

				callStack = append(callStack, biconnFrame{
					u:         v,
					parent:    top.u,
					neighbors: collectNeighborIDs(g, v),
				})
			} else if v != top.parent && disc[v] < disc[top.u] {
				*stack = append(*stack, Edge{From: top.u, To: v})
				if disc[v] < low[top.u] {
					low[top.u] = disc[v]
				}
			}
		} else {
			// Done with this node; pop and update parent.
			finished := *top
			callStack = callStack[:len(callStack)-1]

			if len(callStack) > 0 {
				parentFrame := &callStack[len(callStack)-1]
				if low[finished.u] < low[parentFrame.u] {
					low[parentFrame.u] = low[finished.u]
				}

				// Check if parentFrame.u is an articulation point via this child.
				isRoot := parentFrame.parent == -1
				if (isRoot && parentFrame.children > 1) || (!isRoot && low[finished.u] >= disc[parentFrame.u]) {
					var comp []Edge
					for {
						if len(*stack) == 0 {
							break
						}
						e := (*stack)[len(*stack)-1]
						*stack = (*stack)[:len(*stack)-1]
						comp = append(comp, e)
						if e.From == parentFrame.u && e.To == finished.u {
							break
						}
					}
					if len(comp) > 0 {
						*components = append(*components, comp)
					}
				}
			}
		}
	}
}

func collectNeighborIDs(g graph.Undirected, id int64) []int64 {
	var result []int64
	it := g.From(id)
	for it.Next() {
		result = append(result, it.Node().ID())
	}
	return result
}

// ArticulationPoints returns all articulation points (cut vertices) in an
// undirected graph. An articulation point is a node whose removal disconnects
// the graph.
func ArticulationPoints(ctx context.Context, g graph.Undirected) []int64 {
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

// apFrame is a stack frame for iterative articulation-point DFS.
type apFrame struct {
	u         int64
	parent    int64
	neighbors []int64
	idx       int
	children  int
}

func apDFS(g graph.Undirected, root, parent int64, visited map[int64]bool, disc, low map[int64]int, isAP map[int64]bool, timer *int) {
	// Iterative DFS using explicit stack to avoid stack overflow.
	callStack := []apFrame{{u: root, parent: parent, neighbors: collectNeighborIDs(g, root)}}
	visited[root] = true
	disc[root] = *timer
	low[root] = *timer
	*timer++

	for len(callStack) > 0 {
		top := &callStack[len(callStack)-1]

		if top.idx < len(top.neighbors) {
			v := top.neighbors[top.idx]
			top.idx++

			if !visited[v] {
				top.children++

				visited[v] = true
				disc[v] = *timer
				low[v] = *timer
				*timer++

				callStack = append(callStack, apFrame{
					u:         v,
					parent:    top.u,
					neighbors: collectNeighborIDs(g, v),
				})
			} else if v != top.parent {
				if disc[v] < low[top.u] {
					low[top.u] = disc[v]
				}
			}
		} else {
			// Done with this node; pop and update parent.
			finished := *top
			callStack = callStack[:len(callStack)-1]

			if len(callStack) > 0 {
				parentFrame := &callStack[len(callStack)-1]
				if low[finished.u] < low[parentFrame.u] {
					low[parentFrame.u] = low[finished.u]
				}
				if parentFrame.parent == -1 && parentFrame.children > 1 {
					isAP[parentFrame.u] = true
				}
				if parentFrame.parent != -1 && low[finished.u] >= disc[parentFrame.u] {
					isAP[parentFrame.u] = true
				}
			}
		}
	}
}
