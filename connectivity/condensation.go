// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
)

// CondensedEdge represents a directed edge between two SCCs in the condensation DAG.
type CondensedEdge struct {
	From, To int
}

// Condensation returns the DAG condensation of a directed graph: each strongly
// connected component is collapsed into a single node. Returns:
//   - components: each SCC as a slice of original node IDs
//   - edges: directed edges between SCCs (no duplicates)
//   - nodeToSCC: mapping from original node ID to SCC index
//
// The condensation is always a DAG (directed acyclic graph).
//
// Reference: Standard algorithm — compute SCCs via Tarjan, then contract.
func Condensation(g graph.Directed) (components [][]int64, edges []CondensedEdge, nodeToSCC map[int64]int) {
	// Compute SCCs.
	sccs := topo.TarjanSCC(g)

	components = make([][]int64, len(sccs))
	nodeToSCC = make(map[int64]int)
	for i, scc := range sccs {
		ids := make([]int64, len(scc))
		for j, n := range scc {
			ids[j] = n.ID()
			nodeToSCC[n.ID()] = i
		}
		components[i] = ids
	}

	// Build condensation edges.
	edgeSet := make(map[[2]int]bool)
	for _, scc := range sccs {
		for _, n := range scc {
			fromSCC := nodeToSCC[n.ID()]
			to := g.From(n.ID())
			for to.Next() {
				toSCC := nodeToSCC[to.Node().ID()]
				if fromSCC != toSCC {
					key := [2]int{fromSCC, toSCC}
					if !edgeSet[key] {
						edgeSet[key] = true
						edges = append(edges, CondensedEdge{From: fromSCC, To: toSCC})
					}
				}
			}
		}
	}

	return components, edges, nodeToSCC
}
