package engine

import (
	"testing"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// MockSink for testing
type MockSink struct {
	LastEntry *models.LogEntry
	WriteCount int
	Closed     bool
}

func (m *MockSink) Write(entry *models.LogEntry) error {
	m.WriteCount++
	m.LastEntry = entry
	return nil
}

func (m *MockSink) Close() error {
	m.Closed = true
	return nil
}

// MockNotifier for testing
type MockNotifier struct {
	LastMsg *models.NotifMessage
	NotifyCount int
	Closed      bool
}

func (m *MockNotifier) Notify(msg *models.NotifMessage) error {
	m.NotifyCount++
	m.LastMsg = msg
	return nil
}

func (m *MockNotifier) Close() error {
	m.Closed = true
	return nil
}

func TestLogEngine_Levels(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:  mockSink,
		Level: models.LevelInfo,
		Name:  "TestLogger",
	}

	// 1. Debug (Should be filtered)
	engine.Debug("Debug message")
	if mockSink.WriteCount != 0 {
		t.Errorf("Expected 0 writes for Debug level, got %d", mockSink.WriteCount)
	}

	// 2. Info (Should pass)
	engine.Info("Info message")
	if mockSink.WriteCount != 1 {
		t.Errorf("Expected 1 write for Info level, got %d", mockSink.WriteCount)
	}
	if mockSink.LastEntry.Message != "Info message" {
		t.Errorf("Expected message 'Info message', got '%s'", mockSink.LastEntry.Message)
	}

	// 3. Error (Should pass and trigger notifier)
	mockNotif := &MockNotifier{}
	engine.Notifier = mockNotif
	engine.Error("Error message")
	if mockSink.WriteCount != 2 {
		t.Errorf("Expected 2 writes, got %d", mockSink.WriteCount)
	}
	if mockNotif.NotifyCount != 1 {
		t.Errorf("Expected 1 notification for Error level, got %d", mockNotif.NotifyCount)
	}
	if mockNotif.LastMsg.Message != "Error message" {
		t.Errorf("Expected notification message 'Error message', got '%s'", mockNotif.LastMsg.Message)
	}
}

func TestLogEngine_Close(t *testing.T) {
	mockSink := &MockSink{}
	mockNotif := &MockNotifier{}
	engine := &LogEngine{
		Sink:     mockSink,
		Notifier: mockNotif,
	}

	engine.Close()
	if !mockSink.Closed {
		t.Error("Expected sink to be closed")
	}
	if !mockNotif.Closed {
		t.Error("Expected notifier to be closed")
	}
}
