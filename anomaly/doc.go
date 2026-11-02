// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

// Package anomaly provides graph-based anomaly detection algorithms. It
// identifies structurally unusual nodes using a combination of degree
// z-scores, local clustering coefficient deviation, and neighbor degree
// variance. FRAUDAR finds one dense bipartite block while downweighting
// popular target nodes. FindNearBicliqueBlocks finds several bounded,
// approximately complete provider/entity groups around supplied seed pairs.
//
// These methods are useful for fraud detection, network intrusion detection,
// and identifying unusual entities in knowledge graphs.
package anomaly
