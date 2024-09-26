// Package proxy implements various proxy related utilities
package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"

	"darvaza.org/core"
)

// CloseWriter represents a connection that can close its Write stream
type CloseWriter interface {
	CloseWrite() error
}

// Forward will take a context, a "downstream" net.Conn and a netip.Addr it will
// create a new connection "upstream" and it will move bytes between the two.
// Practically it will proxy between the two connections
func Forward(ctx context.Context, conn net.Conn, addr netip.AddrPort) error {
	defer conn.Close()
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	if !addr.IsValid() {
		return fmt.Errorf("invalid upstream address")
	}

	upstream, err := net.Dial("tcp", addr.String())
	if err != nil {
		// TODO: Need to retry?
		return err
	}
	defer upstream.Close()

	var wg core.WaitGroup

	wg.Go(func() error {
		return copyConn(conn, upstream)
	})

	if err := copyConn(upstream, conn); err != nil {
		return err
	}

	_ = wg.Wait()
	return nil
}

func copyConn(from, to net.Conn) error {
	// We just want to CloseWrite to signal that we finished writing
	// but naked net.Conn does not have it so we transform to TCPConn
	if w, ok := from.(CloseWriter); ok {
		defer func() {
			_ = w.CloseWrite()
		}()
	}

	_, err := io.Copy(from, to)
	return err
}
