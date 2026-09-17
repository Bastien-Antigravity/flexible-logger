package profiles

// =============================================================================
// ESSENTIAL PROCESS: Development logging profile providing synchronous, human-readable console and local file output at Debug level.
//
// DATA FLOW:
//   1. Configures synchronous ConsoleSink with TextSerializer.
//   2. Configures synchronous local FileSink at resolved path.
//   3. Combines sinks in MultiSink with Debug severity threshold.
//
// KEY PARAMETERS:
//   - name: Application identifier.
//   - useLocalNotif: Notification routing flag.
// =============================================================================

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/helpers"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/notifier"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
)

// -----------------------------------------------------------------------------
// NewDevelLogger creates a development logger with:
// - Console output (Sync)
// - Local file (Sync) - Path derived from executable or defaults
func NewDevelLogger(name string, useLocalNotif bool) interfaces.Logger {
	// 1. Console Sink
	consoleSink := sink.NewConsoleSink()

	// 2. File Sink
	logPath := helpers.GetLogPath(name)
	var fileSink interfaces.Sink
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DevelLogger: Failed to open log file %s: %v, falling back to console\n", logPath, err)
		fileSink = consoleSink
	} else {
		// Use TextSerializer for development logs so they are readable in the file
		fileSink = sink.NewWriterSink(f, serializers.NewTextSerializer())
	}

	// 3. MultiSink (Fan-out)
	// Both are sync.
	multi := sink.NewMultiSink(consoleSink, fileSink)

	// 4. Wrappers
	// SyncPooledSink is obsolete. MultiSink and children handle lifecycle via ref counting.
	// 4. Engine
	logger := factory.CreateLogEngine(name, models.LevelDebug, multi, true, 1.0).(*engine.LogEngine)

	// 5. Notifier
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
