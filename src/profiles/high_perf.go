package profiles

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/network_manager"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------
// NewHighPerfLogger creates a high performance logger with:
// - Network (Async)
// - Notif (Async)
func NewHighPerfLogger(name string, config *distributed_config.Config) interfaces.Logger {
	// 1. Network (Async)
	nm := network_manager.NewNetworkManager()

	if config.Capabilities.Logger == nil {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: Logger configuration missing\n")
		os.Exit(1)
	}
	ipPtr := &config.Capabilities.Logger.IP
	portPtr := &config.Capabilities.Logger.Port

	// Default public IP (as pointer to handle dynamic update requirement, though static here for now)
	publicIP := "127.0.0.1"

	conn, err := nm.ConnectWithRetry(ipPtr, portPtr, &publicIP, "tcp")
	var networkSink interfaces.Sink
	if err == nil {
		ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
		networkSink = sink.NewAsyncSink(ns, 16384) // Larger buffer
	} else {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: Failed to connect to log server: %v\n", err)
		os.Exit(1)
	}

	// 2. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, networkSink).(*engine.LogEngine)

	// 3. Notifier (Async)
	if config.Capabilities.Notification == nil {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: Notification configuration missing\n")
		os.Exit(1)
	}
	notifIpPtr := &config.Capabilities.Notification.IP
	notifPortPtr := &config.Capabilities.Notification.Port

	logger.Notifier = notifier.NewRemoteNotifier(notifIpPtr, notifPortPtr, &publicIP)

	return logger
}
