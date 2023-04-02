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
		// We just want to CloseWrite to signal that we finished writing
		// but naked net.Conn does not have it so we transform to TCPConn
		if cw, ok := conn.(interface{ CloseWrite() error }); ok {
			defer cw.CloseWrite()
		}
		_, err = io.Copy(conn, upstream)
		return err
	})

	if cwu, ok := upstream.(interface{ CloseWrite() error }); ok {
		defer cwu.CloseWrite()
	}
	_, err = io.Copy(upstream, conn)
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}
