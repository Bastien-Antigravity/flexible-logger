package main

import (
	"fmt"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------
func main() {
	fmt.Println("=== Logger Demo ===")

	// Create Distributed Config
	distConf := distributed_config.New("standalone")

	// 3. High Performance Logger (Async everything)
	fmt.Println("\n--- High Perf Logger ---")
	perfLog := profiles.NewHighPerfLogger("PerfApp2", distConf)

	start := time.Now()
	for i := 0; i < 1_000_000; i++ {
		perfLog.Info(fmt.Sprintf("HighPerf log message %d", i))
	}

	// Close flushing async buffer
	perfLog.Close()
	fmt.Printf("Wrote 1_000_000 logs in %v\n", time.Since(start))
	fmt.Println("Check logs in ./log directory")
}
