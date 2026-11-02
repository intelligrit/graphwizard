// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package paths

import (
	"context"
	"errors"

	"gonum.org/v1/gonum/graph"
)

// ExpandOptions bounds and filters path expansion from a start node.
// Hops are edges, so a path containing only the start node has zero hops.
type ExpandOptions struct {
	MinHops  int
	MaxHops  int
	MaxPaths int

	// Target, when non-nil, causes ExpandPaths to emit only paths ending at
	// that node. Expansion stops at the target.
	Target *int64

	// AllowStep, when non-nil, decides whether the directed step from one
	// node ID to the next may be taken. The start node is not filtered.
	AllowStep func(from, to int64) bool
}

type expandFrame struct {
	id        int64
	entered   bool
	neighbors graph.Nodes
}

// ExpandPaths visits simple paths from start in depth-first order. A simple
// path never visits the same node twice. Directed graphs follow outgoing
// edges; undirected graphs may traverse an edge in either direction.
//
// MaxHops must be at least MinHops, and MaxPaths must be positive. Each path
// passed to visit is an independent slice that the caller may retain. Returning
// false from visit stops expansion. The result counts paths passed to visit.
// A missing start or target yields zero paths without an error.
//
// ExpandPaths does not copy the graph or collect the results. Its working set
// contains the current path, its visited nodes, and one neighbor iterator per
// active hop. A graph implementation may buffer neighbors inside an iterator.
// The graph must not be mutated during traversal.
func ExpandPaths(ctx context.Context, g graph.Graph, start int64, opts ExpandOptions, visit func([]int64) bool) (int, error) {
	if opts.MinHops < 0 || opts.MaxHops < opts.MinHops {
		return 0, errors.New("paths: require 0 <= MinHops <= MaxHops")
	}
	if opts.MaxPaths <= 0 {
		return 0, errors.New("paths: MaxPaths must be positive")
	}
	if visit == nil {
		return 0, errors.New("paths: visit must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if g.Node(start) == nil || (opts.Target != nil && g.Node(*opts.Target) == nil) {
		return 0, nil
	}

	current := []int64{start}
	onPath := map[int64]bool{start: true}
	stack := []expandFrame{{id: start}}
	emitted := 0

	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			return emitted, err
		}
		frame := &stack[len(stack)-1]
		if !frame.entered {
			frame.entered = true
			hops := len(current) - 1
			atTarget := opts.Target != nil && frame.id == *opts.Target
			if hops >= opts.MinHops && (opts.Target == nil || atTarget) {
				path := append([]int64(nil), current...)
				emitted++
				if !visit(path) || emitted == opts.MaxPaths {
					return emitted, nil
				}
			}
			if hops == opts.MaxHops || atTarget {
				delete(onPath, frame.id)
				current = current[:len(current)-1]
				stack = stack[:len(stack)-1]
				continue
			}
			frame.neighbors = g.From(frame.id)
		}

		if frame.neighbors.Next() {
			next := frame.neighbors.Node().ID()
			if onPath[next] || (opts.AllowStep != nil && !opts.AllowStep(frame.id, next)) {
				continue
			}
			onPath[next] = true
			current = append(current, next)
			stack = append(stack, expandFrame{id: next})
			continue
		}
		delete(onPath, frame.id)
		current = current[:len(current)-1]
		stack = stack[:len(stack)-1]
	}
	return emitted, nil
}
