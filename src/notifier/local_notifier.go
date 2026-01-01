package notifier

import (
	"fmt"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// LocalNotifier sends notifications to a local channel instead of a network socket.
type LocalNotifier struct {
	notifChan chan *models.NotifMessage
}

// -----------------------------------------------------------------------------
func NewLocalNotifier() *LocalNotifier {
	return &LocalNotifier{}
}

// -----------------------------------------------------------------------------
func (ln *LocalNotifier) SetQueue(q chan *models.NotifMessage) {
	ln.notifChan = q
}

// -----------------------------------------------------------------------------
func (ln *LocalNotifier) Notify(n *models.NotifMessage) error {
	if ln.notifChan == nil {
		// Drop or error if no queue bound?
		// For now, silent drop or error.
		return fmt.Errorf("no local queue bound")
	}

	select {
	case ln.notifChan <- n:
		return nil
	default:
		return fmt.Errorf("local notification buffer full")
	}
}

// -----------------------------------------------------------------------------
func (ln *LocalNotifier) Close() error {
	// Needed for interface
	// We do not close the channel as it is injected (owned by caller).
	return nil
}
