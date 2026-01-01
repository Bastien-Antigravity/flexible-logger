package main

import (
	"fmt"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------
func main() {
	fmt.Println("=== Logger Benchmark ===")

	// HighPerf Logger (Network Async)
	// Mock Config
	distConf := distributed_config.New("standalone")
	// config object no longer needed
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
