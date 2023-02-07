package net

import (
	"net"
)

// ListenUpgrader represents a tool that keep account of listening ports
// but allows us to provide our own helper to do the last step
type ListenUpgrader interface {
	ListenWithCallback(network, addr string,
		callback func(network, addr string) (net.Listener, error)) (net.Listener, error)
	ListenPacketWithCallback(network, addr string,
		callback func(netowkr, addr string) (net.PacketConn, error)) (net.PacketConn, error)
}

// ListenUpgraderConfig represents an object equivalent to a ListenConfig but
// using a ListenUpgrader in the middle
type ListenUpgraderConfig struct {
	conf ListenConfig
	upg  ListenUpgrader
}

// WithUpgrader creates a new ListenConfig using the provided ListenUpgrader
func (lc ListenConfig) WithUpgrader(upg ListenUpgrader) *ListenUpgraderConfig {
	return &ListenUpgraderConfig{
		conf: lc,
		upg:  upg,
	}
}

// Listen acts like the standard net.Listen but using our ListenConfig and via
// an Upgrader tool
func (lu ListenUpgraderConfig) Listen(network, addr string) (net.Listener, error) {
	return lu.upg.ListenWithCallback(network, addr, lu.conf.Listen)
}

// ListenPacket acts like the standard net.ListenPacket but using our
// ListenConfig and via an Upgrader tool
func (lu ListenUpgraderConfig) ListenPacket(network, addr string) (net.PacketConn, error) {
	return lu.upg.ListenPacketWithCallback(network, addr, lu.conf.ListenPacket)
}

// ListenTCP acts like the standard net.ListenTCP but using our ListenConfig and
// the ListenUpgrader
func (lu ListenUpgraderConfig) ListenTCP(network string, laddr *net.TCPAddr) (
	*net.TCPListener, error) {
	if laddr == nil {
		laddr = &net.TCPAddr{}
	}

	ln, err := lu.Listen(network, laddr.String())
	if err != nil {
		return nil, err
	}
	return ln.(*net.TCPListener), nil
}

// ListenUDP acts like the standard net.ListenUDP but using our ListenConfig and
// the ListenUpgrader
func (lu ListenUpgraderConfig) ListenUDP(network string, laddr *net.UDPAddr) (
	*net.UDPConn, error) {
	if laddr == nil {
		laddr = &net.UDPAddr{}
	}

	ln, err := lu.ListenPacket(network, laddr.String())
	if err != nil {
		return nil, err
	}
	return ln.(*net.UDPConn), nil
}

// ListenAll acts like Listen but on a list of addresses
func (lu ListenUpgraderConfig) ListenAll(network string, addrs []string) ([]net.Listener, error) {
	out := make([]net.Listener, 0, len(addrs))

	for _, addr := range addrs {
		lsn, err := lu.Listen(network, addr)
		if err != nil {
			for _, lsn := range out {
				_ = lsn.Close()
			}
			return nil, err
		}
		out = append(out, lsn)
	}

	return out, nil
}

// ListenAllPacket acts like ListenPacket but on a list of addresses
func (lu ListenUpgraderConfig) ListenAllPacket(network string, addrs []string) (
	[]net.PacketConn, error) {
	out := make([]net.PacketConn, 0, len(addrs))

	for _, addr := range addrs {
		lsn, err := lu.ListenPacket(network, addr)
		if err != nil {
			for _, lsn := range out {
				_ = lsn.Close()
			}
			return nil, err
		}
		out = append(out, lsn)
	}

	return out, nil
}

// ListenAllTCP acts like ListenTCP but on a list of addresses
func (lu ListenUpgraderConfig) ListenAllTCP(network string, laddrs []*net.TCPAddr) (
	[]*net.TCPListener, error) {
	out := make([]*net.TCPListener, 0, len(laddrs))

	for _, addr := range laddrs {
		lsn, err := lu.ListenTCP(network, addr)
		if err != nil {
			for _, lsn := range out {
				_ = lsn.Close()
			}
			return nil, err
		}
		out = append(out, lsn)
	}

	return out, nil
}

// ListenAllUDP acts like ListenUDP but on a list of addresses
func (lu ListenUpgraderConfig) ListenAllUDP(network string, laddrs []*net.UDPAddr) (
	[]*net.UDPConn, error) {
	out := make([]*net.UDPConn, 0, len(laddrs))

	for _, addr := range laddrs {
		lsn, err := lu.ListenUDP(network, addr)
		if err != nil {
			for _, lsn := range out {
				_ = lsn.Close()
			}
			return nil, err
		}
		out = append(out, lsn)
	}

	return out, nil
}
