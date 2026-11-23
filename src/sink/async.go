// -----------------------------------------------------------------------------
// ESSENTIAL PROCESS:
// Asynchronous channel-buffered sink decoupling caller execution from
// underlying I/O latency.
//
// DATA FLOW:
// 1. Receives LogEntry in non-blocking select buffer.
// 2. Background worker goroutine pops entries and retains references.
// 3. Delegates to downstream Sink and releases reference after write.
//
// KEY PARAMETERS:
// - buffer: Buffered channel queue.
// - next: Downstream target sink.
// - droppedCount: Atomic counter tracking entries dropped on buffer saturation.
// - OnError: Optional callback for write failures.
// -----------------------------------------------------------------------------

package sink

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/error_handler"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// AsyncSink wraps another sink and writes asynchronously using a buffered channel.
type AsyncSink struct {
	next              interfaces.Sink
	buffer            chan *models.LogEntry
	wg                sync.WaitGroup
	bufferSize        int
	drainTimeout      time.Duration
	closed            atomic.Bool
	droppedCount      atomic.Uint64
	lastAlertNano     atomic.Int64
	lastAlertedDrops  atomic.Uint64
	dropAlertInterval atomic.Int64
	dropNotifier      interfaces.Notifier
	serviceName       string
	mu                sync.RWMutex
	OnError           func(error, *models.LogEntry)                     // Optional error handler
	OnDropAlert       func(droppedInWindow uint64, totalDropped uint64) // Optional throttled drop alert callback
}

// -----------------------------------------------------------------------------
func NewAsyncSink(next interfaces.Sink, bufferSize int) *AsyncSink {
	s := &AsyncSink{
		next:         next,
		buffer:       make(chan *models.LogEntry, bufferSize),
		bufferSize:   bufferSize,
		drainTimeout: 2 * time.Second,
	}
	s.dropAlertInterval.Store(int64(60 * time.Second))
	s.wg.Add(1)
	go s.worker()
	return s
}

// -----------------------------------------------------------------------------
func (s *AsyncSink) SetOnError(onError func(error, *models.LogEntry)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OnError = onError
}

// -----------------------------------------------------------------------------
// SetDropAlertInterval overrides the minimum duration between throttled buffer-drop alerts (default 60s).
func (s *AsyncSink) SetDropAlertInterval(interval time.Duration) {
	if interval < 0 {
		interval = 0
	}
	s.dropAlertInterval.Store(int64(interval))
}

// -----------------------------------------------------------------------------
// SetOnDropAlert registers an optional callback invoked at most once per dropAlertInterval when logs are dropped.
func (s *AsyncSink) SetOnDropAlert(fn func(droppedInWindow uint64, totalDropped uint64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OnDropAlert = fn
}

// -----------------------------------------------------------------------------
// SetDropNotifier attaches a Notifier (e.g., RemoteNotifier / Telegram) to receive throttled buffer saturation alerts.
func (s *AsyncSink) SetDropNotifier(n interfaces.Notifier, serviceName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropNotifier = n
	s.serviceName = serviceName
}

// -----------------------------------------------------------------------------
// SetDrainTimeout overrides the maximum duration Close() will wait for the buffer to drain.
func (s *AsyncSink) SetDrainTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drainTimeout = timeout
}

// -----------------------------------------------------------------------------

func (s *AsyncSink) worker() {
	defer s.wg.Done()
	for entry := range s.buffer {
		// Retain the entry so it is not recycled back into the pool
		// by s.next.Write's deferred Release() while we are still processing it.
		entry.Retain()

		err := s.next.Write(entry)
		if err != nil {
			if s.OnError != nil {
				s.OnError(err, entry)
			} else {
				// Safely report error since we still hold a reference
				error_handler.ReportInternalError(entry.LoggerName, "AsyncSink.worker", err, entry.Message)
			}
		}

		// Release our own reference to recycle the entry back to the pool
		entry.Release()
	}
}

// -----------------------------------------------------------------------------
// DroppedCount returns the total number of log entries dropped due to buffer saturation.
func (s *AsyncSink) DroppedCount() uint64 {
	return s.droppedCount.Load()
}

// -----------------------------------------------------------------------------

func (s *AsyncSink) Write(entry *models.LogEntry) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed.Load() {
		entry.Release()
		return fmt.Errorf("sink closed")
	}

	select {
	case s.buffer <- entry:
		return nil
	default:
		total := s.droppedCount.Add(1)
		loggerName := entry.LoggerName
		entry.Release() // Drop = Release ownership
		s.maybeTriggerDropAlert(total, loggerName)
		return fmt.Errorf("buffer full, log dropped")
	}
}

// -----------------------------------------------------------------------------

func (s *AsyncSink) maybeTriggerDropAlert(totalDropped uint64, loggerName string) {
	interval := s.dropAlertInterval.Load()
	now := time.Now().UnixNano()
	last := s.lastAlertNano.Load()

	if last != 0 && interval > 0 && (now-last) < interval {
		return
	}

	if !s.lastAlertNano.CompareAndSwap(last, now) {
		return
	}

	prevDrops := s.lastAlertedDrops.Swap(totalDropped)
	droppedInWindow := totalDropped
	if totalDropped >= prevDrops {
		droppedInWindow = totalDropped - prevDrops
	}
	if droppedInWindow == 0 {
		droppedInWindow = 1
	}

	svc := s.serviceName
	if svc == "" {
		svc = loggerName
	}
	if svc == "" {
		svc = "flexible-logger"
	}

	alertMsg := fmt.Sprintf("⚠️ [%s] AsyncSink buffer saturated: %d log(s) dropped in window (total dropped: %d)", svc, droppedInWindow, totalDropped)
	error_handler.ReportInternalError(svc, "AsyncSink.BufferOverflow", fmt.Errorf("%s", alertMsg), "")

	if s.OnDropAlert != nil {
		s.OnDropAlert(droppedInWindow, totalDropped)
	}

	if s.dropNotifier != nil {
		_ = s.dropNotifier.Notify(&models.NotifMessage{
			Message: alertMsg,
			Level:   models.LevelWarning.String(),
			Tags:    []string{"telegram"},
		})
	}
}

// -----------------------------------------------------------------------------

func (s *AsyncSink) Close() error {
	s.mu.Lock()
	if s.closed.Swap(true) {
		s.mu.Unlock()
		return nil // Already closed
	}
	close(s.buffer)
	timeout := s.drainTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	s.mu.Unlock()

	drained := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(drained)
	}()

	select {
	case <-drained:
		return s.next.Close()
	case <-time.After(timeout):
		// Timeout reached while draining; close underlying sink to abort hanging downstream operations
		closeErr := s.next.Close()
		if closeErr != nil {
			return fmt.Errorf("async sink drain timeout (%v): downstream close failed: %w", timeout, closeErr)
		}
		return fmt.Errorf("async sink drain timeout (%v): shutdown completed with remaining buffer aborted", timeout)
	}
}
