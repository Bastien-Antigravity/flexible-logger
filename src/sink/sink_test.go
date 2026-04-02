package sink

import (
	"bytes"
	"sync"
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
	LastEntry  *models.LogEntry
	WriteCount int
	Closed     bool
	mu         sync.Mutex
}

func (m *MockSink) Write(entry *models.LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WriteCount++
	m.LastEntry = entry
	return nil
}

func (m *MockSink) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Closed = true
	return nil
}

func (m *MockSink) GetWriteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.WriteCount
}

func (m *MockSink) GetLastEntry() *models.LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.LastEntry
}

func (m *MockSink) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Closed
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

	if mockSink.GetWriteCount() != 1 {
		t.Errorf("Expected 1 write, got %d", mockSink.GetWriteCount())
	}
	if mockSink.GetLastEntry().Message != "Async message" {
		t.Errorf("Expected 'Async message', got '%s'", mockSink.GetLastEntry().Message)
	}

	asyncSink.Close()
	if !mockSink.IsClosed() {
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

	if sink1.GetWriteCount() != 1 || sink2.GetWriteCount() != 1 {
		t.Errorf("Expected 1 write in both sinks, got sink1=%d, sink2=%d", sink1.GetWriteCount(), sink2.GetWriteCount())
	}
}
