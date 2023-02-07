package net

import "net"

// ListenerConfig mimicks net.ListenConfig providing a configuration
// context for net.Listen() and net.ListenPacket() alternatives
type ListenerConfig interface {
	Listen(network, addr string) (net.Listener, error)
	ListenPacket(network, addr string) (net.PacketConn, error)
}

// AllListenerConfig is equivalent to ListenerConfig but takes an
// array of addresses
type AllListenerConfig interface {
	ListenAll(network string, addr []string) ([]net.Listener, error)
	ListenAllPacket(network string, addr []string) ([]net.PacketConn, error)
}

// TCPListenerConfig provides a context-aware alternative to net.ListenTCP
type TCPListenerConfig interface {
	ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
}

// AllTCPListenerConfig is equivalent to TCPListenerConfig but takes an
// array of addresses
type AllTCPListenerConfig interface {
	ListenAllTCP(network string, ladders []*net.TCPAddr) ([]*net.TCPListener, error)
}

// UDPListenerConfig provides a context-aware alternative to net.ListenUDP
type UDPListenerConfig interface {
	ListenUDP(network string, laddr *net.UDPAddr) (*net.UDPConn, error)
}

// AllUDPListenerConfig is equivalent to UDPListenerConfig but takes an
// array of addresses
type AllUDPListenerConfig interface {
	ListenAllUDP(network string, ladders []*net.UDPAddr) ([]*net.UDPConn, error)
}
