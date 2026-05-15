package serializers

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	logger_schema "github.com/Bastien-Antigravity/flexible-logger/src/schemas/capnp/logger"

	capnp "capnproto.org/go/capnp/v3"
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
	
	levels := []models.Level{
		models.LevelDebug,
		models.LevelStream,
		models.LevelInfo,
		models.LevelLogon,
		models.LevelLogout,
		models.LevelTrade,
		models.LevelSchedule,
		models.LevelReport,
		models.LevelWarning,
		models.LevelError,
		models.LevelCritical,
	}

	for _, l := range levels {
		entry := &models.LogEntry{
			Timestamp:    time.Now().UTC(),
			Level:        l,
			Message:      "test message " + l.String(),
			Hostname:     "host1",
			LoggerName:   "logger1",
			ProcessID:    "123",
			Filename:     "file.go",
			FunctionName: "main",
			LineNumber:   "10",
		}

		data, err := s.Serialize(entry)
		if err != nil {
			t.Fatalf("Level %v: Failed to serialize: %v", l, err)
		}

		// Deserialize and verify
		msg, err := capnp.UnmarshalPacked(data)
		if err != nil {
			t.Fatalf("Level %v: Failed to unmarshal packed: %v", l, err)
		}

		loggerMsg, err := logger_schema.ReadRootLoggerMsg(msg)
		if err != nil {
			t.Fatalf("Level %v: Failed to read root LoggerMsg: %v", l, err)
		}

		gotMsg, err := loggerMsg.Message_()
		if err != nil {
			t.Fatalf("Level %v: Failed to get message: %v", l, err)
		}
		if gotMsg != entry.Message {
			t.Errorf("Level %v: expected message '%s', got '%s'", l, entry.Message, gotMsg)
		}

		if got := loggerMsg.Level(); uint16(got) != uint16(l) {
			t.Errorf("Level %v: expected level enum %v, got %v", l, uint16(l), uint16(got))
		}
		
		gotHost, err := loggerMsg.Hostname()
		if err != nil {
			t.Fatalf("Level %v: Failed to get hostname: %v", l, err)
		}
		if gotHost != entry.Hostname {
			t.Errorf("Level %v: expected hostname '%s', got '%s'", l, entry.Hostname, gotHost)
		}
	}
}
