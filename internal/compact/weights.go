// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package compact

import (
	"encoding/binary"
	"math"
	"strconv"
)

// Block boundaries depend only on edge position, keeping random access O(1).
const weightBlockSize = 128

// A kind is also the number of bytes per entry; constants store one float64.
const (
	constantWeight byte = 0
	byteWeight     byte = 1
	shortWeight    byte = 2
	intWeight      byte = 4
	floatWeight    byte = 8
)

// Weights preserves float64 bit patterns, using a constant or the narrowest
// unsigned integer array when possible. All other values retain float64 storage.
type Weights struct {
	constant     float64
	u8           []uint8
	u16          []uint16
	u32          []uint32
	f64          []float64
	blockOffsets []int
	blockKinds   []byte
	blockData    []byte
}

// NewWeights reads n weights from get without requiring a temporary float array.
// Width selection draws on Besta et al., "Log(Graph): A Near-Optimal
// High-Performance Graph Representation", PACT 2018, Section 3.3:
// https://arxiv.org/abs/2010.15879
// Here integer widths are byte-aligned; other values retain exact float64 bits.
func NewWeights(n int, get func(int) float64) Weights {
	return newWeights(n, get, false)
}

// NewBlockWeights chooses an exact representation per 128-entry block, but
// falls back to graph-wide encoding unless the complete payload is smaller.
// get must be repeatable: sizing and encoding make multiple passes.
//
// Group-level width selection is inspired by Shun, Dhulipala, and Blelloch,
// "Smaller and Faster: Parallel Processing of Compressed Graphs with Ligra+",
// DCC 2015, Section 3 (Encoding and Weighted Graphs):
// https://www.cs.cmu.edu/~blelloch/papers/SDB15.pdf
// Fixed 128-entry blocks, constants, float64 fallback, and separate weight
// storage are local choices. No external implementation is used.
func NewBlockWeights(n int, get func(int) float64) Weights {
	return newWeights(n, get, true)
}

func newWeights(n int, get func(int) float64, blockwise bool) Weights {
	if n == 0 {
		return Weights{constant: 1}
	}
	kind, first := classifyWeights(0, n, get)
	if kind == constantWeight {
		return Weights{constant: first}
	}
	if blockwise {
		if s, ok := buildBlockWeights(n, get, int64(n)*int64(kind)); ok {
			return s
		}
	}
	var s Weights
	switch kind {
	case byteWeight:
		s.u8 = make([]uint8, n)
	case shortWeight:
		s.u16 = make([]uint16, n)
	case intWeight:
		s.u32 = make([]uint32, n)
	default:
		s.f64 = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		w := get(i)
		switch kind {
		case byteWeight:
			s.u8[i] = uint8(w)
		case shortWeight:
			s.u16[i] = uint16(w)
		case intWeight:
			s.u32[i] = uint32(w)
		default:
			s.f64[i] = w
		}
	}
	return s
}

func classifyWeights(start, end int, get func(int) float64) (byte, float64) {
	first := get(start)
	constant, integral := true, true
	maxValue := 0.0
	for i := start; i < end; i++ {
		w := get(i)
		constant = constant && math.Float64bits(w) == math.Float64bits(first)
		if !(w >= 0 && w <= math.MaxUint32 && math.Trunc(w) == w) || math.Signbit(w) {
			integral = false
		}
		maxValue = max(maxValue, w)
	}
	if constant {
		return constantWeight, first
	}
	switch {
	case integral && maxValue <= math.MaxUint8:
		return byteWeight, first
	case integral && maxValue <= math.MaxUint16:
		return shortWeight, first
	case integral:
		return intWeight, first
	default:
		return floatWeight, first
	}
}

func buildBlockWeights(n int, get func(int) float64, globalBytes int64) (Weights, bool) {
	blocks := (n-1)/weightBlockSize + 1
	size := 0
	for start := 0; start < n; start += weightBlockSize {
		end := min(n, start+weightBlockSize)
		kind, _ := classifyWeights(start, end, get)
		if kind == constantWeight {
			size += 8
		} else {
			size += (end - start) * int(kind)
		}
	}
	// One offset and one kind byte per block; include all retained metadata.
	if int64(size)+int64(blocks)*(strconv.IntSize/8+1) >= globalBytes {
		return Weights{}, false
	}
	s := Weights{
		blockOffsets: make([]int, blocks),
		blockKinds:   make([]byte, blocks),
		blockData:    make([]byte, size),
	}
	p := 0
	for b := range blocks {
		start, end := b*weightBlockSize, min(n, (b+1)*weightBlockSize)
		kind, first := classifyWeights(start, end, get)
		s.blockOffsets[b], s.blockKinds[b] = p, kind
		if kind == constantWeight {
			binary.LittleEndian.PutUint64(s.blockData[p:], math.Float64bits(first))
			p += 8
			continue
		}
		for i := start; i < end; i++ {
			w := get(i)
			switch kind {
			case byteWeight:
				s.blockData[p] = byte(w)
			case shortWeight:
				binary.LittleEndian.PutUint16(s.blockData[p:], uint16(w))
			case intWeight:
				binary.LittleEndian.PutUint32(s.blockData[p:], uint32(w))
			default:
				binary.LittleEndian.PutUint64(s.blockData[p:], math.Float64bits(w))
			}
			p += int(kind)
		}
	}
	return s, true
}

// At returns the exact original weight at edge position i.
func (s *Weights) At(i int) float64 {
	switch {
	case s.u8 != nil:
		return float64(s.u8[i])
	case s.u16 != nil:
		return float64(s.u16[i])
	case s.u32 != nil:
		return float64(s.u32[i])
	case s.f64 != nil:
		return s.f64[i]
	case s.blockKinds != nil:
		b := i / weightBlockSize
		kind := s.blockKinds[b]
		p := s.blockOffsets[b] + (i%weightBlockSize)*int(kind)
		switch kind {
		case byteWeight:
			return float64(s.blockData[p])
		case shortWeight:
			return float64(binary.LittleEndian.Uint16(s.blockData[p:]))
		case intWeight:
			return float64(binary.LittleEndian.Uint32(s.blockData[p:]))
		default: // constant or float64
			return math.Float64frombits(binary.LittleEndian.Uint64(s.blockData[p:]))
		}
	default:
		return s.constant
	}
}

// StorageBytes returns the retained array payload, excluding object headers.
func (s *Weights) StorageBytes() int64 {
	return int64(cap(s.u8)) + 2*int64(cap(s.u16)) + 4*int64(cap(s.u32)) + 8*int64(cap(s.f64)) +
		int64(cap(s.blockOffsets))*(strconv.IntSize/8) + int64(cap(s.blockKinds)) + int64(cap(s.blockData))
}
