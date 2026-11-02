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
no saving. `multigraph.BlockWeights` applies the same encoding to relationship
columns; diskgraph weights stay in SQLite.

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

Choose options according to the limiting resource:

| Setting | Good fit | Cost or limitation |
|---|---|---|
| Default flat neighbors and graph-wide weights | A graph that fits comfortably in memory; frequent scans or random edge lookups. | Uses four bytes per neighbor even when smaller representations would fit. Exact narrow/constant weight encoding is already automatic. |
| `PackNeighbors24` | Predictable neighbor savings with modest decoding work. | Saves 25% of neighbor payload only; dense indexes must fit in 24 bits. Each access unpacks three bytes. |
| `CompressNeighbors` | Memory-constrained analyses with compressible neighbor gaps and predominantly sequential traversal. | Signed varints and restart decoding cost CPU. Random lookup decodes within restart blocks; some graphs fall back without saving payload. |
| `BlockWeights` | Weight columns with locally small ranges, constant runs, or isolated outliers. | Adds metadata, encoding passes, and lookup work. Already narrow uniform weights or widely scattered outliers may yield no saving. |

The savings check for delta neighbors and block weights protects retained
payload size. It does not guarantee faster construction or traversal, or refund
the work spent evaluating an encoding that falls back. Test options separately
if you need to attribute a saving to neighbor encoding versus weight encoding.
`multigraph.BlockWeights` configures the immutable store; options passed to
`ProjectUndirected` independently configure each result. Neighbor encoding does
not compress the store's 16-byte relationship metadata rows.

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

## Multiple weights and relationship projections

`multigraph` is an opt-in in-memory relationship store for analyses that need
separate observations and more than one metric. A record contains endpoint IDs,
a type string, an int32 year, and any number of named float64 weights. Each call
to `Builder.AddRelationship` creates a distinct record, including exact
duplicates. Its returned ID is an insertion position within that store, not a
cross-database identifier. `Store.Relationships` and `Store.Weight` expose the
original records without reduction. Empty type and zero year denote untyped
and undated records; there is no automatic date inference.

Existing `densegraph` and `diskgraph` builders still store one scalar weight per
endpoint pair and retain their existing duplicate handling. Choose the new
store explicitly when repeated pairs must survive.

For database input, return four metadata columns followed by the named weights:

```go
relationships, err := loader.LoadRelationships(db, `
    SELECT provider_id, drug_id, 'prescribed', year, claims, dollars
    FROM annual_relationships
    ORDER BY provider_id, drug_id, year
`, []string{"claims", "dollars"}, multigraph.BlockWeights)
if err != nil {
    return err
}
```

Every SQL row survives. NULL weights remain missing, including in an entirely
NULL column; NULL type/year become empty type/zero year. `DefineWeight` can
register a metric with no supplied values when using the builder directly.
Weights use float64 values supplied by the driver, so numeric conversion from
other SQL types happens before storage. Include a unique tie-breaker in `ORDER BY`
if duplicate rows need stable IDs and floating-point reduction order.
To retain isolated nodes or preallocate capacity, create a `multigraph.Builder`,
call `AddNode`, then pass SQL rows to `loader.ScanRelationships`. That function
appends to the builder; discard it if an error requires all-or-nothing loading.
Input queries must preserve the distinctions you need: grouping across years
before loading cannot be undone by the graph library.

Algorithms consume an explicit scalar-weight projection:

```go
g, err := relationships.ProjectUndirected(multigraph.Projection{
    Weight: "claims",
    Filter: func(r multigraph.Relationship) bool {
        return r.Type == "prescribed" && r.Year >= 2020 && r.Year <= 2023
    },
    Combine: multigraph.Sum,
}, densegraph.PackNeighbors24, densegraph.BlockWeights)
```

`ProjectUndirected` returns a `*densegraph.Undirected`, usable directly by gonum
and GraphWizard algorithms. Filtering happens before missing-weight checks and
combining. The projection groups both endpoint orientations together and keeps
all registered nodes, including isolates. Self relationships are counted once;
the resulting graph retains gonum's zero self-weight convention for `Weight`.

| Choice | Behavior |
|---|---|
| `Combine: RejectParallel` (default) | Error if more than one selected, retained record connects a pair. |
| `Combine: Sum` | Sum weights in insertion order; appropriate when the chosen metric is additive. |
| `Combine: Minimum` / `Maximum` | Keep the smallest/largest selected value per pair. |
| `Weight: ""` | One unit per relationship; together with `Sum`, counts relationships. |
| `Missing: ErrorMissing` (default) | Error for a selected record without the requested weight. |
| `Missing: SkipMissing` | Omit that record from the projection. |
| `Missing: ZeroMissing` | Include that record with weight zero, keeping its edge. |

