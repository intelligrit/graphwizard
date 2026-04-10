// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package progress provides a lightweight progress-reporting interface for
// graph algorithms. Callers attach a Reporter to a context.Context and pass
// that context to any algorithm. The algorithm calls Report at natural
// checkpoints (each iteration, each phase, etc.), and the Reporter receives
// those updates.
//
// Algorithms that receive a plain context.Background() incur only a single
// context.Value lookup (nil check) per Report call — effectively zero overhead.
//
// # Usage
//
//	handler := progress.Func(func(p progress.Progress) {
//	    fmt.Printf("[%s] %d/%d\n", p.Phase, p.Step, p.Total)
//	})
//	ctx := progress.With(context.Background(), handler)
//	result := community.Leiden(ctx, g, 1.0, rng)
package progress

import "context"

// Progress describes the current state of a running algorithm.
type Progress struct {
	// Phase names the current execution stage (e.g., "build", "iterate",
	// "refine", "converge"). Algorithms document their phase names.
	Phase string

	// Step is the 0-based index of the current step within Phase.
	Step int

	// Total is the expected number of steps in Phase, or -1 when unknown
	// (e.g., algorithms that run until convergence).
	Total int

	// Message is an optional human-readable description of the current step.
	Message string
}

// Reporter receives progress updates from a running algorithm.
type Reporter interface {
	Report(p Progress)
}

// Func is a function that satisfies Reporter.
type Func func(p Progress)

// Report implements Reporter.
func (f Func) Report(p Progress) { f(p) }

type contextKey struct{}

// With returns a copy of ctx carrying r. Pass the returned context to any
// algorithm to receive progress updates. Algorithms that receive a plain
// context.Background() emit no events.
func With(ctx context.Context, r Reporter) context.Context {
	return context.WithValue(ctx, contextKey{}, r)
}

// Report delivers p to the Reporter attached to ctx, if any.
// It is a no-op when ctx carries no Reporter.
func Report(ctx context.Context, p Progress) {
	if r, ok := ctx.Value(contextKey{}).(Reporter); ok {
		r.Report(p)
	}
}
