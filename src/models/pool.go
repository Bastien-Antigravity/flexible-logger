package models

import "sync"

// -----------------------------------------------------------------------------
// Global Pool for LogEntries
var EntryPool = sync.Pool{
	New: func() interface{} {
		return &LogEntry{}
	},
}
