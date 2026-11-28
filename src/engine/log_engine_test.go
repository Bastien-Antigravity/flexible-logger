package engine

// =============================================================================
// ESSENTIAL PROCESS: Unit tests validating log level filtering, dynamic caller metadata enrichment, sampling, and sink lifecycle in LogEngine.
//
// DATA FLOW:
//   1. Constructs mock sinks and test LogEngine instances.
//   2. Executes logging methods across all severity levels.
//   3. Asserts level suppression, caller frame capture, and object recycling.
//
// KEY PARAMETERS:
//   - t: Testing harness pointer.
//   - mockSink: Test sink capturing emitted LogEntry instances.
// =============================================================================

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/sink"
)

// MockSink for testing
type MockSink struct {
	LastEntry     *models.LogEntry
	WriteCount    int
	Closed        bool
	ErrorToReturn error
	mu            sync.Mutex
}

func (m *MockSink) Write(entry *models.LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ErrorToReturn != nil {
		if entry != nil {
			entry.Release()
		}
		return m.ErrorToReturn
	}
	m.WriteCount++
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

// MockNotifier for testing
type MockNotifier struct {
	LastMsg       *models.NotifMessage
	NotifyCount   int
	Closed        bool
	ErrorToReturn error
	mu            sync.Mutex
}

func (m *MockNotifier) Notify(msg *models.NotifMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ErrorToReturn != nil {
		return m.ErrorToReturn
	}
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
	if count < 4400 || count > 5600 {
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

func TestLogEngine_NotifierError_DoesNotBlockSink(t *testing.T) {
	mockSink := &MockSink{}
	mockNotif := &MockNotifier{ErrorToReturn: fmt.Errorf("notifier error")}
	engine := &LogEngine{
		Sink:         mockSink,
		Notifier:     mockNotif,
		Level:        models.LevelInfo,
		SamplingRate: 1.0,
	}

	engine.Error("Test error")

	if mockSink.GetWriteCount() != 1 {
		t.Errorf("Expected log to be written to sink even if notifier fails, got %d writes", mockSink.GetWriteCount())
	}
}

func TestLogEngine_ErrorCollectsCallerEvenWhenDisabled(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:              mockSink,
		Level:             models.LevelInfo,
		CollectCallerInfo: false, // Disabled for general logs
		SamplingRate:      1.0,
	}

	engine.Error("Error level log")
	entry := mockSink.GetLastEntry()

	if entry.Filename == "source-context" {
		t.Error("Expected real filename for Error level even when CollectCallerInfo is false")
	}
	if entry.Filename == "" {
		t.Error("Expected non-empty filename")
	}
}

func emitLogAndHelperFromSameFunction(engine *LogEngine) (directEntry models.LogEntry, helperEntry models.LogEntry) {
	mockSink := engine.Sink.(*MockSink)
	engine.Log(models.LevelError, "Direct Log() call")
	directEntry = *mockSink.GetLastEntry()

	engine.Error("Helper Error() call")
	helperEntry = *mockSink.GetLastEntry()
	return directEntry, helperEntry
}

func TestLogEngine_DirectLogVsHelper_IdenticalCallerAttribution(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:              mockSink,
		Level:             models.LevelDebug,
		CollectCallerInfo: true,
		SamplingRate:      1.0,
	}

	directEntry, helperEntry := emitLogAndHelperFromSameFunction(engine)

	if directEntry.Filename != "log_engine_test.go" {
		t.Errorf("Expected direct Log() Filename 'log_engine_test.go', got %q", directEntry.Filename)
	}
	if directEntry.Filename != helperEntry.Filename {
		t.Errorf("Filename mismatch between direct Log() (%q) and helper Error() (%q)", directEntry.Filename, helperEntry.Filename)
	}
	if directEntry.FunctionName != helperEntry.FunctionName {
		t.Errorf("FunctionName mismatch between direct Log() (%q) and helper Error() (%q)", directEntry.FunctionName, helperEntry.FunctionName)
	}
	if directEntry.FunctionName != "engine.emitLogAndHelperFromSameFunction" {
		t.Errorf("Expected caller FunctionName 'engine.emitLogAndHelperFromSameFunction', got %q", directEntry.FunctionName)
	}
}

// sampleWrapperFacade simulates a facade library (like universal-logger) that wraps LogEngine.
func sampleWrapperFacade(eng *LogEngine, msg string) {
	eng.Info("%s", msg)
}

func TestLogEngine_SetCallerSkip_WrapperAttribution(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:              mockSink,
		Level:             models.LevelDebug,
		CollectCallerInfo: true,
		SamplingRate:      1.0,
	}

	// 1. Without caller skip: resolves inside sampleWrapperFacade
	sampleWrapperFacade(engine, "Without skip")
	entryWithoutSkip := mockSink.GetLastEntry()
	if entryWithoutSkip.FunctionName != "engine.sampleWrapperFacade" {
		t.Errorf("Expected caller 'engine.sampleWrapperFacade', got %q", entryWithoutSkip.FunctionName)
	}

	// 2. With CallerSkip = 1: skips the wrapper and resolves to TestLogEngine_SetCallerSkip_WrapperAttribution
	engine.SetCallerSkip(1)
	sampleWrapperFacade(engine, "With skip=1")
	entryWithSkip := mockSink.GetLastEntry()
	if entryWithSkip.FunctionName != "engine.TestLogEngine_SetCallerSkip_WrapperAttribution" {
		t.Errorf("Expected caller 'engine.TestLogEngine_SetCallerSkip_WrapperAttribution', got %q", entryWithSkip.FunctionName)
	}
}

