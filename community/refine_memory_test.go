// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package community

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"runtime"
	"testing"

	"github.com/intelligrit/graphwizard/densegraph"
)

func refinementFixture(n, groups int) (weightedAdj, []float64, []int) {
	b := densegraph.NewUndirectedBuilderSized(n, n*10)
	for i := n - 1; i >= 0; i-- {
		b.AddNode(int64(i))
	}
	comm := make([]int, n)
	for i := range n {
		comm[i] = i % groups
		if i >= n-5 {
			comm[i] = i
			// Singleton communities consume no seeds.
			continue
		}
		for step := 1; step <= 4; step++ {
			j := i + step*groups
			if j < n-5 {
				b.AddWeightedEdge(int64(i), int64(j), float64((i+j)%5))
			}
		}
		// Neighbors from other communities must not access their scratch slots.
		if i+1 < n-5 {
			b.AddWeightedEdge(int64(i), int64(i+1), 10.25)
		}
	}
	_, adj, degree, _ := buildWeightedAdj(b.Build())
	return adj, degree, comm
}

func refinementDigest(labels []int) string {
	h := sha256.New()
	var buf [8]byte
	for _, label := range labels {
		binary.LittleEndian.PutUint64(buf[:], uint64(label))
		_, _ = h.Write(buf[:])
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestRefinementReusePreservesResults(t *testing.T) {
	adj, degree, comm := refinementFixture(96, 7)
	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)
	for _, tc := range []struct {
		seed   int64
		digest string
		next   int64
	}{
		// Captured before sharing scratch buffers or reusing RNG state.
		{0, "204de2c5f4e6a5f30e28fbfc736f2b240813fa454e74409a63452f78bf977091", 6044372234677422456},
		{42, "c6bde8b6f547396a98200dbf5cb1dbbe9fc8b560a4f6421d0a41a501221857a5", 3545887102062614208},
		{-9, "a75176bb8362d4d494c746f640f7b72da895cfcea36a3ad21e73349fe18c2be6", 2825193896916117448},
	} {
		t.Run(fmt.Sprint(tc.seed), func(t *testing.T) {
			for _, workers := range []int{0, 1, 2, 8} {
				rng := rand.New(rand.NewSource(tc.seed))
				var got []int
				if workers == 0 {
					got = refine(adj, degree, comm, len(comm), rng)
				} else {
					runtime.GOMAXPROCS(workers)
					got = refineParallel(adj, degree, comm, len(comm), rng)
				}
				digest, next := refinementDigest(got), rng.Int63()
				if digest != tc.digest || next != tc.next {
					t.Fatalf("workers=%d digest=%s next=%d", workers, digest, next)
				}
				for i, label := range got {
					if comm[label] != comm[i] {
						t.Fatal("refinement crossed original community boundary")
					}
				}
			}
		})
	}
}

func BenchmarkRefinementScratch(b *testing.B) {
	adj, degree, comm := refinementFixture(10000, 1250)
	for _, parallel := range []bool{false, true} {
		name := "serial"
		if parallel {
			name = "parallel"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				rng := rand.New(rand.NewSource(42))
				var labels []int
				if parallel {
					labels = refineParallel(adj, degree, comm, len(comm), rng)
				} else {
					labels = refine(adj, degree, comm, len(comm), rng)
				}
				if len(labels) != len(comm) {
					b.Fatal("missing nodes")
				}
			}
		})
	}
}
