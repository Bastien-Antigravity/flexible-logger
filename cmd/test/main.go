package main

import (
	"fmt"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
	"github.com/Bastien-Antigravity/flexible-logger/src/test_utils"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------
func main() {
	fmt.Println("=== Logger Benchmark ===")

	// HighPerf Logger (Network Async)
	// Mock Config
	distConf := distributed_config.New("standalone")

	// Start Mock Servers
	logIp, logPort, stopLog := test_utils.StartMockServer("LogServer")
	defer stopLog()
	notifIp, notifPort, stopNotif := test_utils.StartMockServer("NotifServer")
	defer stopNotif()

	// Override Config with Mock Addresses
	if distConf.Capabilities == nil {
		distConf.Capabilities = make(map[string]interface{})
	}
	distConf.Capabilities["log_server"] = map[string]interface{}{
		"ip":   logIp,
		"port": logPort,
	}
	distConf.Capabilities["notif_server"] = map[string]interface{}{
		"ip":   notifIp,
		"port": notifPort,
	}

	prodLog := profiles.NewHighPerfLogger("BenchApp", distConf, false)

	count := 1_000_000
	fmt.Printf("Logging %d messages...\n", count)

	start := time.Now()
	for i := 0; i < count; i++ {
		prodLog.Info("Benchmark message payload")
	}

	// Separate close time (flushing) from log time if desired
	prodLog.Close()

	duration := time.Since(start)
	fmt.Printf("Total time: %v\n", duration)
	fmt.Printf("Throughput: %.2f logs/sec\n", float64(count)/duration.Seconds())
	fmt.Printf("Time per log: %v\n", duration/time.Duration(count))
}
