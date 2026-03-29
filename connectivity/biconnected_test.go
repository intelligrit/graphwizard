// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"sort"
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestBiconnectedComponents_TwoTrianglesWithBridge(t *testing.T) {
	// Two triangles connected by a bridge: {0,1,3} -- bridge 1-2 -- {2,4,5}
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(2)))

	comps := BiconnectedComponents(g)
	// Expect 3 components: left triangle (3 edges), bridge (1 edge), right triangle (3 edges)
	if len(comps) != 3 {
		t.Fatalf("expected 3 biconnected components, got %d", len(comps))
	}

	sizes := make([]int, len(comps))
	for i, c := range comps {
		sizes[i] = len(c)
	}
	sort.Ints(sizes)
	// Should be [1, 3, 3]
	if sizes[0] != 1 || sizes[1] != 3 || sizes[2] != 3 {
		t.Errorf("expected component sizes [1,3,3], got %v", sizes)
	}
}

func TestBiconnectedComponents_SingleTriangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	comps := BiconnectedComponents(g)
	if len(comps) != 1 {
		t.Fatalf("expected 1 biconnected component, got %d", len(comps))
	}
	if len(comps[0]) != 3 {
		t.Errorf("expected 3 edges in component, got %d", len(comps[0]))
	}
}

func TestArticulationPoints_TwoTrianglesWithBridge(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(4)))
	g.SetEdge(g.NewEdge(simple.Node(4), simple.Node(5)))
	g.SetEdge(g.NewEdge(simple.Node(5), simple.Node(2)))

	aps := ArticulationPoints(g)
	sort.Slice(aps, func(i, j int) bool { return aps[i] < aps[j] })
	// Nodes 1 and 2 are articulation points (bridge endpoints)
	if len(aps) != 2 {
		t.Fatalf("expected 2 articulation points, got %d: %v", len(aps), aps)
	}
	if aps[0] != 1 || aps[1] != 2 {
		t.Errorf("expected articulation points [1,2], got %v", aps)
	}
}

func TestArticulationPoints_Triangle(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	aps := ArticulationPoints(g)
	if len(aps) != 0 {
		t.Errorf("triangle has no articulation points, got %v", aps)
	}
}
