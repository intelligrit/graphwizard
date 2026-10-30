// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"fmt"
	"log"
	"strconv"

	"github.com/intelligrit/graphwizard/internal/compact"
)

// Option configures how a graph is opened.
type Option func(*openConfig)

type openConfig struct {
	preload      bool
	forcePreload bool
	neighbors    compact.NeighborEncoding
}

// Preload enables adjacency preloading at open time. This caches dense-indexed
// adjacency in Go memory. From() iterates without decoding rows and
// HasEdgeBetween() uses binary search in a sorted neighbor list. This is
// recommended for algorithms that call HasEdgeBetween in tight loops
// (e.g., ClusteringCoefficient).
//
// If the estimated memory cost exceeds 70% of available system memory,
// a warning is logged and preloading is skipped. Use ForcePreload to
// override this check.
//
// Memory cost is 4 bytes per directed edge entry plus machine-sized CSR
// offsets and 8 bytes per node ID. SQLite weights remain on disk.
func Preload(c *openConfig) {
	c.preload = true
}

// CompressedPreload enables preloading with lossless block delta encoding when
// its payload is smaller than flat int32 targets. Neighbors and From decode
// lazily; DenseNeighbors allocates a decoded row. Combine with ForcePreload to
// override the available-memory check. Compression is performed after building
// flat adjacency, so peak preload memory exceeds the retained compressed size.
// If combined with Packed24Preload, the last encoding option wins.
func CompressedPreload(c *openConfig) {
	c.preload = true
	c.neighbors = compact.DeltaNeighbors
}

// Packed24Preload enables undirected adjacency preloading with three bytes per
// dense neighbor index. If any index exceeds 16,777,215, flat int32 storage is
// used instead. Original node IDs and SQLite weights are unchanged. Lookups
// read packed indexes directly; DenseNeighbors allocates a decoded row.
// Combine with ForcePreload to override the available-memory check. Packing
// temporarily overlaps flat adjacency. If combined with CompressedPreload,
// the last encoding option wins.
func Packed24Preload(c *openConfig) {
	c.preload = true
	c.neighbors = compact.Packed24Neighbors
}

// ForcePreload enables adjacency preloading and skips the memory safety
// check. Use with caution on very large graphs.
func ForcePreload(c *openConfig) {
	c.preload = true
	c.forcePreload = true
}

// memoryThreshold is the fraction of available memory that the adjacency
// preload is allowed to consume. 0.7 = preload if it fits in 70% of
// available memory, leaving 30% headroom.
const memoryThreshold = 0.70

// estimateAdjBytes estimates the memory needed to preload adjacency in CSR
// format. Keep this estimate conservative: construction also needs row
// cursors, and optional compression temporarily overlaps the flat payload.
// Existing node IDs and per-node overhead must be accounted for separately.
func estimateAdjBytes(adjBucketSize int64) uint64 {
	rows := uint64(adjBucketSize) / 16 // adjBucketSize = count * 16
	return rows * 10
}

// tryPreload attempts to preload adjacency data, checking available
// memory unless forced.
func tryPreload(g interface {
	preloadAdj()
	adjBucketSize() int64
	NumNodes() int
}, cfg openConfig) {
	rawSize := g.adjBucketSize()
	estimated := estimateAdjBytes(rawSize) + uint64(g.NumNodes()+1)*uint64(3*strconv.IntSize/8)

	if !cfg.forcePreload {
		avail := availableMemory()
		if avail == 0 {
			log.Printf("diskgraph: unable to determine available system memory — skipping adjacency preload (estimated %s needed). Call PreloadAdjacency() manually or use ForcePreload to override.",
				formatBytes(estimated),
			)
			return
		}
		threshold := uint64(float64(avail) * memoryThreshold)
		if estimated > threshold {
			log.Printf("diskgraph: skipping adjacency preload — estimated %s needed, %s available (%s with %.0f%% safety margin). Use ForcePreload to override, or call PreloadAdjacency() manually.",
				formatBytes(estimated),
				formatBytes(avail),
				formatBytes(threshold),
				memoryThreshold*100,
			)
			return
		}
	}

	g.preloadAdj()
}

func formatBytes(b uint64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
