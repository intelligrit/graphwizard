// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package loader

import (
	"testing"
)

// TestWriteRows_Empty verifies no error with empty rows.
func TestWriteRows_Empty(t *testing.T) {
	// WriteRows with nil rows slice should return nil immediately since
	// the loop body never executes.
	err := WriteRows(nil, "INSERT INTO t VALUES ($1)", nil)
	if err != nil {
		t.Errorf("expected no error for empty rows, got %v", err)
	}
}

// TestWriteRows_EmptySlice verifies no error with empty slice.
func TestWriteRows_EmptySlice(t *testing.T) {
	err := WriteRows(nil, "INSERT INTO t VALUES ($1)", [][]interface{}{})
	if err != nil {
		t.Errorf("expected no error for empty rows, got %v", err)
	}
}
