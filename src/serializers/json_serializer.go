package serializers

// =============================================================================
// ESSENTIAL PROCESS: Structured JSON serialization engine for containerized log processors and file archives.
//
// DATA FLOW:
//   1. Encodes LogEntry struct to JSON bytes.
//   2. Appends standard Unix newline terminator for stream delineation.
//   3. Returns serialized payload.
//
// KEY PARAMETERS:
//   - JSONSerializer: Serializer struct.
// =============================================================================

import (
	"encoding/json"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------

// JSONSerializer serializes logs to a documented JSON format.
type JSONSerializer struct{}

// -----------------------------------------------------------------------------

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// -----------------------------------------------------------------------------

func (s *JSONSerializer) Serialize(entry *models.LogEntry) ([]byte, error) {
	// Simple JSON marshaling using the tags added to LogEntry
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	// Add a newline for container and stream log compatibility (Unix LF)
	data = append(data, '\n')
	return data, nil
}
