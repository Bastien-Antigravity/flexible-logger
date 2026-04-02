package sink

import (
	"fmt"
	"sync"

	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/error_handler"
)

// -----------------------------------------------------------------------------
// AsyncSink wraps another sink and writes asynchronously using a buffered channel.
type AsyncSink struct {
	next       interfaces.Sink
	buffer     chan *models.LogEntry
	wg         sync.WaitGroup
	bufferSize int
	OnError    func(error, *models.LogEntry) // Optional error handler
}

// -----------------------------------------------------------------------------
func NewAsyncSink(next interfaces.Sink, bufferSize int) *AsyncSink {
	s := &AsyncSink{
		next:       next,
		buffer:     make(chan *models.LogEntry, bufferSize),
		bufferSize: bufferSize,
	}
	s.wg.Add(1)
	go s.worker()
	return s
}

// -----------------------------------------------------------------------------
func (s *AsyncSink) SetOnError(onError func(error, *models.LogEntry)) {
	s.OnError = onError
}

// -----------------------------------------------------------------------------

func (s *AsyncSink) worker() {
	defer s.wg.Done()
	for entry := range s.buffer {
		if err := s.next.Write(entry); err != nil {
			if s.OnError != nil {
				s.OnError(err, entry)
			} else {
				// Use global reporter by default
				error_handler.ReportInternalError(entry.LoggerName, "AsyncSink.worker", err, entry.Message)
			}
		}
	}
}

// -----------------------------------------------------------------------------

func (s *AsyncSink) Write(entry *models.LogEntry) error {
	select {
	case s.buffer <- entry:
		return nil
	default:
		entry.Release() // Drop = Release ownership
		return fmt.Errorf("buffer full, log dropped")
	}
}

// -----------------------------------------------------------------------------

func (s *AsyncSink) Close() error {
	close(s.buffer)
	s.wg.Wait()
	return s.next.Close()
}
