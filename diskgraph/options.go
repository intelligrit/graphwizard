// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"fmt"
	"log"
)

// Option configures how a graph is opened.
type Option func(*openConfig)

type openConfig struct {
	noPreload    bool
	forcePreload bool
}

// NoPreload disables automatic adjacency preloading. Use this when the
// graph's adjacency structure is too large to fit in memory.
func NoPreload(c *openConfig) {
	c.noPreload = true
}

// ForcePreload forces adjacency preloading even when the estimated memory
// cost exceeds available memory. Use with caution — this may cause swap
// pressure on very large graphs.
func ForcePreload(c *openConfig) {
	c.forcePreload = true
}

// memoryThreshold is the fraction of available memory that the adjacency
// preload is allowed to consume. 0.7 = preload if it fits in 70% of
// available memory, leaving 30% headroom.
const memoryThreshold = 0.70

// estimateAdjBytes estimates the memory needed to preload adjacency data
// for a graph stored in the given bolt adjacency bucket. Each edge is stored
// as two int64 entries (one per direction for undirected), plus map overhead.
//
// Rough formula: adjBucketBytes gives the raw packed data size. The in-memory
// representation uses ~2.5x that (slice headers, map buckets, set entries).
func estimateAdjBytes(adjBucketSize int64) uint64 {
	// Each adjacency entry is 8 bytes in bolt. In memory we store:
	// - adjCache: map[int64][]int64 — 8 bytes per neighbor + slice header
	// - adjSet: map[int64]map[int64]struct{} — ~40 bytes per entry (map overhead)
	// Conservative estimate: 5x the raw bolt data.
	return uint64(adjBucketSize) * 5
}

// tryAutoPreload attempts to preload adjacency data. It checks available
// memory and logs a warning if the graph is too large, unless forced.
func tryAutoPreload(g interface {
	preloadAdj()
	adjBucketSize() int64
}, cfg openConfig) {
	if cfg.noPreload {
		return
	}

	rawSize := g.adjBucketSize()
	estimated := estimateAdjBytes(rawSize)

	if !cfg.forcePreload {
		avail := availableMemory()
		if avail == 0 {
			log.Printf("diskgraph: unable to determine available system memory — skipping adjacency preload (estimated %s needed). Use ForcePreload to override, or call PreloadAdjacency() manually.",
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
