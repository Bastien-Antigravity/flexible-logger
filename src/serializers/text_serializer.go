package serializers

import (
	"fmt"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// TextSerializer serializes logs to a human-readable text format.
type TextSerializer struct{}

// -----------------------------------------------------------------------------
func NewTextSerializer() *TextSerializer {
	return &TextSerializer{}
}

// -----------------------------------------------------------------------------
func (s *TextSerializer) Serialize(entry *models.LogEntry) ([]byte, error) {
	// Format: [TIMESTAMP] [LEVEL] LOGGER: MESSAGE\n
	str := fmt.Sprintf("[%s] [%d] %s: %s\n",
		entry.Timestamp.UTC().Format(time.RFC3339),
		entry.Level,
		entry.LoggerName,
		entry.Message,
	)
	return []byte(str), nil
}
