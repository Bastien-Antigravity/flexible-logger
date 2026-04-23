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
// NewAuditLogger creates a strictly reliable logger for audit trails:
// - Console output (Sync)
// - Local file (Sync)
// - Network (Sync / Blocking) -> Ensures server receipt before returning
// - Notif (Async)
func NewAuditLogger(name string, config *distributed_config.Config, useLocalNotif bool) interfaces.Logger {
	// 1. Console (Sync)
	consoleSink := sink.NewConsoleSink()

	// 2. File (Sync)
	logPath := helpers.GetDefaultLogPath()
	var fileSink interfaces.Sink
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		fileSink = sink.NewWriterSink(f, serializers.NewTextSerializer())
	} else {
		fileSink = consoleSink
	}

	// 3. Network (Sync / Blocking)
	nm := conn_manager.NewCriticalStrategy(nil)
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

		// Use Connect with ModeIndefinite for Audit trail
		conn := nm.Connect(&lsCap.IP, &lsCap.Port, &publicIP, "tcp-hello:"+name, conn_manager.ModeIndefinite)

		// IMPORTANT: Wrap directly in WriterSink WITHOUT an AsyncSink wrapper.
		// This makes the log.Info() call WAIT for the socket write to complete.
		networkSink := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())

		// Combine
		multi := sink.NewMultiSink(consoleSink, fileSink, networkSink)

		// 4. Engine
		// Audit logs NEVER use sampling (1.0 rate)
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

		// Async is fine for notifs even in audit mode, as logs are the primary trail
		var nsCap ServerCap
		if err := config.GetCapability("notif_server", &nsCap); err == nil && nsCap.IP != "" {
			logger.Notifier = notifier.NewRemoteNotifier(&nsCap.IP, &nsCap.Port, &publicIP, name)
		}

		return logger
	}

	// Fallback
	multi := sink.NewMultiSink(consoleSink, fileSink)
	return factory.CreateLogEngine(name, models.LevelInfo, multi, true, 1.0)
}
