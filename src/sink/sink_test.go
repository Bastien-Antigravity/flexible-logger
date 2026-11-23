package sink

// =============================================================================
// ESSENTIAL PROCESS: Unit tests for WriterSink, AsyncSink, ConsoleSink, and MultiSink reference counting and backpressure behavior.
//
// DATA FLOW:
//   1. Executes concurrent writes against buffered and unbuffered sinks.
//   2. Verifies buffer saturation drop behavior and reference count accounting.
//
// KEY PARAMETERS:
//   - t: Testing harness.
// =============================================================================

import (
	"bytes"
	"fmt"
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

type blockingMockSink struct {
	unblock chan struct{}
	started chan struct{}
	once    sync.Once
}

func (b *blockingMockSink) Write(entry *models.LogEntry) error {
	b.once.Do(func() {
		close(b.started)
	})
	<-b.unblock
	if entry != nil {
		entry.Release()
	}
	return nil
}

func (b *blockingMockSink) Close() error {
	return nil
}

func TestAsyncSink_BufferFull_DropsLog(t *testing.T) {
	blocker := &blockingMockSink{
		unblock: make(chan struct{}),
		started: make(chan struct{}),
	}
	// Buffer capacity = 1
	asyncSink := NewAsyncSink(blocker, 1)
	defer func() {
		close(blocker.unblock)
		_ = asyncSink.Close()
	}()

	if asyncSink.DroppedCount() != 0 {
		t.Fatalf("Expected initial DroppedCount() == 0, got %d", asyncSink.DroppedCount())
	}

	// 1. Entry 1 is picked up by the worker goroutine and blocks on blocker.unblock
	e1 := models.EntryPool.Get().(*models.LogEntry)
	e1.Reset()
	e1.Message = "Msg 1"
	if err := asyncSink.Write(e1); err != nil {
		t.Fatalf("Expected Msg 1 to be accepted, got %v", err)
	}
	<-blocker.started

	// 2. Entry 2 fills the buffered channel (capacity 1)
	e2 := models.EntryPool.Get().(*models.LogEntry)
	e2.Reset()
	e2.Message = "Msg 2"
	if err := asyncSink.Write(e2); err != nil {
		t.Fatalf("Expected Msg 2 to fill buffer, got %v", err)
	}

	// 3. Entry 3 & Entry 4 overflow the saturated buffer and must increment DroppedCount()
	e3 := models.EntryPool.Get().(*models.LogEntry)
	e3.Reset()
	e3.Message = "Msg 3"
	if err := asyncSink.Write(e3); err == nil {
		t.Error("Expected error when buffer is full on Msg 3, got nil")
	}

	e4 := models.EntryPool.Get().(*models.LogEntry)
	e4.Reset()
	e4.Message = "Msg 4"
	if err := asyncSink.Write(e4); err == nil {
		t.Error("Expected error when buffer is full on Msg 4, got nil")
	}

	if got := asyncSink.DroppedCount(); got != 2 {
		t.Errorf("Expected DroppedCount() == 2 after 2 dropped entries, got %d", got)
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

func TestMultiSink_OneSink(t *testing.T) {
	sink1 := &MockSink{}
	multi := NewMultiSink(sink1)

	entry := models.EntryPool.Get().(*models.LogEntry)
	entry.Reset()
	entry.Message = "Single child sink message"

	err := multi.Write(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if sink1.GetWriteCount() != 1 {
		t.Errorf("Expected 1 write in sink1, got %d", sink1.GetWriteCount())
	}
	if sink1.GetLastEntry().Message != "Single child sink message" {
		t.Errorf("Expected 'Single child sink message', got %s", sink1.GetLastEntry().Message)
	}
}

type errorFailingSink struct {
	err error
}

func (e *errorFailingSink) Write(entry *models.LogEntry) error {
	if entry != nil {
		entry.Release()
	}
	return e.err
}

func (e *errorFailingSink) Close() error {
	return nil
}

func TestMultiSink_ChildError_ContinuesAndReleases(t *testing.T) {
	failSink := &errorFailingSink{err: fmt.Errorf("disk I/O error")}
	okSink := &MockSink{}
	multi := NewMultiSink(failSink, okSink)

	entry := models.EntryPool.Get().(*models.LogEntry)
	entry.Reset()
	entry.Message = "Error resilience message"

	err := multi.Write(entry)
	if err == nil {
		t.Fatal("Expected error from multi.Write when child sink fails, got nil")
	}

	// Verify that the second sink still received the entry
	if okSink.GetWriteCount() != 1 {
		t.Errorf("Expected okSink to receive entry despite previous sink error, got %d writes", okSink.GetWriteCount())
	}
	if okSink.GetLastEntry().Message != "Error resilience message" {
		t.Errorf("Expected okSink message 'Error resilience message', got %s", okSink.GetLastEntry().Message)
	}
}

func TestMultiSink_MixedSyncAndAsync(t *testing.T) {
	syncMock := &MockSink{}
	asyncMock := &MockSink{}
	asyncSink := NewAsyncSink(asyncMock, 100)
	multi := NewMultiSink(syncMock, asyncSink)

	totalLogs := 50
	for i := 0; i < totalLogs; i++ {
		entry := models.EntryPool.Get().(*models.LogEntry)
		entry.Reset()
		entry.Message = fmt.Sprintf("Mixed message %d", i)
		if err := multi.Write(entry); err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
	}

	// Close multi, which closes asyncSink and waits for worker to drain
	if err := multi.Close(); err != nil {
		t.Fatalf("MultiSink.Close failed: %v", err)
	}

	if syncMock.GetWriteCount() != totalLogs {
		t.Errorf("Expected syncMock to receive %d logs, got %d", totalLogs, syncMock.GetWriteCount())
	}
	if asyncMock.GetWriteCount() != totalLogs {
		t.Errorf("Expected asyncMock to receive %d logs, got %d", totalLogs, asyncMock.GetWriteCount())
	}
}

func TestAsyncSink_Close_DrainsAllEnqueuedEntries(t *testing.T) {
	mockSink := &MockSink{}
	asyncSink := NewAsyncSink(mockSink, 100)

	totalLogs := 30
	for i := 0; i < totalLogs; i++ {
		entry := models.EntryPool.Get().(*models.LogEntry)
		entry.Reset()
		entry.Message = fmt.Sprintf("Drain message %d", i)
		if err := asyncSink.Write(entry); err != nil {
			t.Fatalf("Async write %d failed: %v", i, err)
		}
	}

	// Close must wait for worker to process all 30 entries
	if err := asyncSink.Close(); err != nil {
		t.Fatalf("AsyncSink.Close failed: %v", err)
	}

	if mockSink.GetWriteCount() != totalLogs {
		t.Errorf("Expected mockSink to have drained %d logs, got %d", totalLogs, mockSink.GetWriteCount())
	}
}

func TestAsyncSink_Close_DrainTimeout(t *testing.T) {
	blocker := &blockingMockSink{
		unblock: make(chan struct{}),
		started: make(chan struct{}),
	}
	asyncSink := NewAsyncSink(blocker, 5)
	asyncSink.SetDrainTimeout(50 * time.Millisecond)

	entry := models.EntryPool.Get().(*models.LogEntry)
	entry.Reset()
	entry.Message = "Stuck message"
	if err := asyncSink.Write(entry); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	<-blocker.started

	// Close should not hang forever; it should time out after ~50ms
	start := time.Now()
	err := asyncSink.Close()
	elapsed := time.Since(start)

	// Clean up blocker
	close(blocker.unblock)

	if err == nil {
		t.Error("Expected timeout error from AsyncSink.Close(), got nil")
	}
	if elapsed < 30*time.Millisecond || elapsed > 1*time.Second {
		t.Errorf("Expected Close() to return in ~50ms, took %v", elapsed)
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

type mockDropNotifier struct {
	mu       sync.Mutex
	messages []*models.NotifMessage
}

func (m *mockDropNotifier) Notify(n *models.NotifMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, n)
	return nil
}

func (m *mockDropNotifier) Close() error {
	return nil
}

func (m *mockDropNotifier) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

func (m *mockDropNotifier) Last() *models.NotifMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.messages) == 0 {
		return nil
	}
	return m.messages[len(m.messages)-1]
}

func TestAsyncSink_ThrottledDropAlert_PreventsTelegramStorm(t *testing.T) {
	blocker := &blockingMockSink{
		unblock: make(chan struct{}),
		started: make(chan struct{}),
	}
	asyncSink := NewAsyncSink(blocker, 1)
	asyncSink.SetDropAlertInterval(80 * time.Millisecond)

	notif := &mockDropNotifier{}
	asyncSink.SetDropNotifier(notif, "order-gateway")

	var callbackCalls int
	var lastWindowDrops, lastTotalDrops uint64
	asyncSink.SetOnDropAlert(func(droppedInWindow uint64, totalDropped uint64) {
		callbackCalls++
		lastWindowDrops = droppedInWindow
		lastTotalDrops = totalDropped
	})

	defer func() {
		close(blocker.unblock)
		_ = asyncSink.Close()
	}()

	// 1. Occupy worker
	e1 := models.EntryPool.Get().(*models.LogEntry)
	e1.Reset()
	e1.Message = "Active in worker"
	_ = asyncSink.Write(e1)
	<-blocker.started

	// 2. Fill channel buffer (size = 1)
	e2 := models.EntryPool.Get().(*models.LogEntry)
	e2.Reset()
	e2.Message = "Queued in buffer"
	_ = asyncSink.Write(e2)

	// 3. Burst of 10 dropped logs inside the 80ms throttle window:
	// Only the FIRST drop should emit a Telegram notification; the remaining 9 must be suppressed.
	for i := 0; i < 10; i++ {
		e := models.EntryPool.Get().(*models.LogEntry)
		e.Reset()
		e.LoggerName = "order-gateway"
		e.Message = fmt.Sprintf("Overflow %d", i)
		_ = asyncSink.Write(e)
	}

	if got := asyncSink.DroppedCount(); got != 10 {
		t.Fatalf("Expected DroppedCount() == 10, got %d", got)
	}
	if notif.Count() != 1 {
		t.Fatalf("Expected exactly 1 throttled Telegram alert during burst, got %d", notif.Count())
	}
	if callbackCalls != 1 || lastWindowDrops != 1 || lastTotalDrops != 1 {
		t.Errorf("Expected first alert callback (1, 1), got calls=%d (%d, %d)", callbackCalls, lastWindowDrops, lastTotalDrops)
	}

	// 4. Wait for throttle window to expire, then trigger 1 more drop.
	// The second alert must aggregate all 10 drops since the first alert (9 suppressed + 1 new = 10 in window, 11 total).
	time.Sleep(100 * time.Millisecond)
	eLast := models.EntryPool.Get().(*models.LogEntry)
	eLast.Reset()
	eLast.LoggerName = "order-gateway"
	eLast.Message = "Post-window overflow"
	_ = asyncSink.Write(eLast)

	if got := asyncSink.DroppedCount(); got != 11 {
		t.Fatalf("Expected DroppedCount() == 11, got %d", got)
	}
	if notif.Count() != 2 {
		t.Fatalf("Expected 2nd throttled Telegram alert after window expiry, got %d", notif.Count())
	}
	if callbackCalls != 2 || lastWindowDrops != 10 || lastTotalDrops != 11 {
		t.Errorf("Expected second alert callback (10 in window, 11 total), got calls=%d (%d, %d)", callbackCalls, lastWindowDrops, lastTotalDrops)
	}
	lastMsg := notif.Last()
	if lastMsg == nil || len(lastMsg.Tags) == 0 || lastMsg.Tags[0] != "telegram" {
		t.Errorf("Expected Telegram tag on drop alert, got %+v", lastMsg)
	}
}
