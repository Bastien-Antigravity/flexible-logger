package notifier

// =============================================================================
// ESSENTIAL PROCESS: In-process notification dispatcher routing alerts to local Go channels for reactive event handling.
//
// DATA FLOW:
//   1. Accepts incoming NotifMessage from LogEngine.
//   2. Performs non-blocking delivery to subscriber channel.
//   3. Logs diagnostic warning if channel is full or unassigned.
//
// KEY PARAMETERS:
//   - queue: User-provided channel receiving notifications.
//   - dropped: Atomic counter tracking overflow drops.
// =============================================================================

import (
	"fmt"
	"sync"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------

// LocalNotifier sends notifications to a local channel instead of a network socket.
type LocalNotifier struct {
	mu        sync.RWMutex
	notifChan chan *models.NotifMessage
}

// -----------------------------------------------------------------------------

func NewLocalNotifier() *LocalNotifier {
	return &LocalNotifier{}
}

// -----------------------------------------------------------------------------

func (ln *LocalNotifier) SetQueue(q chan *models.NotifMessage) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.notifChan = q
}

// -----------------------------------------------------------------------------

func (ln *LocalNotifier) Notify(n *models.NotifMessage) error {
	ln.mu.RLock()
	ch := ln.notifChan
	ln.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("no local queue bound")
	}

	select {
	case ch <- n:
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
