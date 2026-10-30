// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package loader

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/intelligrit/graphwizard/multigraph"
)

// LoadRelationships executes a query returning (from_id, to_id, type, year,
// weights...). weightNames names the remaining columns in order; each name must
// be nonempty and unique. Every row becomes a distinct relationship, including
// rows with identical endpoints, type, and year. Use ORDER BY when stable record
// IDs or floating-point reduction order matter.
//
// NULL weights remain missing; NULL type/year become empty type/zero year. Years
// must fit int32. Values are represented as the float64 returned by database/sql.
// Nodes are inferred from endpoints. To retain isolated nodes or preallocate for
// large inputs, use a multigraph.Builder with AddNode and ScanRelationships.
// Build options affect the stored columns, not subsequent graph projections.
func LoadRelationships(db *sql.DB, query string, weightNames []string, opts ...multigraph.BuildOption) (*multigraph.Store, error) {
	if err := validateWeightNames(weightNames); err != nil {
		return nil, err
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("loader: relationship query: %w", err)
	}
	defer rows.Close()
	b := multigraph.NewBuilder()
	if err := ScanRelationships(rows, b, weightNames); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("loader: close relationship rows: %w", err)
	}
	return b.Build(opts...), nil
}

// ScanRelationships appends rows with the layout documented by LoadRelationships
// to b. It registers all named weights, including entirely NULL columns. The
// caller owns the rows and must close them if needed. On error, earlier rows
// remain in the builder; callers requiring all-or-nothing loading should discard
// it. This function does not freeze the builder or change existing relationships.
func ScanRelationships(rows RowScanner, b *multigraph.Builder, weightNames []string) error {
	if err := validateWeightNames(weightNames); err != nil {
		return err
	}
	for _, name := range weightNames {
		b.DefineWeight(name)
	}
	var from, to int64
	var kind sql.NullString
	var year sql.NullInt64
	values := make([]sql.NullFloat64, len(weightNames))
	dest := make([]interface{}, 4+len(weightNames))
	dest[0], dest[1], dest[2], dest[3] = &from, &to, &kind, &year
	for i := range values {
		dest[4+i] = &values[i]
	}
	weights := make(map[string]float64, len(weightNames))
	row := 0
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return fmt.Errorf("loader: relationship row %d: %w", row, err)
		}
		if year.Int64 < math.MinInt32 || year.Int64 > math.MaxInt32 {
			return fmt.Errorf("loader: relationship row %d year %d does not fit int32", row, year.Int64)
		}
		clear(weights)
		for i, name := range weightNames {
			if values[i].Valid {
				weights[name] = values[i].Float64
			}
		}
		b.AddRelationship(from, to, multigraph.Attributes{Type: kind.String, Year: int32(year.Int64)}, weights)
		row++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("loader: read relationships: %w", err)
	}
	return nil
}

func validateWeightNames(names []string) error {
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if name == "" {
			return fmt.Errorf("loader: empty relationship weight name")
		}
		if seen[name] {
			return fmt.Errorf("loader: duplicate relationship weight name %q", name)
		}
		seen[name] = true
	}
	return nil
}
