<p align="center">
  <img src="logo.svg" alt="GraphWizard" width="200">
</p>

<h1 align="center">GraphWizard</h1>

<p align="center">
  <strong>A complete graph algorithm library for Go</strong>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/intelligrit/graphwizard">
    <img src="https://pkg.go.dev/badge/github.com/intelligrit/graphwizard.svg" alt="Go Reference">
  </a>
  <img src="https://img.shields.io/badge/coverage-97.3%25-brightgreen" alt="Coverage">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License">
  <img src="https://img.shields.io/badge/go-%3E%3D1.22-00ADD8" alt="Go Version">
</p>

---

GraphWizard provides 40+ graph algorithms through a clean, consistent API
built on [gonum/graph](https://pkg.go.dev/gonum.org/v1/gonum/graph)
interfaces. Cleanroom implementations from academic papers, plus unified
wrappers around gonum's built-in algorithms — one import per domain,
no iterator gymnastics required.

**New:** The `diskgraph` package provides a bbolt-backed graph that is
**faster than in-memory gonum graphs** for most algorithms while also
supporting persistence and out-of-core processing. See
[Recommended: diskgraph](#recommended-diskgraph) below.

## Origin Story

GraphWizard was born out of the [ACT-IAC](https://actiac.org/) 2026 AI
Hackathon, where our team at [Intelligrit](https://intelligrit.com) built
**Integrity** — an AI-powered system that detects anomalous Medicare provider
billing patterns using only public data. We won the hackathon.

Integrity's graph analysis pipeline relied on Memgraph and its MAGE algorithm
library for fraud network detection: community clustering, centrality scoring,
bridge detection, and co-prescribing anomaly analysis across 5.8 million
provider nodes and 82.6 million edges. When we evaluated our long-term
architecture, we realized we needed a pure Go solution — no external graph
database, no Docker dependency, no Bolt protocol. Just algorithms that run
in-process on gonum graph structures loaded from DuckDB.

We looked at the Go ecosystem and found gonum covers the basics well, but
there's no comprehensive library that fills the gaps: no Leiden, no Katz, no
bipartite matching, no bridge detection, no MST, no embeddings. So we built
one — including `diskgraph`, a bbolt-backed graph store that turned out to be
faster than gonum's in-memory graphs for most algorithms, while also handling
our 82.6M edge dataset without requiring everything in RAM.

Every algorithm is implemented cleanroom from the original academic papers.
Every exported function has godoc, examples, and tests. The result is a
library we needed — and that we think the Go graph community needs too.

## Install

```
go get github.com/intelligrit/graphwizard
```

## Packages

| Package | Algorithms | Source |
|---|---|---|
| **diskgraph** | Disk-backed Undirected/Directed graphs with auto-preloading, bbolt storage | Custom |
| **centrality** | PageRank, Betweenness, Closeness, Harmonic, HITS, Katz, Degree, Personalized PageRank, Eccentricity, Diameter, Radius, Influence Maximization | Custom + gonum |
| **community** | Leiden, Louvain, Label Propagation, Spectral Clustering | Custom + gonum |
| **connectivity** | Bridges, Biconnected Components, Articulation Points, WCC, SCC, Cycles, Union-Find, DAG Condensation, K-Core, Degeneracy, Topological Sort | Custom + gonum |
| **embedding** | Node2Vec Walks, DeepWalk Walks, SVD Embedding | Custom |
| **flow** | Max Flow (Dinic), Min Cut | Custom + gonum |
| **matching** | Hopcroft-Karp Bipartite Matching | Custom |
| **paths** | Dijkstra, Bellman-Ford, Floyd-Warshall, A*, Yen's K Shortest | Custom + gonum |
| **similarity** | Jaccard, Overlap, Cosine, SimRank, Common Neighbors, Adamic-Adar, Preferential Attachment, Link Prediction | Custom |
| **structure** | Clustering Coefficient, Triangle Count, Cliques, Coloring, Set Cover, Kruskal MST, Prim MST, TSP, Bipartite Projection | Custom + gonum |
| **traverse** | BFS, DFS, BFS Path | gonum |

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/intelligrit/graphwizard/centrality"
    "github.com/intelligrit/graphwizard/community"
    "github.com/intelligrit/graphwizard/connectivity"
    "github.com/intelligrit/graphwizard/diskgraph"
)

func main() {
    // Build a graph on disk (persisted to a file).
    b, _ := diskgraph.NewUndirectedBuilder("example.db")
    b.Batch(func(tx *diskgraph.UndirectedTx) error {
        tx.AddEdge(0, 1)
        tx.AddEdge(1, 2)
        tx.AddEdge(2, 0)
        tx.AddEdge(2, 3)
        return nil
    })
    b.Close()

    // Open the graph — adjacency is auto-preloaded for speed.
    g, _ := diskgraph.OpenUndirected("example.db")
    defer g.Close()

    // Find bridge edges.
    bridges := connectivity.Bridges(g)
    fmt.Printf("Bridges: %d\n", len(bridges)) // 1 (edge 2-3)

    // Degree centrality.
    deg := centrality.Degree(g)
    fmt.Printf("Node 2 centrality: %.2f\n", deg[2]) // 1.00

    // Community detection.
    comms := community.Louvain(g, 1.0, nil)
    fmt.Printf("Communities: %v\n", comms)
}
```

## Recommended: diskgraph

We recommend `diskgraph` as the default graph storage for all GraphWizard
workloads. It replaces `gonum/graph/simple` with a bbolt-backed
implementation that is:

- **Faster** — benchmarks show diskgraph matches or beats in-memory
  `simple.UndirectedGraph` on 40+ algorithms. Algorithms like
  PreferentialAttachment run 4x faster; most run 10-30% faster.

- **Persistent** — build the graph once, reopen it instantly across
  program runs. No serialization, no import step. The bolt file is the
  database.

- **Scalable** — graphs too large for RAM work automatically. Pass
  `NoPreload` and the graph reads directly from disk via memory-mapped
  I/O.

- **Tunable** — pass `Preload` to cache adjacency data in Go memory
  for O(1) `HasEdgeBetween` on algorithms like ClusteringCoefficient.
  Without it, reads go through bolt's memory-mapped file with zero
  heap allocation.

```go
// Build once (use Batch for best write performance).
b, _ := diskgraph.NewUndirectedBuilder("providers.db")
b.Batch(func(tx *diskgraph.UndirectedTx) error {
    for _, edge := range edges {
        tx.AddWeightedEdge(edge.From, edge.To, edge.Weight)
    }
    return nil
})
b.Close()

// Open anywhere, any time — file IS the graph.
g, _ := diskgraph.OpenUndirected("providers.db")
defer g.Close()

// All GraphWizard algorithms work unchanged.
comms := community.Leiden(g, 1.0, rng)
scores := centrality.Degree(g)
bridges := connectivity.Bridges(g)
```

**When to use `simple.UndirectedGraph` instead:** Only when you need
a mutable graph (adding/removing edges after construction) or when
building ephemeral throwaway graphs in tests. For all analytical
workloads, use `diskgraph`.

## Best Practices

See **[BEST_PRACTICES.md](BEST_PRACTICES.md)** for detailed guidance on:
choosing algorithm variants for your graph size, loading from SQL,
anomaly detection pipelines, streaming updates, performance tips, and
a complexity reference table for every algorithm.

## Design Principles

- **One import per domain.** No need to learn gonum's iterator patterns or
  juggle multiple package imports. `centrality.PageRank(g, 0.85, 1e-6)` just
  works.

- **Consistent return types.** Centrality measures return `map[int64]float64`.
  Components return `[][]int64`. No custom iterator types to unpack.

- **Standard interfaces.** Every function accepts `graph.Graph`,
  `graph.Undirected`, or `graph.Directed` from gonum. Build your graph with
  `diskgraph` (recommended) or `simple.NewUndirectedGraph()` and pass it
  to any algorithm.

- **Cleanroom implementations.** Custom algorithms are implemented from
  academic papers, not ported from existing libraries. Each function documents
  its reference paper.

- **Tested and documented.** 97.3% test coverage. 36 runnable examples. Every
  exported function has godoc comments.

## Academic References

| Algorithm | Paper |
|---|---|
| Leiden | Traag, Waltman, van Eck. "From Louvain to Leiden." *Scientific Reports*, 2019 |
| Label Propagation | Raghavan, Albert, Kumara. "Near linear time algorithm to detect community structures." *Physical Review E*, 2007 |
| Katz Centrality | Katz. "A New Status Index Derived from Sociometric Analysis." *Psychometrika*, 1953 |
| Personalized PageRank | Haveliwala. "Topic-Sensitive PageRank." *WWW*, 2002 |
| Hopcroft-Karp | Hopcroft, Karp. "An n^(5/2) Algorithm for Maximum Matchings in Bipartite Graphs." *SIAM J. Computing*, 1973 |
| Yen's K Shortest | Yen. "Finding the K Shortest Loopless Paths in a Network." *Management Science*, 1971 |
| Bridges | Tarjan. "A Note on Finding the Bridges of a Graph." *Information Processing Letters*, 1974 |
| Biconnected Components | Hopcroft, Tarjan. "Algorithm 447." *Communications of the ACM*, 1973 |
| Kruskal MST | Kruskal. "On the Shortest Spanning Subtree." *Proceedings of the AMS*, 1956 |
| Prim MST | Prim. "Shortest Connection Networks." *Bell System Technical Journal*, 1957 |
| Node2Vec | Grover, Leskovec. "node2vec: Scalable Feature Learning for Networks." *KDD*, 2016 |
| DeepWalk | Perozzi, Al-Rfou, Skiena. "DeepWalk: Online Learning of Social Representations." *KDD*, 2014 |
| SVD Embedding | Levy, Goldberg. "Neural Word Embedding as Implicit Matrix Factorization." *NIPS*, 2014 |
| Clustering Coefficient | Watts, Strogatz. "Collective dynamics of 'small-world' networks." *Nature*, 1998 |
| Set Cover | Chvatal. "A Greedy Heuristic for the Set-Covering Problem." *Mathematics of Operations Research*, 1979 |
| TSP 2-opt | Croes. "A Method for Solving Traveling-Salesman Problems." *Operations Research*, 1958 |
| Adamic-Adar | Adamic, Adar. "Friends and neighbors on the Web." *Social Networks*, 2003 |
| SimRank | Jeh, Widom. "SimRank: A Measure of Structural-Context Similarity." *KDD*, 2002 |
| Influence Maximization | Kempe, Kleinberg, Tardos. "Maximizing the Spread of Influence." *KDD*, 2003 |
| CELF Optimization | Leskovec et al. "Cost-effective Outbreak Detection in Networks." *KDD*, 2007 |
| Spectral Clustering | Ng, Jordan, Weiss. "On Spectral Clustering." *NIPS*, 2001 |
| Min Cut | Ford, Fulkerson. "Maximal Flow Through a Network." *Canadian J. Math*, 1956 |

## Stats

- 50+ source files, 50+ test files
- 7,500+ lines of Go
- 220+ test and example functions
- 97%+ statement coverage (5 packages at 100%)
- Dependencies: gonum (algorithms), bbolt (disk storage)

## License

MIT. See [LICENSE](LICENSE).

## About Intelligrit

[Intelligrit LLC](https://intelligrit.com) is a technology company focused on
AI-powered solutions for government program integrity. GraphWizard is one piece
of the analytical toolkit we built to detect fraud, waste, and abuse in federal
healthcare programs.
