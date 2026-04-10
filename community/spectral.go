// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"context"
	"math"

	"github.com/intelligrit/graphwizard/progress"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/mat"
)

// SpectralClustering partitions an undirected graph into k clusters using the
// normalized Laplacian eigenvectors and k-means.
//
// The algorithm:
//  1. Computes the normalized graph Laplacian L = I - D^{-1/2} A D^{-1/2}.
//  2. Finds the k smallest eigenvectors of L.
//  3. Clusters the rows (one per node) using k-means.
//
// Reference: A. Ng, M. Jordan, Y. Weiss, "On Spectral Clustering: Analysis
// and an Algorithm", NIPS 2001.
func SpectralClustering(ctx context.Context, g graph.Undirected, k int) map[int64]int {
	nodes := g.Nodes()
	var ids []int64
	for nodes.Next() {
		ids = append(ids, nodes.Node().ID())
	}
	n := len(ids)
	if n == 0 || k <= 0 {
		return make(map[int64]int)
	}
	if k > n {
		k = n
	}

	idx := make(map[int64]int, n)
	for i, id := range ids {
		idx[id] = i
	}

	// Build adjacency matrix and degree.
	progress.Report(ctx, progress.Progress{Phase: "build", Step: 0, Total: 3})
	A := mat.NewDense(n, n, nil)
	deg := make([]float64, n)
	for i, id := range ids {
		it := g.From(id)
		for it.Next() {
			j := idx[it.Node().ID()]
			w := 1.0
			if wg, ok := g.(graph.Weighted); ok {
				if ew, ok := wg.Weight(id, ids[j]); ok {
					w = ew
				}
			}
			A.Set(i, j, w)
			deg[i] += w
		}
	}

	// Normalized Laplacian: L = I - D^{-1/2} A D^{-1/2}
	lData := make([]float64, n*n)
	for i := 0; i < n; i++ {
		lData[i*n+i] = 1.0
		for j := 0; j < n; j++ {
			aij := A.At(i, j)
			if aij == 0 {
				continue
			}
			if deg[i] > 0 && deg[j] > 0 {
				lData[i*n+j] -= aij / (math.Sqrt(deg[i]) * math.Sqrt(deg[j]))
			}
		}
	}
	L := mat.NewSymDense(n, lData)

	// Eigendecomposition.
	progress.Report(ctx, progress.Progress{Phase: "eigen", Step: 1, Total: 3})
	var eig mat.EigenSym
	if !eig.Factorize(L, true) {
		// Fallback: each node in its own cluster.
		result := make(map[int64]int, n)
		for i, id := range ids {
			result[id] = i % k
		}
		return result
	}

	// Extract the k eigenvectors corresponding to the k smallest eigenvalues.
	var evecs mat.Dense
	eig.VectorsTo(&evecs)
	vals := make([]float64, n)
	eig.Values(vals)

	// vals are sorted ascending by gonum, so first k columns are what we want.
	embedding := mat.NewDense(n, k, nil)
	for i := 0; i < n; i++ {
		for j := 0; j < k; j++ {
			embedding.Set(i, j, evecs.At(i, j))
		}
	}

	// Normalize rows.
	for i := 0; i < n; i++ {
		norm := 0.0
		for j := 0; j < k; j++ {
			v := embedding.At(i, j)
			norm += v * v
		}
		norm = math.Sqrt(norm)
		if norm > 0 {
			for j := 0; j < k; j++ {
				embedding.Set(i, j, embedding.At(i, j)/norm)
			}
		}
	}

	// K-means clustering on the rows.
	progress.Report(ctx, progress.Progress{Phase: "cluster", Step: 2, Total: 3})
	labels := kmeans(embedding, n, k, 100)

	result := make(map[int64]int, n)
	for i, id := range ids {
		result[id] = labels[i]
	}
	return result
}

// kmeans clusters n rows of the embedding matrix into k clusters.
func kmeans(data *mat.Dense, n, k, maxIter int) []int {
	_, cols := data.Dims()
	labels := make([]int, n)

	// Initialize centroids: farthest-first (k-means++ lite).
	centroids := make([][]float64, k)
	centroids[0] = make([]float64, cols)
	for j := 0; j < cols; j++ {
		centroids[0][j] = data.At(0, j)
	}
	for c := 1; c < k; c++ {
		// Pick the row farthest from all existing centroids.
		bestRow := 0
		bestDist := -1.0
		for i := 0; i < n; i++ {
			minDist := math.Inf(1)
			for cc := 0; cc < c; cc++ {
				dist := 0.0
				for j := 0; j < cols; j++ {
					d := data.At(i, j) - centroids[cc][j]
					dist += d * d
				}
				if dist < minDist {
					minDist = dist
				}
			}
			if minDist > bestDist {
				bestDist = minDist
				bestRow = i
			}
		}
		centroids[c] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			centroids[c][j] = data.At(bestRow, j)
		}
	}

	for iter := 0; iter < maxIter; iter++ {
		changed := false
		// Assign each point to nearest centroid.
		for i := 0; i < n; i++ {
			best := 0
			bestDist := math.Inf(1)
			for c := 0; c < k; c++ {
				dist := 0.0
				for j := 0; j < cols; j++ {
					d := data.At(i, j) - centroids[c][j]
					dist += d * d
				}
				if dist < bestDist {
					bestDist = dist
					best = c
				}
			}
			if labels[i] != best {
				labels[i] = best
				changed = true
			}
		}

		if !changed {
			break
		}

		// Update centroids.
		counts := make([]int, k)
		for c := range centroids {
			for j := range centroids[c] {
				centroids[c][j] = 0
			}
		}
		for i := 0; i < n; i++ {
			c := labels[i]
			counts[c]++
			for j := 0; j < cols; j++ {
				centroids[c][j] += data.At(i, j)
			}
		}
		for c := 0; c < k; c++ {
			if counts[c] > 0 {
				for j := 0; j < cols; j++ {
					centroids[c][j] /= float64(counts[c])
				}
			}
		}
	}

	return labels
}
