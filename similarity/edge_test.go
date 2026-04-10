// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestJaccard_OneIsolated(t *testing.T) {
	// Node 0 has neighbors, node 1 has none.
	// N(0)={2}, N(1)={} => union={2}, intersection={} => J=0
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))
	g.AddNode(simple.Node(1))

	score := Jaccard(context.Background(), g, 0, 1)
	if score != 0 {
		t.Errorf("expected J=0 when one node isolated, got %f", score)
	}
}
