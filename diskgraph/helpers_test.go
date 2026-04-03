// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph_test

import (
	"os"

	"gonum.org/v1/gonum/graph"
)

func tempDir() string {
	d, err := os.MkdirTemp("", "diskgraph-example-*")
	if err != nil {
		panic(err)
	}
	return d
}

func collectNodeIDs(it graph.Nodes) []int64 {
	var ids []int64
	for it.Next() {
		ids = append(ids, it.Node().ID())
	}
	return ids
}
