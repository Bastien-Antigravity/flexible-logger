package models

import (
	"sync/atomic"
	"time"
)

// -----------------------------------------------------------------------------
// LogEntry represents a single log message.
// It is designed to be pooled to reduce allocations.
type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Level        Level     `json:"level"`
	Message      string    `json:"message"`
	Hostname     string    `json:"hostname"`
	LoggerName   string    `json:"loggerName"`
	Module       string    `json:"module,omitempty"`
	Filename     string    `json:"filename"`
	FunctionName string    `json:"functionName"`
	LineNumber   string    `json:"lineNumber"`
	ProcessID    string    `json:"processID"`
	ProcessName  string    `json:"processName"`
	ThreadID     string    `json:"threadID,omitempty"`
	ThreadName   string    `json:"threadName,omitempty"`
	StackTrace   string    `json:"stackTrace,omitempty"`
	ServiceName  string    `json:"serviceName,omitempty"`
	PathName     string    `json:"pathName,omitempty"`

	// refCount manages the lifecycle of the entry across multiple sinks
	refCount int32
}

// -----------------------------------------------------------------------------
// Reset clears the LogEntry for reuse.
func (e *LogEntry) Reset() {
	// Zero out all fields in one optimized operation
	*e = LogEntry{}

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
