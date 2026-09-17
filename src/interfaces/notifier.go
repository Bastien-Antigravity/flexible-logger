package interfaces

// =============================================================================
// ESSENTIAL PROCESS: Interface contract for multi-channel alert and notification dispatchers.
//
// DATA FLOW:
//   1. Defines Notify method accepting NotifMessage models.
//   2. Specifies Close method for graceful flush and channel cleanup.
//
// KEY PARAMETERS:
//   - Notifier: Contract implemented by LocalNotifier and RemoteNotifier.
// =============================================================================

import "github.com/Bastien-Antigravity/flexible-logger/src/models"

// -----------------------------------------------------------------------------
// Notifier defines a component capable of sending notifications
type Notifier interface {
	// -------------------------------------------------------------------------
	// Notify sends a notification.
	Notify(n *models.NotifMessage) error

	// -------------------------------------------------------------------------
	// Close closes the notifier.
	Close() error
}
