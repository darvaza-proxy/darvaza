package bind

import "net"

// Listener mimics net.ListenConfig providing a configuration
// context for net.Listen() and net.ListenPacket() alternatives
type Listener interface {
	Listen(network, addr string) (net.Listener, error)
	ListenPacket(network, addr string) (net.PacketConn, error)
}

// AllListener is equivalent to Listener but takes an
// array of addresses
type AllListener interface {
	ListenAll(network string, addr []string) ([]net.Listener, error)
	ListenAllPacket(network string, addr []string) ([]net.PacketConn, error)
}

// TCPListener provides a context-aware alternative to net.ListenTCP
type TCPListener interface {
	ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
}

// AllTCPListener is equivalent to TCPListener but takes an
// array of addresses
type AllTCPListener interface {
	ListenAllTCP(network string, ladders []*net.TCPAddr) ([]*net.TCPListener, error)
}

// UDPListener provides a context-aware alternative to net.ListenUDP
type UDPListener interface {
	ListenUDP(network string, laddr *net.UDPAddr) (*net.UDPConn, error)
}

// AllUDPListener is equivalent to UDPListener but takes an
// array of addresses
type AllUDPListener interface {
	ListenAllUDP(network string, ladders []*net.UDPAddr) ([]*net.UDPConn, error)
}

// TCPUDPListener provides the callbacks used by Bind().
// ListenTCP() and ListenUDP()
type TCPUDPListener interface {
	TCPListener
	UDPListener
}
