package engine

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/error_handler"
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

	// Metadata configuration
	ProcessID           int
	ProcessName         string
	CollectCallerInfo   bool // If true, collect caller info for ALL levels. If false, only for Error/Critical.
	CallerSkip          int  // Optional offset to skip more frames (for wrapper libraries)
	SamplingRate        float64
	AlwaysCollectCaller bool // (Redundant if we use logic below, let's keep it simple)
}

// -----------------------------------------------------------------------------
// SetLevel sets the current log level.
func (l *LogEngine) SetLevel(level models.Level) {
	l.Level = level
}

// -----------------------------------------------------------------------------
// SetCallerSkip sets the number of stack frames to skip.
func (l *LogEngine) SetCallerSkip(skip int) {
	l.CallerSkip = skip
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Close() {
	if err := l.Sink.Close(); err != nil {
		error_handler.ReportInternalError(l.Name, "sink.Close", err, "")
	}
	if l.Notifier != nil {
		if err := l.Notifier.Close(); err != nil {
			error_handler.ReportInternalError(l.Name, "notifier.Close", err, "")
		}
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

	// Static Metadata
	e.ProcessID = strconv.Itoa(l.ProcessID)
	e.ProcessName = l.ProcessName

	// Dynamic Metadata (Caller Info)
	// Smart/Selective: Always on for Error/Critical, or if explicitly enabled for all levels
	if l.CollectCallerInfo || level >= models.LevelError {
		// Default skip is 3: getEntry -> Log -> Info/Debug/etc. -> User Code
		// Add CallerSkip to handle wrapping libraries
		pc, file, line, ok := runtime.Caller(3 + l.CallerSkip)
		if ok {
			e.Filename = filepath.Base(file) // Recommendation 1A: Short Filename
			e.LineNumber = strconv.Itoa(line)
			if fn := runtime.FuncForPC(pc); fn != nil {
				e.FunctionName = filepath.Base(fn.Name())
			}
			e.PathName = file // Keep relative/full path in PathName field just in case
		}
	} else {
		e.Filename = "source-context"
		e.LineNumber = "0"
		e.FunctionName = "runtime-caller-skipped"
	}

	return e
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Log(level models.Level, format string, args ...any) {
	if level < l.Level {
		return
	}

	// Sampling logic
	// We only sample non-critical logs (below Warning) if rate is set
	// Note: 0.0 is treated as 100% (not set) for backward compatibility with zero-initialized structs
	if l.SamplingRate > 0 && l.SamplingRate < 1.0 && level < models.LevelWarning {
		if rand.Float64() > l.SamplingRate {
			return
		}
	}

	msg := fmt.Sprintf(format, args...)
	e := l.getEntry(level, msg)
	if err := l.Sink.Write(e); err != nil {
		error_handler.ReportInternalError(l.Name, "sink", err, msg)
	}

	// Check for Notification triggers
	// Example strategy: Notify on Warning or above, or specific rules
	// In govenv this is map-based. Here we do simple level check for demo.
	if l.Notifier != nil && level >= models.LevelWarning {
		n := &models.NotifMessage{
			Message: msg,
			Tags:    []string{"alert"}, // Default tag
		}
		if err := l.Notifier.Notify(n); err != nil {
			error_handler.ReportInternalError(l.Name, "notifier", err, msg)
		}
	}
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Debug(format string, args ...any)  { l.Log(models.LevelDebug, format, args...) }
func (l *LogEngine) Info(format string, args ...any)   { l.Log(models.LevelInfo, format, args...) }
func (l *LogEngine) Stream(format string, args ...any) { l.Log(models.LevelStream, format, args...) }
func (l *LogEngine) Logon(format string, args ...any)  { l.Log(models.LevelLogon, format, args...) }
func (l *LogEngine) Logout(format string, args ...any) { l.Log(models.LevelLogout, format, args...) }
func (l *LogEngine) Trade(format string, args ...any)  { l.Log(models.LevelTrade, format, args...) }
func (l *LogEngine) Schedule(format string, args ...any) {
	l.Log(models.LevelSchedule, format, args...)
}
func (l *LogEngine) Report(format string, args ...any)  { l.Log(models.LevelReport, format, args...) }
func (l *LogEngine) Warning(format string, args ...any) { l.Log(models.LevelWarning, format, args...) }
func (l *LogEngine) Error(format string, args ...any)   { l.Log(models.LevelError, format, args...) }
func (l *LogEngine) Critical(format string, args ...any) {
	l.Log(models.LevelCritical, format, args...)
}
