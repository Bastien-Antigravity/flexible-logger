package profiles

// =============================================================================
// ESSENTIAL PROCESS: Audit logging profile enforcing strict synchronous file and network delivery without sampling.
//
// DATA FLOW:
//   1. Initializes synchronous Console and local File sinks.
//   2. Connects to log-server in blocking mode (ModeIndefinite).
//   3. Wraps connection in synchronous WriterSink with Cap'n Proto serializer.
//   4. Combines into MultiSink and returns fully synchronous audit logger.
//
// KEY PARAMETERS:
//   - name: Service or component name.
//   - config: Distributed configuration carrying log_server coordinates.
//   - useLocalNotif: Boolean controlling local channel versus remote notification.
// =============================================================================

import (
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
// NewAuditLogger creates a strictly reliable logger for audit trails:
// - Console output (Sync)
// - Local file (Sync)
// - Network (Sync / Blocking) -> Ensures server receipt before returning
// - Notif (Async)
func NewAuditLogger(name string, config *distributed_config.Config, useLocalNotif bool) interfaces.Logger {
	// 1. Console (Sync)
	consoleSink := sink.NewConsoleSink()
	sinks := []interfaces.Sink{consoleSink}

	// 2. File (Sync)
	logPath := helpers.GetLogPath(name)
	if f, err := sink.NewRollingFileWriter(logPath, 10*1024*1024, 5); err == nil {
		fileSink := sink.NewWriterSink(f, serializers.NewTextSerializer())
		sinks = append(sinks, fileSink)
	}

	// 3. Network (Sync / Blocking)
	type ServerCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	publicIP := "127.0.0.1"

	var lsCap ServerCap
	if config != nil && config.GetCapability("log_server", &lsCap) == nil && lsCap.IP != "" {
		nm := conn_manager.NewCriticalStrategy(nil)
		nm.OnError = func(attempt int, err error, source string, msg string) {
			error_handler.ReportInternalError(name, source, err, msg)
		}

		// Use Connect with ModeIndefinite for Audit trail
		conn := nm.Connect(&lsCap.IP, &lsCap.Port, &publicIP, "tcp-hello:"+name, conn_manager.ModeIndefinite)

		// IMPORTANT: Wrap directly in WriterSink WITHOUT an AsyncSink wrapper.
		// This makes the log.Info() call WAIT for the socket write to complete.
		networkSink := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
		sinks = append(sinks, networkSink)
	}

	// Combine
	multi := sink.NewMultiSink(sinks...)

	// 4. Engine
	// Audit logs NEVER use sampling (1.0 rate)
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi, true, 1.0).(*engine.LogEngine)

	// 5. Notifier (Initialized independently of log_server)
	if useLocalNotif {
		localNotif := notifier.NewLocalNotifier()
		logger.Notifier = localNotif
		return &NotifLoggerWrapper{
			Logger:        logger,
			localNotifier: localNotif,
		}
	}

	// Async is fine for notifs even in audit mode, as logs are the primary trail
	var nsCap ServerCap
	if config != nil && config.GetCapability("notif_server", &nsCap) == nil && nsCap.IP != "" {
		logger.Notifier = notifier.NewRemoteNotifier(&nsCap.IP, &nsCap.Port, &publicIP, name)
	}

	return logger
}
