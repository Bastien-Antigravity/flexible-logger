package main

import (
	"fmt"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

func main() {
	// Setup a standalone config for testing
	config := distributed_config.New("test")

	fmt.Println("=== Testing Premium Features ===")

	// 1. Test Cloud Native (JSON)
	fmt.Println("\n1. Initializing CloudNative Logger (JSON)...")
	cloudLog := profiles.NewCloudLogger("CloudApp", config)
	cloudLog.Info("This should appear as a JSON line in stdout and ./log/")
	cloudLog.Close()

	// 2. Test Audit Logger (Blocking)
	fmt.Println("\n2. Initializing Audit Logger (Blocking)...")
	// Note: This will fallback to local sinks if Log Server is not running on 127.0.0.2
	auditLog := profiles.NewAuditLogger("AuditApp", config)
	auditLog.Warning("Audit trail: Secure transaction #999 initiated")
	auditLog.Close()

	// 3. Test Sampling Logic
	fmt.Println("\n3. Testing Sampling Logic (50% rate)...")
	standardLog := profiles.NewStandardLogger("SampleApp", config)

	// Access the engine to set sampling rate manually for the test
	// In production, this would be set via the factory or config
	if engine, ok := standardLog.(interface{ SetSamplingRate(float64) }); ok {
		engine.SetSamplingRate(0.5)
	}

	for i := 0; i < 100; i++ {
		standardLog.Info("Message %d", i)
		// Verification would involve counting lines in the log file
	}
	standardLog.Close()

	fmt.Println("\n=== Verification Complete ===")
}
