// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package embedding_test

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/intelligrit/graphwizard/embedding"
	"gonum.org/v1/gonum/graph/simple"
)

func ExampleNode2VecWalks() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	rng := rand.New(rand.NewSource(42))
	walks := embedding.Node2VecWalks(context.Background(), g, embedding.WalkParams{
		WalkLength:   5,
		WalksPerNode: 2,
		P:            1.0,
		Q:            1.0,
	}, rng)

	fmt.Printf("walks: %d\n", len(walks))
	// Output: walks: 6
}

func ExampleEmbed() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	rng := rand.New(rand.NewSource(42))
	walks := embedding.DeepWalkWalks(context.Background(), g, 10, 20, rng)
	emb := embedding.Embed(context.Background(), walks, []int64{0, 1, 2}, 2, 3)

	fmt.Printf("dim: %d\n", len(emb[0]))
	// Output: dim: 2
}

func ExampleNode2VecWalksParallel() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	walks := embedding.Node2VecWalksParallel(context.Background(), g, embedding.WalkParams{
		WalkLength:   5,
		WalksPerNode: 2,
		P:            1.0,
		Q:            1.0,
	}, 42)

	fmt.Printf("walks: %d\n", len(walks))
	// Output: walks: 6
}

func ExampleDeepWalkWalksParallel() {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	walks := embedding.DeepWalkWalksParallel(context.Background(), g, 10, 3, 42)
	fmt.Printf("walks: %d\n", len(walks))
	// Output: walks: 9
}
