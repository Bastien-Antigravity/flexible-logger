package interfaces

import "github.com/Bastien-Antigravity/flexible-logger/src/models"

// -----------------------------------------------------------------------------
// Logger is the main interface for logging
type Logger interface {
	// -------------------------------------------------------------------------
	// Debug logs a message at Debug level.
	Debug(format string, args ...any)

	// -------------------------------------------------------------------------
	// Info logs a message at Info level.
	Info(format string, args ...any)

	// -------------------------------------------------------------------------
	// Warning logs a message at Warning level.
	Warning(format string, args ...any)

	// -------------------------------------------------------------------------
	// Error logs a message at Error level.
	Error(format string, args ...any)

	// -------------------------------------------------------------------------
	// Critical logs a message at Critical level.
	Critical(format string, args ...any)

	// -------------------------------------------------------------------------
	// Log logs a message at a specific level.
	Log(level models.Level, format string, args ...any)

	// -------------------------------------------------------------------------
	// Close flushes any buffered logs and closes the handler.
	Close()
}
