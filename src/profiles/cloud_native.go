package profiles

import (
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
// NewCloudLogger creates a logger optimized for containerized/cloud environments:
// - Console output (Async JSON) -> For Fluentd/Datadog/etc.
// - Local file (Async JSON)
// - Network (Async Capnp) -> For centralized Log Server
// - Notif (Async)
func NewCloudLogger(name string, config *distributed_config.Config, useLocalNotif bool) interfaces.Logger {
	// 1. Console (Async JSON)
	consoleSink := sink.NewConsoleSink()
	jsonConsole := sink.NewWriterSink(os.Stdout, serializers.NewJSONSerializer())
	asyncConsole := sink.NewAsyncSink(jsonConsole, 2048)

	// 2. File (Async JSON)
	logPath := helpers.GetDefaultLogPath()
	var fileSink interfaces.Sink
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		jsonFile := sink.NewWriterSink(f, serializers.NewJSONSerializer())
		fileSink = sink.NewAsyncSink(jsonFile, 4096)
	} else {
		fileSink = consoleSink
	}

	// 3. Network (Async Capnp)
	nm := conn_manager.NewNetworkManager(-1, 200, 5000, 2000, 2.0, 0.1)
	nm.OnError = func(attempt int, err error, source string, msg string) {
		error_handler.ReportInternalError(name, source, err, msg)
	}

	type ServerCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	var lsCap ServerCap
	if err := config.GetCapability("log-server", &lsCap); err == nil && lsCap.IP != "" {
		publicIP := "127.0.0.1"
		conn, err := nm.ConnectWithRetry(&lsCap.IP, &lsCap.Port, &publicIP, "tcp-hello:"+name)
		if err == nil {
			ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
			networkSink := sink.NewAsyncSink(ns, 8192)

			// Combine
			multi := sink.NewMultiSink(asyncConsole, fileSink, networkSink)

			// 4. Engine
			logger := factory.CreateLogEngine(name, models.LevelInfo, multi, true, 1.0).(*engine.LogEngine)

			// 5. Notifier
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
	}

	// Fallback if network config missing
	multi := sink.NewMultiSink(asyncConsole, fileSink)
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi, true, 1.0).(*engine.LogEngine)

	if useLocalNotif {
		localNotif := notifier.NewLocalNotifier()
		logger.Notifier = localNotif
		return &NotifLoggerWrapper{
			Logger:        logger,
			localNotifier: localNotif,
		}
	}

	return logger
}
