// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

//go:build !darwin && !linux

package diskgraph

func totalSystemMemory() uint64 {
	return 0 // unknown — preload will always be attempted
}
