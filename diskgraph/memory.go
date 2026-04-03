// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

package diskgraph

import (
	"runtime"
)

// availableMemory returns an estimate of the memory available for use,
// in bytes. It subtracts the current Go heap usage from the total system
// memory. Returns 0 if the system memory cannot be determined.
func availableMemory() uint64 {
	total := totalSystemMemory()
	if total == 0 {
		return 0
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	used := ms.Sys
	if used >= total {
		return 0
	}
	return total - used
}
