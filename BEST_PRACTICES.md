# GraphWizard Best Practices

Practical guidance for using GraphWizard effectively on real-world
graphs, from small exploratory analysis to 5M+ node production
workloads.

## Use diskgraph by Default

The `diskgraph` package is the recommended graph storage for all
GraphWizard workloads. It is faster than gonum's `simple.UndirectedGraph`
for most algorithms, persists to disk automatically, and handles
graphs too large for memory.

```go
// Build once.
b, _ := diskgraph.NewUndirectedBuilder("graph.db")
b.Batch(func(tx *diskgraph.UndirectedTx) error {
    // Add edges in bulk — single transaction for speed.
    tx.AddEdge(0, 1)
    tx.AddEdge(1, 2)
    return nil
})
b.Close()

// Open and use with any algorithm.
g, _ := diskgraph.OpenUndirected("graph.db")
defer g.Close()
```

**Memory behavior:** By default, adjacency is read using SQLite queries.
Pass `Preload` to keep dense-indexed adjacency in Go memory. Edge weights
remain in SQLite:

```go
g, _ := diskgraph.OpenUndirected("graph.db", diskgraph.Preload)
```

Flat preloaded adjacency costs `4*D + sizeof(int)*(N+1) + 8*N` bytes,
where `D` is the number of directed edge entries and `N` the number of
nodes. An undirected edge normally contributes two entries. On a 64-bit
system, 5.8M nodes and 82.6M undirected edges need approximately **0.75 GB**
of slice payloads, excluding SQLite caches and algorithm working memory.

If the estimated cost exceeds 70% of available RAM, a warning is logged
and preloading is skipped. Use `ForcePreload` to override.

**When to use `simple.UndirectedGraph` instead:** Only when you need a
mutable graph (adding/removing edges after construction) or for
ephemeral throwaway graphs in tests.

## Memory and compression

For immutable in-memory analysis, `densegraph` stores the same dense neighbor
indexes and keeps weights inline. Its flat slice payload is:

```
4*D + sizeof(int)*(N+1) + 8*N + weightBytes
```

Exact weight storage is selected automatically after duplicate edges are
resolved:

| Weight values | Additional bytes per directed entry |
|---|---:|
| All identical, including unit weights | 0 (one graph-wide constant) |
| Nonnegative integers up to 255 | 1 |
| Nonnegative integers up to 65,535 | 2 |
| Nonnegative integers up to 4,294,967,295 | 4 |
| Other values | 8 (`float64`) |

Float64 fallback preserves negative values, fractions, signed zero, infinities,
and NaN payloads. Constant storage also preserves the original bit pattern.
These encodings preserve every edge and weight; they do not approximate the
graph.

Two independent options can reduce storage further:

```go
g := builder.Build(densegraph.PackNeighbors24, densegraph.BlockWeights)
```

`PackNeighbors24` uses three bytes per dense neighbor index, replacing `4*D`
with `3*D` in the storage formula. This saves exactly 25% of the neighbor array,
not 25% of the entire graph. All indexes must fit in 24 bits (maximum
16,777,215); otherwise the graph keeps flat int32 indexes. This always fits
graphs with at most 16,777,216 nodes. Original int64 IDs do not affect packing.
For 82.6M undirected edges, packing saves about 165.2 MB of neighbor payload.
Binary search reads each packed index directly, without decoding preceding
neighbors. Scans still pay for unpacking three bytes. Use
`diskgraph.Packed24Preload` for the same option on undirected SQLite adjacency.

`BlockWeights` selects a constant, uint8, uint16, uint32, or float64 separately
for each consecutive block of 128 directed entries. A fractional outlier thus
widens only its block. Blocks cross node boundaries and store one machine-sized
byte offset plus a one-byte kind (9 bytes per block on a 64-bit system).
Constant blocks store one float64; other blocks use 1, 2, 4, or 8 bytes per entry.
Selection includes all this metadata and keeps the graph-wide encoding unless
blocks save payload. Unit-weight graphs remain a single implicit constant.
Weight lookup is still O(1), but reads block metadata and decodes bytes.
Construction makes extra passes over weights to size and select the encoding
before allocating the final arrays. Scattered outliers in every block may give
no saving. This option affects densegraph only; diskgraph weights stay in SQLite.

