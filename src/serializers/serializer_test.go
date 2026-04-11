package serializers

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

func TestJSONSerializer(t *testing.T) {
	s := NewJSONSerializer()
	entry := &models.LogEntry{
		Timestamp:    time.Now().UTC(),
		Level:        models.LevelInfo,
		Message:      "test message",
		LoggerName:   "test-logger",
		Filename:     "test.go",
		LineNumber:   "42",
		ProcessID:    "123",
	}

	data, err := s.Serialize(entry)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Verify valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify key fields
	if parsed["message"] != "test message" {
		t.Errorf("Expected message 'test message', got '%v'", parsed["message"])
	}
	if parsed["loggerName"] != "test-logger" {
		t.Errorf("Expected loggerName 'test-logger', got '%v'", parsed["loggerName"])
	}
}

func TestTextSerializer(t *testing.T) {
	s := NewTextSerializer()
	entry := &models.LogEntry{
		Timestamp:    time.Now().UTC(),
		Level:        models.LevelInfo,
		Message:      "test message",
		LoggerName:   "test-logger",
		Filename:     "test.go",
		LineNumber:   "42",
		ProcessID:    "123",
	}

	data, err := s.Serialize(entry)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	str := string(data)
	// Format: [TIMESTAMP] [LEVEL] [PID] [FILE:LINE] LOGGER: MESSAGE\n
	if !strings.Contains(str, "[INFO") {
		t.Errorf("Expected level [INFO], got %s", str)
	}
	if !strings.Contains(str, "[123]") {
		t.Errorf("Expected PID [123], got %s", str)
	}
	if !strings.Contains(str, "[test.go:42]") {
		t.Errorf("Expected source [test.go:42], got %s", str)
	}
	if !strings.Contains(str, "test-logger: test message") {
		t.Errorf("Expected logger/message, got %s", str)
	}
}

func TestCapnpSerializer(t *testing.T) {
	s := NewCapnpSerializer()
	entry := &models.LogEntry{
		Level:   models.LevelInfo,
		Message: "test",
	}

	data, err := s.Serialize(entry)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	if len(data) == 0 {
		t.Error("Capnp output is empty")
	}
	// Basic check for Cap'n Proto framing (segment count least significant byte 0 for single segment)
	if len(data) < 8 {
		t.Error("Capnp output too short for header")
	}
}
