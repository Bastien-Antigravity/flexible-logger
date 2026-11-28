package profiles

// =============================================================================
// ESSENTIAL PROCESS: High-performance logging profile designed for ultra-low latency via non-blocking asynchronous Cap'n Proto network streaming.
//
// DATA FLOW:
//   1. Connects to central log-server using non-blocking connection mode.
//   2. Wraps socket in 16,384-capacity AsyncSink with CapnpSerializer.
//   3. Bypasses disk I/O entirely to deliver maximum throughput.
//
// KEY PARAMETERS:
//   - name: Microservice name.
//   - config: Distributed configuration provider.
//   - useLocalNotif: Alert dispatching selector.
// =============================================================================

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/error_handler"
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
	"github.com/Bastien-Antigravity/microservice-toolbox/go/pkg/conn_manager"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------
// NewHighPerfLogger creates a high performance logger with:
// - Network (Async)
// - Notif (Async)
func NewHighPerfLogger(name string, config *distributed_config.Config, useLocalNotif bool) interfaces.Logger {
	// 1. Network (Async)
	nm := conn_manager.NewPerformanceStrategy(nil)
	nm.OnError = func(attempt int, err error, source string, msg string) {
		error_handler.ReportInternalError(name, source, err, msg)
	}

	var asyncNetSink *sink.AsyncSink
	var networkSink interfaces.Sink

	if lsCap, ok := resolveServerCap(config, "log_server"); ok {
		conn := nm.Connect(&lsCap.IP, &lsCap.Port, publicIPPtr(config), "tcp-hello:"+name, conn_manager.ModeNonBlocking)
		if conn != nil {
			ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
			asyncNetSink = sink.NewAsyncSink(ns, 16384) // Larger buffer
			networkSink = asyncNetSink
		}
	}

	if networkSink == nil {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: log_server connection failed or config missing, falling back to console sink\n")
		networkSink = sink.NewConsoleSink()
	}

	// 2. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, networkSink, false, 1.0).(*engine.LogEngine)

	// 3. Notifier + throttled drop alerts
	return attachNotifier(name, config, useLocalNotif, logger, asyncNetSink)
}
