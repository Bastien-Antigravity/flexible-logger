package profiles

// =============================================================================
// ESSENTIAL PROCESS: Integration tests validating all 8 logger profiles, mock TCP handshakes, and local notification delivery.
//
// DATA FLOW:
//   1. Spawns loopback mock servers.
//   2. Instantiates Minimal, Devel, Audit, Standard, and Notif profiles.
//   3. Asserts non-blocking throughput, blocking audit semantics, and alert delivery.
//
// KEY PARAMETERS:
//   - t: Testing harness.
// =============================================================================

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// probeLoopback verifies if loopback socket listening and dialing are permitted
// in the test execution environment without polluting any test listener's backlog.
func probeLoopback(t *testing.T) {
	t.Helper()
	probeLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Loopback socket listen not permitted in sandbox: %v", err)
	}
	defer probeLn.Close()

	probeConn, probeErr := net.DialTimeout("tcp", probeLn.Addr().String(), 100*time.Millisecond)
	if probeErr != nil {
		t.Skipf("Loopback socket dial not permitted in sandbox environment: %v", probeErr)
	}
	probeConn.Close()
}

// TestAllProfiles verifies that all 8 logger profiles boot cleanly (including nil config) and log without panicking.
func TestAllProfiles(t *testing.T) {
	t.Setenv("LOG_DIR", t.TempDir())

	tests := []struct {
		name    string
		factory func() interfaces.Logger
	}{
		{"Minimal", func() interfaces.Logger { return NewMinimalLogger("test-min", false) }},
		{"Developer", func() interfaces.Logger { return NewDevelLogger("test-dev", false) }},
		{"Standard", func() interfaces.Logger { return NewStandardLogger("test-std", nil, false) }},
		{"Audit", func() interfaces.Logger { return NewAuditLogger("test-audit", nil, false) }},
		{"CloudNative", func() interfaces.Logger { return NewCloudLogger("test-cloud", nil, false) }},
		{"HighPerf", func() interfaces.Logger { return NewHighPerfLogger("test-hp", nil, false) }},
		{"NoLock", func() interfaces.Logger { return NewNoLockLogger("test-nolock", nil, false) }},
		{"NotifLogger", func() interfaces.Logger { return NewNotifLogger("test-notif-nil", nil, true) }},
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

func TestAuditLogger_MissingConfig_WarnsAndDoesNotPanic(t *testing.T) {
	t.Setenv("LOG_DIR", t.TempDir())
	// Calling NewAuditLogger with nil config should warn on stderr, attach console and file fallback, and not panic
	logger := NewAuditLogger("audit-missing-test", nil, false)
	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}
	defer logger.Close()

	logger.Info("Audit log message under missing remote config")
}

func TestAuditLogger_EmergencyFileFallback(t *testing.T) {
	// Point LOG_DIR at a regular file so opening the primary audit log inside it fails,
	// forcing NewAuditLogger to fall back to filepath.Join(os.TempDir(), "emergency_audit.log").
	blockingFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}
	t.Setenv("LOG_DIR", blockingFile)

	emergencyPath := filepath.Join(os.TempDir(), "emergency_audit.log")
	_ = os.Remove(emergencyPath)
	defer func() { _ = os.Remove(emergencyPath) }()

	logger := NewAuditLogger("emergency-fallback-test", nil, false)
	if logger == nil {
		t.Fatal("Expected non-nil logger when falling back to emergency audit file")
	}

	const sentinel = "Emergency fallback log entry"
	logger.Info(sentinel)
	logger.Close()

	content, err := os.ReadFile(emergencyPath)
	if err != nil {
		t.Fatalf("Expected emergency audit file at %s, got error: %v", emergencyPath, err)
	}
	if !strings.Contains(string(content), sentinel) {
		t.Fatalf("Expected emergency audit file to contain %q, got %q", sentinel, string(content))
	}
}

// Note: Profiles like Standard, Audit, and CloudNative require a valid
// distributed_config.Config and a network connection, so they are typically
// tested using "Integration Tests" with a local mock server.

func startTestServer(t *testing.T) (string, string, func()) {
	probeLoopback(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Loopback socket listen not permitted in sandbox: %v", err)
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
				// Absorb all incoming data until connection closes
				_, _ = io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	return ip, port, func() {
		close(stopChan)
		ln.Close()
	}
}

func TestNotifLogger_LocalQueue(t *testing.T) {
	t.Setenv("LOG_DIR", t.TempDir())
	ip, port, stop := startTestServer(t)
	defer stop()

	// 1. Setup Mock Config with Capability
	// Using the core struct directly to avoid complex strategy loading in tests
	configData := &core.Config{
		Common: core.CommonConfig{PublicIP: ip},
		Capabilities: map[string]interface{}{
			"log_server": map[string]interface{}{
				"ip":   ip,
				"port": port,
			},
		},
	}
	cfg := &distributed_config.Config{Config: configData}

	// 2. Instantiate NotifLogger (Always use true for this profile)
	logger := NewNotifLogger("test-notif", cfg, true)
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

func TestProfile_AppNamePropagation(t *testing.T) {
	t.Setenv("LOG_DIR", t.TempDir())
	probeLoopback(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Loopback socket listen not permitted in sandbox: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	parts := strings.Split(addr, ":")
	ip, port := parts[0], parts[1]

	expectedName := "test-app-123"
	detectedNameChan := make(chan string, 1)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 2048)
				n, _ := c.Read(buf)
				if strings.Contains(string(buf[:n]), expectedName) {
					select {
					case detectedNameChan <- expectedName:
					default:
					}
				}
				_, _ = io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	configData := &core.Config{
		Common: core.CommonConfig{PublicIP: ip},
		Capabilities: map[string]interface{}{
			"log_server": map[string]interface{}{
				"ip":   ip,
				"port": port,
			},
			"notif_server": map[string]interface{}{
				"ip":   ip,
				"port": port,
			},
		},
	}
	cfg := &distributed_config.Config{Config: configData}

	// This should trigger ConnectBlocking which sends the Hello handshake
	logger := NewStandardLogger(expectedName, cfg, false)
	if logger == nil {
		t.Fatal("Failed to create StandardLogger")
	}
	defer logger.Close()

	select {
	case name := <-detectedNameChan:
		if name != expectedName {
			t.Errorf("Expected name %q to be sent, but it wasn't detected in handshake", expectedName)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timed out waiting for handshake with app name")
	}
}
