package sink

// =============================================================================
// ESSENTIAL PROCESS: Standard console output sink routing formatted text logs to os.Stderr.
//
// DATA FLOW:
//   1. Wraps WriterSink targeting os.Stderr.
//   2. Emits logs formatted with TextSerializer.
//
// KEY PARAMETERS:
//   - ConsoleSink: Sink struct.
//   - inner: Internal WriterSink instance.
// =============================================================================

import (
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
)

// -----------------------------------------------------------------------------
// ConsoleSink writes to stdout in a human-readable format.
type ConsoleSink struct {
	inner *WriterSink
}

// -----------------------------------------------------------------------------
func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{
		inner: NewWriterSink(os.Stderr, serializers.NewTextSerializer()),
	}
}

// -----------------------------------------------------------------------------
func (s *ConsoleSink) Write(entry *models.LogEntry) error {
	// WriterSink.Write already handles entry.Release()
	return s.inner.Write(entry)
}

// -----------------------------------------------------------------------------
func (s *ConsoleSink) Close() error {
	return s.inner.Close()
}
