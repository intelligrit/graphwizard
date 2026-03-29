// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package connectivity

// UnionFind implements a disjoint-set data structure with union by rank and
// path compression, giving nearly O(1) amortized operations.
//
// Reference: R. Tarjan, "Efficiency of a Good But Not Linear Set Union
// Algorithm", JACM, 1975.
type UnionFind struct {
	parent map[int64]int64
	rank   map[int64]int
}

// NewUnionFind creates a new UnionFind with no elements. Elements are added
// implicitly on first use via Find or Union.
func NewUnionFind() *UnionFind {
	return &UnionFind{
		parent: make(map[int64]int64),
		rank:   make(map[int64]int),
	}
}

// Find returns the representative (root) of the set containing x, creating a
// singleton set if x has not been seen before. Uses path compression.
func (uf *UnionFind) Find(x int64) int64 {
	if _, ok := uf.parent[x]; !ok {
		uf.parent[x] = x
		uf.rank[x] = 0
	}
	if uf.parent[x] != x {
		uf.parent[x] = uf.Find(uf.parent[x])
	}
	return uf.parent[x]
}

// Union merges the sets containing x and y. Returns true if x and y were in
// different sets (i.e., a merge actually happened).
func (uf *UnionFind) Union(x, y int64) bool {
	rx := uf.Find(x)
	ry := uf.Find(y)
	if rx == ry {
		return false
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
	return true
}

// Connected returns true if x and y are in the same set.
func (uf *UnionFind) Connected(x, y int64) bool {
	return uf.Find(x) == uf.Find(y)
}

// Sets returns all disjoint sets as a map from representative to members.
func (uf *UnionFind) Sets() map[int64][]int64 {
	groups := make(map[int64][]int64)
	for x := range uf.parent {
		root := uf.Find(x)
		groups[root] = append(groups[root], x)
	}
	return groups
}