Both options are off by default. `BlockWeights` combines with either neighbor
encoding; when `PackNeighbors24` and `CompressNeighbors` are both supplied, the
last wins. The analogous diskgraph encoding options also use the last selection;
`Preload` and `ForcePreload` do not reset it.

These independent implementations draw on Besta et al.,
[Log(Graph): A Near-Optimal High-Performance Graph Representation](https://arxiv.org/abs/2010.15879)
(PACT 2018; arXiv version 2020), Sections 3.2–3.3, for sizing vertex IDs and
integer weights to their value ranges; and Shun, Dhulipala, and Blelloch,
[Smaller and Faster: Parallel Processing of Compressed Graphs with Ligra+](https://www.cs.cmu.edu/~blelloch/papers/SDB15.pdf)
(DCC 2015), Section 3, for selecting widths for groups of entries. Our 24-bit
width, fixed 128-entry weight blocks, constant representation, and exact float64
fallback are local design choices. The papers' measurements are not performance
claims for GraphWizard; the measurements below come from our own benchmarks.

Use `builder.Build(densegraph.CompressNeighbors)` or
`diskgraph.OpenUndirected(path, diskgraph.CompressedPreload)` to request
lossless block delta encoding. Each block of 32 dense indexes stores an
absolute starting value followed by signed delta varints. Restart positions
bound lookup decoding. Blocks may cross node boundaries to avoid per-node
compression metadata. The encoder keeps flat storage when compression would
increase the payload. Gap encoding and lazy traversal are discussed in Boldi
and Vigna, [The WebGraph Framework I: Compression Techniques](https://vigna.di.unimi.it/ftp/papers/WebGraphI.pdf)
(WWW 2004). Our 32-entry restarts and signed varints are local choices; this
independent implementation does not reproduce WebGraph's reference or interval
compression schemes.

Delta compression depends on neighbor gaps and ordering. It increases decoding
CPU cost, especially for random edge lookups. The default remains flat storage
and graph-wide exact weights.
`From`, `Neighbors`, and weighted neighbor traversal decode without retaining
a second adjacency copy. `DenseNeighbors` remains compatible, but allocates a
fresh row when packed or compressed. Concurrent readers have independent decoder state;
there is no mutable shared decompression cache.

Leiden and parallel Leiden share the original weighted adjacency through
`DenseWeightedAdjacency`. They still allocate node state, aggregate edge maps,
and coarsened adjacency. Disk graphs keep weights in SQLite and use a single
edge scan to construct the algorithm's weighted adjacency, avoiding per-edge
SQL lookups. Katz variants use reusable neighbor cursors rather than repeatedly
materializing compressed rows. Each worker has independent decoder state.
Parallel refinement shares its node labels and weight scratch across workers:
each community accesses only its own nodes and refined labels. On a 64-bit
system, these two arrays use `16 × V` bytes total instead of
`8 × V + 16 × V × workers`. Membership lists, permutations, cursors, and other
algorithm state are additional. Serial and parallel refinement reuse RNG state
with the same per-community seeds, preserving permutations and results.
Lower scratch allocation does not guarantee faster refinement: worker
scheduling and cache locality still affect throughput. Measure the whole
analysis as well as allocation volume.

Measure retained arrays with `densegraph.Undirected.StorageBytes()` or
`diskgraph.Undirected.AdjacencyBytes()`. These exclude headers, allocator
rounding, database caches, and algorithm state. For densegraph, the builder
also retains a 16-byte entry per directed edge until construction finishes
(32 bytes per ordinary undirected edge, before final arrays). It maps original
IDs to temporary `int32` indexes during ingestion, then remaps those indexes
once into sorted node-ID order. CSR construction uses the dense endpoints
directly, avoiding per-edge binary searches and a separate degree array.
The ID map, node IDs, remapping table, spare builder capacity, and final arrays
also need memory. Compact temporary edges help most on edge-heavy graphs;
the index map and remapping table add per-node overhead during construction.
Use `NewUndirectedBuilderSized` when counts are available to
avoid growing and copying the edge buffer; its edge hint counts directed entries.
Neighbor packing or compression temporarily overlaps flat and encoded arrays.
Construction still buffers the graph; lower retained memory does not imply
the same reduction in peak construction memory.

Reproduce storage, traversal, construction, and algorithm allocation benchmarks:

```sh
go test ./internal/compact ./densegraph ./community -run '^$' \
  -bench 'BenchmarkNeighbors|BenchmarkCompactGraph|BenchmarkWeightedAdjacencyBuild|BenchmarkSharedLeiden|BenchmarkRefinementScratch' \
  -benchmem
```

The storage metrics describe the benchmark fixtures, not a guaranteed reduction
for a production graph. Existing `BenchmarkGraphMemory` benchmarks keep the
graph alive across GC so retained heap measurements include its storage.
`BenchmarkCompactGraph` compares all six neighbor/weight combinations on small
integer weights; `BenchmarkCompactGraphOutlier` adds one fractional edge to
show when block weights help. Both report retained payload, traversal time,
and construction allocations.

For the 10,000-node outlier fixture (319,728 directed entries), the retained
slice payloads measured on a 64-bit system are:

| Optional settings | Retained bytes |
|---|---:|
| None | 3,996,744 |
| `PackNeighbors24` | 3,677,016 |
| `BlockWeights` | 1,782,026 |
| Both | 1,462,298 |

Both options save about 63% for this fixture, which deliberately has mostly
small integer weights. Uniformly small integer weights already use one byte
each, so the block option falls back without a storage saving. Time benchmarks
should be run on representative graph sizes and degree/weight distributions;
these payload reductions do not predict traversal speed or peak RSS.

## Choosing the Right Algorithm Variant

Most algorithms have both sequential and parallel versions. Use this
guide:

| Graph Size | Recommendation |
|---|---|
| < 1K nodes | Sequential — goroutine overhead exceeds benefit |
| 1K–100K nodes | Parallel for O(V^2) algorithms (triangles, clustering, similarity) |
| 100K+ nodes | Always use parallel variants; avoid O(V^2) algorithms entirely |

**Specific guidance:**

- **Betweenness centrality**: Use `ApproximateBetweenness` with
  k=1000 for any graph over 10K nodes. Exact betweenness is O(VE)
  and takes years on million-node graphs.

- **Community detection**: `Leiden` is the best default. `Louvain`
  is faster but doesn't guarantee well-connected communities.
  `LabelPropagation` is fastest but non-deterministic and can produce
  poor results on some topologies. `SpectralClustering` is best for
  small graphs (< 10K) where you know the number of clusters.

- **All-pairs similarity** (`JaccardAll`, `PredictLinks`): Only
  practical for graphs under 50K nodes (O(V^2) pairs). For larger
  graphs, compute pairwise similarity only for nodes sharing a
  neighbor.

## Loading Graphs from SQL

For small-to-medium graphs, use the `loader` package to load directly
into memory:

```go
import (
    "database/sql"
    "github.com/intelligrit/graphwizard/loader"
    _ "github.com/marcboeker/go-duckdb/v2"
)

db, _ := sql.Open("duckdb", "analytics.duckdb")

// Unweighted graph.
g, err := loader.LoadUndirected(db,
    "SELECT from_id, to_id FROM affiliations")

// Weighted graph.
wg, err := loader.LoadWeightedUndirected(db,
    "SELECT provider_npi, drug_name, total_claims FROM prescriptions")
```

For large graphs or when you want persistence, load into diskgraph:

```go
b, _ := diskgraph.NewUndirectedBuilder("affiliations.db")
rows, _ := db.Query("SELECT from_id, to_id FROM affiliations")
b.Batch(func(tx *diskgraph.UndirectedTx) error {
    for rows.Next() {
        var from, to int64
        rows.Scan(&from, &to)
        tx.AddEdge(from, to)
    }
    return nil
})
b.Close()

// Now open for analysis — persisted for future runs.
g, _ := diskgraph.OpenUndirected("affiliations.db")
defer g.Close()
```

The loader query must return exactly 2 columns (from_id, to_id) for
unweighted, or 3 columns (from_id, to_id, weight) for weighted.
Column types must be scannable to int64 and float64.

## Writing Results Back

After running algorithms, write results back to your database:

```go
scores := centrality.ApproximateBetweenness(g, 1000, rng)
loader.WriteResults(db, "betweenness_scores", scores)

comms := community.Leiden(g, 1.0, rng)
loader.WriteCommunities(db, "communities", comms)
```

## Subgraph Extraction for Dashboards

Don't load the entire graph for per-provider views. Extract the
relevant neighborhood:

```go
// Get 2-hop neighborhood of a provider.
sub := subgraph.NHopNeighborhoodUndirected(fullGraph, providerNPI, 2)

// Run algorithms on just this subgraph.
scores := centrality.Degree(sub)
```

## Streaming Updates

For real-time analysis where the graph changes incrementally:

```go
sg := stream.New()
sg.AddEdge(100, 200, 1.0)
sg.AddEdge(200, 300, 2.0)

// Get the graph for algorithm use.
g := sg.Graph()
comms := community.Louvain(g, 1.0, nil)

// Later, apply changes without rebuilding.
sg.AddEdge(300, 400, 1.5)
sg.RemoveEdge(100, 200)

// Check what changed.
changes := sg.Changes()
```

## Anomaly Detection Pipeline

For fraud detection, combine multiple signals:

```go
// 1. Structural anomaly scores.
iso := anomaly.IsolationScore(g)
outliers := anomaly.StructuralOutliers(g, 100)

// 2. Community-based: nodes bridging multiple communities.
comms := community.Leiden(g, 1.0, rng)
bridges := connectivity.Bridges(g)

// 3. Centrality-based: unusual influence patterns.
ppr := centrality.PersonalizedPageRank(dg, seedNode, 0.85, 1e-6, 100)
approxBetween := centrality.ApproximateBetweenness(g, 1000, rng)

// 4. Link prediction: find hidden connections.
predicted := similarity.PredictLinks(g, 100, similarity.AdamicAdar)
```

## Graph Diffing for Change Detection

Compare snapshots to detect changes over time:

```go
before, _ := loader.LoadUndirected(db, "SELECT * FROM edges_jan")
after, _ := loader.LoadUndirected(db, "SELECT * FROM edges_feb")

result := diff.Compare(before, after)
fmt.Printf("New edges: %d, Removed: %d\n",
    len(result.AddedEdges), len(result.RemovedEdges))
```

## Performance Tips

1. **Use `diskgraph`** as your default graph storage. It is faster than
   in-memory gonum graphs for most algorithms, persists automatically,
   and handles out-of-core workloads. See "Use diskgraph by Default"
   above.

2. **Use `Batch`** when building disk graphs. Writing edges one at a
   time in separate transactions is orders of magnitude slower than
   batching them.

3. **Pre-build neighbor sets** if you're calling multiple similarity
   functions on the same graph. The parallel variants do this
   internally. Diskgraph can preload compact sorted adjacency with
   `Preload` or `CompressedPreload`; similarity algorithms may still build
   additional working sets.

4. **Use `runtime.GOMAXPROCS`** to control parallelism. The parallel
   variants auto-detect available cores but you can tune this for
   shared environments.

5. **Graph implementations must be safe for concurrent reads** when
   using parallel functions. Both `diskgraph` and gonum's
   `simple.*Graph` types satisfy this requirement.

6. **For DuckDB**, use `?access_mode=READ_ONLY` when loading to
   allow concurrent readers.

7. **Memory estimation**: use `StorageBytes` / `AdjacencyBytes` for retained
   array payloads and profile the whole process for peak memory. Preload
   stores dense int32 neighbors; its cost depends on both node count and
   directed edge count. Without preloading, SQL queries and database caches
   still use memory. See [Memory and compression](#memory-and-compression).

## Algorithm Complexity Reference

| Algorithm | Time | Space | Parallelizable |
|---|---|---|---|
| BFS/DFS | O(V+E) | O(V) | No |
| Bridges | O(V+E) | O(V) | Per-component |
| WCC/SCC | O(V+E) | O(V) | No |
| PageRank | O(k(V+E)) | O(V) | Per-iteration |
| Betweenness (exact) | O(VE) | O(V+E) | Per-source |
| Betweenness (approx) | O(kE) | O(V+E) | Per-source |
| Leiden/Louvain | ~O(V+E) | O(V+E) | Limited |
| Katz/PPR | O(k(V+E)) | O(V) | Per-node |
| Triangle count | O(V·d²) | O(V+E) | Per-node |
| Clustering coeff | O(V·d²) | O(V+E) | Per-node |
| MST (Kruskal) | O(E log E) | O(V+E) | No |
| Hopcroft-Karp | O(E√V) | O(V+E) | No |
| Yen K-shortest | O(KV(V log V + E)) | O(V+E) | Per-spur |
| SimRank | O(kV²d²) | O(V²) | Per-iteration |
| Node2Vec walks | O(walks·length) | O(walks·length) | Per-walk |
| Spectral clustering | O(V³) | O(V²) | Limited |
