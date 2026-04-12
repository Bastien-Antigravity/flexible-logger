package engine

import (
	"sync"
	"testing"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// MockSink for testing
type MockSink struct {
	LastEntry *models.LogEntry
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

// MockNotifier for testing
type MockNotifier struct {
	LastMsg *models.NotifMessage
	NotifyCount int
	Closed      bool
	mu          sync.Mutex
}

func (m *MockNotifier) Notify(msg *models.NotifMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.NotifyCount++
	m.LastMsg = msg
	return nil
}

func (m *MockNotifier) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Closed = true
	return nil
}

func (m *MockNotifier) GetNotifyCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.NotifyCount
}

func (m *MockNotifier) GetLastMsg() *models.NotifMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.LastMsg
}

func (m *MockNotifier) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Closed
}

func TestLogEngine_Levels(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:         mockSink,
		Level:        models.LevelInfo,
		Name:         "TestLogger",
		SamplingRate: 1.0,
	}

	// 1. Debug (Should be filtered)
	engine.Debug("Debug message")
	if mockSink.WriteCount != 0 {
		t.Errorf("Expected 0 writes for Debug level, got %d", mockSink.WriteCount)
	}

	// 2. Info (Should pass)
	engine.Info("Info message")
	if mockSink.GetWriteCount() != 1 {
		t.Errorf("Expected 1 write for Info level, got %d", mockSink.GetWriteCount())
	}
	lastEntry := mockSink.GetLastEntry()
	if lastEntry == nil {
		t.Fatal("Expected LastEntry not to be nil")
	}
	if lastEntry.Message != "Info message" {
		t.Errorf("Expected message 'Info message', got '%s'", lastEntry.Message)
	}

	// 3. Error (Should pass and trigger notifier)
	mockNotif := &MockNotifier{}
	engine.Notifier = mockNotif
	engine.Error("Error message")
	if mockSink.GetWriteCount() != 2 {
		t.Errorf("Expected 2 writes, got %d", mockSink.GetWriteCount())
	}
	if mockNotif.GetNotifyCount() != 1 {
		t.Errorf("Expected 1 notification for Error level, got %d", mockNotif.GetNotifyCount())
	}
	if mockNotif.GetLastMsg().Message != "Error message" {
		t.Errorf("Expected notification message 'Error message', got '%s'", mockNotif.GetLastMsg().Message)
	}
}

func TestLogEngine_Close(t *testing.T) {
	mockSink := &MockSink{}
	mockNotif := &MockNotifier{}
	engine := &LogEngine{
		Sink:         mockSink,
		Notifier:     mockNotif,
		SamplingRate: 1.0,
	}

	engine.Close()
	if !mockSink.IsClosed() {
		t.Error("Expected sink to be closed")
	}
	if !mockNotif.IsClosed() {
		t.Error("Expected notifier to be closed")
	}
}

func TestLogEngine_Sampling(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:         mockSink,
		Level:        models.LevelDebug,
		SamplingRate: 0.5, // Keep 50%
	}

	total := 10000
	for i := 0; i < total; i++ {
		engine.Info("Test")
	}

	count := mockSink.GetWriteCount()
	// Statistical check: should be roughy 5000. We use a range of +/- 500 (10 sigma) to ensure stability.
	if count < 4500 || count > 5500 {
		t.Errorf("Sampling logic seems off. Expected ~5000 logs, got %d", count)
	}

	// Critical logs should NEVER be sampled
	mockSink.WriteCount = 0
	for i := 0; i < 100; i++ {
		engine.Error("Critical error")
	}
	if mockSink.GetWriteCount() != 100 {
		t.Errorf("Error logs were sampled! Expected 100, got %d", mockSink.GetWriteCount())
	}
}

func TestLogEngine_MetadataFallback(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:              mockSink,
		Level:             models.LevelInfo,
		CollectCallerInfo: false, // Disabled
		SamplingRate:      1.0,
	}

	engine.Info("Hello")
	entry := mockSink.GetLastEntry()
	if entry == nil {
		t.Fatal("Expected entry not to be nil")
	}

	if entry.Filename != "source-context" {
		t.Errorf("Expected fallback filename 'source-context', got '%s'", entry.Filename)
	}
	if entry.FunctionName != "runtime-caller-skipped" {
		t.Errorf("Expected fallback function 'runtime-caller-skipped', got '%s'", entry.FunctionName)
	}
}

func BenchmarkLogEngine_NoSampling(b *testing.B) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:              mockSink,
		Level:             models.LevelInfo,
		CollectCallerInfo: false,
		SamplingRate:      1.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Info("Benchmark message %d", i)
	}
}
