// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package stream_test

import (
	"fmt"

	"github.com/intelligrit/graphwizard/stream"
)

func ExampleStreamGraph() {
	sg := stream.New()
	sg.AddNode(1)
	sg.AddNode(2)
	sg.AddEdge(1, 2, 1.5)

	fmt.Printf("changes: %d\n", len(sg.Changes()))
	sg.Flush()
	fmt.Printf("after flush: %d\n", len(sg.Changes()))

	g := sg.Graph()
	w, _ := g.Weight(1, 2)
	fmt.Printf("weight(1,2): %.1f\n", w)
	// Output:
	// changes: 3
	// after flush: 0
	// weight(1,2): 1.5
}
