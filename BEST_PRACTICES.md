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

**Memory behavior:** By default, adjacency data is preloaded into
memory at open time. For a graph with E edges, this uses roughly
`E * 40` bytes. If the estimated cost exceeds 70% of available system
memory, it logs a warning and falls back to pure disk reads.

| Edges | Preload Memory |
|---|---|
| 1M | ~40 MB |
| 10M | ~400 MB |
| 83M | ~3.3 GB |
| 500M | ~20 GB |

For very large graphs, pass `diskgraph.NoPreload`:

```go
g, _ := diskgraph.OpenUndirected("huge.db", diskgraph.NoPreload)
```

**When to use `simple.UndirectedGraph` instead:** Only when you need a
mutable graph (adding/removing edges after construction) or for
ephemeral throwaway graphs in tests.

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
   internally. With diskgraph, adjacency is auto-preloaded so this
   is already handled.

4. **Use `runtime.GOMAXPROCS`** to control parallelism. The parallel
   variants auto-detect available cores but you can tune this for
   shared environments.

5. **Graph implementations must be safe for concurrent reads** when
   using parallel functions. Both `diskgraph` and gonum's
   `simple.*Graph` types satisfy this requirement.

6. **For DuckDB**, use `?access_mode=READ_ONLY` when loading to
   allow concurrent readers.

7. **Memory estimation**: `diskgraph` with preloading uses ~40 bytes
   per edge for adjacency data. The bolt file on disk uses ~50 bytes
   per edge. A 5.8M node / 82.6M edge graph needs ~3.3 GB for preload
   plus the bolt file. Without preloading (`NoPreload`), the graph uses
   only the OS page cache — no Go heap allocation.

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
