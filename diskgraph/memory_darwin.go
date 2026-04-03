// Copyright (c) 2026 Intelligrit. MIT License. See LICENSE.

//go:build darwin

package diskgraph

import (
	"encoding/binary"
	"syscall"
)

func totalSystemMemory() uint64 {
	raw, err := syscall.Sysctl("hw.memsize")
	if err != nil {
		return 0
	}
	// Sysctl returns the value as a raw byte string (may be 7 or 8 bytes
	// due to null-terminator stripping). Copy into a fixed-size buffer.
	buf := make([]byte, 8)
	copy(buf, []byte(raw))
	return binary.LittleEndian.Uint64(buf)
}
