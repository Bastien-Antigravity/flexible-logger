package notifier

import (
	"fmt"
	"sync"

	"github.com/Bastien-Antigravity/flexible-logger/src/models"
	"github.com/Bastien-Antigravity/flexible-logger/src/network_manager"
	notifie_schema "github.com/Bastien-Antigravity/flexible-logger/src/schemas/notifie_msg"

	capnp "capnproto.org/go/capnp/v3"
)

// -----------------------------------------------------------------------------

// RemoteNotifier sends notifications to a remote govenv NotifServer.
type RemoteNotifier struct {
	ip         *string
	port       *string
	publicIP   *string
	notifChan  chan *models.Notification
	wg         sync.WaitGroup
	netManager *network_manager.NetworkManager
}

// -----------------------------------------------------------------------------

func NewRemoteNotifier(ip, port, publicIP *string) *RemoteNotifier {
	rn := &RemoteNotifier{
		ip:         ip,
		port:       port,
		publicIP:   publicIP,
		notifChan:  make(chan *models.Notification, 1000),
		netManager: network_manager.NewNetworkManager(),
	}
	rn.wg.Add(1)
	go rn.worker()
	return rn
}

// -----------------------------------------------------------------------------

func (rn *RemoteNotifier) Notify(n *models.Notification) error {
	select {
	case rn.notifChan <- n:
		return nil
	default:
		return fmt.Errorf("notification buffer full")
	}
}

// -----------------------------------------------------------------------------

func (rn *RemoteNotifier) Close() error {
	close(rn.notifChan)
	rn.wg.Wait()
	return nil
}

// -----------------------------------------------------------------------------

func (rn *RemoteNotifier) worker() {
	defer rn.wg.Done()

	// Initial Connection
	// Pointers are already verified (by caller ideally, or nil check here if overly cautious)
	// But let's assume valid pointers passed from profile which does the check.

	conn, err := rn.netManager.ConnectWithRetry(rn.ip, rn.port, rn.publicIP, "tcp-hello")
	if err != nil {
		fmt.Printf("RemoteNotifier: Fatal error connecting: %v\n", err)
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
			fmt.Printf("RemoteNotifier: Failed to send notification: %v\n", err)
			// Connection is still valid (it's the managed wrapper), so just continue
		}
	}
}

// -----------------------------------------------------------------------------

// serialize converts Notification to NotifieMsg Cap'n Proto format.
// It uses the locally replicated NotifieMsg schema.
func (rn *RemoteNotifier) serialize(n *models.Notification) []byte {
	// Create a new Cap'n Proto message
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		// Should not happen for new message
		return nil
	}

	// Create Root Struct
	notifMsg, err := notifie_schema.NewRootNotifieMsg(seg)
	if err != nil {
		return nil
	}

	// Set Fields
	_ = notifMsg.SetMessage_(n.Message)
	_ = notifMsg.SetAttachment(n.Attachment)

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
