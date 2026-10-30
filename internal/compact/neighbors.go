// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package compact implements exact, immutable graph storage primitives.
package compact

import (
	"encoding/binary"
	"slices"
	"strconv"

	"gonum.org/v1/gonum/graph"
)

// Each block starts with an absolute dense index. Signed delta varints permit
// blocks to cross CSR row boundaries. Gap encoding and lazy traversal are
// discussed in Boldi and Vigna, "The WebGraph Framework I: Compression
// Techniques", WWW 2004: https://vigna.di.unimi.it/ftp/papers/WebGraphI.pdf
// The 32-entry restarts and signed varint layout are local design choices;
// this does not implement WebGraph's reference or interval compression.
const blockSize = 32

// NeighborEncoding selects an exact representation of dense neighbor indexes.
type NeighborEncoding uint8

const (
	FlatNeighbors NeighborEncoding = iota
	DeltaNeighbors
	Packed24Neighbors
)

// Neighbors stores sorted CSR rows, flat, packed into 24 bits, or delta encoded.
// Offsets and edge positions use int so edge counts are not restricted to
// signed 32 bits.
type Neighbors struct {
	offsets   []int
	flat      []int32
	packed24  []byte
	bases     []int32
	positions []int
	data      []byte
}

// NewNeighbors takes ownership of offsets and sorted targets. Compression is
// retained only when its complete payload is smaller than the flat targets.
//
// Fixed-width packing draws on Besta et al., "Log(Graph): A Near-Optimal
// High-Performance Graph Representation", PACT 2018, Section 3.2.1:
// https://arxiv.org/abs/2010.15879
// Our byte-aligned 24-bit mode and int32 fallback are local choices rather
// than the paper's arbitrary bit widths. Implemented independently from papers.
func NewNeighbors(offsets []int, targets []int32, encoding NeighborEncoding) *Neighbors {
	a := &Neighbors{offsets: offsets, flat: targets}
	if len(targets) == 0 {
		return a
	}
	if encoding == Packed24Neighbors {
		// Validate before allocating: no truncation, including negative indexes.
		for _, v := range targets {
			if v < 0 || v > 0xffffff {
				return a
			}
		}
		a.packed24 = make([]byte, 3*len(targets))
		for i, v := range targets {
			p := 3 * i
			a.packed24[p], a.packed24[p+1], a.packed24[p+2] = byte(v), byte(v>>8), byte(v>>16)
		}
		a.flat = nil
		return a
	}
	if encoding != DeltaNeighbors {
		return a
	}
	blocks := (len(targets)-1)/blockSize + 1
	size := 0
	var buf [binary.MaxVarintLen32]byte
	for i, v := range targets {
		if i%blockSize != 0 {
			size += binary.PutVarint(buf[:], int64(v)-int64(targets[i-1]))
		}
	}
	// positions are machine ints; use their actual size without unsafe.
	if int64(size)+int64(blocks)*(4+strconv.IntSize/8) >= int64(len(targets))*4 {
		return a
	}
	a.bases = make([]int32, blocks)
	a.positions = make([]int, blocks)
	a.data = make([]byte, size)
	p := 0
	for i, v := range targets {
		if i%blockSize == 0 {
			a.bases[i/blockSize], a.positions[i/blockSize] = v, p
		} else {
			p += binary.PutVarint(a.data[p:], int64(v)-int64(targets[i-1]))
		}
	}
	a.flat = nil
	return a
}

func (a *Neighbors) at24(pos int) int32 {
	p := 3 * pos
	return int32(a.packed24[p]) | int32(a.packed24[p+1])<<8 | int32(a.packed24[p+2])<<16
}

// Bounds returns the edge-position range of row i, or an empty range.
func (a *Neighbors) Bounds(i int) (int, int) {
	if i < 0 || i >= len(a.offsets)-1 {
		return 0, 0
	}
	return a.offsets[i], a.offsets[i+1]
}

func (a *Neighbors) at(pos int) (int32, int) {
	if a.flat != nil {
		return a.flat[pos], 0
	}
	b := pos / blockSize
	v, p := a.bases[b], a.positions[b]
	for k := b * blockSize; k < pos; k++ {
		d, n := binary.Varint(a.data[p:])
		p += n
		v += int32(d)
	}
	return v, p
}

