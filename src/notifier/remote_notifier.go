package notifier

// =============================================================================
// ESSENTIAL PROCESS: Asynchronous Cap'n Proto notification dispatcher streaming alerts over SafeSocket TCP to notif-server.
//
// DATA FLOW:
//   1. Queues NotifMessage instances into internal buffered channel.
//   2. Background worker serializes message using Cap'n Proto NotifierMsg schema.
//   3. Transmits binary frames over managed SafeSocket connection.
//
// KEY PARAMETERS:
//   - notifChan: Internal channel buffering alert messages.
//   - netManager: Connection manager handling exponential backoff reconnects.
//   - connReady: Synchronization channel signaling connection initialization.
// =============================================================================

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Bastien-Antigravity/flexible-logger/src/error_handler"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	notifier_schema "github.com/Bastien-Antigravity/flexible-logger/src/schemas/capnp/notifier"
	"github.com/Bastien-Antigravity/microservice-toolbox/go/pkg/conn_manager"

	capnp "capnproto.org/go/capnp/v3"
)

// -----------------------------------------------------------------------------

// RemoteNotifier sends notifications to a remote govenv NotifServer.
type RemoteNotifier struct {
	ip         *string
	port       *string
	publicIP   *string
	appName    string
	notifChan  chan *models.NotifMessage
	wg         sync.WaitGroup
	netManager *conn_manager.NetworkManager
	conn       io.WriteCloser // stored reference for clean shutdown
	connReady  chan struct{}  // signals that conn has been initialized
	closeOnce  sync.Once
}

// -----------------------------------------------------------------------------

func NewRemoteNotifier(ip, port, publicIP *string, appName string) *RemoteNotifier {
	// We use the Performance Strategy for notifications as they should be non-blocking
	nm := conn_manager.NewPerformanceStrategy(nil)
	nm.OnError = func(attempt int, err error, source string, msg string) {
		error_handler.ReportInternalError("RemoteNotifier", source, err, msg)
	}

	rn := &RemoteNotifier{
		ip:         ip,
		port:       port,
		publicIP:   publicIP,
		appName:    appName,
		notifChan:  make(chan *models.NotifMessage, 1000),
		netManager: nm,
		connReady:  make(chan struct{}),
	}
	rn.wg.Add(1)
	go rn.worker()
	return rn
}

// -----------------------------------------------------------------------------

func (rn *RemoteNotifier) Notify(n *models.NotifMessage) error {
	select {
	case rn.notifChan <- n:
		return nil
	default:
		return fmt.Errorf("notification buffer full")
	}
}

// -----------------------------------------------------------------------------

func (rn *RemoteNotifier) Close() error {
	rn.closeOnce.Do(func() {
		close(rn.notifChan)
		<-rn.connReady

		// Allow worker to drain remaining queued notifications
		drained := make(chan struct{})
		go func() {
			rn.wg.Wait()
			close(drained)
		}()

		select {
		case <-drained:
			// Drained cleanly
		case <-time.After(2 * time.Second):
			// Timeout reached while draining; force close connection to break worker
			if rn.conn != nil {
				rn.conn.Close()
			}
			<-drained
		}

		if rn.conn != nil {
			rn.conn.Close()
		}
	})
	return nil
}

// -----------------------------------------------------------------------------

func (rn *RemoteNotifier) worker() {
	defer rn.wg.Done()

	conn := rn.netManager.Connect(rn.ip, rn.port, rn.publicIP, "tcp-hello:"+rn.appName, conn_manager.ModeNonBlocking)

	// Store reference and signal readiness for Close() to use.
	// This must happen even if conn is nil, to prevent Close() from blocking on <-connReady.
	rn.conn = conn
	close(rn.connReady)

	if conn == nil {
		error_handler.ReportInternalError("RemoteNotifier", "worker.connect", fmt.Errorf("failed to initialize connection"), "")
		return
	}
	defer conn.Close()

	for n := range rn.notifChan {
		data := rn.serialize(n)
		if data == nil {
			continue
		}

		// ManagedConnection handles reconnection internally
		_, err := conn.Write(data)
		if err != nil {
			error_handler.ReportInternalError("RemoteNotifier", "worker.write", err, n.Message)
		}
	}
}

// -----------------------------------------------------------------------------

// serialize converts Notification to NotifierMsg Cap'n Proto format.
// It uses the locally replicated NotifierMsg schema.
func (rn *RemoteNotifier) serialize(n *models.NotifMessage) []byte {
	// Create a new Cap'n Proto message
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		// Should not happen for new message
		return nil
	}

	// Create Root Struct
	notifMsg, err := notifier_schema.NewRootNotifierMsg(seg)
	if err != nil {
		return nil
	}

	// Set Fields
	_ = notifMsg.SetMessage_(n.Message)
	_ = notifMsg.SetAttachment(n.Attachment)
	_ = notifMsg.SetLevel(n.Level)

	// Set Tags
	if len(n.Tags) > 0 {
		tagList, err := notifMsg.NewTags(int32(len(n.Tags)))
		if err == nil {
			for i, t := range n.Tags {
				_ = tagList.Set(i, t)
			}
		}
	}

	// Marshal to bytes
	data, err := msg.MarshalPacked()
	if err != nil {
		return nil
	}
	return data
}
