package profiles

import (
	"fmt"
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/helpers"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
)

// -----------------------------------------------------------------------------
// NewDevelLogger creates a development logger with:
// - Console output (Sync)
// - Local file (Sync) - Path derived from executable or defaults
func NewDevelLogger(name string) interfaces.Logger {
	// 1. Console Sink
	consoleSink := sink.NewConsoleSink()

	// 2. File Sink
	logPath := helpers.GetDefaultLogPath()
	var fileSink interfaces.Sink
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DevelLogger: Failed to open log file %s: %v\n", logPath, err)
		os.Exit(1)
	} else {
		// Use TextSerializer for development logs so they are readable in the file
		fileSink = sink.NewWriterSink(f, serializers.NewTextSerializer())
	}

	// 3. MultiSink (Fan-out)
	// Both are sync.
	multi := sink.NewMultiSink(consoleSink, fileSink)

	// 4. Wrappers
	// SyncPooledSink is obsolete. MultiSink and children handle lifecycle via ref counting.
	return factory.CreateLogEngine(name, models.LevelDebug, multi)
}
