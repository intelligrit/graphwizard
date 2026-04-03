// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"math/rand"
	"os"
	"path/filepath"

	"github.com/intelligrit/graphwizard/diskgraph"
)

// DiskErdosRenyi generates a random undirected disk-backed graph with n nodes
// where each possible edge exists independently with probability p.
func DiskErdosRenyi(n int, p float64, rng *rand.Rand, dir string) (*diskgraph.Undirected, error) {
	path := filepath.Join(dir, "er.db")
	b, err := diskgraph.NewUndirectedBuilder(path)
	if err != nil {
		return nil, err
	}
	err = b.Batch(func(tx *diskgraph.UndirectedTx) error {
		for i := int64(0); i < int64(n); i++ {
			if err := tx.AddNode(i); err != nil {
				return err
			}
		}
		for i := int64(0); i < int64(n); i++ {
			for j := i + 1; j < int64(n); j++ {
				if rng.Float64() < p {
					if err := tx.AddEdge(i, j); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		b.Close()
		return nil, err
	}
	if err := b.Close(); err != nil {
		return nil, err
	}
	return diskgraph.OpenUndirected(path)
}

// DiskBarabasiAlbert generates a scale-free undirected disk-backed graph.
func DiskBarabasiAlbert(n, m int, rng *rand.Rand, dir string) (*diskgraph.Undirected, error) {
	path := filepath.Join(dir, "ba.db")
	b, err := diskgraph.NewUndirectedBuilder(path)
	if err != nil {
		return nil, err
	}
	if n <= 0 || m <= 0 {
		b.Close()
		return diskgraph.OpenUndirected(path)
	}

	m0 := m + 1
	if m0 > n {
		m0 = n
	}

	degree := make([]int, n)
	totalDegree := 0
	for i := 0; i < m0; i++ {
		degree[i] = m0 - 1
		totalDegree += m0 - 1
	}

	err = b.Batch(func(tx *diskgraph.UndirectedTx) error {
		// Start with m0 fully connected nodes.
		for i := int64(0); i < int64(m0); i++ {
			for j := i + 1; j < int64(m0); j++ {
				if err := tx.AddEdge(i, j); err != nil {
					return err
				}
			}
		}

		// Add remaining nodes via preferential attachment.
		for i := m0; i < n; i++ {
			if err := tx.AddNode(int64(i)); err != nil {
				return err
			}
			targets := make(map[int]bool)
			for len(targets) < m && len(targets) < i {
				r := rng.Intn(totalDegree)
				cumulative := 0
				for j := 0; j < i; j++ {
					cumulative += degree[j]
					if r < cumulative {
						targets[j] = true
						break
					}
				}
			}
			for t := range targets {
				if err := tx.AddEdge(int64(i), int64(t)); err != nil {
					return err
				}
				degree[i]++
				degree[t]++
				totalDegree += 2
			}
		}
		return nil
	})
	if err != nil {
		b.Close()
		return nil, err
	}

	if err := b.Close(); err != nil {
		return nil, err
	}
	return diskgraph.OpenUndirected(path)
}

// DiskTwoClusterGraph generates a two-cluster disk-backed graph.
func DiskTwoClusterGraph(clusterSize int, pIn, pOut float64, rng *rand.Rand, dir string) (*diskgraph.Undirected, error) {
	path := filepath.Join(dir, "tc.db")
	b, err := diskgraph.NewUndirectedBuilder(path)
	if err != nil {
		return nil, err
	}
	n := int64(clusterSize)

	err = b.Batch(func(tx *diskgraph.UndirectedTx) error {
		for i := int64(0); i < n; i++ {
			for j := i + 1; j < n; j++ {
				if rng.Float64() < pIn {
					if err := tx.AddEdge(i, j); err != nil {
						return err
					}
				}
			}
		}
		for i := n; i < 2*n; i++ {
			for j := i + 1; j < 2*n; j++ {
				if rng.Float64() < pIn {
					if err := tx.AddEdge(i, j); err != nil {
						return err
					}
				}
			}
		}
		for i := int64(0); i < n; i++ {
			for j := n; j < 2*n; j++ {
				if rng.Float64() < pOut {
					if err := tx.AddEdge(i, j); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		b.Close()
		return nil, err
	}

	if err := b.Close(); err != nil {
		return nil, err
	}
	return diskgraph.OpenUndirected(path)
}

// DiskWeightedErdosRenyi generates a random weighted undirected disk-backed graph.
func DiskWeightedErdosRenyi(n int, p float64, maxWeight float64, rng *rand.Rand, dir string) (*diskgraph.Undirected, error) {
	path := filepath.Join(dir, "wer.db")
	b, err := diskgraph.NewUndirectedBuilder(path)
	if err != nil {
		return nil, err
	}
	err = b.Batch(func(tx *diskgraph.UndirectedTx) error {
		for i := int64(0); i < int64(n); i++ {
			if err := tx.AddNode(i); err != nil {
				return err
			}
		}
		for i := int64(0); i < int64(n); i++ {
			for j := i + 1; j < int64(n); j++ {
				if rng.Float64() < p {
					w := rng.Float64() * maxWeight
					if err := tx.AddWeightedEdge(i, j, w); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		b.Close()
		return nil, err
	}
	if err := b.Close(); err != nil {
		return nil, err
	}
	return diskgraph.OpenUndirected(path)
}

// DiskKarateClub builds the Zachary Karate Club as a disk-backed graph.
func DiskKarateClub(dir string) (*diskgraph.Undirected, error) {
	path := filepath.Join(dir, "karate.db")
	b, err := diskgraph.NewUndirectedBuilder(path)
	if err != nil {
		return nil, err
	}
	edges := [][2]int64{
		{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8},
		{0, 10}, {0, 11}, {0, 12}, {0, 13}, {0, 17}, {0, 19}, {0, 21}, {0, 31},
		{1, 2}, {1, 3}, {1, 7}, {1, 13}, {1, 17}, {1, 19}, {1, 21}, {1, 30},
		{2, 3}, {2, 7}, {2, 8}, {2, 9}, {2, 13}, {2, 27}, {2, 28}, {2, 32},
		{3, 7}, {3, 12}, {3, 13},
		{4, 6}, {4, 10},
		{5, 6}, {5, 10}, {5, 16},
		{6, 16},
		{8, 30}, {8, 32}, {8, 33},
		{9, 33},
		{13, 33},
		{14, 32}, {14, 33},
		{15, 32}, {15, 33},
		{18, 32}, {18, 33},
		{19, 33},
		{20, 32}, {20, 33},
		{22, 32}, {22, 33},
		{23, 25}, {23, 27}, {23, 29}, {23, 32}, {23, 33},
		{24, 25}, {24, 27}, {24, 31},
		{25, 31},
		{26, 29}, {26, 33},
		{27, 33},
		{28, 31}, {28, 33},
		{29, 32}, {29, 33},
		{30, 32}, {30, 33},
		{31, 32}, {31, 33},
		{32, 33},
	}
	err = b.Batch(func(tx *diskgraph.UndirectedTx) error {
		for _, e := range edges {
			if err := tx.AddEdge(e[0], e[1]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		b.Close()
		return nil, err
	}
	if err := b.Close(); err != nil {
		return nil, err
	}
	return diskgraph.OpenUndirected(path)
}

// diskTempDir creates a temporary directory for disk graph files.
func diskTempDir(prefix string) (string, func()) {
	dir, err := os.MkdirTemp("", "diskgraph-"+prefix+"-*")
	if err != nil {
		panic(err)
	}
	return dir, func() { os.RemoveAll(dir) }
}
