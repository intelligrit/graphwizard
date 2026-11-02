# Annual bipartite analysis for Integrity

GraphWizard can keep the full relationship `multigraph.Store` resident and
analyze several observed years in one process. The APIs serve different
questions:

| API | Question | Additional live memory |
|---|---|---|
| `multigraph.Store.VisitAnnualChanges` | Which provider–entity connections, amounts, and concentration changed between two years? | Selected row IDs for two years plus provider summaries; edge results stream to a callback. |
| `multigraph.Store.VisitAnnualTrajectories` | Which providers show a sustained change across several years relative to degree peers? | One bounded annual series per provider; results stream to a callback. No graph projection. |
| `similarity.VisitBatchBipartiteOverlap` | Which provider pairs share several uncommon drugs or manufacturers? | Node degrees and one provider's capped candidates; results stream to a callback. |
| `anomaly.FRAUDAR` | Is there a dense provider–entity block after discounting popular entities? | Node arrays and a priority heap; no pair projection or edge copy. |
| `anomaly.FindNearBicliqueBlocks` | Are several bounded, nearly complete provider–entity groups present around overlap pairs? | Node degrees plus capped local candidate sets and the returned blocks. |

These are complementary to existing Leiden communities, local personalized
PageRank, and the per-pair `BipartiteOverlapScorer`. The batch scorer does not
return p-values: candidate selection and caps would invalidate an uncorrected
pairwise test. `DroppedPairEncounters` and `TruncatedLeftNodes` quantify
skipped work; each match carries `CandidateSearchTruncated` so investigators
can see when its provider's top-K may be incomplete. Ranking uses the sum of
`1 / bridge degree` for
shared eligible entities, divided by the geometric mean of provider degrees.
It returns all shared entities as evidence, including hubs excluded from pair
discovery. Compare against raw shared count on the same candidate set.

## Historical rules

- Prescribing observations cover **2013–2023**. Only compare adjacent years
  within that range. Use `Type: "prescribed"` and `Weight: "claims"` or
  `Weight: "dollars"`.
- Payments cover **2015–2024**. Use `TypePrefix: "payment/"` and
  `Weight: "dollars"`. The annual edge key includes the exact payment type.
- Affiliations are **undated and current**. They cannot be included in a
  historical annual comparison or interpreted as existing in 2023.
- `VisitAnnualChanges` selects only the two explicit positive years and, by
  default, rejects a requested type with no rows in either year. A recorded
  zero has `Present=true`; a missing named amount has `Present=false` and
  `Complete=false`. Partial provider totals are not scored as complete.
  `AnnualMetric.Value` is an explicit sum; `Store.Weight` retains the original
  exact float64 value and its presence bit for each relationship ID.
- The local fixture aggregates source rows within provider, entity, type,
  and year. Its relationships are annual aggregates, not payment events.
- `VisitAnnualTrajectories` requires an observed source row in **every**
  requested year and at least two complete years on each side of a candidate
  break by default. It does not treat absent, missing, and recorded-zero
  values as equivalent. Its peer cohort contains only providers with a
  complete series and uses first-year relationship-count bins. A requested
  year with no source rows is an error rather than an inferred zero.
- An annual change edge is a provider–entity–**exact type** key. Its payment
  concentration is across those typed connections. A payment overlap graph
  filtered by `payment/` explicitly combines payment types at the same
  provider–manufacturer pair; its pair evidence identifies manufacturers.
  Join those IDs back to the unchanged store when an investigator needs the
  exact payment type, amount, and year for each underlying relationship.

## Integration pattern

Keep `store` alive through the whole run. For each year, build one filtered
projection, run the graph algorithms, then let that projection go before the
next year. The annual change API reads the store directly and needs no graph
projection.

