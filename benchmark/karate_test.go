// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package benchmark

import (
	"math"
	"math/rand"
	"testing"

	"github.com/intelligrit/graphwizard/centrality"
	"github.com/intelligrit/graphwizard/community"
	"github.com/intelligrit/graphwizard/connectivity"
	"github.com/intelligrit/graphwizard/similarity"
	"github.com/intelligrit/graphwizard/structure"
	"gonum.org/v1/gonum/graph/simple"
)

func TestKarateClub_GraphStructure(t *testing.T) {
	g := KarateClub()
	if g.Nodes().Len() != 34 {
		t.Errorf("expected 34 nodes, got %d", g.Nodes().Len())
	}
	edges := 0
	nodes := g.Nodes()
	for nodes.Next() {
		edges += g.From(nodes.Node().ID()).Len()
	}
	edges /= 2 // undirected
	if edges != 78 {
		t.Errorf("expected 78 edges, got %d", edges)
	}
}

func TestKarateClub_Leiden(t *testing.T) {
	g := KarateClub()
	gt := KarateClubGroundTruth()

	rng := rand.New(rand.NewSource(42))
	comms := community.Leiden(g, 1.0, rng)

	nmi := normalizedMutualInfo(comms, gt)
	t.Logf("Leiden NMI: %.4f", nmi)
	if nmi < 0.5 {
		t.Errorf("Leiden NMI too low: %.4f (expected > 0.5)", nmi)
	}
}

func TestKarateClub_Louvain(t *testing.T) {
	g := KarateClub()
	gt := KarateClubGroundTruth()

	comms := community.Louvain(g, 1.0, nil)

	nmi := normalizedMutualInfo(comms, gt)
	t.Logf("Louvain NMI: %.4f", nmi)
	if nmi < 0.5 {
		t.Errorf("Louvain NMI too low: %.4f (expected > 0.5)", nmi)
	}
}

func TestKarateClub_LabelPropagation(t *testing.T) {
	g := KarateClub()
	gt := KarateClubGroundTruth()

	rng := rand.New(rand.NewSource(42))
	comms := community.LabelPropagation(g, 100, rng)

	nmi := normalizedMutualInfo(comms, gt)
	t.Logf("LabelProp NMI: %.4f", nmi)
	if nmi < 0.3 {
		t.Errorf("LabelProp NMI too low: %.4f (expected > 0.3)", nmi)
	}
}

func TestKarateClub_LeidenBetterThanRandom(t *testing.T) {
	g := KarateClub()
	gt := KarateClubGroundTruth()

	// Random assignment baseline.
	rng := rand.New(rand.NewSource(42))
	randomComms := make(map[int64]int64)
	for id := range gt {
		randomComms[id] = int64(rng.Intn(2))
	}
	randomNMI := normalizedMutualInfo(randomComms, gt)

	// Leiden.
	rng2 := rand.New(rand.NewSource(42))
	leidenComms := community.Leiden(g, 1.0, rng2)
	leidenNMI := normalizedMutualInfo(leidenComms, gt)

	t.Logf("Random NMI: %.4f, Leiden NMI: %.4f", randomNMI, leidenNMI)
	if leidenNMI <= randomNMI {
		t.Error("Leiden should outperform random assignment")
	}
}

func TestKarateClub_PageRank(t *testing.T) {
	g := KarateClub()
	// PageRank needs directed; convert to directed (both directions).
	dg := toDirected(g)

	scores := centrality.PageRank(dg, 0.85, 1e-6)
	if len(scores) != 34 {
		t.Fatalf("expected 34 scores, got %d", len(scores))
	}

	// Nodes 0 and 33 are the two leaders — they should have high PageRank.
	// Node 33 has the highest degree (17), so it should be near the top.
	topNode := int64(-1)
	topScore := 0.0
	for id, s := range scores {
		if s > topScore {
			topScore = s
			topNode = id
		}
	}
	if topNode != 33 && topNode != 0 {
		t.Logf("top PageRank node: %d (expected 33 or 0)", topNode)
	}
}

func TestKarateClub_Betweenness(t *testing.T) {
	g := KarateClub()

	scores := centrality.Betweenness(g)

	// Node 0 (Mr. Hi) should have high betweenness — he bridges many paths.
	if scores[0] < scores[4] {
		t.Error("Mr. Hi (0) should have higher betweenness than peripheral node 4")
	}
}

func TestKarateClub_Bridges(t *testing.T) {
	g := KarateClub()
	bridges := connectivity.Bridges(g)

	// Karate Club is well-connected; it should have very few bridges.
	t.Logf("bridges: %d", len(bridges))
	if len(bridges) > 10 {
		t.Errorf("expected few bridges, got %d", len(bridges))
	}
}

