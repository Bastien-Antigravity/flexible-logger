package engine

import (
	"fmt"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// LogEngine is the concrete core implementation of the Logger interface.
// It orchestrates populating LogEntries and dispatching them to Sinks.
type LogEngine struct {
	Sink        interfaces.Sink
	Notifier    interfaces.Notifier // Optional notifier
	Level       models.Level
	Name        string
	Hostname    string
	ServiceName string
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Close() {
	l.Sink.Close()
	if l.Notifier != nil {
		l.Notifier.Close()
	}
}

// -----------------------------------------------------------------------------
func (l *LogEngine) getEntry(level models.Level, msg string) *models.LogEntry {
	e := models.EntryPool.Get().(*models.LogEntry)
	e.Reset()
	e.Timestamp = time.Now().UTC()
	e.Level = level
	e.Message = msg
	e.LoggerName = l.Name
	e.Hostname = l.Hostname
	e.ServiceName = l.ServiceName
	return e
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Log(level models.Level, format string, args ...any) {
	if level < l.Level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	e := l.getEntry(level, msg)
	l.Sink.Write(e)

	// Check for Notification triggers
	// Example strategy: Notify on Warning or above, or specific rules
	// In govenv this is map-based. Here we do simple level check for demo.
	if l.Notifier != nil && level >= models.LevelWarning {
		n := &models.NotifMessage{
			Message: msg,
			Tags:    []string{"alert"}, // Default tag
		}
		l.Notifier.Notify(n)
	}
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Debug(format string, args ...any)    { l.Log(models.LevelDebug, format, args...) }
func (l *LogEngine) Info(format string, args ...any)     { l.Log(models.LevelInfo, format, args...) }
func (l *LogEngine) Warning(format string, args ...any)  { l.Log(models.LevelWarning, format, args...) }
func (l *LogEngine) Error(format string, args ...any)    { l.Log(models.LevelError, format, args...) }
func (l *LogEngine) Critical(format string, args ...any) { l.Log(models.LevelCritical, format, args...) }
