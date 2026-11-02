// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package multigraph preserves distinct relationships with named weights and
// optional type/year attributes, then projects them into gonum-compatible graphs
// for analysis. Repeated endpoints never overwrite a relationship in the store.
//
// A Store is an immutable relationship table, not an implicit scalar-weight
// graph. ProjectUndirected explicitly selects a weight and a rule for combining
// parallel relationships before returning a densegraph.Undirected. Existing
// single-weight graph builders and their storage costs are unchanged.
//
// Relationship IDs are insertion positions within one store. Endpoint direction
// is preserved in the store; an undirected projection combines both orientations.
// Type and year are application-supplied labels, not inferred temporal semantics.
// Zero year conventionally denotes an undated relationship.
// VisitAnnualChanges compares two explicitly selected positive years without
// constructing or retaining two graph projections. VisitAnnualTrajectories
// scores sustained provider-level changes across a requested annual series.
package multigraph
