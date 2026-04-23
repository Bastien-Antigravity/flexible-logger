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
	// Extra functions
	// -------------------------------------------------------------------------

	// -------------------------------------------------------------------------
	// Stream logs a message at Stream level.
	Stream(format string, args ...any)

	// -------------------------------------------------------------------------
	// Logon logs a message at Logon level.
	Logon(format string, args ...any)

	// -------------------------------------------------------------------------
	// Logout logs a message at Logout level.
	Logout(format string, args ...any)

	// -------------------------------------------------------------------------
	// Trade logs a message at Trade level.
	Trade(format string, args ...any)

	// -------------------------------------------------------------------------
	// Schedule logs a message at Schedule level.
	Schedule(format string, args ...any)

	// -------------------------------------------------------------------------
	// Report logs a message at Report level.
	Report(format string, args ...any)

	// -------------------------------------------------------------------------
	// Log logs a message at a specific level.
	Log(level models.Level, format string, args ...any)

	// -------------------------------------------------------------------------
	// SetLevel sets the current log level.
	SetLevel(level models.Level)

	// GetLevel returns the current log level.
	GetLevel() models.Level

	// -------------------------------------------------------------------------
	// SetCallerSkip sets the number of stack frames to skip when detecting source info.
	SetCallerSkip(skip int)

	// -------------------------------------------------------------------------
	// Close flushes any buffered logs and closes the handler.
	Close()
}
