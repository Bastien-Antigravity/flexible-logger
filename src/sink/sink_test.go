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
	// Note: We copy the message because the entry might be reused after Release
	if entry != nil {
		e := *entry
		m.LastEntry = &e
		entry.Release()
	}
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

	err := multi.Write(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sink1.GetWriteCount() != 1 || sink2.GetWriteCount() != 1 {
		t.Errorf("Expected 1 write in both sinks, got sink1=%d, sink2=%d", sink1.GetWriteCount(), sink2.GetWriteCount())
	}
}

func TestAsyncSink_BufferFull_DropsLog(t *testing.T) {
	mockSink := &MockSink{}
	// Small buffer
	asyncSink := NewAsyncSink(mockSink, 1)

	// Fill buffer and worker's current processing
	e1 := models.EntryPool.Get().(*models.LogEntry)
	e1.Reset()
	e1.Message = "Msg 1"
	_ = asyncSink.Write(e1)

	e2 := models.EntryPool.Get().(*models.LogEntry)
	e2.Reset()
	e2.Message = "Msg 2"
	_ = asyncSink.Write(e2)

	// Third one should be dropped
	e3 := models.EntryPool.Get().(*models.LogEntry)
	e3.Reset()
	e3.Message = "Msg 3"
	err := asyncSink.Write(e3)

	if err == nil {
		t.Error("Expected error when buffer is full, got nil")
	}
}

func TestMultiSink_ZeroSinks(t *testing.T) {
	multi := NewMultiSink()

	entry := models.EntryPool.Get().(*models.LogEntry)
	entry.Reset()

	err := multi.Write(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestMultiSink_Close_AllSinksClosed(t *testing.T) {
	sink1 := &MockSink{}
	sink2 := &MockSink{}
	multi := NewMultiSink(sink1, sink2)

	_ = multi.Close()

	if !sink1.IsClosed() || !sink2.IsClosed() {
		t.Error("Expected all underlying sinks to be closed")
	}
}
