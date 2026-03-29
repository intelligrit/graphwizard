// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/graph"
)

// PredictedLink represents a predicted edge between two unconnected nodes
// with an associated score.
type PredictedLink struct {
	A, B  int64
	Score float64
}

// CommonNeighbors returns the number of shared neighbors between two nodes.
// Higher values suggest a higher likelihood of a future connection.
//
// CN(u, v) = |N(u) ∩ N(v)|
func CommonNeighbors(g graph.Undirected, u, v int64) int {
	nu := neighborSet(g, u)
	nv := neighborSet(g, v)
	count := 0
	for id := range nu {
		if _, ok := nv[id]; ok {
			count++
		}
	}
	return count
}

// AdamicAdar returns the Adamic-Adar index between two nodes. Shared
// neighbors with fewer connections contribute more, making this a weighted
// version of common neighbors that emphasizes rare shared connections.
//
// AA(u, v) = Σ_{w ∈ N(u) ∩ N(v)} 1 / log(|N(w)|)
//
// Reference: L. Adamic and E. Adar, "Friends and neighbors on the Web",
// Social Networks, 2003.
func AdamicAdar(g graph.Undirected, u, v int64) float64 {
	nu := neighborSet(g, u)
	nv := neighborSet(g, v)
	score := 0.0
	for w := range nu {
		if _, ok := nv[w]; !ok {
			continue
		}
		deg := degree(g, w)
		if deg > 1 {
			score += 1.0 / math.Log(float64(deg))
		}
	}
	return score
}

// PreferentialAttachment returns the preferential attachment score between
// two nodes. Nodes with high degree are more likely to form new connections.
//
// PA(u, v) = |N(u)| * |N(v)|
//
// Reference: A. Barabasi and R. Albert, "Emergence of Scaling in Random
// Networks", Science, 1999.
func PreferentialAttachment(g graph.Undirected, u, v int64) int {
	return degree(g, u) * degree(g, v)
}

// PredictLinks returns the top-k predicted links for the entire graph using
// the given scoring function. Only pairs that are NOT already connected and
// where A < B are returned, sorted by score descending.
//
// The scorer function should be one of CommonNeighbors (cast to float64),
// AdamicAdar, or a custom function with signature func(g, u, v) float64.
func PredictLinks(g graph.Undirected, k int, scorer func(graph.Undirected, int64, int64) float64) []PredictedLink {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	var candidates []PredictedLink
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			u, v := ids[i], ids[j]
			if g.HasEdgeBetween(u, v) {
				continue
			}
			score := scorer(g, u, v)
			if score > 0 {
				candidates = append(candidates, PredictedLink{A: u, B: v, Score: score})
			}
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if k > 0 && k < len(candidates) {
		candidates = candidates[:k]
	}
	return candidates
}

// Cosine returns the cosine similarity between two nodes based on their
// neighbor sets (binary vectors over the node space).
//
// Cosine(u, v) = |N(u) ∩ N(v)| / sqrt(|N(u)| * |N(v)|)
//
// Returns 0 if either node has no neighbors.
func Cosine(g graph.Undirected, u, v int64) float64 {
	nu := neighborSet(g, u)
	nv := neighborSet(g, v)
	if len(nu) == 0 || len(nv) == 0 {
		return 0
	}
	intersection := 0
	for id := range nu {
		if _, ok := nv[id]; ok {
			intersection++
		}
	}
	return float64(intersection) / math.Sqrt(float64(len(nu))*float64(len(nv)))
}

func degree(g graph.Undirected, id int64) int {
	it := g.From(id)
	d := 0
	for it.Next() {
		d++
	}
	return d
}
