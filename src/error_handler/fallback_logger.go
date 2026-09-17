package error_handler

// =============================================================================
// ESSENTIAL PROCESS: Isolated diagnostic fallback logger preventing infinite recursion when logger infrastructure fails.
//
// DATA FLOW:
//   1. Catches infrastructure errors from sinks, buffers, or network managers.
//   2. Formats internal error using TextSerializer into a LogEntry.
//   3. Writes emergency diagnostic message directly to os.Stderr.
//
// KEY PARAMETERS:
//   - loggerName: Subsystem or logger name reporting the failure.
//   - source: Internal process or method where error occurred.
//   - err: Root cause error instance.
//   - originalMsg: Original log message payload that failed delivery.
// =============================================================================

import (
	"fmt"
	"os"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
)

// -----------------------------------------------------------------------------
// ReportInternalError is the centralized way to report errors in the logger itself.
// It formats the error as a LogEntry to maintain consistency.
func ReportInternalError(loggerName string, source string, err error, originalMsg string) {
	serializer := serializers.NewTextSerializer()

	e := &models.LogEntry{
		Timestamp:  time.Now().UTC(),
		Level:      models.LevelError,
		LoggerName: loggerName,
		Message:    fmt.Sprintf("INTERNAL ERROR [%s]: %v (Original: %s)", source, err, originalMsg),
	}

	data, sErr := serializer.Serialize(e)
	if sErr != nil {
		// Absolute fallback if serialization also fails
		fmt.Fprintf(os.Stderr, "[%s] [ERROR] %s: INTERNAL ERROR [%s]: %v (Original: %s)\n",
			time.Now().UTC().Format(time.RFC3339), loggerName, source, err, originalMsg)
		return
	}

	os.Stderr.Write(data)
}
