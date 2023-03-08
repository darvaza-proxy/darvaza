package bind

import (
	"net"
)

var (
	_ Listener       = (*UpgraderListenConfig)(nil)
	_ AllListener    = (*UpgraderListenConfig)(nil)
	_ TCPListener    = (*UpgraderListenConfig)(nil)
	_ UDPListener    = (*UpgraderListenConfig)(nil)
	_ AllTCPListener = (*UpgraderListenConfig)(nil)
	_ AllUDPListener = (*UpgraderListenConfig)(nil)
)

// Upgrader represents a tool that keep account of listening ports
// but allows us to provide our own helper to do the last step
type Upgrader interface {
	ListenWithCallback(network, addr string,
		callback func(network, addr string) (net.Listener, error)) (net.Listener, error)
	ListenPacketWithCallback(network, addr string,
		callback func(netowkr, addr string) (net.PacketConn, error)) (net.PacketConn, error)
}

// UpgraderListenConfig represents an object equivalent to a ListenConfig but
// using a Upgrader in the middle
type UpgraderListenConfig struct {
	conf ListenConfig
	upg  Upgrader
}

// WithUpgrader creates a new ListenConfig using the provided Upgrader
func (lc ListenConfig) WithUpgrader(upg Upgrader) *UpgraderListenConfig {
	return &UpgraderListenConfig{
		conf: lc,
		upg:  upg,
	}
}

// Listen acts like the standard net.Listen but using our ListenConfig and via
// an Upgrader tool
func (lu UpgraderListenConfig) Listen(network, addr string) (net.Listener, error) {
	return lu.upg.ListenWithCallback(network, addr, lu.conf.Listen)
}

// ListenPacket acts like the standard net.ListenPacket but using our
// ListenConfig and via an Upgrader tool
func (lu UpgraderListenConfig) ListenPacket(network, addr string) (net.PacketConn, error) {
	return lu.upg.ListenPacketWithCallback(network, addr, lu.conf.ListenPacket)
}

// ListenTCP acts like the standard net.ListenTCP but using our ListenConfig and
// the UpgraderListen
func (lu UpgraderListenConfig) ListenTCP(network string, laddr *net.TCPAddr) (
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
// the UpgraderListen
func (lu UpgraderListenConfig) ListenUDP(network string, laddr *net.UDPAddr) (
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
func (lu UpgraderListenConfig) ListenAll(network string, addrs []string) ([]net.Listener, error) {
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
func (lu UpgraderListenConfig) ListenAllPacket(network string, addrs []string) (
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
func (lu UpgraderListenConfig) ListenAllTCP(network string, laddrs []*net.TCPAddr) (
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
func (lu UpgraderListenConfig) ListenAllUDP(network string, laddrs []*net.UDPAddr) (
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
