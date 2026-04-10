// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package embedding

import (
	"context"
	"math"

	"gonum.org/v1/gonum/mat"
)

// Embedding maps node IDs to fixed-dimensional vectors.
type Embedding map[int64][]float64

// Embed computes node embeddings from random walks using a co-occurrence
// matrix and truncated SVD.
//
// The algorithm:
//  1. Builds a co-occurrence matrix from the walks using a sliding window.
//  2. Applies shifted positive PMI (PPMI) weighting.
//  3. Computes a truncated SVD to produce dim-dimensional embeddings.
//
// This is a lightweight alternative to training a skip-gram model and produces
// comparable results for many downstream tasks.
//
// Reference: O. Levy and Y. Goldberg, "Neural Word Embedding as Implicit
// Matrix Factorization", NIPS 2014.
func Embed(ctx context.Context, walks [][]int64, nodeIDs []int64, dim, windowSize int) Embedding {
	n := len(nodeIDs)
	if n == 0 || dim <= 0 {
		return make(Embedding)
	}

	idx := make(map[int64]int, n)
	for i, id := range nodeIDs {
		idx[id] = i
	}

	// Build co-occurrence counts.
	cooccur := mat.NewDense(n, n, nil)
	nodeCount := make([]float64, n)
	totalPairs := 0.0

	for _, walk := range walks {
		for i, id := range walk {
			ii, ok := idx[id]
			if !ok {
				continue
			}
			nodeCount[ii]++

			for j := i + 1; j < len(walk) && j <= i+windowSize; j++ {
				jj, ok := idx[walk[j]]
				if !ok {
					continue
				}
				cooccur.Set(ii, jj, cooccur.At(ii, jj)+1)
				cooccur.Set(jj, ii, cooccur.At(jj, ii)+1)
				totalPairs += 2
			}
		}
	}

	if totalPairs == 0 {
		// No co-occurrences; return zero vectors.
		result := make(Embedding, n)
		for _, id := range nodeIDs {
			result[id] = make([]float64, dim)
		}
		return result
	}

	// Compute PPMI matrix.
	ppmi := mat.NewDense(n, n, nil)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			cij := cooccur.At(i, j)
			if cij == 0 {
				continue
			}
			pmi := math.Log(cij*totalPairs/(nodeCount[i]*nodeCount[j]))
			if pmi > 0 {
				ppmi.Set(i, j, pmi)
			}
		}
	}

	// Truncated SVD.
	if dim > n {
		dim = n
	}
	var svd mat.SVD
	if !svd.Factorize(ppmi, mat.SVDThin) {
		// Fallback: return zero vectors.
		result := make(Embedding, n)
		for _, id := range nodeIDs {
			result[id] = make([]float64, dim)
		}
		return result
	}

	var u mat.Dense
	svd.UTo(&u)
	vals := svd.Values(nil)

	// Embedding = U[:, :dim] * sqrt(Sigma[:dim])
	result := make(Embedding, n)
	for i, id := range nodeIDs {
		vec := make([]float64, dim)
		for d := 0; d < dim && d < len(vals); d++ {
			vec[d] = u.At(i, d) * math.Sqrt(vals[d])
		}
		result[id] = vec
	}
	return result
}
