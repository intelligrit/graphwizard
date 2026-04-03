// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph_test

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/intelligrit/graphwizard/diskgraph"
)

func Example() {
	dir := tempDir()
	path := filepath.Join(dir, "example.db")

	// Build a small social graph on disk.
	b, _ := diskgraph.NewUndirectedBuilder(path)
	b.AddEdge(0, 1)
	b.AddEdge(1, 2)
	b.AddEdge(2, 0)
	b.Close()

	// Open read-only — this memory-maps the file.
	g, _ := diskgraph.OpenUndirected(path)
	defer g.Close()

	// Use standard gonum/graph interface methods.
	fmt.Println("Node 0 exists:", g.Node(0) != nil)
	fmt.Println("Edge 0-1:", g.HasEdgeBetween(0, 1))
	fmt.Println("Edge 0-99:", g.HasEdgeBetween(0, 99))

	w, ok := g.Weight(0, 1)
	fmt.Printf("Weight(0,1): %.1f, ok=%v\n", w, ok)

	// Output:
	// Node 0 exists: true
	// Edge 0-1: true
	// Edge 0-99: false
	// Weight(0,1): 1.0, ok=true
}

func Example_directed() {
	dir := tempDir()
	path := filepath.Join(dir, "dag.db")

	b, _ := diskgraph.NewDirectedBuilder(path)
	b.AddEdge(0, 1)
	b.AddEdge(0, 2)
	b.AddEdge(1, 2)
	b.Close()

	g, _ := diskgraph.OpenDirected(path)
	defer g.Close()

	// Forward: 0 -> {1, 2}
	fwd := collectNodeIDs(g.From(0))
	sort.Slice(fwd, func(i, j int) bool { return fwd[i] < fwd[j] })
	fmt.Println("From(0):", fwd)

	// Reverse: 2 <- {0, 1}
	rev := collectNodeIDs(g.To(2))
	sort.Slice(rev, func(i, j int) bool { return rev[i] < rev[j] })
	fmt.Println("To(2):", rev)

	fmt.Println("0->1:", g.HasEdgeFromTo(0, 1))
	fmt.Println("1->0:", g.HasEdgeFromTo(1, 0))

	// Output:
	// From(0): [1 2]
	// To(2): [0 1]
	// 0->1: true
	// 1->0: false
}

func Example_weighted() {
	dir := tempDir()
	path := filepath.Join(dir, "weighted.db")

	b, _ := diskgraph.NewUndirectedBuilder(path)
	b.AddWeightedEdge(0, 1, 0.5)
	b.AddWeightedEdge(1, 2, 1.5)
	b.Close()

	g, _ := diskgraph.OpenUndirected(path)
	defer g.Close()

	e := g.WeightedEdge(0, 1)
	fmt.Printf("Edge 0->1 weight: %.1f\n", e.Weight())

	e = g.WeightedEdgeBetween(1, 0)
	fmt.Printf("Edge 1->0 weight: %.1f\n", e.Weight())

	// Output:
	// Edge 0->1 weight: 0.5
	// Edge 1->0 weight: 0.5
}
