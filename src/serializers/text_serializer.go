package serializers

import (
	"fmt"

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
	// Format: [TIMESTAMP] [LEVEL] [PID] [FILE:LINE] LOGGER: MESSAGE\n
	str := fmt.Sprintf("[%s] [%-8s] [%s] [%s:%s] %s: %s\n",
		entry.Timestamp.UTC().Format("2006-01-02 15:04:05.000"),
		entry.Level.String(),
		entry.ProcessID,
		entry.Filename,
		entry.LineNumber,
		entry.LoggerName,
		entry.Message,
	)
	return []byte(str), nil
}
