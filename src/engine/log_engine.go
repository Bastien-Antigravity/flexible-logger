// -----------------------------------------------------------------------------
// ESSENTIAL PROCESS:
// Core logging engine coordinating log entry acquisition, dynamic metadata
// sampling, filtering, and sink dispatch.
//
// DATA FLOW:
// 1. Receives log invocation (Level, format, arguments).
// 2. Evaluates severity threshold and optional probabilistic sampling.
// 3. Borrows LogEntry from sync.Pool and enriches with caller metadata.
// 4. Dispatches entry to configured primary Sink.
// 5. Triggers Notifier on Warning or Error severity levels.
//
// KEY PARAMETERS:
// - Sink: Destination sink handling log persistence or network routing.
// - Notifier: Optional alert dispatcher for high-severity events.
// - Level: Configured minimum logging severity threshold.
// - SamplingRate: Statistical sampling ratio for non-critical entries.
// -----------------------------------------------------------------------------

package engine

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/error_handler"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
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
	ProcessID         int
	ProcessIDStr      string
	ProcessName       string
	CollectCallerInfo bool // If true, collect caller info for ALL levels. If false, only for Error/Critical.
	CallerSkip        int  // Optional offset to skip more frames (for wrapper libraries)
	SamplingRate      float64

	closeOnce          sync.Once
	atomicLevel        atomic.Uint32
	atomicCallerSkip   atomic.Int32
	callerSkipSet      atomic.Bool
	atomicSamplingRate atomic.Uint64
	samplingRateSet    atomic.Bool
}

// -----------------------------------------------------------------------------
func (l *LogEngine) getProcessIDStr() string {
	if l.ProcessIDStr != "" {
		return l.ProcessIDStr
	}
	if l.ProcessID != 0 {
		return strconv.Itoa(l.ProcessID)
	}
	return "0"
}

// -----------------------------------------------------------------------------
// SetLevel sets the current log level.
func (l *LogEngine) SetLevel(level models.Level) {
	l.atomicLevel.Store(uint32(level) + 1)
}

// GetLevel returns the current log level.
func (l *LogEngine) GetLevel() models.Level {
	if val := l.atomicLevel.Load(); val != 0 {
		return models.Level(val - 1)
	}
	return l.Level
}

// -----------------------------------------------------------------------------
// SetCallerSkip sets the number of stack frames to skip.
func (l *LogEngine) SetCallerSkip(skip int) {
	l.atomicCallerSkip.Store(int32(skip))
	l.callerSkipSet.Store(true)
}

func (l *LogEngine) getCallerSkip() int {
	if l.callerSkipSet.Load() {
		return int(l.atomicCallerSkip.Load())
	}
	return l.CallerSkip
}

// -----------------------------------------------------------------------------
// SetSamplingRate updates the non-critical log sampling rate atomically.
func (l *LogEngine) SetSamplingRate(rate float64) {
	l.atomicSamplingRate.Store(math.Float64bits(rate))
	l.samplingRateSet.Store(true)
}

