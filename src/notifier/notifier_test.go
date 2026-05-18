package notifier

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
