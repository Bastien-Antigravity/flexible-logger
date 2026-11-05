package profiles

// =============================================================================
// ESSENTIAL PROCESS: Cloud-native logging profile emitting asynchronous JSON to stdout and rotating files for container log collectors.
//
// DATA FLOW:
//   1. Initializes asynchronous JSON sinks for stderr and local file.
//   2. Optionally connects to centralized log-server via Cap'n Proto.
//   3. Combines sinks using MultiSink for high throughput in Kubernetes pods.
//
// KEY PARAMETERS:
//   - name: Microservice name.
//   - config: Distributed configuration provider.
//   - useLocalNotif: Notification queue mode flag.
// =============================================================================

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
	jsonConsole := sink.NewWriterSink(os.Stderr, serializers.NewJSONSerializer())
	asyncConsole := sink.NewAsyncSink(jsonConsole, 2048)
	sinks := []interfaces.Sink{asyncConsole}

	// 2. File (Async JSON)
	logPath := helpers.GetLogPath(name)
	if f, err := sink.NewRollingFileWriter(logPath, 10*1024*1024, 5); err == nil {
		jsonFile := sink.NewWriterSink(f, serializers.NewJSONSerializer())
		fileSink := sink.NewAsyncSink(jsonFile, 4096)
		sinks = append(sinks, fileSink)
	}

	// 3. Network (Async Capnp)
	type ServerCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	publicIP := "127.0.0.1"

	var lsCap ServerCap
	if config != nil && config.GetCapability("log_server", &lsCap) == nil && lsCap.IP != "" {
		nm := conn_manager.NewNetworkManager(-1, 200, 5000, 2000, 2.0, 0.1)
		nm.OnError = func(attempt int, err error, source string, msg string) {
			error_handler.ReportInternalError(name, source, err, msg)
		}
		conn, err := nm.ConnectWithRetry(&lsCap.IP, &lsCap.Port, &publicIP, "tcp-hello:"+name)
		if err == nil {
			ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
			networkSink := sink.NewAsyncSink(ns, 8192)
			sinks = append(sinks, networkSink)
		}
	}

	// Combine
	multi := sink.NewMultiSink(sinks...)

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
	if config != nil && config.GetCapability("notif_server", &nsCap) == nil && nsCap.IP != "" {
		logger.Notifier = notifier.NewRemoteNotifier(&nsCap.IP, &nsCap.Port, &publicIP, name)
	}

	return logger
}
