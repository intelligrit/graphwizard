// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths

import (
	"container/heap"
	"context"
	"math"
	"sort"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// distItem is an entry in the priority queue for dijkstraFiltered.
type distItem struct {
	id   int64
	dist float64
}

// distHeap implements heap.Interface for Dijkstra's priority queue.
type distHeap []distItem

func (h distHeap) Len() int            { return len(h) }
func (h distHeap) Less(i, j int) bool  { return h[i].dist < h[j].dist }
func (h distHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *distHeap) Push(x interface{}) { *h = append(*h, x.(distItem)) }
func (h *distHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// WeightedPath represents a path with its total weight.
type WeightedPath struct {
	Nodes  []graph.Node
	Weight float64
}

// YenKShortest finds the K shortest loopless paths from source to target in a
// weighted directed graph using Yen's algorithm.
//
// The graph must implement graph.Weighted for edge weights. Returns up to K
// paths sorted by weight. If fewer than K paths exist, all are returned.
//
// Time complexity: O(KN(N log N + E)) where N=nodes, E=edges.
//
// Reference: J. Yen, "Finding the K Shortest Loopless Paths in a Network",
// Management Science, 1971.
func YenKShortest(ctx context.Context, g graph.WeightedDirected, source, target int64, k int) []WeightedPath {
	if k <= 0 {
		return nil
	}

	// Find the first shortest path using Dijkstra.
	shortest := path.DijkstraFrom(g.Node(source), g)
	firstPath, firstWeight := shortest.To(target)
	if math.IsInf(firstWeight, 1) || len(firstPath) == 0 {
		return nil
	}

	A := []WeightedPath{{Nodes: firstPath, Weight: firstWeight}}
	var B []WeightedPath

	for ki := 1; ki < k; ki++ {
		progress.Report(ctx, progress.Progress{Phase: "yen", Step: ki, Total: k})
		prevPath := A[ki-1].Nodes

		for i := 0; i < len(prevPath)-1; i++ {
			spurNode := prevPath[i]
			rootPath := make([]graph.Node, i+1)
			copy(rootPath, prevPath[:i+1])

			// Collect edges and nodes to exclude.
			removedEdges := make(map[[2]int64]bool)
			for _, p := range A {
				if len(p.Nodes) > i && pathPrefixEqual(p.Nodes[:i+1], rootPath) {
					removedEdges[[2]int64{p.Nodes[i].ID(), p.Nodes[i+1].ID()}] = true
				}
			}

			removedNodes := make(map[int64]bool)
			for _, n := range rootPath[:len(rootPath)-1] {
				removedNodes[n.ID()] = true
			}

			// Find spur path in filtered graph.
			spurPath, spurWeight := dijkstraFiltered(g, spurNode.ID(), target, removedEdges, removedNodes)
			if spurPath == nil {
				continue
			}

			// Compute root path weight.
			rootWeight := 0.0
			for j := 0; j < len(rootPath)-1; j++ {
				w, ok := g.Weight(rootPath[j].ID(), rootPath[j+1].ID())
				if !ok {
					rootWeight = math.Inf(1)
					break
				}
				rootWeight += w
			}
			if math.IsInf(rootWeight, 1) {
				continue
			}

			// Combine root + spur (skip duplicate spur node).
			combined := make([]graph.Node, len(rootPath)+len(spurPath)-1)
			copy(combined, rootPath)
			copy(combined[len(rootPath):], spurPath[1:])
			totalWeight := rootWeight + spurWeight

			// Add to B if not duplicate.
			if !containsPath(B, combined) {
				B = append(B, WeightedPath{Nodes: combined, Weight: totalWeight})
			}
		}

		if len(B) == 0 {
			break
		}

		// Sort B and pick the shortest.
		sort.Slice(B, func(i, j int) bool { return B[i].Weight < B[j].Weight })
		A = append(A, B[0])
		B = B[1:]
	}

	return A
}

func pathPrefixEqual(a, b []graph.Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID() != b[i].ID() {
			return false
		}
	}
	return true
}

func containsPath(paths []WeightedPath, candidate []graph.Node) bool {
	for _, p := range paths {
		if len(p.Nodes) != len(candidate) {
			continue
		}
		same := true
		for i := range p.Nodes {
			if p.Nodes[i].ID() != candidate[i].ID() {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}

// dijkstraFiltered runs Dijkstra from source to target, excluding certain
// edges and nodes. Uses a min-heap for O(E log V) instead of O(V^2).
// Returns the path (starting at source) and its weight.
func dijkstraFiltered(g graph.WeightedDirected, source, target int64, removedEdges map[[2]int64]bool, removedNodes map[int64]bool) ([]graph.Node, float64) {
	dist := map[int64]float64{source: 0}
	prev := map[int64]int64{}
	visited := make(map[int64]bool)

	h := &distHeap{{id: source, dist: 0}}
	heap.Init(h)

	for h.Len() > 0 {
		item := heap.Pop(h).(distItem)
		u := item.id

		if visited[u] {
			continue
		}
		if u == target {
			break
		}
		visited[u] = true

		uDist := dist[u]
		to := g.From(u)
		for to.Next() {
			v := to.Node().ID()
			if removedNodes[v] || removedEdges[[2]int64{u, v}] {
				continue
			}
			w, ok := g.Weight(u, v)
			if !ok {
				continue
			}
			alt := uDist + w
			if dv, seen := dist[v]; !seen || alt < dv {
				dist[v] = alt
				prev[v] = u
				heap.Push(h, distItem{id: v, dist: alt})
			}
		}
	}

	if _, ok := dist[target]; !ok {
		return nil, math.Inf(1)
	}

	// Reconstruct path.
	var nodes []int64
	for at := target; at != source; {
		nodes = append(nodes, at)
		p, ok := prev[at]
		if !ok {
			return nil, math.Inf(1)
		}
		at = p
	}
	nodes = append(nodes, source)

	// Reverse.
	result := make([]graph.Node, len(nodes))
	for i, id := range nodes {
		result[len(nodes)-1-i] = g.Node(id)
	}
	return result, dist[target]
}
