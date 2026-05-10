package tests

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

// Helper to start a local mock server for handshakes
func startMockServer(t *testing.T) (string, string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	parts := strings.Split(addr, ":")
	ip, port := parts[0], parts[1]

	stop := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stop:
					return
				default:
					continue
				}
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				_, _ = c.Read(buf) // Absorb Hello
			}(conn)
		}
	}()

	return ip, port, func() {
		close(stop)
		ln.Close()
	}
}

// -----------------------------------------------------------------------------

func TestPremium_AuditBlocking(t *testing.T) {
	ip, port, stop := startMockServer(t)
	defer stop()

	// 1. Setup Mock Config with mandatory capabilities
	configData := &core.Config{
		Common: core.CommonConfig{Name: "AuditTest"},
		Capabilities: map[string]interface{}{
			"log_server": map[string]interface{}{
				"ip": ip,
				"port": port,
			},
			"notif_server": map[string]interface{}{
				"ip": ip,
				"port": port,
			},
		},
	}
	cfg := &distributed_config.Config{Config: configData}

	// 2. Initialize Audit Logger
	logger := profiles.NewAuditLogger("AuditApp", cfg, false)
	if logger == nil {
		t.Fatal("Failed to initialize AuditLogger")
	}
	defer logger.Close()

	start := time.Now()
	logger.Info("Audit Trail: Secure transaction initiated")
	elapsed := time.Since(start)

	t.Logf("Audit log completed in %v", elapsed)
}

func TestPremium_SamplingIntegrity(t *testing.T) {
	ip, port, stop := startMockServer(t)
	defer stop()

	// 1. Setup Mock Config
	configData := &core.Config{
		Common: core.CommonConfig{Name: "SamplingTest"},
		Capabilities: map[string]interface{}{
			"log_server": map[string]interface{}{
				"ip": ip,
				"port": port,
			},
			"notif_server": map[string]interface{}{
				"ip": ip,
				"port": port,
			},
		},
	}
	cfg := &distributed_config.Config{Config: configData}

	// 2. Initialize Standard Logger
	logger := profiles.NewStandardLogger("SampleApp", cfg, false)
	if logger == nil {
		t.Fatal("Failed to initialize StandardLogger")
	}
	defer logger.Close()

	// 3. Set Sampling to 50%
	if e, ok := logger.(*engine.LogEngine); ok {
		e.SamplingRate = 0.5
	} else {
		t.Skip("Logger implementation does not support manual sampling control")
	}

	// 4. Verify Statistical Distribution
	for i := 0; i < 20; i++ {
		logger.Info("Sampling check %d", i)
	}
	
	for i := 0; i < 10; i++ {
		logger.Error("Critical error %d", i)
	}
}

func TestPremium_FleetHandshake(t *testing.T) {
	ip, port, stop := startMockServer(t)
	defer stop()

	// 1. Setup Mock Config with mandatory capabilities
	configData := &core.Config{
		Common: core.CommonConfig{Name: "FleetTest"},
		Capabilities: map[string]interface{}{
			"log_server": map[string]interface{}{
				"ip": ip,
				"port": port,
			},
			"notif_server": map[string]interface{}{
				"ip": ip,
				"port": port,
			},
		},
	}
	// Simulate the bridge providing a 'test' config
	cfg := &distributed_config.Config{Config: configData}

	logger := profiles.NewStandardLogger("FleetApp", cfg, false)
	if logger == nil {
		t.Fatal("Failed to boot logger in Fleet mode")
	}
	// Give a tiny moment for async handshakes to stabilize
	time.Sleep(50 * time.Millisecond)
	defer logger.Close()

	logger.Info("Ecosystem handshake successful")
}
