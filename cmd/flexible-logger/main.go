package main

import (
	"fmt"
	"time"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

// -----------------------------------------------------------------------------
func main() {
	fmt.Println("=== Logger Demo ===")

	// Create Distributed Config
	distConf := distributed_config.New("standalone")

	// 3. High Performance Logger (Async everything)
	fmt.Println("\n--- High Perf Logger ---")
	perfLog := profiles.NewStandardLogger("mini", distConf)

	x := 1_000

	start := time.Now()
	for i := 0; i < x; i++ {
		perfLog.Info("HighPerf log message %d", i)
	}

	// Close flushing async buffer
	perfLog.Close()
	fmt.Printf("Wrote %d logs in %v\n", x, time.Since(start))
	fmt.Println("Check logs in ./log directory")
}
