package profiles

// =============================================================================
// ESSENTIAL PROCESS: Standard fleet logging profile balancing synchronous local file/console persistence with asynchronous remote network streaming.
//
// DATA FLOW:
//   1. Initializes synchronous Console and local File sinks.
//   2. Resolves log_server capability and attaches buffered AsyncSink.
//   3. Configures LogEngine with Info threshold and attaches RemoteNotifier.
//
// KEY PARAMETERS:
//   - name: Microservice name.
//   - config: Distributed configuration provider.
//   - useLocalNotif: Alert routing mode selector.
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
// NewStandardLogger creates a standard logger with:
// - Console output (Sync)
// - Local file (Sync) - Path derived from executable or defaults
// - Network (Async) - Address from Config
// - Notif (Async) - Address from Config
func NewStandardLogger(name string, config *distributed_config.Config, useLocalNotif bool) interfaces.Logger {
	// 1. Console (Sync)
	consoleSink := sink.NewConsoleSink()
	sinks := []interfaces.Sink{consoleSink}

	// 2. File (Sync)
	logPath := helpers.GetLogPath(name)
	if f, err := sink.NewRollingFileWriter(logPath, 10*1024*1024, 5); err == nil {
		sinks = append(sinks, sink.NewWriterSink(f, serializers.NewTextSerializer()))
	} else {
		fmt.Fprintf(os.Stderr, "StandardLogger: Failed to open log file %s: %v, falling back to console only\n", logPath, err)
	}

	// 3. Network (Async)
	var networkSink *sink.AsyncSink
	if lsCap, ok := resolveServerCap(config, "log_server"); ok {
		nm := conn_manager.NewStandardStrategy(nil)
		nm.OnError = func(attempt int, err error, source string, msg string) {
			error_handler.ReportInternalError(name, source, err, msg)
		}
		conn := nm.Connect(&lsCap.IP, &lsCap.Port, publicIPPtr(config), "tcp-hello:"+name, conn_manager.ModeIndefinite)
		if conn != nil {
			ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
			networkSink = sink.NewAsyncSink(ns, 4096)
			sinks = append(sinks, networkSink)
		}
	} else {
		fmt.Fprintf(os.Stderr, "StandardLogger: Logger configuration missing for log_server, running in local-only mode\n")
	}

	// 4. Combine
	multi := sink.NewMultiSink(sinks...)

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi, true, 1.0).(*engine.LogEngine)

	// 6. Notifier + throttled drop alerts
	return attachNotifier(name, config, useLocalNotif, logger, networkSink)
}
