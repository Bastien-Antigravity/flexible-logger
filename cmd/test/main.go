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
	if distConf.Capabilities.LogServer == nil {
		panic("Config error: LogServer capability missing. Check your YAML tags (should be log_server).")
	}
	distConf.Capabilities.LogServer.IP = logIp
	distConf.Capabilities.LogServer.Port = logPort

	if distConf.Capabilities.NotifServer == nil {
		panic("Config error: NotifServer capability missing. Check your YAML tags (should be notif_server).")
	}
	distConf.Capabilities.NotifServer.IP = notifIp
	distConf.Capabilities.NotifServer.Port = notifPort

	prodLog := profiles.NewHighPerfLogger("BenchApp", distConf)

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
