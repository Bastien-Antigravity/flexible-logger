package serializers

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
	
	// Add a newline for file/console stream compatibility
	data = append(data, '\n')
	return data, nil
}
