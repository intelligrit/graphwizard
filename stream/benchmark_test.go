// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package stream

import (
	"sync"
	"testing"
)

func BenchmarkAddEdge_10K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sg := New()
		for e := int64(0); e < 10000; e++ {
			sg.AddEdge(e, e+1, 1.0)
		}
	}
}

func BenchmarkRemoveEdge_1K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sg := New()
		// Build graph with 1K edges.
		for e := int64(0); e < 1000; e++ {
			sg.AddEdge(e, e+1, 1.0)
		}
		// Remove all edges.
		for e := int64(0); e < 1000; e++ {
			sg.RemoveEdge(e, e+1)
		}
	}
}

func BenchmarkConcurrentAddEdge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sg := New()
		var wg sync.WaitGroup
		for g := 0; g < 10; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				base := int64(goroutineID * 1000)
				for e := int64(0); e < 1000; e++ {
					sg.AddEdge(base+e, base+e+1, 1.0)
				}
			}(g)
		}
		wg.Wait()
	}
}
