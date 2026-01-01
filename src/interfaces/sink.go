package interfaces

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
