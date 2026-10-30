// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"math"
	"path/filepath"
	"slices"
	"testing"

	"github.com/intelligrit/graphwizard/internal/compact"
)

func TestCompressedPreloadMatchesSQL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "compressed.db")
	b, err := NewUndirectedBuilder(path)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Batch(func(tx *UndirectedTx) error {
		for i := range 128 {
			if err := tx.AddNode(int64(i)*1000003 - 9000000); err != nil {
				return err
			}
			for j := i + 1; j < min(128, i+17); j++ {
				if err := tx.AddWeightedEdge(int64(i)*1000003-9000000, int64(j)*1000003-9000000, float64(i+j)); err != nil {
					return err
				}
			}
		}
		return tx.AddNode(math.MaxInt64)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	sql, err := OpenUndirected(path)
	if err != nil {
		t.Fatal(err)
	}
	defer sql.Close()
	flat, err := OpenUndirected(path, ForcePreload)
	if err != nil {
		t.Fatal(err)
	}
	defer flat.Close()
	packed, err := OpenUndirected(path, CompressedPreload, ForcePreload)
	if err != nil {
		t.Fatal(err)
	}
	defer packed.Close()
	if packed.AdjacencyBytes() >= flat.AdjacencyBytes() {
		t.Fatal("fixture should compress")
	}
	packed24, err := OpenUndirected(path, Packed24Preload, ForcePreload)
	if err != nil {
		t.Fatal(err)
	}
	defer packed24.Close()
	if packed24.AdjacencyBytes() >= flat.AdjacencyBytes() {
		t.Fatal("24-bit packing should save memory")
	}
	for i, id := range flat.NodeIDs() {
		want := collectIDs(sql.From(id))
		for _, g := range []*Undirected{flat, packed, packed24} {
			if got := collectIDs(g.From(id)); !slices.Equal(got, want) {
				t.Fatalf("From(%d) differs", id)
			}
			if !slices.Equal(g.DenseNeighbors(i), flat.DenseNeighbors(i)) {
				t.Fatal("dense row differs")
			}
			var streamed []int32
			for j := range g.Neighbors(i) {
				streamed = append(streamed, j)
			}
			if !slices.Equal(streamed, flat.DenseNeighbors(i)) {
				t.Fatal("streamed row differs")
			}
			for _, dst := range flat.NodeIDs() {
				if g.HasEdgeBetween(id, dst) != sql.HasEdgeBetween(id, dst) {
					t.Fatal("edge existence differs")
				}
			}
			it := g.From(id)
			if it.Node() != nil {
				t.Fatal("iterator initially positioned")
			}
			var first []int64
			for it.Next() {
				first = append(first, it.Node().ID())
			}
			it.Reset()
			var second []int64
			for it.Next() {
				second = append(second, it.Node().ID())
			}
			if !slices.Equal(first, second) {
				t.Fatal("reset differs")
			}
		}
	}
	for _, g := range []*Undirected{flat, packed, packed24} {
		cursor := g.NewNeighborCursor()
		for i := range g.NumNodes() {
			cursor.Reset(i)
			for _, want := range g.DenseNeighbors(i) {
				if !cursor.Next() || cursor.Neighbor() != want {
					t.Fatal("cursor differs")
				}
			}
			if cursor.Next() {
				t.Fatal("cursor not exhausted")
			}
		}
		cursor.Reset(-1)
		if cursor.Next() {
			t.Fatal("invalid cursor row")
		}
		if g.From(math.MinInt64).Next() || g.HasEdgeBetween(math.MinInt64, math.MaxInt64) {
			t.Fatal("missing ID exists")
		}
		for range g.Neighbors(math.MaxInt) {
			t.Fatal("invalid row yielded")
		}
		for range g.Neighbors(0) {
			break
		}
	}
	for range sql.Neighbors(0) {
		t.Fatal("unpreloaded graph yielded dense neighbors")
	}
	cursor := sql.NewNeighborCursor()
	cursor.Reset(0)
	if cursor.Next() {
		t.Fatal("unpreloaded cursor yielded")
	}
}

func TestPreloadEncodingOptionOrder(t *testing.T) {
	for _, tc := range []struct {
		opts []Option
		want compact.NeighborEncoding
	}{
		{[]Option{Preload}, compact.FlatNeighbors},
		{[]Option{CompressedPreload, Packed24Preload, ForcePreload, Preload}, compact.Packed24Neighbors},
		{[]Option{Packed24Preload, CompressedPreload, ForcePreload}, compact.DeltaNeighbors},
	} {
		var cfg openConfig
		for _, opt := range tc.opts {
			opt(&cfg)
		}
		if !cfg.preload || cfg.neighbors != tc.want {
			t.Fatal("incorrect preload encoding")
		}
	}
}
