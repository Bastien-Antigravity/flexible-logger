package interfaces

import "github.com/Bastien-Antigravity/flexible-logger/src/models"

// -----------------------------------------------------------------------------
// Serializer converts entry to bytes
type Serializer interface {
	// -------------------------------------------------------------------------
	// Serialize converts a log entry into a byte slice.
	Serialize(entry *models.LogEntry) ([]byte, error)
}
