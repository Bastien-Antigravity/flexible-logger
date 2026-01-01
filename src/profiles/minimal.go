package profiles

import (
	"github.com/Bastien-Antigravity/flexible-logger/src/factory"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
)

// -----------------------------------------------------------------------------
// NewMinimalLogger creates a minimal logger with:
// - Console output (Async)
func NewMinimalLogger(name string) interfaces.Logger {
	// 1. Console Sink
	consoleSink := sink.NewConsoleSink()

	// 2. Async Wrapper
	// Buffer size can be small for minimal
	asyncConsole := sink.NewAsyncSink(consoleSink, 1024)

	return factory.CreateLogEngine(name, models.LevelInfo, asyncConsole)
}
