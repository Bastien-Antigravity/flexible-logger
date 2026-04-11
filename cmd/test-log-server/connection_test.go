package main

import (
	"testing"
	"time"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

func TestLogServerConnection(t *testing.T) {
	// Setup Config
	distConf := distributed_config.New("test-cli")
	if distConf.Capabilities == nil {
		distConf.Capabilities = make(map[string]interface{})
	}
	distConf.Capabilities["log_server"] = map[string]interface{}{
		"ip":   "127.0.0.2",
		"port": "9020",
	}
	distConf.Capabilities["notif_server"] = map[string]interface{}{
		"ip":   "127.0.0.2",
		"port": "10080",
	}

	// Initialize Logger - This will attempt to connect
	// Note: If the server is not running, this might fail depending on retry policy
	logger := profiles.NewNoLockLogger("TestLogServer", distConf)
	defer logger.Close()

	// Send a few messages
	for i := 0; i < 5; i++ {
		logger.Info("Test message from connection_test.go")
	}

	// Small sleep to allow async logs to be processed (though logger.Close does this too)
	time.Sleep(1000 * time.Millisecond)
}

func BenchmarkLogServerThroughput(b *testing.B) {
	// Setup Config
	distConf := distributed_config.New("bench-cli")
	if distConf.Capabilities == nil {
		distConf.Capabilities = make(map[string]interface{})
	}
	distConf.Capabilities["log_server"] = map[string]interface{}{
		"ip":   "127.0.0.2",
		"port": "9020",
	}
	distConf.Capabilities["notif_server"] = map[string]interface{}{
		"ip":   "127.0.0.2",
		"port": "10080",
	}

	logger := profiles.NewHighPerfLogger("BenchmarkClient", distConf)
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message")
	}
}
