package models

import (
	"sync/atomic"
	"time"
)

// -----------------------------------------------------------------------------
// LogEntry represents a single log message.
// It is designed to be pooled to reduce allocations.
type LogEntry struct {
	Timestamp    time.Time
	Level        Level
	Message      string
	Hostname     string
	LoggerName   string
	Module       string
	Filename     string
	FunctionName string
	LineNumber   string
	ProcessID    string
	ProcessName  string
	ThreadID     string
	ThreadName   string
	StackTrace   string
	ServiceName  string
	PathName     string

	// refCount manages the lifecycle of the entry across multiple sinks
	refCount int32
}

// -----------------------------------------------------------------------------
// Reset clears the LogEntry for reuse.
func (e *LogEntry) Reset() {
	e.Timestamp = time.Time{}
	e.Level = LevelNotSet
	e.Message = ""
	e.Hostname = ""
	e.LoggerName = ""
	e.Module = ""
	e.Filename = ""
	e.FunctionName = ""
	e.LineNumber = ""
	e.ProcessID = ""
	e.ProcessName = ""
	e.ThreadID = ""
	e.ThreadName = ""
	e.StackTrace = ""
	e.ServiceName = ""
	e.PathName = ""

	// Reset refCount to 1 (owned by whoever got it from pool)
	atomic.StoreInt32(&e.refCount, 1)
}

// -----------------------------------------------------------------------------
// Retain increments the reference count.
// Must be called when passing the entry to an additional async consumer.
func (e *LogEntry) Retain() {
	atomic.AddInt32(&e.refCount, 1)
}

// -----------------------------------------------------------------------------
// Release decrements the reference count and returns the entry to the pool if 0.
func (e *LogEntry) Release() {
	if atomic.AddInt32(&e.refCount, -1) == 0 {
		EntryPool.Put(e)
	}
}
