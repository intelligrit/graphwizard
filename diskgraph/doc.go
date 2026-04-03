// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package diskgraph provides disk-backed graph implementations using bbolt
// that satisfy gonum/graph interfaces. Graphs are memory-mapped from disk,
// so they can exceed available RAM — the OS page cache handles hot data
// automatically.
//
// The package provides four graph types: [Undirected], [Directed],
// [WeightedUndirected], and [WeightedDirected]. Each is built once using
// a corresponding Builder and then opened read-only for algorithm use.
//
//	// Build a graph on disk.
//	b, _ := diskgraph.NewUndirectedBuilder("social.db")
//	b.AddNode(0)
//	b.AddNode(1)
//	b.AddEdge(0, 1)
//	b.Close() // finalizes the file
//
//	// Open for read-only use with any graphwizard algorithm.
//	g, _ := diskgraph.OpenUndirected("social.db")
//	defer g.Close()
//	deg := centrality.Degree(g)
package diskgraph