func TestLogEngine_AllSeverityHelpersAndGetLevel(t *testing.T) {
	mockSink := &MockSink{}
	mockNotif := &MockNotifier{}
	eng := &LogEngine{
		Sink:         mockSink,
		Notifier:     mockNotif,
		Level:        models.LevelInfo,
		Name:         "SeverityTest",
		SamplingRate: 1.0,
	}

	if got := eng.GetLevel(); got != models.LevelInfo {
		t.Fatalf("Expected initial GetLevel() == LevelInfo, got %v", got)
	}
	eng.SetLevel(models.LevelDebug)
	if got := eng.GetLevel(); got != models.LevelDebug {
		t.Fatalf("Expected updated GetLevel() == LevelDebug, got %v", got)
	}

	cases := []struct {
		name        string
		call        func(string, ...any)
		wantLevel   models.Level
		expectNotif bool
	}{
		{"Debug", eng.Debug, models.LevelDebug, false},
		{"Stream", eng.Stream, models.LevelStream, false},
		{"Info", eng.Info, models.LevelInfo, false},
		{"Logon", eng.Logon, models.LevelLogon, false},
		{"Logout", eng.Logout, models.LevelLogout, false},
		{"Trade", eng.Trade, models.LevelTrade, false},
		{"Schedule", eng.Schedule, models.LevelSchedule, false},
		{"Report", eng.Report, models.LevelReport, false},
		{"Warning", eng.Warning, models.LevelWarning, true},
		{"Error", eng.Error, models.LevelError, true},
		{"Critical", eng.Critical, models.LevelCritical, true},
	}

	wantNotifs := 0
	for i, tc := range cases {
		tc.call("helper-%s-%d", tc.name, i)
		entry := mockSink.GetLastEntry()
		if entry == nil || entry.Level != tc.wantLevel {
			t.Fatalf("%s: expected level %v, got %+v", tc.name, tc.wantLevel, entry)
		}
		if tc.expectNotif {
			wantNotifs++
		}
	}

	if mockSink.GetWriteCount() != len(cases) {
		t.Fatalf("Expected %d sink writes, got %d", len(cases), mockSink.GetWriteCount())
	}
	if mockNotif.GetNotifyCount() != wantNotifs {
		t.Fatalf("Expected %d notifications (Warning/Error/Critical), got %d", wantNotifs, mockNotif.GetNotifyCount())
	}
}

func TestLogEngine_SetSamplingRate_DynamicUpdate(t *testing.T) {
	mockSink := &MockSink{}
	eng := &LogEngine{
		Sink:         mockSink,
		Level:        models.LevelDebug,
		SamplingRate: 1.0,
	}

	eng.SetSamplingRate(0.01)
	for i := 0; i < 500; i++ {
		eng.Info("sampled-%d", i)
	}
	sampledWrites := mockSink.GetWriteCount()
	if sampledWrites >= 50 {
		t.Fatalf("Expected < 50 Info writes with SetSamplingRate(0.01), got %d", sampledWrites)
	}

	beforeCritical := mockSink.GetWriteCount()
	for i := 0; i < 25; i++ {
		eng.Error("err-%d", i)
		eng.Critical("crit-%d", i)
	}
	if got := mockSink.GetWriteCount() - beforeCritical; got != 50 {
		t.Fatalf("Expected all 50 Error/Critical logs bypass sampling, got %d", got)
	}

	eng.SetSamplingRate(1.0)
	beforeFull := mockSink.GetWriteCount()
	for i := 0; i < 50; i++ {
		eng.Info("full-%d", i)
	}
	if got := mockSink.GetWriteCount() - beforeFull; got != 50 {
		t.Fatalf("Expected all 50 Info writes after SetSamplingRate(1.0), got %d", got)
	}

	// Backward-compat: rate == 0.0 behaves as unset (100% pass-through)
	eng.SetSamplingRate(0.0)
	beforeZero := mockSink.GetWriteCount()
	for i := 0; i < 50; i++ {
		eng.Info("zero-%d", i)
	}
	if got := mockSink.GetWriteCount() - beforeZero; got != 50 {
		t.Fatalf("Expected all 50 Info writes under backward-compat SetSamplingRate(0.0), got %d", got)
	}
}

