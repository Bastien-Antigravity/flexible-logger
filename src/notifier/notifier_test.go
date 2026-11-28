package notifier

// =============================================================================
// ESSENTIAL PROCESS: Unit tests for LocalNotifier and RemoteNotifier under synchronous and asynchronous conditions.
//
// DATA FLOW:
//   1. Tests local channel delivery and queue overflow handling.
//   2. Validates RemoteNotifier background worker initialization and teardown.
//
// KEY PARAMETERS:
//   - t: Testing harness.
// =============================================================================

import (
	"testing"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

func TestLocalNotifier_Notify(t *testing.T) {
	ln := NewLocalNotifier()
	ch := make(chan *models.NotifMessage, 1)
	ln.SetQueue(ch)

	msg := &models.NotifMessage{Message: "Test Notif"}
	err := ln.Notify(msg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	received := <-ch
	if received.Message != "Test Notif" {
		t.Errorf("Expected 'Test Notif', got '%s'", received.Message)
	}
}

func TestLocalNotifier_NoQueue(t *testing.T) {
	ln := NewLocalNotifier()
	msg := &models.NotifMessage{Message: "Test Notif"}
	err := ln.Notify(msg)
	if err == nil {
		t.Error("Expected error when no queue is bound, got nil")
	}
}

func TestRemoteNotifier_Notify(t *testing.T) {
	// RemoteNotifier starts a worker and attempts to connect.
	// For unit testing Notify() without a real server, we can verify it queues correctly.
	ip := "127.0.0.1"
	port := "9999"
	publicIP := "127.0.0.1"

	rn := NewRemoteNotifier(&ip, &port, &publicIP, "test-app")
	// Ensure we close to avoid goroutine leak
	defer rn.Close()

	msg := &models.NotifMessage{Message: "Remote Test"}
	err := rn.Notify(msg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// -----------------------------------------------------------------------------

func TestLocalNotifier_BufferFull_IncrementsDroppedCount(t *testing.T) {
	ln := NewLocalNotifier()
	ch := make(chan *models.NotifMessage, 1)
	ln.SetQueue(ch)

	if err := ln.Notify(&models.NotifMessage{Message: "ok"}); err != nil {
		t.Fatalf("Unexpected error on first Notify: %v", err)
	}
	if err := ln.Notify(&models.NotifMessage{Message: "dropped-1"}); err == nil {
		t.Fatal("Expected error when local queue is full")
	}
	if err := ln.Notify(&models.NotifMessage{Message: "dropped-2"}); err == nil {
		t.Fatal("Expected error when local queue is full")
	}

	if got := ln.DroppedCount(); got != 2 {
		t.Fatalf("Expected DroppedCount() == 2, got %d", got)
	}
}

// -----------------------------------------------------------------------------

func TestRemoteNotifier_BufferFull_IncrementsDroppedCount(t *testing.T) {
	rn := &RemoteNotifier{
		notifChan: make(chan *models.NotifMessage, 1),
	}

	if err := rn.Notify(&models.NotifMessage{Message: "queued"}); err != nil {
		t.Fatalf("Expected first Notify into 1-slot buffer to succeed, got: %v", err)
	}
	if err := rn.Notify(&models.NotifMessage{Message: "overflow"}); err == nil {
		t.Fatal("Expected error when RemoteNotifier buffer is full")
	}
	if got := rn.DroppedCount(); got != 1 {
		t.Fatalf("Expected DroppedCount() == 1 on overflow, got %d", got)
	}
}
