package profiles

// =============================================================================
// ESSENTIAL PROCESS: Fully asynchronous lock-free profile fanning out logs across async console, async file, and async network sinks.
//
// DATA FLOW:
//   1. Wraps Console, File, and Network sinks in dedicated AsyncSink buffers.
//   2. Combines into MultiSink utilizing atomic reference counting for fan-out.
//   3. Guarantees that caller goroutines never block on underlying I/O.
//
// KEY PARAMETERS:
//   - name: Microservice name.
//   - config: Distributed configuration provider.
//   - useLocalNotif: Notification routing flag.
// =============================================================================

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
// NewNoLockLogger creates a fully async logger with:
// - Console output (Async)
// - Local file (Async)
// - Network (Async)
// - Notif (Async)
func NewNoLockLogger(name string, config *distributed_config.Config, useLocalNotif bool) interfaces.Logger {
	// 1. Console (Async)
	consoleSink := sink.NewConsoleSink()
	asyncConsole := sink.NewAsyncSink(consoleSink, 1024)

	// 2. File (Async)
	logPath := helpers.GetLogPath(name)
	var fileSink interfaces.Sink
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fileSink = sink.NewConsoleSink()
	} else {
		fileSink = sink.NewWriterSink(f, serializers.NewTextSerializer())
	}
	asyncFile := sink.NewAsyncSink(fileSink, 4096)

	// 3. Network (Async)
	nm := conn_manager.NewPerformanceStrategy(nil)
	nm.OnError = func(attempt int, err error, source string, msg string) {
		error_handler.ReportInternalError(name, source, err, msg)
	}

	type ServerCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	var lsCap ServerCap
	var networkSink interfaces.Sink
	publicIP := "127.0.0.1"

	if err := config.GetCapability("log_server", &lsCap); err == nil && lsCap.IP != "" {
		ipPtr := &lsCap.IP
		portPtr := &lsCap.Port

		conn, err := nm.ConnectWithRetry(ipPtr, portPtr, &publicIP, "tcp-hello:"+name)
		if err == nil {
			ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
			networkSink = sink.NewAsyncSink(ns, 8192)
		} else {
			fmt.Fprintf(os.Stderr, "NoLockLogger: Failed to connect to log server: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "NoLockLogger: Logger configuration missing for log_server, running local-only\n")
	}

	// 4. MultiSink
	var sinks []interfaces.Sink
	sinks = append(sinks, asyncConsole, asyncFile)
	if networkSink != nil {
		sinks = append(sinks, networkSink)
	}
	multi := sink.NewMultiSink(sinks...)

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi, false, 1.0).(*engine.LogEngine)

	// 6. Notifier
	if useLocalNotif {
		localNotif := notifier.NewLocalNotifier()
		logger.Notifier = localNotif
		return &NotifLoggerWrapper{
			Logger:        logger,
			localNotifier: localNotif,
		}
	}

	var nsCap ServerCap
	if err := config.GetCapability("notif_server", &nsCap); err == nil && nsCap.IP != "" {
		logger.Notifier = notifier.NewRemoteNotifier(&nsCap.IP, &nsCap.Port, &publicIP, name)
	}

	return logger
}
