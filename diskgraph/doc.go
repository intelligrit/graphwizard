// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package diskgraph provides disk-backed graph implementations using bbolt
// that satisfy gonum/graph interfaces. Graphs are persisted to a single file,
// memory-mapped for reads, and can exceed available RAM.
//
// diskgraph is the recommended graph storage for GraphWizard. Benchmarks show
// it is faster than gonum's in-memory simple.UndirectedGraph for the majority
// of algorithms, while also providing persistence and out-of-core support.
//
// # Building a graph
//
// Use a Builder to create a graph file. For best performance, use Batch to
// write many edges in a single transaction:
//
//	b, _ := diskgraph.NewUndirectedBuilder("social.db")
//	b.Batch(func(tx *diskgraph.UndirectedTx) error {
//	    tx.AddEdge(0, 1)
//	    tx.AddEdge(1, 2)
//	    tx.AddWeightedEdge(2, 3, 0.75)
//	    return nil
//	})
//	b.Close() // persists to disk
//
// # Opening and using a graph
//
// Open the file for read-only use with any GraphWizard algorithm:
//
//	g, _ := diskgraph.OpenUndirected("social.db")
//	defer g.Close()
//	deg := centrality.Degree(g)
//	bridges := connectivity.Bridges(g)
//
// The graph file is persistent — build once, open as many times as needed
// across program runs.
//
// # Adjacency preloading
//
// For algorithms that call HasEdgeBetween in tight loops (e.g.,
// ClusteringCoefficient), pass Preload to cache adjacency data in
// Go memory for O(1) lookups:
//
//	g, _ := diskgraph.OpenUndirected("graph.db", diskgraph.Preload)
//
// Or call PreloadAdjacency() after opening:
//
//	g, _ := diskgraph.OpenUndirected("graph.db")
//	g.PreloadAdjacency()
//
// Preload costs roughly E * 40 bytes (83M edges ~ 3.3 GB). If the
// estimated cost exceeds 70% of available RAM, a warning is logged and
// preloading is skipped. Use ForcePreload to override.
package diskgraph
