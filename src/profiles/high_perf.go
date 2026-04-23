package profiles

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
	if err := config.GetCapability("log_server", &lsCap); err != nil || lsCap.IP == "" {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: Logger configuration missing\n")
		os.Exit(1)
	}
	ipPtr := &lsCap.IP
	portPtr := &lsCap.Port

	// Default public IP (as pointer to handle dynamic update requirement, though static here for now)
	publicIP := "127.0.0.1"

	// Use Connect with ModeNonBlocking
	conn := nm.Connect(ipPtr, portPtr, &publicIP, "tcp-hello:"+name, conn_manager.ModeNonBlocking)
	var networkSink interfaces.Sink
	if conn != nil {
		ns := sink.NewWriterSink(conn, serializers.NewCapnpSerializer())
		networkSink = sink.NewAsyncSink(ns, 16384) // Larger buffer
	} else {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: Failed to initialize connection manager for log server\n")
		os.Exit(1)
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
	if err := config.GetCapability("notif_server", &nsCap); err != nil || nsCap.IP == "" {
		fmt.Fprintf(os.Stderr, "HighPerfLogger: Notification configuration missing\n")
		os.Exit(1)
	}
	notifIpPtr := &nsCap.IP
	notifPortPtr := &nsCap.Port

	logger.Notifier = notifier.NewRemoteNotifier(notifIpPtr, notifPortPtr, &publicIP, name)

	return logger
}
