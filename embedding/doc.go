// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package embedding provides graph embedding algorithms that produce
// fixed-dimensional vector representations of nodes.
//
// # Custom implementations (from academic papers)
//
//   - [Node2VecWalks] — biased random walks with return parameter p and
//     in-out parameter q (Grover & Leskovec, KDD 2016).
//   - [DeepWalkWalks] — uniform random walks (Perozzi et al., KDD 2014).
//     Equivalent to Node2Vec with p=1, q=1.
//   - [Embed] — compute node embeddings from random walks via pointwise
//     mutual information (PMI) matrix and truncated SVD. This is a
//     lightweight alternative to skip-gram that runs entirely in Go.
package embedding
