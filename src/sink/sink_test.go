package sink

import (
	"bytes"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// MockSerializer for testing
type MockSerializer struct {
	SerializeFunc func(entry *models.LogEntry) ([]byte, error)
}

func (m *MockSerializer) Serialize(entry *models.LogEntry) ([]byte, error) {
	if m.SerializeFunc != nil {
		return m.SerializeFunc(entry)
	}
	return []byte(entry.Message), nil
}

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

func TestWriterSink_Write(t *testing.T) {
	buf := &bytes.Buffer{}
	mockSer := &MockSerializer{}
	sink := NewWriterSink(buf, mockSer)

	entry := models.EntryPool.Get().(*models.LogEntry)
	entry.Reset()
	entry.Message = "Test message"

	err := sink.Write(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.String() != "Test message" {
		t.Errorf("Expected 'Test message', got '%s'", buf.String())
	}
}

func TestAsyncSink_Write(t *testing.T) {
	mockSink := &MockSink{}
	asyncSink := NewAsyncSink(mockSink, 10)

	entry := models.EntryPool.Get().(*models.LogEntry)
	entry.Reset()
	entry.Message = "Async message"

	err := asyncSink.Write(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Wait for worker to process
	time.Sleep(50 * time.Millisecond)

	if mockSink.WriteCount != 1 {
		t.Errorf("Expected 1 write, got %d", mockSink.WriteCount)
	}
	if mockSink.LastEntry.Message != "Async message" {
		t.Errorf("Expected 'Async message', got '%s'", mockSink.LastEntry.Message)
	}

	asyncSink.Close()
	if !mockSink.Closed {
		t.Error("Expected underlying sink to be closed")
	}
}

func TestMultiSink_Write(t *testing.T) {
	sink1 := &MockSink{}
	sink2 := &MockSink{}
	multi := NewMultiSink(sink1, sink2)

	entry := models.EntryPool.Get().(*models.LogEntry)
	entry.Reset()
	entry.Message = "Multi message"

	multi.Write(entry)

	if sink1.WriteCount != 1 || sink2.WriteCount != 1 {
		t.Errorf("Expected 1 write in both sinks, got sink1=%d, sink2=%d", sink1.WriteCount, sink2.WriteCount)
	}
}