```go
_, err := store.VisitAnnualChanges(ctx, multigraph.AnnualChangeOptions{
    Earlier: 2022, Later: 2023, Type: "prescribed", Weight: "claims",
}, visitProviderChange, visitEdgeChange)
if err != nil { return err }

_, err = store.VisitAnnualTrajectories(ctx, multigraph.TrajectoryOptions{
    FirstYear: 2013, LastYear: 2023,
    Type: "prescribed", Weight: "claims",
}, visitTrajectory)
if err != nil { return err }

for _, year := range []int32{2022, 2023} {
    if err := func() error {
        var seeds []anomaly.NearBicliqueSeed // bounded pairs collected below
        g, err := store.ProjectUndirected(multigraph.Projection{
            Combine: multigraph.Sum,
            Filter: func(r multigraph.Relationship) bool {
                return r.Type == "prescribed" && r.Year == year
            },
        })
        if err != nil { return err }
        _, err = similarity.VisitBatchBipartiteOverlap(ctx, g, isProvider,
            similarity.BatchOverlapOptions{
                MaxBridgeDegree: 1024, MaxCandidatesPerLeft: 512,
                TopK: 3, MinSeedShared: 2,
            }, func(m similarity.BatchOverlapMatch) error {
                if err := visitPair(m); err != nil { return err }
                if len(seeds) < 200 {
                    seeds = append(seeds, anomaly.NearBicliqueSeed{A: m.A, B: m.B})
                }
                return nil
            })
        if err != nil { return err }
        if year == 2023 {
            _, err = anomaly.FRAUDAR(ctx, g, isProvider, 2000)
            if err != nil { return err }
            _, _, err = anomaly.FindNearBicliqueBlocks(ctx, g, isProvider,
                seeds, anomaly.DefaultNearBicliqueOptions())
        }
        return err
    }(); err != nil { return err }
}
```

`isProvider` must use Integrity's node-kind mapping. The temporal API expects
`From` to be the provider and `To` the entity for the selected relationship
types. The graph algorithms require a provider/entity bipartition. For
payments, use a filter such as `strings.HasPrefix(r.Type, "payment/")` with
an exact year. Do not add `affiliated` or future rows to that projection.

`FindNearBicliqueBlocks` is a bounded seeded heuristic. Each reported provider
and entity meets the requested local coverage threshold, and every block
reports density, rarity-weighted density, supporting entity IDs and global
degrees, and whether candidate caps were hit. It is not an exhaustive maximum
biclique or k-biplex solver. Seed it with overlap pairs from the **same annual
projection**. Keep a mix of high overlap scores and a deterministic sample of
lower-ranked pairs to reduce selection bias; the selected seeds still limit
what the search can find. Set `MaxProviders`, `MaxBridges`, candidate limits,
and `TopK` to bound both work and output.

The overlap cap is a coverage and runtime control, not a statistical test.
Use a larger `MaxCandidatesPerLeft` on high-priority providers or the full
projection when resources allow, and report `TruncatedLeftNodes` alongside
results. Do not report p-values from either selected overlap pairs or seeded
blocks without a separately specified testing universe and selection-aware
multiple-testing correction.

## Cross-network limits

The current graph has provider–drug and provider–manufacturer relationships,
but no dated drug–manufacturer crosswalk. A local DuckDB audit found product
manufacturer observations only for 2023, with many unmatched or ambiguous
drug names and very little exact manufacturer-name agreement with payment
nodes. Current `ndc` labels and current affiliations cannot establish past
ownership or affiliation. Keep prescribing and payment analyses separate
until a validated, dated product–company mapping is available. Annual
aggregates also cannot support event-time lockstep methods such as CopyCatch.

## Full-scale check

On the local 290,789,017-relationship fixture, one packed/block store load
retained 6.99 GiB of payload. Four year-pair analyses and three sequential
annual projections completed in 245 seconds end to end; peak process RSS was
43.14 GiB, reached during ingestion and essentially the prior 43.1 GiB load
baseline. The 2022–2023 prescribing comparison took 2.82 seconds and selected
exactly 48,713,390 rows, matching the fixture manifest. The 2023–2024
payment comparison took 1.45 seconds and selected 7,396,619 rows. Each
projection released back to within 30 KiB of the store-only heap baseline.

