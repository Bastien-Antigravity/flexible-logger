package models

import (
	"sync/atomic"
	"testing"
)

func TestLogEntry_Reset(t *testing.T) {
	e := &LogEntry{
		Message:    "dirty",
		LoggerName: "dirty",
	}
	e.refCount = 10

	e.Reset()

	if e.Message != "" {
		t.Errorf("Expected empty Message, got %s", e.Message)
	}
	if atomic.LoadInt32(&e.refCount) != 1 {
		t.Errorf("Expected refCount 1, got %d", atomic.LoadInt32(&e.refCount))
	}
}

func TestLogEntry_RetainRelease(t *testing.T) {
	e := EntryPool.Get().(*LogEntry)
	e.Reset()

	if atomic.LoadInt32(&e.refCount) != 1 {
		t.Errorf("Initial refCount should be 1, got %d", atomic.LoadInt32(&e.refCount))
	}

	e.Retain()
	if atomic.LoadInt32(&e.refCount) != 2 {
		t.Errorf("After Retain, refCount should be 2, got %d", atomic.LoadInt32(&e.refCount))
	}

	e.Release()
	if atomic.LoadInt32(&e.refCount) != 1 {
		t.Errorf("After first Release, refCount should be 1, got %d", atomic.LoadInt32(&e.refCount))
	}

	// Final release returns it to pool
	e.Release()
	if atomic.LoadInt32(&e.refCount) != 0 {
		t.Errorf("After final Release, refCount should be 0, got %d", atomic.LoadInt32(&e.refCount))
	}
}

func TestEntryPool_Reuse(t *testing.T) {
	// Get an entry, modify it, release it, then get another and check if it's clean
	e1 := EntryPool.Get().(*LogEntry)
	e1.Reset()
	e1.Message = "entry 1"
	e1.Release()

	// sync.Pool doesn't guarantee reuse, but for testing purposes in a single thread it usually does
	e2 := EntryPool.Get().(*LogEntry)
	e2.Reset()
	if e2.Message != "" {
		t.Errorf("Expected fresh entry from pool to be clean after Reset, but got Message: %s", e2.Message)
	}
	e2.Release()
}
