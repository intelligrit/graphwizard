// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/coloring"
)

// GraphColoring returns a valid vertex coloring of an undirected graph using
// the DSatur exact algorithm. The result maps node IDs to color indices
// (0-based), and k is the chromatic number (minimum colors needed).
//
// Wraps gonum/graph/coloring.DsaturExact.
func GraphColoring(ctx context.Context, g graph.Undirected) (map[int64]int, int, error) {
	k, colors, err := coloring.DsaturExact(nil, g)
	return colors, k, err
}
