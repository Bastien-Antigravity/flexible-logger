package profiles

// =============================================================================
// ESSENTIAL PROCESS: Shared profile helpers resolving server capabilities, public IP pointers, and notifier/drop-alert wiring across all logger profiles.
//
// DATA FLOW:
//   1. Safely inspects optional *distributed_config.Config for server endpoints.
//   2. Attaches LocalNotifier or RemoteNotifier to LogEngine.
//   3. Wires throttled buffer-drop alerting across all asynchronous sinks.
//
// KEY PARAMETERS:
//   - config: Optional distributed-config pointer (nil-safe).
//   - useLocalNotif: Selects in-process channel vs remote notif-server alerts.
// =============================================================================

import (
	distributed_config "github.com/Bastien-Antigravity/distributed-config"
	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
)

// -----------------------------------------------------------------------------

type serverCap struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// -----------------------------------------------------------------------------

func resolveServerCap(config *distributed_config.Config, key string) (serverCap, bool) {
	if config == nil || config.Config == nil {
		return serverCap{}, false
	}
	var capCfg serverCap
	if err := config.GetCapability(key, &capCfg); err == nil && capCfg.IP != "" {
		return capCfg, true
	}
	return serverCap{}, false
}

// -----------------------------------------------------------------------------

func publicIPPtr(config *distributed_config.Config) *string {
	if config != nil && config.Config != nil {
		return &config.Common.PublicIP
	}
	empty := ""
	return &empty
}

// -----------------------------------------------------------------------------

func attachNotifier(name string, config *distributed_config.Config, useLocalNotif bool, logger *engine.LogEngine, asyncSinks ...*sink.AsyncSink) interfaces.Logger {
	wireDropNotifier := func(n interfaces.Notifier) {
		for _, s := range asyncSinks {
			if s != nil {
				s.SetDropNotifier(n, name)
			}
		}
	}

	if useLocalNotif {
		localNotif := notifier.NewLocalNotifier()
		logger.Notifier = localNotif
		wireDropNotifier(localNotif)
		return &NotifLoggerWrapper{
			Logger:        logger,
			localNotifier: localNotif,
		}
	}

	if nsCap, ok := resolveServerCap(config, "notif_server"); ok {
		remoteNotif := notifier.NewRemoteNotifier(&nsCap.IP, &nsCap.Port, publicIPPtr(config), name)
		logger.Notifier = remoteNotif
		wireDropNotifier(remoteNotif)
	}

	return logger
}