An unknown metric name is always an error, even for an empty selection. The
errors `ErrUnknownWeight`, `ErrMissingWeight`, and `ErrParallelRelationships`
support `errors.Is`. A stored zero is never treated as absent. Combining values
uses ordinary float64 arithmetic and is an intentional transformation; it does
not promise lossless aggregation, avoid overflow, or infer compatible units.
The original store remains exact and unchanged. Use a metric's meaning when
choosing a reducer: an additive count and a shortest-path cost may need different
rules. Weight-aware algorithms use the selected metric; topology-only algorithms
such as the current Katz implementations ignore edge weights.

Storage has one 16-byte metadata row per relationship, original node IDs stored
once, and independently compressed weight columns. A partially populated column
adds one presence bit per relationship, rounded to whole uint64 words. Complete
columns need no bitmap, and constant columns need no per-record value array.
There are no retained per-relationship maps or duplicated reverse records.
`multigraph.BlockWeights` opts into exact block encoding for store columns;
`densegraph.BlockWeights` on a projection separately configures that graph.

`Store.StorageBytes` reports retained numeric array payload, excluding strings,
dictionaries, and object headers. A materialized projection adds its own graph
storage and temporary sorting/build buffers. Keeping several projected graphs
simultaneously therefore adds memory. `NewBuilderSized` accepts a relationship
count, with each observation counted once. Existing single-weight graphs incur
none of this metadata overhead.

The new store is immutable and supports concurrent readers and independent
projections. It does not add persistence or change the existing diskgraph schema;
retain the source relationships in your database when they must survive restart.

```sh
go test ./multigraph -run '^$' -bench BenchmarkRelationships -benchmem
```

### Optional storage modes at scale

A full-scale local validation used 290,789,017 relationships, 3,243,863 nodes,
and three named columns: claims, dollars, and a constant unit weight. Records
preserved year and relationship type; source rows within an endpoint/type/year
group were summed before loading. These are annual aggregates plus undated
relationships, rather than individual transaction events. The source fixture
and run artifacts are not distributed with the repository; the tables contain
aggregate measurements only.

Three configurations were tested in separate processes:

| Mode | Store build options | Projection build options |
|---|---|---|
| Default | None | None |
| Packed + blocks | `multigraph.BlockWeights` | `densegraph.PackNeighbors24`, `densegraph.BlockWeights` |
| Delta + blocks | `multigraph.BlockWeights` | `densegraph.CompressNeighbors`, `densegraph.BlockWeights` |

The count and dollar-sum graphs each had 81,545,079 undirected pairs. The
claims-sum graph had 65,609,451 pairs. All projections retained the full node
set, including isolates. Retained numeric payloads, in GiB (2^30 bytes), were:

| Mode | Relationship store | Count graph | Claims graph | Dollar graph |
|---|---:|---:|---:|---:|
| Default | 7.675 | 0.808 | 1.026 | 1.871 |
| Packed + blocks | 6.989 | 0.656 | 0.665 | 1.719 |
| Delta + blocks | 6.989 | 0.445 | 0.484 | 1.509 |

Store block weights saved about 9%. The combined projection options saved
about 35% (packed) or 53% (delta) on the claims graph, but only about 8% or 19%
on the dollar graph. Fractional dollar weights still require exact float64
storage, whereas many claim weights fit narrower integer blocks. The two
optional modes have identical store sizes because neighbor encoding applies
only to projections. These combined-mode measurements do not isolate the CPU
cost of block weights from neighbor decoding.

Elapsed times for the count graph were:

| Mode | Freeze store (s) | Build projection (s) | Katz (s) | Parallel Leiden (s) |
|---|---:|---:|---:|---:|
| Default | 3.52 | 47.20 | 2.68 | 16.54 |
| Packed + blocks | 4.98 | 47.36 | 2.69 | 17.46 |
| Delta + blocks | 5.09 | 47.50 | 6.58 | 20.84 |

For the claims graph, parallel Leiden took 16.71, 18.59, and 21.06 seconds,
respectively. Delta reduced retained memory the most, while Leiden took about
26% longer on both views and count-graph Katz took about 2.5 times as long.
The small differences in projection construction times should not be treated
as reliable speed rankings from a single run.

The machine was an Apple M5 Max with 128 GiB RAM, macOS arm64, Go 1.27.3,
and DuckDB 1.4.1. Each process used `GOMAXPROCS=8`, a 40 GiB Go soft memory
limit, and DuckDB configured for four threads and a 4 GB memory limit.
Katz used `alpha = 0.5/(maxDegree+1)`, beta 1, tolerance 1e-8, and at most
40 iterations. Parallel Leiden used resolution 1 and seed 42. There was one
full run per mode, not repeated-run medians or a guarantee for other graphs.
Katz checks topology; Leiden uses the selected weights.

