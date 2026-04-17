package main

import (
	"fmt"
	"time"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/profiles"
)

func main() {
	fmt.Println("=== Log Server Test Client ===")
	fmt.Println("Connecting to 127.0.0.2:9020...")

	// 1. Setup Config
	distConf := distributed_config.New("test-client")
	if distConf.Capabilities == nil {
		distConf.Capabilities = make(map[string]interface{})
	}
	distConf.Capabilities["log_server"] = map[string]interface{}{
		"ip":   "127.0.0.2",
		"port": "9020",
	}
	// Add a dummy notif_server to satisfy profiles.NewHighPerfLogger
	distConf.Capabilities["notif_server"] = map[string]interface{}{
		"ip":   "127.0.0.2",
		"port": "10080",
	}

	test_func(distConf)

	fmt.Println("Done.")
}

func test_func(distConf *distributed_config.Config) {
	logger := profiles.NewNoLockLogger("TestLogServer", distConf, false)
	defer logger.Close()

	// 3. Send Messages
	count := 10
	fmt.Printf("Sending %d test messages...\n", count)
	for i := 1; i <= count; i++ {
		msg := fmt.Sprintf("Test message #%d from test-log-server client", i)
		logger.Info(msg)
		fmt.Printf("Sent: %s\n", msg)
		logger.Error(msg)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Done.")
}
