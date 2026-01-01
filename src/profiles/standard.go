package profiles

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/helpers"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/network_manager"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------
// NewStandardLogger creates a standard logger with:
// - Console output (Sync)
// - Local file (Sync) - Path derived from executable or defaults
// - Network (Async) - Address from Config
// - Notif (Async) - Address from Config
func NewStandardLogger(name string, config *distributed_config.Config) interfaces.Logger {
	// 1. Console (Sync)
	consoleSink := sink.NewConsoleSink()

	// 2. File (Sync)
	logPath := helpers.GetDefaultLogPath()
	var fileSink interfaces.Sink
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "StandardLogger: Failed to open log file %s: %v\n", logPath, err)
		os.Exit(1)
	} else {
		fileSink = sink.NewWriterSink(f, serializers.NewCapnpSerializer())
	}

	// 3. Network (Async)
	// We block now until connection is established, per requirements.
	nm := network_manager.NewNetworkManager()

	if config.Capabilities.Logger == nil {
		fmt.Fprintf(os.Stderr, "StandardLogger: Logger configuration missing\n")
		os.Exit(1)
	}
	ipPtr := &config.Capabilities.Logger.IP
	portPtr := &config.Capabilities.Logger.Port

	// Default public IP
	publicIP := "127.0.0.1"

	// Block until connected. ConnectBlocking returns io.WriteCloser.
	conn := nm.ConnectBlocking(ipPtr, portPtr, &publicIP, "tcp")

	// Create WriterSink with CapnpSerializer
	// sink.NewWriterSink(io.WriteCloser, Serializer)
	ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
	// Use AsyncSink for network as per requirement (Async)
	networkSink := sink.NewAsyncSink(ns, 4096)

	// 4. Combine
	multi := sink.NewMultiSink(consoleSink, fileSink, networkSink)

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi).(*engine.LogEngine)

	// 6. Notifier (Async)
	// RemoteNotifier handles its own connection/retry logic.
	if config.Capabilities.Notification == nil {
		fmt.Fprintf(os.Stderr, "StandardLogger: Notification configuration missing\n")
		os.Exit(1)
	}
	notifIpPtr := &config.Capabilities.Notification.IP
	notifPortPtr := &config.Capabilities.Notification.Port

	logger.Notifier = notifier.NewRemoteNotifier(notifIpPtr, notifPortPtr, &publicIP)

	return logger
}