func TestLogEngine_BackpressureAndUnboundQueue_SuppressStderr(t *testing.T) {
	multiOnlyFull := fmt.Sprintf("multisink write errors: %s; %s", sink.ErrBufferFull.Error(), sink.ErrBufferFull.Error())
	multiMixed := fmt.Sprintf("multisink write errors: %s; disk IO error", sink.ErrBufferFull.Error())

	if !isOnlyBufferFullErr(sink.ErrBufferFull) {
		t.Fatal("Expected isOnlyBufferFullErr(sink.ErrBufferFull) == true")
	}
	if !isOnlyBufferFullErr(errors.New("multisink write errors: " + sink.ErrBufferFull.Error())) {
		t.Fatal("Expected single MultiSink ErrBufferFull error to be suppressed")
	}
	if !isOnlyBufferFullErr(errors.New(multiOnlyFull)) {
		t.Fatal("Expected multiple MultiSink ErrBufferFull errors to be suppressed")
	}
	if isOnlyBufferFullErr(errors.New(multiMixed)) {
		t.Fatal("Expected mixed MultiSink error containing disk IO failure NOT to be suppressed")
	}
	if isOnlyBufferFullErr(errors.New("disk IO error")) {
		t.Fatal("Expected unrelated error NOT to be suppressed")
	}

	captureStderr := func(fn func()) string {
		oldStderr := os.Stderr
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Failed to create stderr pipe: %v", err)
		}
		os.Stderr = w
		fn()
		_ = w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		return buf.String()
	}

	// 1. Expected ErrBufferFull & unbound local queue errors must NOT emit INTERNAL ERROR to stderr
	mockSink := &MockSink{ErrorToReturn: sink.ErrBufferFull}
	mockNotif := &MockNotifier{ErrorToReturn: errors.New("no local queue bound")}
	eng := &LogEngine{
		Name:         "SuppressTest",
		Sink:         mockSink,
		Notifier:     mockNotif,
		Level:        models.LevelDebug,
		SamplingRate: 1.0,
	}

	suppressedOut := captureStderr(func() {
		eng.Error("backpressure event")
	})
	if strings.Contains(suppressedOut, "INTERNAL ERROR") {
		t.Fatalf("Expected no INTERNAL ERROR output for expected backpressure/unbound queue, got: %q", suppressedOut)
	}

	// 2. Genuine sink or notifier errors MUST emit INTERNAL ERROR to stderr
	mockSink.ErrorToReturn = errors.New("disk write failed")
	mockNotif.ErrorToReturn = errors.New("notifier socket failure")
	genuineOut := captureStderr(func() {
		eng.Error("real failure")
	})
	if !strings.Contains(genuineOut, "disk write failed") || !strings.Contains(genuineOut, "notifier socket failure") {
		t.Fatalf("Expected real sink and notifier errors on stderr, got: %q", genuineOut)
	}
}

func TestLogEngine_ConcurrentStress(t *testing.T) {
	mockSink := &MockSink{}
	engine := &LogEngine{
		Sink:         mockSink,
		Level:        models.LevelDebug,
		SamplingRate: 1.0,
	}

	numGoroutines := 100
	numLogsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines + 4)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numLogsPerGoroutine; j++ {
				engine.Info("Log from %d: %d", id, j)
			}
		}(i)
	}

	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			engine.SetLevel(models.LevelDebug)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			engine.SetCallerSkip(i % 2)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			engine.SetSamplingRate(1.0)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = engine.GetLevel()
		}
	}()

	wg.Wait()

	expected := numGoroutines * numLogsPerGoroutine
	if mockSink.GetWriteCount() != expected {
		t.Errorf("Expected %d logs, got %d", expected, mockSink.GetWriteCount())
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
