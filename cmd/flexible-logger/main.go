package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"

	utilconf "github.com/Bastien-Antigravity/microservice-toolbox/go/pkg/config"
	"github.com/Bastien-Antigravity/microservice-toolbox/go/pkg/lifecycle"
)

// -----------------------------------------------------------------------------
func main() {
	fmt.Println("=== Logger Migration Demo ===")

	// 1. Initialize Toolbox Config
	appConfig, err := utilconf.LoadConfig("standalone", nil)
	if err != nil {
		fmt.Printf("Critical Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 2. High Performance Logger (Async everything)
	fmt.Println("\n--- High Perf Logger (Network Aware) ---")
	perfLog := profiles.NewHighPerfLogger("PerfApp2", appConfig.Config)

	start := time.Now()
	for i := 0; i < 1_000_000; i++ {
		perfLog.Info("HighPerf log message %d", i)
	}

	// 3. Graceful Shutdown via Toolbox
	lm := lifecycle.NewManager()
	lm.Register("LoggerFlusher", func() error {
		fmt.Println("Flushing logger buffer...")
		perfLog.Close()
		fmt.Printf("Wrote 1_000_000 logs in %v\n", time.Since(start))
		return nil
	})

	// For a demo, we trigger shutdown manually or just wait
	go func() {
		time.Sleep(1 * time.Second)
		lm.Wait(context.Background())
	}()

	lm.Wait(context.Background())
	fmt.Println("Check logs in ./log directory")
}
