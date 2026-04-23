package profiles

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/error_handler"
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/helpers"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
	"github.com/Bastien-Antigravity/microservice-toolbox/go/pkg/conn_manager"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------
// NewStandardLogger creates a standard logger with:
// - Console output (Sync)
// - Local file (Sync) - Path derived from executable or defaults
// - Network (Async) - Address from Config
// - Notif (Async) - Address from Config
func NewStandardLogger(name string, config *distributed_config.Config, useLocalNotif bool) interfaces.Logger {
	// 1. Console (Sync)
	consoleSink := sink.NewConsoleSink()

	// 2. File (Sync)
	logPath := helpers.GetDefaultLogPath()
	var fileSink interfaces.Sink
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "StandardLogger: Failed to open log file %s: %v\n", logPath, err)
		os.Exit(1)
	} else {
		fileSink = sink.NewWriterSink(f, serializers.NewTextSerializer())
	}

	// 3. Network (Async)
	// We block now until connection is established, per requirements.
	nm := conn_manager.NewCriticalStrategy(nil)
	nm.OnError = func(attempt int, err error, source string, msg string) {
		error_handler.ReportInternalError(name, source, err, msg)
	}

	type ServerCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	var lsCap ServerCap
	if err := config.GetCapability("log_server", &lsCap); err != nil || lsCap.IP == "" {
		fmt.Fprintf(os.Stderr, "StandardLogger: Logger configuration missing\n")
		os.Exit(1)
	}

	// Default public IP
	publicIP := "127.0.0.1"

	// Block until connected.
	conn := nm.Connect(&lsCap.IP, &lsCap.Port, &publicIP, "tcp-hello:"+name, conn_manager.ModeIndefinite)

	// Create WriterSink with CapnpSerializer
	// sink.NewWriterSink(io.WriteCloser, Serializer)
	ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
	// Use AsyncSink for network as per requirement (Async)
	networkSink := sink.NewAsyncSink(ns, 4096)

	// 4. Combine
	multi := sink.NewMultiSink(consoleSink, fileSink, networkSink)

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi, true, 1.0).(*engine.LogEngine)

	// 6. Notifier
	if useLocalNotif {
		localNotif := notifier.NewLocalNotifier()
		logger.Notifier = localNotif
		return &NotifLoggerWrapper{
			Logger:        logger,
			localNotifier: localNotif,
		}
	}

	// RemoteNotifier handles its own connection/retry logic.
	var nsCap ServerCap
	if err := config.GetCapability("notif_server", &nsCap); err != nil || nsCap.IP == "" {
		fmt.Fprintf(os.Stderr, "StandardLogger: Notification configuration missing\n")
		os.Exit(1)
	}
	notifIpPtr := &nsCap.IP
	notifPortPtr := &nsCap.Port

	logger.Notifier = notifier.NewRemoteNotifier(notifIpPtr, notifPortPtr, &publicIP, name)

	return logger
}
