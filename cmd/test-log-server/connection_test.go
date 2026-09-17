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
	"testing"
	"time"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
	"github.com/Bastien-Antigravity/flexible-logger/src/test_utils"
)

func TestLogServerConnection(t *testing.T) {
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
