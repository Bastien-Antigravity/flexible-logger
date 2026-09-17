package models

// =============================================================================
// ESSENTIAL PROCESS: Global sync.Pool registry for zero-allocation recycling of LogEntry objects.
//
// DATA FLOW:
//   1. Provides synchronized object caching for LogEntry structs.
//   2. Reduces garbage collection overhead on high-frequency logging paths.
//
// KEY PARAMETERS:
//   - EntryPool: Package-level sync.Pool instance.
// =============================================================================

import "sync"

// -----------------------------------------------------------------------------
// Global Pool for LogEntries
var EntryPool = sync.Pool{
	New: func() interface{} {
		return &LogEntry{}
	},
}
