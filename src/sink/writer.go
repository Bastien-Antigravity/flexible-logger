package sink

import (
	"io"
	"sync"

	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// WriterSink wraps an io.Writer (like a file or socket).
type WriterSink struct {
	w          io.Writer
	serializer interfaces.Serializer
	mu         sync.Mutex
}

// -----------------------------------------------------------------------------
func NewWriterSink(w io.Writer, serializer interfaces.Serializer) *WriterSink {
	return &WriterSink{
		w:          w,
		serializer: serializer,
	}
}

// -----------------------------------------------------------------------------
func (s *WriterSink) Write(entry *models.LogEntry) error {
	defer entry.Release() // Release ownership
	data, err := s.serializer.Serialize(entry)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.w.Write(data)
	return err
}

// -----------------------------------------------------------------------------
func (s *WriterSink) Close() error {
	if closer, ok := s.w.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