Loading and auditing the input took about 150–152 seconds per process. Peak
process RSS was about 43.1 GiB for every mode, despite the 7.0–7.7 GiB retained
store. That peak includes Go, DuckDB, builder capacity, temporary buffers, and
algorithm allocations; it is not `StorageBytes`. The Go and DuckDB limits
are separate controls and do not cap whole-process RSS. A compressed store
that fits in RAM does not imply that loading it or building a projection fits.

Validation covered all records and every projected edge, with independent SQL
references for count sums, claim sums, dollar sums, a year/type selection, and
payment maxima. Raw record hashes, projected edge hashes, Katz scores, and
seeded Leiden partitions matched exactly across storage modes. SQL dollar
SUM used a tolerance for differences in reduction order; its maximum observed
absolute difference was 1.17e-10. This does not relax the bit-exact checks of
stored values or cross-mode results.

Six projections were built from each store, including an annual view repeated
after the other passes. Results remained identical, the store's final hash was
unchanged, and retained Go heap returned within 30 KiB of its store-only baseline
after releasing each graph and forcing GC. This checks sequential reuse;
retaining multiple projections concurrently adds their graph storage and
overlapping build/algorithm working memory.

For repeated analyses, reuse a projected graph while its weight, filter, and
reducer stay the same. Build another view when those choices change, and release
unneeded graphs before the next large build. Measure ingestion, freeze,
projection construction, traversal, and peak RSS separately on your own data.
Use the public synthetic benchmarks above for repeatable encoding comparisons;
their memory savings and timings need not match this large fixture.

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

- **Fraud seed propagation**: `LocalPersonalizedPageRank` accepts weighted
  seeds and explores only nodes reached by local residual pushes. Tune
  `epsilon` to trade precision for work; the returned scores are lower
  bounds and need not sum to one. Its working maps grow with the explored
  neighborhood, not the whole graph. One minus the sum of scores estimates
  unpushed probability mass. On a highly connected graph, a tighter
  `epsilon` can make the explored neighborhood approach the whole graph;
  measure both runtime and peak memory at the chosen setting.

- **Shared device or identifier evidence**: Build a
  `BipartiteOverlapScorer` once, then call `Score` for selected account
  pairs. It reports rarity-weighted overlap and raw hypergeometric
  probabilities within bridge-degree bins without constructing a projected
  graph. Correct the probabilities for all pairs and bins tested before
  calling them statistically significant. The scorer retains a histogram
  of bridge degrees; each pair uses memory proportional to one account's
  neighbors and the degree bins it touches. Graph implementations may
  buffer an individual node's neighbors, and iterator allocation can raise
  temporary heap usage during the full-graph preparation scan.

- **Large annual bipartite views**: Use `VisitBatchBipartiteOverlap` with a
  bridge-degree cutoff and a measured per-provider candidate cap. Increasing
  the cap from 512 to 2,048 on the 2023 Medicare prescribing projection cut
  capped provider searches from 92,628 to 10,919, while increasing overlap
  time from 9.72 to 11.70 seconds. Keep `TruncatedLeftNodes` in the result
  metadata. Seed `FindNearBicliqueBlocks` from bounded overlap pairs on the
  same year and relationship type; its search is local and non-exhaustive.
  `VisitAnnualTrajectories` reads the immutable relationship store directly,
  requires observed years, and uses memory proportional to providers times
  the requested year count. See the [Integrity annual analysis guide](docs/integrity-annual-analysis.md)
  for boundaries and the full-scale benchmark.

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
| Batch bipartite overlap | O(E + pair encounters + P·C log C), with candidate cap C | O(V+C+D), with eligible bridge-degree cap D | Per-provider |
| Seeded near-biclique blocks | O(E + bounded local scans per seed) | O(V + candidate caps + retained blocks) | Per-seed |
| Annual provider trajectories | O(E + P·Y²), Y ≤ 16 | O(P·Y) | Per-provider after peer summary |
| Triangle count | O(V·d²) | O(V+E) | Per-node |
| Clustering coeff | O(V·d²) | O(V+E) | Per-node |
| MST (Kruskal) | O(E log E) | O(V+E) | No |
| Hopcroft-Karp | O(E√V) | O(V+E) | No |
| Bounded path expansion | O(explored path steps), potentially exponential in hop limit | O(hop limit + active neighbor iterators) | No |
| Yen K-shortest | O(KV(V log V + E)) | O(V+E) | Per-spur |
| SimRank | O(kV²d²) | O(V²) | Per-iteration |
| Node2Vec walks | O(walks·length) | O(walks·length) | Per-walk |
| Spectral clustering | O(V³) | O(V²) | Limited |
