package interfaces

// =============================================================================
// ESSENTIAL PROCESS: Destination sink interface defining standard Write and Close contracts for log routing.
//
// DATA FLOW:
//   1. Ingests LogEntry pointer and forwards to transport medium.
//   2. Enforces entry reference count lifecycle management.
//   3. Closes underlying files, network streams, or buffer workers.
//
// KEY PARAMETERS:
//   - Sink: Unified destination abstraction for Console, Writer, Async, and Multi sinks.
// =============================================================================

import "github.com/Bastien-Antigravity/flexible-logger/src/models"

// -----------------------------------------------------------------------------
// Sink defines where log entries are written
type Sink interface {
	// -------------------------------------------------------------------------
	// Write writes a log entry to the sink.
	Write(entry *models.LogEntry) error

	// -------------------------------------------------------------------------
	// Close closes the sink.
	Close() error
}
