package sink

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
		inner: NewWriterSink(os.Stdout, serializers.NewTextSerializer()),
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
