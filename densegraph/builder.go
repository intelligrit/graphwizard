// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package densegraph

import (
	"math"
	"slices"

	"github.com/intelligrit/graphwizard/internal/compact"
)

// BuildOption configures the immutable in-memory representation.
type BuildOption func(*buildConfig)

type buildConfig struct {
	neighbors    compact.NeighborEncoding
	blockWeights bool
}

// CompressNeighbors enables lossless block delta encoding when it saves memory.
// From and Neighbors decode lazily; DenseNeighbors returns a decoded row rather
// than caching it. This trades additional CPU work for lower retained memory.
// If combined with PackNeighbors24, the last neighbor option wins.
func CompressNeighbors(c *buildConfig) { c.neighbors = compact.DeltaNeighbors }

// PackNeighbors24 stores dense neighbor indexes in three bytes instead of four.
// It falls back to flat int32 storage if any index exceeds 16,777,215. Original
// int64 node IDs are unaffected. Lookups read packed indexes directly; traversal
// decodes lazily and DenseNeighbors allocates a decoded row. If combined with
// CompressNeighbors, the last neighbor option wins.
func PackNeighbors24(c *buildConfig) { c.neighbors = compact.Packed24Neighbors }

// BlockWeights selects exact constant, uint8, uint16, uint32, or float64 storage
// separately for each block of 128 directed entries. It is retained only when
// its data and metadata are smaller than the default graph-wide encoding. This
// can isolate outliers, at the cost of extra construction work and decoding.
// It can be combined with either neighbor option.
func BlockWeights(c *buildConfig) { c.blockWeights = true }

type directedEntry struct {
	src, dst int32
	weight   float64
}

// UndirectedBuilder collects nodes and edges, then freezes them into a
// compact CSR-backed Undirected graph via Build.
type UndirectedBuilder struct {
	// Temporary indexes follow insertion order. Build remaps them to sorted
	// original-ID order once, avoiding two int64 IDs in every buffered entry.
	nodeIndex map[int64]int32
	edges     []directedEntry
}

// NewUndirectedBuilder creates a builder for an in-memory undirected graph.
func NewUndirectedBuilder() *UndirectedBuilder {
	return &UndirectedBuilder{
		nodeIndex: make(map[int64]int32),
	}
}

// NewUndirectedBuilderSized creates a builder with pre-allocated capacity.
// edgeHint is the expected number of directed edge entries (2× undirected
// edges if using AddEdge/AddWeightedEdge, 1× if using AddDirectedEntry).
func NewUndirectedBuilderSized(nodeHint, edgeHint int) *UndirectedBuilder {
	return &UndirectedBuilder{
		nodeIndex: make(map[int64]int32, nodeHint),
		edges:     make([]directedEntry, 0, edgeHint),
	}
}

// AddNode adds an isolated node.
func (b *UndirectedBuilder) AddNode(id int64) {
	b.index(id)
}

func (b *UndirectedBuilder) index(id int64) int32 {
	if i, ok := b.nodeIndex[id]; ok {
		return i
	}
	if len(b.nodeIndex) >= math.MaxInt32 {
		panic("densegraph: more than MaxInt32 nodes")
	}
	i := int32(len(b.nodeIndex))
	b.nodeIndex[id] = i
	return i
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
	u, v := b.index(uid), b.index(vid)
	b.edges = append(b.edges, directedEntry{src: u, dst: v, weight: weight})
	if uid != vid {
		b.edges = append(b.edges, directedEntry{src: v, dst: u, weight: weight})
	}
}

// AddDirectedEntry adds a single directed edge entry. Use this when
// loading from a source that already stores both directions (e.g., a
// DuckDB/SQLite edges table with (src, dst) and (dst, src) rows).
func (b *UndirectedBuilder) AddDirectedEntry(src, dst int64, weight float64) {
	u, v := b.index(src), b.index(dst)
	b.edges = append(b.edges, directedEntry{src: u, dst: v, weight: weight})
}

// Build freezes the collected data into a compact, read-only Undirected
// graph. The builder should not be reused after calling Build.
func (b *UndirectedBuilder) Build(opts ...BuildOption) *Undirected {
	var cfg buildConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	// Collect and sort node IDs.
	nodeIDs := make([]int64, 0, len(b.nodeIndex))
	for id := range b.nodeIndex {
		nodeIDs = append(nodeIDs, id)
	}
	slices.Sort(nodeIDs)
	n := len(nodeIDs)
	remap := make([]int32, n)
	for i, id := range nodeIDs {
		remap[b.nodeIndex[id]] = int32(i)
	}
	b.nodeIndex = nil
	for i := range b.edges {
		b.edges[i].src = remap[b.edges[i].src]
		b.edges[i].dst = remap[b.edges[i].dst]
	}

	// Sort edges by (src, dst) so CSR neighbor lists are pre-sorted.
	// Stability preserves insertion order for last-weight-wins deduplication.
	slices.SortStableFunc(b.edges, func(a, b directedEntry) int {
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

	// Count directly into the offset table, then prefix-sum in place. Dense
	// endpoints avoid per-edge binary searches and a separate degree array.
	offsets := make([]int, n+1)
	for _, edge := range b.edges {
		offsets[int(edge.src)+1]++
	}
	for i := range n {
		offsets[i+1] += offsets[i]
	}
	totalEdges := int(offsets[n])

	// The sorted edge stream already has CSR order. Store only dense indices;
	// original IDs are recovered lazily through nodeIDs.
	targets := make([]int32, totalEdges)
	for i, edge := range b.edges {
		targets[i] = edge.dst
	}
	weightBuilder := compact.NewWeights
	if cfg.blockWeights {
		weightBuilder = compact.NewBlockWeights
	}
	weights := weightBuilder(totalEdges, func(i int) float64 { return b.edges[i].weight })
	b.edges = nil

	return &Undirected{
		nodeIDs: nodeIDs,
		adj:     compact.NewNeighbors(offsets, targets, cfg.neighbors),
		weights: weights,
	}
}
