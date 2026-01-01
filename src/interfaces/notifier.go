package interfaces

import "github.com/Bastien-Antigravity/flexible-logger/src/models"

// -----------------------------------------------------------------------------
// Notifier defines a component capable of sending notifications
type Notifier interface {
	// -------------------------------------------------------------------------
	// Notify sends a notification.
	Notify(n *models.Notification) error

	// -------------------------------------------------------------------------
	// Close closes the notifier.
	Close() error
}
