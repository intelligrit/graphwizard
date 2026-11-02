// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package similarity

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"

	"gonum.org/v1/gonum/graph"
)

// DegreeBinOverlap describes the overlap of two left-side nodes among bridge
// nodes with the same degree. PValue is the uncorrected hypergeometric upper
// tail probability for observing at least Shared common neighbors in this bin.
type DegreeBinOverlap struct {
	BridgeDegree int
	Population   int
	ACount       int
	BCount       int
	Shared       int
	PValue       float64
}

// BipartiteOverlapResult reports shared neighbors and degree-aware evidence
// for one pair of nodes on the left side of a bipartite graph.
type BipartiteOverlapResult struct {
	A, B        int64
	Shared      int
	RarityScore float64 // Sum of 1/degree for each shared bridge node.
	Bins        []DegreeBinOverlap
}

// BipartiteOverlapScorer evaluates left-side pairs without constructing a
// projected graph. It retains only a histogram of bridge-node degrees. The
// graph and partition function must remain unchanged after construction.
type BipartiteOverlapScorer struct {
	g                graph.Undirected
	isLeft           func(int64) bool
	bridgePopulation map[int]int
}

// NewBipartiteOverlapScorer scans an undirected bipartite graph once to count
// bridge nodes by degree and verify that every edge crosses the partition.
// isLeft classifies nodes in the partition whose pairs will be scored. The
// working memory is proportional to the number of distinct bridge degrees,
// beyond memory used by the graph's own node and neighbor iterators.
func NewBipartiteOverlapScorer(ctx context.Context, g graph.Undirected, isLeft func(int64) bool) (*BipartiteOverlapScorer, error) {
	if isLeft == nil {
		return nil, errors.New("similarity: isLeft must not be nil")
	}
	s := &BipartiteOverlapScorer{
		g: g, isLeft: isLeft, bridgePopulation: make(map[int]int),
	}
	nodes := g.Nodes()
	for nodes.Next() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		id := nodes.Node().ID()
		left := isLeft(id)
		degree := 0
		neighbors := g.From(id)
		for neighbors.Next() {
			next := neighbors.Node().ID()
			if left == isLeft(next) {
				return nil, fmt.Errorf("similarity: edge %d-%d does not cross bipartite partition", id, next)
			}
			degree++
			if degree&1023 == 0 {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
			}
		}
		if !left {
			s.bridgePopulation[degree]++
		}
	}
	return s, nil
}

// Score compares two distinct nodes on the left side. RarityScore discounts
// popular bridge nodes. Each bin's PValue tests overlap against a null model
// that preserves both left-node degrees within that bridge-degree bin. Apply
// a multiple-testing correction across every pair and bin tested before
// treating any PValue as statistically significant.
//
// Score uses memory proportional to A's bridge neighbors and the number of
// degree bins in A and B. It does not enumerate or retain other left-side pairs.
// The graph may buffer a node's neighbors while serving a From call.
//
// Reference: M. Tumminello et al., "Statistically Validated Networks in
// Bipartite Complex Systems", PLOS ONE 6(3), 2011.
func (s *BipartiteOverlapScorer) Score(ctx context.Context, a, b int64) (BipartiteOverlapResult, error) {
	result := BipartiteOverlapResult{A: a, B: b}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if a == b || s.g.Node(a) == nil || s.g.Node(b) == nil || !s.isLeft(a) || !s.isLeft(b) {
		return result, errors.New("similarity: score requires two distinct left-side nodes in the graph")
	}

	aNeighbors := make(map[int64]int)
	aCounts := make(map[int]int)
	bCounts := make(map[int]int)
	sharedCounts := make(map[int]int)

	neighbors := s.g.From(a)
	for neighbors.Next() {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		bridge := neighbors.Node().ID()
		degree, err := s.bridgeDegree(ctx, bridge)
		if err != nil {
			return result, err
		}
		aNeighbors[bridge] = degree
		aCounts[degree]++
	}
	neighbors = s.g.From(b)
	for neighbors.Next() {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		bridge := neighbors.Node().ID()
		degree, common := aNeighbors[bridge]
		if !common {
			var err error
			degree, err = s.bridgeDegree(ctx, bridge)
			if err != nil {
				return result, err
			}
		} else {
			result.Shared++
			result.RarityScore += bridgeRarity(degree)
			sharedCounts[degree]++
		}
		bCounts[degree]++
	}

	degreeSet := make(map[int]bool, len(aCounts)+len(bCounts))
	for degree := range aCounts {
		degreeSet[degree] = true
	}
	for degree := range bCounts {
		degreeSet[degree] = true
	}
	degrees := make([]int, 0, len(degreeSet))
	for degree := range degreeSet {
		degrees = append(degrees, degree)
	}
	slices.Sort(degrees)
	for _, degree := range degrees {
		population := s.bridgePopulation[degree]
		ac, bc, shared := aCounts[degree], bCounts[degree], sharedCounts[degree]
		result.Bins = append(result.Bins, DegreeBinOverlap{
			BridgeDegree: degree,
			Population:   population,
			ACount:       ac,
			BCount:       bc,
			Shared:       shared,
			PValue:       hypergeometricOverlapTail(population, ac, bc, shared),
		})
	}
	return result, nil
}

func bridgeRarity(degree int) float64 { return 1 / float64(degree) }

func (s *BipartiteOverlapScorer) bridgeDegree(ctx context.Context, id int64) (int, error) {
	degree := 0
	neighbors := s.g.From(id)
	for neighbors.Next() {
		degree++
		if degree&1023 == 0 {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}
	}
	return degree, nil
}

// hypergeometricOverlapTail returns P[X >= observed] for a hypergeometric
// random variable with population N and subset sizes a and b.
func hypergeometricOverlapTail(N, a, b, observed int) float64 {
	if observed <= 0 {
		return 1
	}
	low := max(0, a+b-N)
	high := min(a, b)
	if observed <= low {
		return 1
	}
	if observed > high {
		return 0
	}
	// Sum the upper tail directly when overlap exceeds its expectation so
	// small significant probabilities are not lost to subtraction.
	if float64(observed) > float64(a)*float64(b)/float64(N) {
		return math.Min(1, math.Exp(hypergeometricLogSum(N, a, b, observed, high)))
	}
	return math.Max(0, 1-math.Exp(hypergeometricLogSum(N, a, b, low, observed-1)))
}

func hypergeometricLogSum(N, a, b, start, end int) float64 {
	logP := logChoose(a, start) + logChoose(N-a, b-start) - logChoose(N, b)
	logSum := math.Inf(-1)
	for x := start; x <= end; x++ {
		logSum = logAddExp(logSum, logP)
		if x < end {
			logP += math.Log(float64(a-x)) + math.Log(float64(b-x)) -
				math.Log(float64(x+1)) - math.Log(float64(N-a-b+x+1))
		}
	}
	return logSum
}

func logAddExp(a, b float64) float64 {
	if a == math.Inf(-1) {
		return b
	}
	if b > a {
		a, b = b, a
	}
	return a + math.Log1p(math.Exp(b-a))
}

func logChoose(n, k int) float64 {
	k = min(k, n-k)
	if k <= 64 {
		// Avoid cancellation between large log-gamma values when the
		// selected subset is small relative to the population.
		result := 0.0
		for i := 1; i <= k; i++ {
			result += math.Log(float64(n-k+i)) - math.Log(float64(i))
		}
		return result
	}
	a, _ := math.Lgamma(float64(n) + 1)
	b, _ := math.Lgamma(float64(k) + 1)
	c, _ := math.Lgamma(float64(n-k) + 1)
	return a - b - c
}
