package interfaces

// =============================================================================
// ESSENTIAL PROCESS: Interface contract for encoding LogEntry models into byte streams.
//
// DATA FLOW:
//   1. Accepts structured LogEntry pointer.
//   2. Encodes entry to target wire format (Cap'n Proto, JSON, or Text).
//   3. Returns encoded byte slice or serialization error.
//
// KEY PARAMETERS:
//   - Serializer: Serialization abstraction implemented by all formatters.
// =============================================================================

import "github.com/Bastien-Antigravity/flexible-logger/src/models"

// -----------------------------------------------------------------------------
// Serializer converts entry to bytes
type Serializer interface {
	// -------------------------------------------------------------------------
	// Serialize converts a log entry into a byte slice.
	Serialize(entry *models.LogEntry) ([]byte, error)
}
