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

**New:** The `diskgraph` package provides a SQLite-backed graph that is
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
one — including `diskgraph`, a SQLite-backed graph store that turned out to be
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
| **diskgraph** | Disk-backed Undirected/Directed graphs with optional preloading, SQLite storage | Custom |
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

    // Open the graph — add diskgraph.Preload to cache adjacency in memory.
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
workloads. It replaces `gonum/graph/simple` with a SQLite-backed
implementation that is:

- **Faster** — benchmarks show diskgraph matches or beats in-memory
  `simple.UndirectedGraph` on 40+ algorithms. Algorithms like
  PreferentialAttachment run 4x faster; most run 10-30% faster.

- **Persistent** — build the graph once, reopen it instantly across
  program runs. No serialization, no import step. The SQLite file is the
  database.

- **Scalable** — adjacency stays on disk by default, with node IDs retained
  in memory. Enable preloading when the adjacency fits in RAM.

- **Tunable** — pass `Preload` to cache adjacency data in Go memory
  for in-memory neighbor traversal and binary-search edge lookups.
  Without it, prepared SQLite queries read adjacency and weights.

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

## Compact memory storage

`densegraph` keeps one dense `int32` neighbor index per directed edge entry,
with original node IDs stored once per node. Weights are stored exactly using
a single constant, `uint8`, `uint16`, `uint32`, or `float64`, depending on the
values. Constant weights (including ordinary unweighted graphs) need no
per-edge weight array. No precision is discarded.

```go
b := densegraph.NewUndirectedBuilder()
b.AddWeightedEdge(100, 200, 7)
b.AddWeightedEdge(200, 300, 12)
// Both options are independent and off by default. Build() keeps flat int32
// neighbors and chooses one exact weight representation for the whole graph.
g := b.Build(densegraph.PackNeighbors24, densegraph.BlockWeights)
```

| Optional setting | Effect |
|---|---|
| `densegraph.PackNeighbors24` | Three bytes per neighbor index, saving 25% of the neighbor array. Falls back to int32 if any index exceeds 16,777,215. Original int64 IDs are preserved. |
| `densegraph.CompressNeighbors` | Alternative block delta encoding, retained only if its payload is smaller than flat indexes. Usually more decoding work than 24-bit packing. |
| `densegraph.BlockWeights` | Selects an exact representation per 128 directed entries, so an outlier need not widen every weight. Retained only if data plus metadata beat the graph-wide representation. |

Call `Build` only once per builder. `BlockWeights` combines with either neighbor
option. If both neighbor options are supplied, the last wins. All modes are
lossless, including fractional weights, signed zero, and NaN bit patterns.
Packed neighbors and block weights add decoding work; benchmark with your
graph before selecting them for CPU-sensitive workloads.
`From` and `Neighbors` decode lazily. The compatibility method `DenseNeighbors`
returns a fresh decoded row in packed or delta mode and does not cache it.

For SQLite-backed undirected graphs, neighbor packing is available during
adjacency preloading. Weights stay in SQLite:

```go
g, err := diskgraph.OpenUndirected("graph.db", diskgraph.Packed24Preload)
```

Use `diskgraph.CompressedPreload` for delta encoding, or `diskgraph.Preload`
for flat indexes. The last encoding option wins; `Preload` and `ForcePreload`
do not change a selected encoding.

`g.StorageBytes()` on a dense graph and `g.AdjacencyBytes()` on a disk graph
report retained slice payloads. These exclude object headers, database caches,
allocator overhead, and algorithm working memory. Building or compressing a
graph temporarily needs additional memory; retained size is not peak RSS.
The dense builder uses 16-byte temporary directed entries, assigning compact
indexes during ingestion and remapping them once into sorted node-ID order.
This preserves original IDs and last-weight-wins duplicate handling while
avoiding per-edge ID searches during construction.

Leiden and parallel Leiden borrow weighted adjacency from graphs implementing
`graphwizard.DenseWeightedAdjacency`, including all densegraph storage modes.
They allocate their own coarsened graphs, without copying the original weighted
edge array. Katz variants use streaming neighbors for compressed storage.
Parallel Leiden refinement shares node scratch arrays across workers, with
each community owning disjoint slots; these arrays no longer scale with worker
count. Serial and parallel refinement reuse random generators while preserving
the seeded results.
Generic gonum graphs remain supported through the existing fallback paths.

See [BEST_PRACTICES.md](BEST_PRACTICES.md#memory-and-compression) for storage
formulas, tradeoffs, and reproducible memory benchmarks.

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

| Algorithm / representation | Paper |
|---|---|
| Packed neighbor IDs and integer weight widths | Besta et al. ["Log(Graph): A Near-Optimal High-Performance Graph Representation."](https://arxiv.org/abs/2010.15879) *PACT*, 2018; Sections 3.2–3.3. |
| Grouped weight encoding | Shun, Dhulipala, Blelloch. ["Smaller and Faster: Parallel Processing of Compressed Graphs with Ligra+."](https://www.cs.cmu.edu/~blelloch/papers/SDB15.pdf) *DCC*, 2015; Section 3. |
| Neighbor gap encoding and lazy traversal | Boldi, Vigna. ["The WebGraph Framework I: Compression Techniques."](https://vigna.di.unimi.it/ftp/papers/WebGraphI.pdf) *WWW*, 2004. |
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

The storage papers inform specific techniques. GraphWizard's 24-bit width,
128-entry weight blocks, exact float64 fallback, and 32-entry delta restarts
are independent design choices. Our storage and timing measurements come from
GraphWizard benchmarks; the papers' benchmark results do not describe this
implementation. See [memory and compression](BEST_PRACTICES.md#memory-and-compression).

## Stats

- 50+ source files, 50+ test files
- 7,500+ lines of Go
- 220+ test and example functions
- 97%+ statement coverage (5 packages at 100%)
- Dependencies: gonum (algorithms), SQLite (disk storage)

## License

MIT. See [LICENSE](LICENSE).

## About Intelligrit

[Intelligrit LLC](https://intelligrit.com) is a technology company focused on
AI-powered solutions for government program integrity. GraphWizard is one piece
of the analytical toolkit we built to detect fraud, waste, and abuse in federal
healthcare programs.
