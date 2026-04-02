package test_utils

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// -----------------------------------------------------------------------------
// StartMockServer starts a TCP listener on a random free port and discards all data.
// It returns the assigned address (host:port) and a cleanup function to stop the server.
func StartMockServer(name string) (ip, port string, stop func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("%s: Failed to listen: %v", name, err))
	}

	addr := ln.Addr().String()
	parts := strings.Split(addr, ":")
	ip = parts[0]
	port = parts[1]

	var wg sync.WaitGroup
	quit := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-quit:
					return
				default:
					fmt.Printf("%s: Accept error: %v\n", name, err)
					continue
				}
			}

			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer c.Close()
				// Discard all incoming data
				_, _ = io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	stop = func() {
		close(quit)
		ln.Close()
		wg.Wait()
	}

	return ip, port, stop
}
