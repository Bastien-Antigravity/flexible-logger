package profiles

// =============================================================================
// ESSENTIAL PROCESS: Specialized profile exposing programmable local event hooks for microservices that react in-process to errors.
//
// DATA FLOW:
//   1. Instantiates asynchronous multi-sink logger engine.
//   2. Attaches LocalNotifier with dedicated memory queue.
//   3. Wraps engine in NotifLoggerWrapper exposing SetLocalNotifQueue.
//
// KEY PARAMETERS:
//   - name: Microservice name.
//   - config: Distributed configuration provider.
//   - useLocalNotif: Selector for notification handling.
// =============================================================================

import (
	"fmt"
	"os"

	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/helpers"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
	"github.com/Bastien-Antigravity/microservice-toolbox/go/pkg/conn_manager"
)

// -----------------------------------------------------------------------------

// NotifLoggerWrapper wraps LogEngine to expose SetLocalNotifQueue
type NotifLoggerWrapper struct {
	interfaces.Logger
	localNotifier *notifier.LocalNotifier
}

// -----------------------------------------------------------------------------

func (nl *NotifLoggerWrapper) SetLocalNotifQueue(notifChan chan *models.NotifMessage) {
	nl.localNotifier.SetQueue(notifChan)
}

// -----------------------------------------------------------------------------

// Unwrap returns the underlying Logger instance.
func (nl *NotifLoggerWrapper) Unwrap() any {
	return nl.Logger
}

// -----------------------------------------------------------------------------

// NewNotifLogger creates a logger similar to NoLockLogger but with LocalNotifier.
func NewNotifLogger(name string, config *distributed_config.Config, useLocalNotif bool) *NotifLoggerWrapper {
	// 1. Console (Async)
	consoleSink := sink.NewConsoleSink()
	asyncConsole := sink.NewAsyncSink(consoleSink, 1024)

	// 2. File (Async)
	logPath := helpers.GetLogPath(name)
	var fileSink interfaces.Sink
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fileSink = sink.NewConsoleSink()
	} else {
		fileSink = sink.NewWriterSink(f, serializers.NewTextSerializer())
	}
	asyncFile := sink.NewAsyncSink(fileSink, 4096)

	// 3. Network (Async)
	nm := conn_manager.NewNetworkManager(-1, 200, 5000, 2000, 2.0, 0.1)

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
			fmt.Fprintf(os.Stderr, "NotifLogger: Failed to connect to log server: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "NotifLogger: Logger configuration missing for log_server, running local-only\n")
	}

	// 4. MultiSink
	var sinks []interfaces.Sink
	sinks = append(sinks, asyncConsole, asyncFile)
	if networkSink != nil {
		sinks = append(sinks, networkSink)
	}
	multi := sink.NewMultiSink(sinks...)

	// 5. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, multi, false, 0).(*engine.LogEngine)

	// 6. Notifier (Local)
	localNotif := notifier.NewLocalNotifier()
	logger.Notifier = localNotif

	return &NotifLoggerWrapper{
		Logger:        logger,
		localNotifier: localNotif,
	}
}