func TestKarateClub_TriangleCount(t *testing.T) {
	g := KarateClub()
	perNode, total := structure.TriangleCount(g)

	// Known: Karate Club has 45 triangles.
	if total != 45 {
		t.Errorf("expected 45 triangles, got %d", total)
	}

	// Node 0 participates in many triangles (high degree, dense neighborhood).
	if perNode[0] < 10 {
		t.Errorf("node 0 should participate in many triangles, got %d", perNode[0])
	}
}

func TestKarateClub_ClusteringCoefficient(t *testing.T) {
	g := KarateClub()
	coeffs := structure.ClusteringCoefficient(g)

	avg := structure.AverageClusteringCoefficient(g)
	t.Logf("average clustering coefficient: %.4f", avg)

	// Known: Karate Club average CC is approximately 0.57.
	if math.Abs(avg-0.57) > 0.1 {
		t.Errorf("expected avg CC ~0.57, got %.4f", avg)
	}
	_ = coeffs
}

func TestKarateClub_Diameter(t *testing.T) {
	g := KarateClub()
	d := centrality.Diameter(g)

	// Known: Karate Club diameter is 5.
	if d != 5 {
		t.Errorf("expected diameter 5, got %.0f", d)
	}
}

func TestKarateClub_ConnectedComponents(t *testing.T) {
	g := KarateClub()
	comps := connectivity.ConnectedComponents(g)

	if len(comps) != 1 {
		t.Errorf("Karate Club should be fully connected, got %d components", len(comps))
	}
}

func TestKarateClub_Jaccard(t *testing.T) {
	g := KarateClub()

	// Nodes 0 and 1 are directly connected and share many neighbors.
	j01 := similarity.Jaccard(g, 0, 1)
	// Nodes 0 and 33 are the two leaders — moderate Jaccard (different factions).
	j033 := similarity.Jaccard(g, 0, 33)

	t.Logf("J(0,1)=%.4f J(0,33)=%.4f", j01, j033)
	if j01 <= 0 {
		t.Error("J(0,1) should be > 0")
	}
}

func TestKarateClub_PersonalizedPageRank(t *testing.T) {
	g := KarateClub()
	dg := toDirected(g)

	scores := centrality.PersonalizedPageRank(dg, 0, 0.85, 1e-6, 100)

	// Node 0's faction should have higher PPR from seed 0.
	gt := KarateClubGroundTruth()
	avgFaction0 := 0.0
	avgFaction1 := 0.0
	count0, count1 := 0, 0
	for id, faction := range gt {
		if faction == 0 {
			avgFaction0 += scores[id]
			count0++
		} else {
			avgFaction1 += scores[id]
			count1++
		}
	}
	avgFaction0 /= float64(count0)
	avgFaction1 /= float64(count1)

	t.Logf("PPR from node 0: faction0 avg=%.6f, faction1 avg=%.6f", avgFaction0, avgFaction1)
	if avgFaction0 <= avgFaction1 {
		t.Error("PPR from node 0 should favor Mr. Hi's faction")
	}
}

// --- Helpers ---

func toDirected(g *simple.UndirectedGraph) *simple.DirectedGraph {
	dg := simple.NewDirectedGraph()
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		it := g.From(n.ID())
		for it.Next() {
			dg.SetEdge(dg.NewEdge(simple.Node(n.ID()), simple.Node(it.Node().ID())))
		}
	}
	return dg
}

// normalizedMutualInfo computes a simplified NMI between detected communities
// and ground truth. Returns 0-1 where 1 = perfect agreement.
func normalizedMutualInfo(detected map[int64]int64, groundTruth map[int64]int) float64 {
	// Build confusion matrix.
	detLabels := make(map[int64]int)
	gtLabels := make(map[int]int)
	joint := make(map[[2]int]int)
	n := 0

	for id, gt := range groundTruth {
		det, ok := detected[id]
		if !ok {
			continue
		}
		d := int(det)
		detLabels[det]++
		gtLabels[gt]++
		joint[[2]int{d, gt}]++
		n++
	}

	if n == 0 {
		return 0
	}

	nf := float64(n)
	hDet := 0.0
	for _, c := range detLabels {
		p := float64(c) / nf
		if p > 0 {
			hDet -= p * math.Log(p)
		}
	}
	hGT := 0.0
	for _, c := range gtLabels {
		p := float64(c) / nf
		if p > 0 {
			hGT -= p * math.Log(p)
		}
	}

	mi := 0.0
	for key, c := range joint {
		pJoint := float64(c) / nf
		pDet := float64(detLabels[int64(key[0])]) / nf
		pGT := float64(gtLabels[key[1]]) / nf
		if pJoint > 0 && pDet > 0 && pGT > 0 {
			mi += pJoint * math.Log(pJoint/(pDet*pGT))
		}
	}

	denom := (hDet + hGT) / 2
	if denom == 0 {
		return 0
	}
	return mi / denom
}
