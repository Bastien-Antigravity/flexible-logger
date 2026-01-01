package network_manager

import (
	"fmt"
	"io"
	"math"
	"time"

	safesocket "github.com/Bastien-Antigravity/safe-socket"
)

// -----------------------------------------------------------------------------
// NetworkManager handles reliable connection establishment with retries.
type NetworkManager struct {
	MaxRetries     int
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	ConnectTimeout time.Duration
}

// -----------------------------------------------------------------------------
// NewNetworkManager creates a manager with default retry policies.
func NewNetworkManager() *NetworkManager {
	return &NetworkManager{
		MaxRetries:     5,
		BaseDelay:      200 * time.Millisecond,
		MaxDelay:       5 * time.Second,
		ConnectTimeout: 2 * time.Second,
	}
}

// -----------------------------------------------------------------------------
// EstablishConnection attempts a single connection to the resolved address.
func (nm *NetworkManager) EstablishConnection(ip, port, publicIP *string, profile string) (io.WriteCloser, error) {
	address := fmt.Sprintf("%s:%s", *ip, *port)
	return safesocket.Create(profile, address, *publicIP, "client", true)
}

// -----------------------------------------------------------------------------
// ConnectWithRetry attempts to connect and returns a ManagedConnection.
// If initial connection fails, it still returns a ManagedConnection (wrapper)
// but with nil internal connection, which will auto-connect on first Write.
func (nm *NetworkManager) ConnectWithRetry(ip, port, publicIP *string, profile string) (io.WriteCloser, error) {
	mc := &ManagedConnection{
		ip:       ip,
		port:     port,
		publicIP: publicIP,
		profile:  profile,
		nm:       nm,
	}

	// Try initial connection
	address := fmt.Sprintf("%s:%s", *ip, *port) // Keep address for logging
	var err error
	for i := 0; i < nm.MaxRetries; i++ {
		conn, err := nm.EstablishConnection(ip, port, publicIP, profile)
		if err == nil {
			mc.currentConn = conn
			return mc, nil
		}

		delay := float64(nm.BaseDelay) * math.Pow(2, float64(i))
		if delay > float64(nm.MaxDelay) {
			delay = float64(nm.MaxDelay)
		}
		fmt.Printf("ManagedConnection: Initial connection to %s failed: %v. Retrying in %v...\n", address, err, time.Duration(delay))
		time.Sleep(time.Duration(delay))
		address = fmt.Sprintf("%s:%s", *ip, *port) // Update addr
	}

	// Previously we returned error. Now we return the wrapper so it can auto-connect later.
	// However, existing contract expects error on failure.
	// To minimize breakage, we return error if initial connection fails after retries,
	// BUT we could return the wrapper if we wanted "start disconnected" support.
	// Let's stick to returning error for strict startup behavior, BUT return the wrapper
	// if successful.

	return nil, fmt.Errorf("failed to connect to %s after %d attempts: %w", address, nm.MaxRetries, err)
}

// -----------------------------------------------------------------------------
// ConnectBlocking indefinitely retries connection until successful and returns ManagedConnection.
func (nm *NetworkManager) ConnectBlocking(ip, port, publicIP *string, profile string) io.WriteCloser {
	mc := &ManagedConnection{
		ip:       ip,
		port:     port,
		publicIP: publicIP,
		profile:  profile,
		nm:       nm,
	}

	// Use internal reconnect logic to establish initial connection
	mc.reconnect()
	return mc
}
