// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package anomaly

import (
	"sort"

	"gonum.org/v1/gonum/graph"
)

// StructuralOutliers returns the top-k most anomalous nodes by isolation
// score. If the graph has fewer than k nodes, all nodes are returned.
// Results are sorted by score descending (most anomalous first).
func StructuralOutliers(g graph.Undirected, k int) []int64 {
	scores := IsolationScore(g)

	type entry struct {
		id    int64
		score float64
	}
	var entries []entry
	for id, s := range scores {
		entries = append(entries, entry{id, s})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].score != entries[j].score {
			return entries[i].score > entries[j].score
		}
		return entries[i].id < entries[j].id
	})

	if k > len(entries) {
		k = len(entries)
	}
	result := make([]int64, k)
	for i := 0; i < k; i++ {
		result[i] = entries[i].id
	}
	return result
}
