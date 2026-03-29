// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"gonum.org/v1/gonum/graph"
)

// NodePairScore holds a similarity score between two nodes.
type NodePairScore struct {
	A, B  graph.Node
	Score float64
}

// Jaccard returns the Jaccard similarity between two nodes based on their
// neighbor sets:
//
//	J(u, v) = |N(u) ∩ N(v)| / |N(u) ∪ N(v)|
//
// Returns 0 if both nodes have no neighbors.
func Jaccard(g graph.Undirected, u, v int64) float64 {
	nu := neighborSet(g, u)
	nv := neighborSet(g, v)

	if len(nu) == 0 && len(nv) == 0 {
		return 0
	}

	intersection := 0
	for id := range nu {
		if _, ok := nv[id]; ok {
			intersection++
		}
	}

	union := len(nu) + len(nv) - intersection
	return float64(intersection) / float64(union)
}

// JaccardAll returns all node pairs with Jaccard similarity at or above the
// given threshold. Only pairs where A.ID() < B.ID() are returned to avoid
// duplicates.
func JaccardAll(g graph.Undirected, threshold float64) []NodePairScore {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}

	var results []NodePairScore
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			score := Jaccard(g, ids[i], ids[j])
			if score >= threshold {
				results = append(results, NodePairScore{
					A:     g.Node(ids[i]),
					B:     g.Node(ids[j]),
					Score: score,
				})
			}
		}
	}

	return results
}

// Overlap returns the overlap coefficient between two nodes:
//
//	O(u, v) = |N(u) ∩ N(v)| / min(|N(u)|, |N(v)|)
//
// Returns 0 if either node has no neighbors.
func Overlap(g graph.Undirected, u, v int64) float64 {
	nu := neighborSet(g, u)
	nv := neighborSet(g, v)

	minSize := len(nu)
	if len(nv) < minSize {
		minSize = len(nv)
	}
	if minSize == 0 {
		return 0
	}

	intersection := 0
	for id := range nu {
		if _, ok := nv[id]; ok {
			intersection++
		}
	}

	return float64(intersection) / float64(minSize)
}

func neighborSet(g graph.Undirected, id int64) map[int64]struct{} {
	s := make(map[int64]struct{})
	it := g.From(id)
	for it.Next() {
		s[it.Node().ID()] = struct{}{}
	}
	return s
}
