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
// NewNoLockLogger creates a fully async logger with:
// - Console output (Async)
// - Local file (Async)
// - Network (Async)
// - Notif (Async)
func NewNoLockLogger(name string, config *distributed_config.Config) interfaces.Logger {
	// 1. Console (Async)
	consoleSink := sink.NewConsoleSink()
	asyncConsole := sink.NewAsyncSink(consoleSink, 1024)

	// 2. File (Async)
	logPath := helpers.GetDefaultLogPath()
	var fileSink interfaces.Sink
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fileSink = sink.NewConsoleSink()
	} else {
		fileSink = sink.NewWriterSink(f, serializers.NewCapnpSerializer())
	}
	asyncFile := sink.NewAsyncSink(fileSink, 4096)

	// 3. Network (Async)
	nm := network_manager.NewNetworkManager()

	if config.Capabilities.Logger == nil {
		fmt.Fprintf(os.Stderr, "NoLockLogger: Logger configuration missing\n")
		os.Exit(1)
	}
	ipPtr := &config.Capabilities.Logger.IP
	portPtr := &config.Capabilities.Logger.Port

	// Default public IP
	publicIP := "127.0.0.1"

	conn, err := nm.ConnectWithRetry(ipPtr, portPtr, &publicIP, "tcp")
	var networkSink interfaces.Sink
	if err == nil {
		ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
		networkSink = sink.NewAsyncSink(ns, 8192)
	} else {
		fmt.Fprintf(os.Stderr, "NoLockLogger: Failed to connect to log server: %v\n", err)
		os.Exit(1)
	}

	// 4. MultiSink
	// Writes to AsyncConsole, AsyncFile, AsyncNet.
	// MultiSink handles Retain() for Fan-Out.
	multi := sink.NewMultiSink(asyncConsole, asyncFile, networkSink)

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi).(*engine.LogEngine)

	// 6. Notifier (Async)
	if config.Capabilities.Notification == nil {
		fmt.Fprintf(os.Stderr, "NoLockLogger: Notification configuration missing\n")
		os.Exit(1)
	}
	notifIpPtr := &config.Capabilities.Notification.IP
	notifPortPtr := &config.Capabilities.Notification.Port

	logger.Notifier = notifier.NewRemoteNotifier(notifIpPtr, notifPortPtr, &publicIP)

	return logger
}
