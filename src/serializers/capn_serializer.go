package serializers

import (
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	logger_schema "github.com/Bastien-Antigravity/flexible-logger/src/schemas/logger_msg"

	capnp "capnproto.org/go/capnp/v3"
)

// -----------------------------------------------------------------------------
// CapnpSerializer implements the Serializer interface manually using LoggerMsg schema.
type CapnpSerializer struct{}

// -----------------------------------------------------------------------------
func NewCapnpSerializer() *CapnpSerializer {
	return &CapnpSerializer{}
}

// -----------------------------------------------------------------------------
// Serialize converts LogEntry to bytes using LoggerMsg Cap'n Proto schema.
func (s *CapnpSerializer) Serialize(entry *models.LogEntry) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	loggerMsg, err := logger_schema.NewRootLoggerMsg(seg)
	if err != nil {
		return nil, err
	}

	// Map Fields

	// @0 Timestamp (Text)
	// Use fixed-width nanoseconds (000000000) to ensure consistent log lines
	_ = loggerMsg.SetTimestamp(entry.Timestamp.UTC().Format("2006-01-02T15:04:05.000000000Z"))

	// @1 Hostname (Text)
	_ = loggerMsg.SetHostname(entry.Hostname)

	// @2 LoggerName (Text)
	_ = loggerMsg.SetLoggerName(entry.LoggerName)

	// @3 Module (Text)
	_ = loggerMsg.SetModule(entry.Module)

	// @4 Level (Enum)
	loggerMsg.SetLevel(mapLevel(entry.Level))

	// @5 Filename (Text)
	_ = loggerMsg.SetFilename(entry.Filename)

	// @6 FunctionName (Text)
	_ = loggerMsg.SetFunctionName(entry.FunctionName)

	// @7 LineNumber (Text)
	_ = loggerMsg.SetLineNumber(entry.LineNumber)

	// @8 Message (Text)
	_ = loggerMsg.SetMessage_(entry.Message)

	// @9 PathName (Text)
	_ = loggerMsg.SetPathName(entry.PathName)

	// @10 ProcessId (Text)
	_ = loggerMsg.SetProcessId(entry.ProcessID)

	// @11 ProcessName (Text)
	_ = loggerMsg.SetProcessName(entry.ProcessName)

	// @12 ThreadId (Text)
	_ = loggerMsg.SetThreadId(entry.ThreadID)

	// @13 ThreadName (Text)
	_ = loggerMsg.SetThreadName(entry.ThreadName)

	// @14 ServiceName (Text)
	_ = loggerMsg.SetServiceName(entry.ServiceName)

	// @15 StackTrace (Text)
	_ = loggerMsg.SetStackTrace(entry.StackTrace)

	return msg.MarshalPacked()
}

// -----------------------------------------------------------------------------
func mapLevel(l models.Level) logger_schema.Level {
	switch l {
	case models.LevelDebug:
		return logger_schema.Level_debug
	case models.LevelInfo:
		return logger_schema.Level_info
	case models.LevelWarning:
		return logger_schema.Level_warning
	case models.LevelError:
		return logger_schema.Level_error
	case models.LevelCritical:
		return logger_schema.Level_critical
	default:
		return logger_schema.Level_notset
	}
}
