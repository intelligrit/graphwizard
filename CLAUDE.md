# GraphWizard

## Project

MIT-licensed Go library providing graph algorithms compatible with gonum/graph interfaces.
Cleanroom reimplementation — learn from papers and algorithm descriptions, never copy code.

## Key Rules

- Every `.go` file needs the copyright header:
  `// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.`
- All algorithms MUST be implemented from academic papers and algorithm descriptions only
- NEVER copy code from any existing graph library (igraph, NetworkX, etc.) or any BSL/GPL source
- All public APIs must work with gonum/graph interfaces (graph.Graph, graph.Undirected, etc.)
- Every algorithm must have tests with known-answer graphs
- Run `go test ./...` before declaring done
- MIT license only — no dependencies with incompatible licenses
