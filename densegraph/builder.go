// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package densegraph

import "slices"

type directedEntry struct {
	src, dst int64
	weight   float64
}

// UndirectedBuilder collects nodes and edges, then freezes them into a
// compact CSR-backed Undirected graph via Build.
type UndirectedBuilder struct {
	nodeSet map[int64]struct{}
	edges   []directedEntry
}

// NewUndirectedBuilder creates a builder for an in-memory undirected graph.
func NewUndirectedBuilder() *UndirectedBuilder {
	return &UndirectedBuilder{
		nodeSet: make(map[int64]struct{}),
	}
}

// NewUndirectedBuilderSized creates a builder with pre-allocated capacity.
// edgeHint is the expected number of directed edge entries (2× undirected
// edges if using AddEdge/AddWeightedEdge, 1× if using AddDirectedEntry).
func NewUndirectedBuilderSized(nodeHint, edgeHint int) *UndirectedBuilder {
	return &UndirectedBuilder{
		nodeSet: make(map[int64]struct{}, nodeHint),
		edges:   make([]directedEntry, 0, edgeHint),
	}
}

// AddNode adds an isolated node.
func (b *UndirectedBuilder) AddNode(id int64) {
	b.nodeSet[id] = struct{}{}
}

// AddEdge adds an unweighted undirected edge (weight 1.0).
// Both directions are stored internally.
func (b *UndirectedBuilder) AddEdge(uid, vid int64) {
	b.AddWeightedEdge(uid, vid, 1.0)
}

// AddWeightedEdge adds a weighted undirected edge. Both directions are
// stored internally. Duplicate edges are deduplicated during Build (last
// weight wins).
func (b *UndirectedBuilder) AddWeightedEdge(uid, vid int64, weight float64) {
	b.nodeSet[uid] = struct{}{}
	b.nodeSet[vid] = struct{}{}
	b.edges = append(b.edges, directedEntry{src: uid, dst: vid, weight: weight})
	if uid != vid {
		b.edges = append(b.edges, directedEntry{src: vid, dst: uid, weight: weight})
	}
}

// AddDirectedEntry adds a single directed edge entry. Use this when
// loading from a source that already stores both directions (e.g., a
// DuckDB/SQLite edges table with (src, dst) and (dst, src) rows).
func (b *UndirectedBuilder) AddDirectedEntry(src, dst int64, weight float64) {
	b.nodeSet[src] = struct{}{}
	b.nodeSet[dst] = struct{}{}
	b.edges = append(b.edges, directedEntry{src: src, dst: dst, weight: weight})
}

// Build freezes the collected data into a compact, read-only Undirected
// graph. The builder should not be reused after calling Build.
func (b *UndirectedBuilder) Build() *Undirected {
	// Collect and sort node IDs.
	nodeIDs := make([]int64, 0, len(b.nodeSet))
	for id := range b.nodeSet {
		nodeIDs = append(nodeIDs, id)
	}
	slices.Sort(nodeIDs)
	n := len(nodeIDs)
	b.nodeSet = nil

	if n == 0 {
		return &Undirected{}
	}

	// Sort edges by (src, dst) so CSR neighbor lists are pre-sorted.
	slices.SortFunc(b.edges, func(a, b directedEntry) int {
		if a.src < b.src {
			return -1
		}
		if a.src > b.src {
			return 1
		}
		if a.dst < b.dst {
			return -1
		}
		if a.dst > b.dst {
			return 1
		}
		return 0
	})

	// Dedup consecutive (src, dst) pairs — last weight wins.
	if len(b.edges) > 0 {
		w := 0
		for r := 1; r < len(b.edges); r++ {
			if b.edges[r].src == b.edges[w].src && b.edges[r].dst == b.edges[w].dst {
				b.edges[w].weight = b.edges[r].weight
			} else {
				w++
				b.edges[w] = b.edges[r]
			}
		}
		b.edges = b.edges[:w+1]
	}

	// First pass: count degree of each node.
	degree := make([]int32, n)
	for i := range b.edges {
		idx := searchID(nodeIDs, b.edges[i].src)
		if idx >= 0 {
			degree[idx]++
		}
	}

	// Build offset table from degree counts.
	offsets := make([]int32, n+1)
	for i := range n {
		offsets[i+1] = offsets[i] + degree[i]
	}
	totalEdges := int(offsets[n])

	// Second pass: fill targets and weights (already sorted by dst
	// within each src group thanks to the pre-sort).
	targets := make([]int64, totalEdges)
	weights := make([]float64, totalEdges)
	cursor := make([]int32, n)
	copy(cursor, offsets[:n])
	for i := range b.edges {
		idx := searchID(nodeIDs, b.edges[i].src)
		if idx >= 0 {
			pos := cursor[idx]
			targets[pos] = b.edges[i].dst
			weights[pos] = b.edges[i].weight
			cursor[idx]++
		}
	}
	b.edges = nil // release builder memory

	// Build dense-index version of targets.
	denseTargets := make([]int32, totalEdges)
	for i, tid := range targets {
		denseTargets[i] = int32(searchID(nodeIDs, tid))
	}

	return &Undirected{
		nodeIDs:      nodeIDs,
		offsets:      offsets,
		targets:      targets,
		weights:      weights,
		denseTargets: denseTargets,
	}
}
