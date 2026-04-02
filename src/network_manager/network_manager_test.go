package network_manager

import (
	"testing"
)

func TestNewNetworkManager(t *testing.T) {
	nm := NewNetworkManager()
	if nm == nil {
		t.Fatal("Expected NewNetworkManager to return a manager instance")
	}
	if nm.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", nm.MaxRetries)
	}
}
