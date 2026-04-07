// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"fmt"
	"log"
)

// Option configures how a graph is opened.
type Option func(*openConfig)

type openConfig struct {
	preload      bool
	forcePreload bool
}

// Preload enables adjacency preloading at open time. This caches all
// adjacency lists and neighbor sets in Go memory, making From() and
// HasEdgeBetween() O(1). This is recommended for algorithms that call
// HasEdgeBetween in tight loops (e.g., ClusteringCoefficient).
//
// If the estimated memory cost exceeds 70% of available system memory,
// a warning is logged and preloading is skipped. Use ForcePreload to
// override this check.
//
// Memory cost is roughly 8 bytes per edge entry plus 4*(N+1) bytes for
// offsets. For example, 83M undirected edges (166M DB rows) ≈ 1.3 GB.
func Preload(c *openConfig) {
	c.preload = true
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
// format. Each edge row contributes 8 bytes (one int64 target entry) plus
// a small per-node overhead for the offset table. We use 10 bytes per row
// as a conservative estimate.
func estimateAdjBytes(adjBucketSize int64) uint64 {
	rows := uint64(adjBucketSize) / 16 // adjBucketSize = count * 16
	return rows * 10
}

// tryPreload attempts to preload adjacency data, checking available
// memory unless forced.
func tryPreload(g interface {
	preloadAdj()
	adjBucketSize() int64
}, cfg openConfig) {
	rawSize := g.adjBucketSize()
	estimated := estimateAdjBytes(rawSize)

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