With a bridge-degree cutoff of 1,024 and 512 candidates per provider, 2023
prescribing overlap took 8.65 seconds and returned 143,066 top matches for
50,653 providers; 92,628 providers reached the candidate cap. The 2024
payment pass took 4.29 seconds, returning 67,512 matches for 24,101 providers;
58,438 reached the cap. FRAUDAR took 6.45 seconds on the 2023 prescribing
view and found a 3,180-provider by 622-drug block with 39.5% of possible
edges. Three overlap stream digests, four annual result counts, and the
FRAUDAR block matched a repeated full-fixture run. The local benchmark report
and JSON remain under ignored `.local/multigraph-large/`.

The provider change API reports new/lost connections, degree-normalized
turnover, log total change, and Herfindahl concentration shift, with degree-bin
peer means and deviations for each. Continuing edges report log change adjusted by the provider's total
change and change in share. These are descriptive features. Before treating
any output as a fraud signal, compare it with simple turnover, raw amount
change, raw shared count, and unweighted density on degree-matched providers
using independently reviewed outcomes. The public Medicare fixture supplies
no adjudicated fraud labels, so a structural result alone cannot establish
precision or incremental investigative yield.

FRAUDAR uses the logarithmic column weighting and minimum-degree peeling
from [Hooi et al., KDD 2016](https://www.cs.cmu.edu/~christos/PUBLICATIONS/kdd16-fraudar.pdf).
It reports one best block per annual projection. It is a lead for review,
not a claim that its members committed fraud.

## Expanded full-scale check

An expanded run loaded the same 290,789,017 relationships once, ran both
multi-year trajectory scans and four adjacent-year comparisons directly on
that store, then released each of three annual projections before building
the next. It completed in 274.6 seconds. Peak process RSS was **43.15 GiB**,
reached during ingestion, versus 43.14 GiB in the earlier run. Retained store
payload remained 6.99 GiB. After each new pass, measured Go heap returned
to within 75 KiB of the store-only baseline.

| Annual view | Overlap cap | Provider searches at cap | Overlap time | Multi-block time |
|---|---:|---:|---:|---:|
| 2023 prescribing | 512 | 92,628 | 9.72 s | — |
| 2023 prescribing | 2,048 | 10,919 | 11.70 s | 0.95 s |
| 2024 payments | 512 | 58,438 | 4.94 s | — |
| 2024 payments | 2,048 | 951 | 5.53 s | 0.52 s |

In a fixed sample of 100 providers with results under both caps, the top
overlap match changed for seven prescribing providers and one payment
provider. The 2023 seeded search returned ten diverse blocks from 152 unique
seeds; its largest high-ranked example had 37 providers, seven drugs, 243
edges, and 93.8% density. The 2024 payment search returned ten blocks from
151 unique seeds. The global
FRAUDAR blocks were much larger and less dense: 3,180 by 622 at 39.5% in
2023 prescribing, and 669 by 88 at 50.9% in 2024 payments. These size and
density differences show that the bounded search exposes smaller structural
groups; they do not establish investigative yield.

The 2013–2023 prescribing trajectory pass took 5.15 seconds and scored
370,247 of 1,166,153 observed providers with complete series and enough peer
variation. The 2015–2024 payment pass took 1.30 seconds and scored 161,286
of 1,135,021 observed providers. Many providers were not observed in every
requested year, so these scores cover a selected longitudinal cohort. The
benchmark verified annual row counts against the fixture manifest. A complete
repeat took 268.2 seconds and peaked at 43.13 GiB RSS; source and store
counts, trajectory score digests, temporal results, overlap digests, selected
block summaries, and FRAUDAR blocks matched. The private JSON and report are
under ignored `.local/multigraph-large/`.

These comparisons establish runtime, bounded search coverage, and structural
differences from simple density and overlap baselines. They cannot measure
fraud precision, recall, or incremental investigator yield without reviewed
outcomes. Do not promote scores to fraud probabilities or automatic flags on
this evidence alone.
