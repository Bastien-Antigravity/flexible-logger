package main

// =============================================================================
// ESSENTIAL PROCESS: Interactive visual inspection tool demonstrating 8-column log formatting.
//
// DATA FLOW:
//   1. Creates a sample LogEntry structure.
//   2. Serializes entry using TextSerializer into canonical 8-column format.
//   3. Emits formatted byte stream directly to standard output.
//
// KEY PARAMETERS:
//   - serializer: TextSerializer instance under demonstration.
//   - entry: Sample LogEntry struct with metadata.
// =============================================================================

import (
	"fmt"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
)

func main() {
	serializer := serializers.NewTextSerializer()
	entry := &models.LogEntry{
		Timestamp:  time.Now().UTC(),
		Level:      models.LevelInfo,
		ProcessID:  "12345",
		Filename:   "main.go",
		LineNumber: "42",
		LoggerName: "TestLogger",
		Message:    "This is an example log message",
	}

	output, _ := serializer.Serialize(entry)
	fmt.Print(string(output))
}
