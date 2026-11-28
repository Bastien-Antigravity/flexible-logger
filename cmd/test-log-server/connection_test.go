package main

// =============================================================================
// ESSENTIAL PROCESS: Integration and benchmark tests for TCP log-server connectivity over SafeSocket.
//
// DATA FLOW:
//   1. Spawns loopback mock server absorbing hello handshakes.
//   2. Configures distributed_config capabilities for log_server.
//   3. Dispatches logs via NoLock and HighPerf logger profiles.
//   4. Measures delivery throughput and connection resilience.
//
// KEY PARAMETERS:
//   - logIp, logPort: Dynamic loopback mock listener coordinates.
//   - distConf: Distributed configuration instance.
// =============================================================================

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
	"github.com/Bastien-Antigravity/flexible-logger/src/test_utils"
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

func probeLoopbackB(b *testing.B) {
	b.Helper()
	probeLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Skipf("Loopback socket listen not permitted in sandbox: %v", err)
	}
	defer probeLn.Close()

	probeConn, probeErr := net.DialTimeout("tcp", probeLn.Addr().String(), 100*time.Millisecond)
	if probeErr != nil {
		b.Skipf("Loopback socket dial not permitted in sandbox environment: %v", probeErr)
	}
	probeConn.Close()
}

func setupIsolatedTestConfig(tb testing.TB) {
	tb.Helper()
	tmpDir := tb.TempDir()
	tb.Setenv("LOG_DIR", filepath.Join(tmpDir, "logs"))
	cfgPath := filepath.Join(tmpDir, "standalone.yaml")
	if err := os.WriteFile(cfgPath, []byte("common:\n  name: test-log-server\n"), 0o644); err != nil {
		tb.Fatalf("Failed to write temp config: %v", err)
	}
	tb.Setenv("CONFIG_PATH", cfgPath)
}

func TestLogServerConnection(t *testing.T) {
	probeLoopback(t)
	setupIsolatedTestConfig(t)

	// 1. Setup Mock Server to avoid connection refusal
	// Rename to notif-mock as per user request
	logIp, logPort, stopLog := test_utils.StartMockServer("notif-mock")
	defer stopLog()

	// 2. Setup Config
	distConf := distributed_config.New("standalone")
	if distConf.Capabilities == nil {
		distConf.Capabilities = make(map[string]interface{})
	}
	distConf.Capabilities["log_server"] = map[string]interface{}{
		"ip":   logIp,
		"port": logPort,
	}
	distConf.Capabilities["notif_server"] = map[string]interface{}{
		"ip":   logIp,
		"port": logPort,
	}

	// Initialize Logger - This will attempt to connect
	// Note: If the server is not running, this might fail depending on retry policy
	logger := profiles.NewNoLockLogger("TestLogServer", distConf, false)
	defer logger.Close()

	// Send a few messages
	for i := 0; i < 5; i++ {
		logger.Info("Test message from connection_test.go")
	}

	// Small sleep to allow async logs to be processed (though logger.Close does this too)
	time.Sleep(1000 * time.Millisecond)
}

func BenchmarkLogServerThroughput(b *testing.B) {
	probeLoopbackB(b)
	setupIsolatedTestConfig(b)

	// Setup Mock
	logIp, logPort, stopLog := test_utils.StartMockServer("notif-mock")
	defer stopLog()

	// Setup Config
	distConf := distributed_config.New("standalone")
	if distConf.Capabilities == nil {
		distConf.Capabilities = make(map[string]interface{})
	}
	distConf.Capabilities["log_server"] = map[string]interface{}{
		"ip":   logIp,
		"port": logPort,
	}
	distConf.Capabilities["notif_server"] = map[string]interface{}{
		"ip":   logIp,
		"port": logPort,
	}

	logger := profiles.NewHighPerfLogger("BenchmarkClient", distConf, false)
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message")
	}
}
