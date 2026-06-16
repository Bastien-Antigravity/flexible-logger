package serializers

import (
	"fmt"
	"strings"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// TextSerializer serializes logs to a human-readable text format.
type TextSerializer struct{}

// -----------------------------------------------------------------------------
func NewTextSerializer() *TextSerializer {
	return &TextSerializer{}
}

// -----------------------------------------------------------------------------
func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// -----------------------------------------------------------------------------
func (s *TextSerializer) Serialize(entry *models.LogEntry) ([]byte, error) {
	// 1. Basic 8 columns (Fixed width) - Mirroring Rust log-server format
	// Rust Format: {:<33} {:<12} {:<22} {:<10} {:<20} {:<25} {:<6} {}

	timestamp := entry.Timestamp.UTC().Format("2006-01-02T15:04:05.000000000Z")

	base := fmt.Sprintf("%-33s %-12s %-22s %-10s %-20s %-25s %-6s %s",
		timestamp,
		truncate(entry.Hostname, 12),
		truncate(entry.LoggerName, 22),
		truncate(entry.Level.String(), 10),
		truncate(entry.Filename, 20),
		truncate(entry.FunctionName, 25),
		truncate(entry.LineNumber, 6),
		entry.Message,
	)

	// 2. Extra metadata (key=value)
	var meta []string
	if entry.Module != "" {
		meta = append(meta, fmt.Sprintf("mod=%s", entry.Module))
	}
	if entry.PathName != "" {
		meta = append(meta, fmt.Sprintf("path=%s", entry.PathName))
	}
	if entry.ProcessID != "" {
		meta = append(meta, fmt.Sprintf("pid=%s", entry.ProcessID))
	}
	if entry.ProcessName != "" {
		meta = append(meta, fmt.Sprintf("pname=%s", entry.ProcessName))
	}
	if entry.ThreadID != "" {
		meta = append(meta, fmt.Sprintf("tid=%s", entry.ThreadID))
	}
	if entry.ThreadName != "" {
		meta = append(meta, fmt.Sprintf("tname=%s", entry.ThreadName))
	}
	if entry.ServiceName != "" {
		meta = append(meta, fmt.Sprintf("svc=%s", entry.ServiceName))
	}

	// Stack trace handling: append it last
	if entry.StackTrace != "" {
		meta = append(meta, fmt.Sprintf("stack=%s", strings.ReplaceAll(entry.StackTrace, "\n", " | ")))
	}

	var result string
	if len(meta) == 0 {
		result = base
	} else {
		result = fmt.Sprintf("%s [metadata: %s]", base, strings.Join(meta, " "))
	}

	return []byte(result + "\n"), nil
}
