package sink

// =============================================================================
// ESSENTIAL PROCESS: Fan-out sink broadcasting a single LogEntry across multiple downstream destination sinks using atomic reference counting.
//
// DATA FLOW:
//   1. Receives LogEntry with initial refCount=1.
//   2. Retains entry for each additional sink in the group.
//   3. Dispatches entry to all registered child sinks.
//   4. Aggregates any write errors for caller reporting.
//
// KEY PARAMETERS:
//   - sinks: Slice of downstream destination sinks.
//   - MultiSink: Fan-out sink manager.
// =============================================================================

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------

// MultiSink broadcasts a LogEntry to multiple underlying Sinks.
// It implements the Fan-Out pattern.
type MultiSink struct {
	sinks     []interfaces.Sink
	closeOnce sync.Once
}

// -----------------------------------------------------------------------------

func NewMultiSink(sinks ...interfaces.Sink) *MultiSink {
	return &MultiSink{
		sinks: sinks,
	}
}

// -----------------------------------------------------------------------------

func (ms *MultiSink) Write(entry *models.LogEntry) error {
	var errs []string

	// Handle zero sinks case to prevent memory leak
	if len(ms.sinks) == 0 {
		entry.Release()
		return nil
	}

	// Fan-out: We need to send the entry to N sinks.
	// The entry currently has RefCount=1 (from caller).
	// We need to consume it N times.
	// So we need TotalRefs = N.
	// We currently have 1. We need to add N-1.
	for i := 0; i < len(ms.sinks)-1; i++ {
		entry.Retain()
	}

	for _, s := range ms.sinks {
		if err := s.Write(entry); err != nil {
			errs = append(errs, err.Error())
			// NOTE: Protocol says Write() MUST consume the ref even on error.
			// Currently Console/Writer sink defer Release(), so they do.
			// AsyncSink (if full) calls Release(), so it does.
			// AsyncSink (buffer) -> worker -> next -> Release.
			// Safe.
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multisink write errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// -----------------------------------------------------------------------------

func (ms *MultiSink) Close() error {
	var err error
	ms.closeOnce.Do(func() {
		var errs []string
		for _, s := range ms.sinks {
			if sErr := s.Close(); sErr != nil {
				errs = append(errs, sErr.Error())
			}
		}
		if len(errs) > 0 {
			err = fmt.Errorf("multisink close errors: %s", strings.Join(errs, "; "))
		}
	})
	return err
}
