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
	sinks := []interfaces.Sink{asyncConsole}

	// 2. File (Async)
	logPath := helpers.GetLogPath(name)
	var asyncFile *sink.AsyncSink
	if f, err := sink.NewRollingFileWriter(logPath, 10*1024*1024, 5); err == nil {
		fileSink := sink.NewWriterSink(f, serializers.NewTextSerializer())
		asyncFile = sink.NewAsyncSink(fileSink, 4096)
		sinks = append(sinks, asyncFile)
	} else {
		fmt.Fprintf(os.Stderr, "NoLockLogger: Failed to open log file %s: %v, continuing without file sink\n", logPath, err)
	}

	// 3. Network (Async)
	var networkSink *sink.AsyncSink
	if lsCap, ok := resolveServerCap(config, "log_server"); ok {
		nm := conn_manager.NewPerformanceStrategy(nil)
		nm.OnError = func(attempt int, err error, source string, msg string) {
			error_handler.ReportInternalError(name, source, err, msg)
		}
		conn, err := nm.ConnectWithRetry(&lsCap.IP, &lsCap.Port, publicIPPtr(config), "tcp-hello:"+name)
		if err == nil {
			ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
			networkSink = sink.NewAsyncSink(ns, 8192)
			sinks = append(sinks, networkSink)
		} else {
			fmt.Fprintf(os.Stderr, "NoLockLogger: Failed to connect to log server: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "NoLockLogger: Logger configuration missing for log_server, running local-only\n")
	}

	// 4. MultiSink
	multi := sink.NewMultiSink(sinks...)

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi, false, 1.0).(*engine.LogEngine)

	// 6. Notifier + throttled drop alerts
	return attachNotifier(name, config, useLocalNotif, logger, asyncConsole, asyncFile, networkSink)
}
