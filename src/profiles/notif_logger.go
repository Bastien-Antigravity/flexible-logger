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
	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
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

// NewNotifLogger creates an asynchronous logger wired with LocalNotifier.
func NewNotifLogger(name string, config *distributed_config.Config, useLocalNotif bool) *NotifLoggerWrapper {
	return NewNoLockLogger(name, config, true).(*NotifLoggerWrapper)
}
