// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestLabelPropagation_TwoClusters(t *testing.T) {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	for i := int64(5); i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))

	rng := rand.New(rand.NewSource(42))
	labels := LabelPropagation(context.Background(), g, 100, rng)

	if labels[0] != labels[1] {
		t.Errorf("nodes 0 and 1 should share a label")
	}
	if labels[5] != labels[6] {
		t.Errorf("nodes 5 and 6 should share a label")
	}
	if labels[0] == labels[5] {
		t.Error("clusters should have different labels")
	}
}

func TestLabelPropagation_Empty(t *testing.T) {
	g := simple.NewUndirectedGraph()
	rng := rand.New(rand.NewSource(42))
	labels := LabelPropagation(context.Background(), g, 100, rng)
	if len(labels) != 0 {
		t.Errorf("expected empty, got %v", labels)
	}
}

func TestLabelPropagation_Isolated(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	g.AddNode(simple.Node(1))

	rng := rand.New(rand.NewSource(42))
	labels := LabelPropagation(context.Background(), g, 100, rng)
	if labels[0] == labels[1] {
		t.Error("isolated nodes should keep different labels")
	}
}