func (l *LogEngine) getSamplingRate() float64 {
	if l.samplingRateSet.Load() {
		return math.Float64frombits(l.atomicSamplingRate.Load())
	}
	return l.SamplingRate
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Close() {
	l.closeOnce.Do(func() {
		if l.Sink != nil {
			if err := l.Sink.Close(); err != nil {
				error_handler.ReportInternalError(l.Name, "sink.Close", err, "")
			}
		}
		if l.Notifier != nil {
			if err := l.Notifier.Close(); err != nil {
				error_handler.ReportInternalError(l.Name, "notifier.Close", err, "")
			}
		}
	})
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
	e.ProcessID = l.getProcessIDStr()
	e.ProcessName = l.ProcessName

	// Dynamic Metadata (Caller Info)
	// Smart/Selective: Always on for Error/Critical, or if explicitly enabled for all levels
	if l.CollectCallerInfo || level >= models.LevelError {
		// Default skip is 3: getEntry -> Log -> Info/Debug/etc. -> User Code
		// Add CallerSkip to handle wrapping libraries
		pc, file, line, ok := runtime.Caller(3 + l.getCallerSkip())
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

func isOnlyBufferFullErr(err error) bool {
	if errors.Is(err, sink.ErrBufferFull) {
		return true
	}
	const prefix = "multisink write errors: "
	msg := err.Error()
	if !strings.HasPrefix(msg, prefix) {
		return false
	}
	for _, part := range strings.Split(strings.TrimPrefix(msg, prefix), "; ") {
		if part != sink.ErrBufferFull.Error() {
			return false
		}
	}
	return true
}

// -----------------------------------------------------------------------------
func (l *LogEngine) dispatchEntry(e *models.LogEntry, level models.Level, msg string) {
	if err := l.Sink.Write(e); err != nil && !isOnlyBufferFullErr(err) {
		error_handler.ReportInternalError(l.Name, "sink", err, msg)
	}

	// Check for Notification triggers
	if l.Notifier != nil && level >= models.LevelWarning {
		n := &models.NotifMessage{
			Message: msg,
			Level:   level.String(), // Send the string representation of the level
		}
		if err := l.Notifier.Notify(n); err != nil && err.Error() != "no local queue bound" {
			error_handler.ReportInternalError(l.Name, "notifier", err, msg)
		}
	}
}

// -----------------------------------------------------------------------------
func (l *LogEngine) logInternal(level models.Level, format string, args ...any) {
	if level < l.GetLevel() {
		return
	}

	// Sampling logic
	// We only sample non-critical logs (below Warning) if rate is set
	// Note: 0.0 is treated as 100% (not set) for backward compatibility with zero-initialized structs
	rate := l.getSamplingRate()
	if rate > 0 && rate < 1.0 && level < models.LevelWarning {
		if rand.Float64() > rate {
			return
		}
	}

	msg := fmt.Sprintf(format, args...)
	e := l.getEntry(level, msg)
	l.dispatchEntry(e, level, msg)
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Log(level models.Level, format string, args ...any) {
	l.logInternal(level, format, args...)
}

// -----------------------------------------------------------------------------
func (l *LogEngine) LogWithCaller(level models.Level, msg, file, line, function, module string) {
	if level < l.GetLevel() {
		return
	}

	// Sampling logic
	rate := l.getSamplingRate()
	if rate > 0 && rate < 1.0 && level < models.LevelWarning {
		if rand.Float64() > rate {
			return
		}
	}

	e := models.EntryPool.Get().(*models.LogEntry)
	e.Reset()
	e.Timestamp = time.Now().UTC()
	e.Level = level
	e.Message = msg
	e.LoggerName = l.Name
	e.Hostname = l.Hostname
	e.ServiceName = l.ServiceName

	// Static Metadata
	e.ProcessID = l.getProcessIDStr()
	e.ProcessName = l.ProcessName

	// Dynamic Metadata from Explicit Caller Arguments
	if file != "" {
		e.Filename = filepath.Base(file)
		e.PathName = file
	} else {
		e.Filename = "unknown"
	}
	if line != "" {
		e.LineNumber = line
	} else {
		e.LineNumber = "0"
	}
	if function != "" {
		e.FunctionName = function
	} else {
		e.FunctionName = "unknown"
	}
	e.Module = module

	l.dispatchEntry(e, level, msg)
}

// -----------------------------------------------------------------------------
func (l *LogEngine) Debug(format string, args ...any) {
	l.logInternal(models.LevelDebug, format, args...)
}
func (l *LogEngine) Info(format string, args ...any) {
	l.logInternal(models.LevelInfo, format, args...)
}
func (l *LogEngine) Stream(format string, args ...any) {
	l.logInternal(models.LevelStream, format, args...)
}
func (l *LogEngine) Logon(format string, args ...any) {
	l.logInternal(models.LevelLogon, format, args...)
}
func (l *LogEngine) Logout(format string, args ...any) {
	l.logInternal(models.LevelLogout, format, args...)
}
func (l *LogEngine) Trade(format string, args ...any) {
	l.logInternal(models.LevelTrade, format, args...)
}
func (l *LogEngine) Schedule(format string, args ...any) {
	l.logInternal(models.LevelSchedule, format, args...)
}
func (l *LogEngine) Report(format string, args ...any) {
	l.logInternal(models.LevelReport, format, args...)
}
func (l *LogEngine) Warning(format string, args ...any) {
	l.logInternal(models.LevelWarning, format, args...)
}
func (l *LogEngine) Error(format string, args ...any) {
	l.logInternal(models.LevelError, format, args...)
}
func (l *LogEngine) Critical(format string, args ...any) {
	l.logInternal(models.LevelCritical, format, args...)
}
