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
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
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

		conn := nm.Connect(ipPtr, portPtr, &publicIP, "tcp-hello:"+name, conn_manager.ModeNonBlocking)
		if conn != nil {
			ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
			networkSink = sink.NewAsyncSink(ns, 16384) // Larger buffer
		}
	}

	if networkSink == nil {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: log_server connection failed or config missing, falling back to console sink\n")
		networkSink = sink.NewConsoleSink()
	}

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, networkSink, false, 1.0).(*engine.LogEngine)

	// 3. Notifier
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
