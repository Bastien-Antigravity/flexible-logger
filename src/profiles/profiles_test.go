package profiles

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/distributed-config/src/core"
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

func startTestServer(t *testing.T) (string, string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	
	addr := ln.Addr().String()
	parts := strings.Split(addr, ":")
	ip := parts[0]
	port := parts[1]
	
	stopChan := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stopChan:
					return
				default:
					continue
				}
			}
			go func(c net.Conn) {
				defer c.Close()
				// Read & Discard (Hello handshake simulation)
				buf := make([]byte, 1024)
				c.Read(buf)
			}(conn)
		}
	}()
	
	return ip, port, func() {
		close(stopChan)
		ln.Close()
	}
}

func TestNotifLogger_LocalQueue(t *testing.T) {
	ip, port, stop := startTestServer(t)
	defer stop()

	// 1. Setup Mock Config with Capability
	// Using the core struct directly to avoid complex strategy loading in tests
	configData := &core.Config{
		Capabilities: map[string]interface{}{
			"log_server": map[string]interface{}{
				"ip": ip,
				"port": port,
			},
		},
	}
	cfg := &distributed_config.Config{Config: configData}

	// 2. Instantiate NotifLogger
	logger := NewNotifLogger("test-notif", cfg)
	if logger == nil {
		t.Fatal("Failed to create NotifLogger")
	}
	defer logger.Close()

	// 3. Setup Local Queue
	notifChan := make(chan *models.NotifMessage, 10)
	logger.SetLocalNotifQueue(notifChan)

	// 4. Trigger Notification (Error level)
	expectedMsg := "Critical failure detected"
	logger.Error(expectedMsg)

	// 5. Verify Receipt with Timeout
	select {
	case msg := <-notifChan:
		if msg.Message != expectedMsg {
			t.Errorf("Expected message %q, got %q", expectedMsg, msg.Message)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for local notification")
	}

	// 6. Verify non-triggering Level (Info)
	logger.Info("Normal operation")
	select {
	case msg := <-notifChan:
		t.Errorf("Received unexpected notification for INFO level: %v", msg.Message)
	case <-time.After(500 * time.Millisecond):
		// Success: nothing received
	}
}

