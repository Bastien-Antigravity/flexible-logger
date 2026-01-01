package sink

import (
	"fmt"
	"sync"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// ConsoleSink writes to stdout in a human-readable format.
type ConsoleSink struct {
	mu sync.Mutex
}

// -----------------------------------------------------------------------------
func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{}
}

// -----------------------------------------------------------------------------
func (s *ConsoleSink) Write(entry *models.LogEntry) error {
	defer entry.Release() // Release ownership
	s.mu.Lock()
	defer s.mu.Unlock()
	// Simple text format
	fmt.Printf("[%s] [%d] %s: %s\n", entry.Timestamp.UTC().Format(time.RFC3339), entry.Level, entry.LoggerName, entry.Message)
	return nil
}

// -----------------------------------------------------------------------------
func (s *ConsoleSink) Close() error {
	return nil
}
