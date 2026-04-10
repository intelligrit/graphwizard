// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package structure

import (
	"context"

	"github.com/intelligrit/graphwizard/progress"
)

// SetCover solves the minimum set cover problem using a greedy approximation.
//
// Given a universe of elements and a collection of sets (each identified by an
// index), returns the indices of a subset of sets that cover all elements in
// the universe. The greedy algorithm achieves an O(ln n) approximation ratio.
//
// Reference: V. Chvatal, "A Greedy Heuristic for the Set-Covering Problem",
// Mathematics of Operations Research, 1979.
func SetCover(ctx context.Context, universe []int64, sets [][]int64) []int {
	uncovered := make(map[int64]bool, len(universe))
	for _, e := range universe {
		uncovered[e] = true
	}

	used := make(map[int]bool)
	var result []int
	iteration := 0

	for len(uncovered) > 0 {
		progress.Report(ctx, progress.Progress{Phase: "greedy", Step: iteration, Total: -1})
		iteration++
		// Pick the set that covers the most uncovered elements.
		bestIdx := -1
		bestCount := 0

		for i, s := range sets {
			if used[i] {
				continue
			}
			count := 0
			for _, e := range s {
				if uncovered[e] {
					count++
				}
			}
			if count > bestCount {
				bestCount = count
				bestIdx = i
			}
		}

		if bestIdx == -1 || bestCount == 0 {
			break // Cannot cover remaining elements.
		}

		used[bestIdx] = true
		result = append(result, bestIdx)
		for _, e := range sets[bestIdx] {
			delete(uncovered, e)
		}
	}

	return result
}
