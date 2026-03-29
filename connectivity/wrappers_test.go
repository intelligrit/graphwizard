// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

import (
	"testing"

	"gonum.org/v1/gonum/graph/simple"
)

func TestConnectedComponents_TwoComponents(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))

	comps := ConnectedComponents(g)
	if len(comps) != 2 {
		t.Fatalf("expected 2 components, got %d", len(comps))
	}
}

func TestConnectedComponents_SingleComponent(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	comps := ConnectedComponents(g)
	if len(comps) != 1 {
		t.Fatalf("expected 1 component, got %d", len(comps))
	}
	if len(comps[0]) != 3 {
		t.Errorf("expected 3 nodes in component, got %d", len(comps[0]))
	}
}

func TestStronglyConnectedComponents_Cycle(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	comps := StronglyConnectedComponents(g)
	if len(comps) != 1 {
		t.Fatalf("expected 1 SCC in cycle, got %d", len(comps))
	}
}

func TestStronglyConnectedComponents_Chain(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	comps := StronglyConnectedComponents(g)
	if len(comps) != 3 {
		t.Fatalf("expected 3 SCCs in chain, got %d", len(comps))
	}
}

func TestDirectedCycles_Triangle(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(0)))

	cycles := DirectedCycles(g)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(cycles))
	}
	if len(cycles[0]) != 4 { // cycle includes return to start
		t.Errorf("expected cycle of length 4 (with return), got %d", len(cycles[0]))
	}
}

func TestDirectedCycles_NoCycles(t *testing.T) {
	g := simple.NewDirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))

	cycles := DirectedCycles(g)
	if len(cycles) != 0 {
		t.Fatalf("expected 0 cycles in DAG, got %d", len(cycles))
	}
}

func TestUndirectedCycles_Square(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(1), simple.Node(2)))
	g.SetEdge(g.NewEdge(simple.Node(2), simple.Node(3)))
	g.SetEdge(g.NewEdge(simple.Node(3), simple.Node(0)))

	cycles := UndirectedCycles(g)
	if len(cycles) == 0 {
		t.Fatal("expected at least 1 cycle basis element")
	}
}

func TestUndirectedCycles_Tree(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(1)))
	g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(2)))

	cycles := UndirectedCycles(g)
	if len(cycles) != 0 {
		t.Fatalf("expected 0 cycles in tree, got %d", len(cycles))
	}
}
