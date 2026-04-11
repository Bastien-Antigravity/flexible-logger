package profiles

import (
	"testing"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// TestAllProfiles verifies that basic standalone profiles boot up and log correctly.
func TestAllProfiles(t *testing.T) {
	tests := []struct {
		name    string
		factory func() interfaces.Logger
	}{
		{"Minimal", func() interfaces.Logger { return NewMinimalLogger("test-min") }},
		{"Developer", func() interfaces.Logger { return NewDevelLogger("test-dev") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tt.factory()
			if logger == nil {
				t.Fatalf("%s: initialization failed", tt.name)
			}
			defer logger.Close()

			// Check that it doesn't crash on standard logging
			logger.Info("Checking %s profile", tt.name)
			logger.Error("Checking error path for %s", tt.name)
		})
	}
}

// SlowSink for testing Audit blocking
type SlowSink struct {
	Delay time.Duration
}

func (s *SlowSink) Write(e *models.LogEntry) error {
	time.Sleep(s.Delay)
	e.Release()
	return nil
}
func (s *SlowSink) Close() error { return nil }

func TestAuditLogger_Blocking(t *testing.T) {
	// We use the LogEngine directly with a SlowSink to simulate NewAuditLogger behavior
	// since NewAuditLogger requires a network connection.
	sink := &SlowSink{Delay: 100 * time.Millisecond}
	logger := engine.LogEngine{
		Sink:         sink,
		Level:        models.LevelInfo,
		SamplingRate: 1.0,
	}

	start := time.Now()
	logger.Info("Audit message")
	elapsed := time.Since(start)

	if elapsed < 100*time.Millisecond {
		t.Errorf("Audit mode did not block! Expected > 100ms, got %v", elapsed)
	}
}

// Note: Profiles like Standard, Audit, and CloudNative require a valid
// distributed_config.Config and a network connection, so they are typically
// tested using "Integration Tests" with a local mock server.
