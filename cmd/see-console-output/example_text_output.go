package main

import (
	"fmt"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/serializers"
)

func main() {
	serializer := serializers.NewTextSerializer()
	entry := &models.LogEntry{
		Timestamp:    time.Now(),
		Level:        models.LevelInfo,
		ProcessID:    "12345",
		Filename:     "main.go",
		LineNumber:   "42",
		LoggerName:   "TestLogger",
		Message:      "This is an example log message",
	}

	output, _ := serializer.Serialize(entry)
	fmt.Print(string(output))
}