// Row returns a read-only view for flat storage, or a fresh decoded row for
// compressed storage. Decoded rows are never cached in the graph.
func (a *Neighbors) Row(i int) []int32 {
	lo, hi := a.Bounds(i)
	if lo == hi {
		return nil
	}
	if a.flat != nil {
		return a.flat[lo:hi]
	}
	row := make([]int32, 0, hi-lo)
	c := a.Cursor(i)
	for c.Next() {
		row = append(row, c.Value())
	}
	return row
}

// Find returns the global edge position of target within sorted row i, or -1.
func (a *Neighbors) Find(i int, target int32) int {
	lo, hi := a.Bounds(i)
	if a.flat != nil {
		p, ok := slices.BinarySearch(a.flat[lo:hi], target)
		if ok {
			return lo + p
		}
		return -1
	}
	if lo == hi {
		return -1
	}
	if a.packed24 != nil {
		end := hi
		for lo < hi {
			mid := lo + (hi-lo)/2
			if a.at24(mid) < target {
				lo = mid + 1
			} else {
				hi = mid
			}
		}
		if lo < end && a.at24(lo) == target {
			return lo
		}
		return -1
	}
	// Binary search restart values within this row, then decode at most one
	// block plus the next restart. The previous block may contain duplicates
	// of the matching restart, so start there to find the first occurrence.
	first, end := (lo+blockSize-1)/blockSize, (hi-1)/blockSize+1
	b, _ := slices.BinarySearch(a.bases[first:end], target)
	b += first
	c := Cursor{a: a, pos: max(lo, (b-1)*blockSize) - 1, end: min(hi, b*blockSize+1), nextByte: -1}
	for c.Next() {
		if c.Value() == target {
			return c.Position()
		}
		if c.Value() > target {
			break
		}
	}
	return -1
}

// Cursor is an independent, allocation-free row decoder.
type Cursor struct {
	a                  *Neighbors
	pos, end, nextByte int
	value              int32
}

// Cursor returns a decoder positioned before the first neighbor of row i.
func (a *Neighbors) Cursor(i int) Cursor {
	lo, hi := a.Bounds(i)
	return Cursor{a: a, pos: lo - 1, end: hi, nextByte: -1}
}

// Reset reuses the decoder for a different row without allocating.
func (c *Cursor) Reset(i int) {
	if c.a == nil {
		*c = Cursor{pos: -1}
		return
	}
	*c = c.a.Cursor(i)
}

// Neighbor returns the current dense neighbor index.
func (c *Cursor) Neighbor() int32 { return c.value }

// Next advances the decoder.
func (c *Cursor) Next() bool {
	c.pos++
	if c.pos >= c.end {
		return false
	}
	if c.a.flat != nil {
		c.value = c.a.flat[c.pos]
		return true
	}
	if c.a.packed24 != nil {
		c.value = c.a.at24(c.pos)
		return true
	}
	if c.nextByte < 0 || c.pos%blockSize == 0 {
		c.value, c.nextByte = c.a.at(c.pos)
	} else {
		d, n := binary.Varint(c.a.data[c.nextByte:])
		c.nextByte += n
		c.value += int32(d)
	}
	return true
}

// Value returns the current dense neighbor index.
func (c *Cursor) Value() int32 { return c.value }

// Position returns the current global edge position, for parallel weights.
func (c *Cursor) Position() int { return c.pos }

// StorageBytes returns the retained slice payload, excluding object headers.
func (a *Neighbors) StorageBytes() int64 {
	intBytes := int64(strconv.IntSize / 8)
	return int64(cap(a.offsets)+cap(a.positions))*intBytes + int64(cap(a.flat)+cap(a.bases))*4 + int64(cap(a.data)) + int64(cap(a.packed24))
}

// Nodes maps dense neighbor indexes back to original IDs lazily.
type Nodes struct {
	ids             []int64
	initial, cursor Cursor
	valid           bool
}

// NewNodes creates a gonum iterator without materializing original neighbor IDs.
func NewNodes(ids []int64, cursor Cursor) graph.Nodes {
	return &Nodes{ids: ids, initial: cursor, cursor: cursor}
}

func (it *Nodes) Next() bool { it.valid = it.cursor.Next(); return it.valid }
func (it *Nodes) Len() int   { return max(0, it.cursor.end-it.cursor.pos-1) }
func (it *Nodes) Reset()     { it.cursor = it.initial; it.valid = false }
func (it *Nodes) Node() graph.Node {
	if !it.valid {
		return nil
	}
	return node(it.ids[it.cursor.Value()])
}

type node int64

func (n node) ID() int64 { return int64(n) }
