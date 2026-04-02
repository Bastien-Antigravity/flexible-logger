package error_handler

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
