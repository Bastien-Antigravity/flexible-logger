package profiles

// =============================================================================
// ESSENTIAL PROCESS: Minimalist logging profile buffering asynchronous text output to console without file or network overhead.
//
// DATA FLOW:
//   1. Initializes ConsoleSink wrapped in lightweight AsyncSink.
//   2. Binds LogEngine with Info severity threshold.
//   3. Provides zero-dependency logger for lightweight CLI tools.
//
// KEY PARAMETERS:
//   - name: CLI or utility identifier.
//   - useLocalNotif: Alert queue selector.
// =============================================================================

import (
	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
)

// -----------------------------------------------------------------------------
// NewMinimalLogger creates a minimal logger with:
// - Console output (Async)
func NewMinimalLogger(name string, useLocalNotif bool) interfaces.Logger {
	// 1. Console Sink
	consoleSink := sink.NewConsoleSink()

	// 2. Async Wrapper
	// Buffer size can be small for minimal
	asyncConsole := sink.NewAsyncSink(consoleSink, 1024)

	// 3. Engine
	logger := factory.CreateLogEngine(name, models.LevelInfo, asyncConsole, false, 1.0).(*engine.LogEngine)

	// 4. Notifier
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
