package sink

// =============================================================================
// ESSENTIAL PROCESS: Thread-safe sink wrapping an arbitrary io.Writer (files, network connections, buffers) with a Serializer.
//
// DATA FLOW:
//   1. Accepts LogEntry and defers entry.Release().
//   2. Serializes entry using configured Serializer.
//   3. Acquires mutex lock and writes byte slice to io.Writer.
//
// KEY PARAMETERS:
//   - w: Target io.Writer destination.
//   - serializer: Formatter encoding LogEntry into bytes.
//   - mu: Mutex synchronizing writes.
// =============================================================================

import (
	"io"
	"os"
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
	closeOnce  sync.Once
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
	var err error
	s.closeOnce.Do(func() {
		if s.w == os.Stdout || s.w == os.Stderr || s.w == os.Stdin {
			return
		}
		if closer, ok := s.w.(io.Closer); ok {
			err = closer.Close()
		}
	})
	return err
}
