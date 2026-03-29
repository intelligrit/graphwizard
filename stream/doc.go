// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package stream provides a thread-safe, mutable graph wrapper that tracks
// incremental changes. It wraps a gonum simple.WeightedUndirectedGraph and
// records all mutations (AddNode, RemoveNode, AddEdge, RemoveEdge) in a
// change log.
//
// This is useful for streaming graph updates where you need to know what
// changed since the last checkpoint, or for feeding incremental updates to
// algorithms that support them.
//
// Usage:
//
//	sg := stream.New()
//	sg.AddNode(1)
//	sg.AddNode(2)
//	sg.AddEdge(1, 2, 1.0)
//	changes := sg.Changes()  // returns the 3 mutations
//	sg.Flush()               // clears the change log
//	g := sg.Graph()          // use with any graphwizard algorithm
package stream
