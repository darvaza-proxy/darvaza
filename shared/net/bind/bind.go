// Package bind provides mechanisms to facilitate listening TCP and UDP ports
package bind

import (
	"net"

	"github.com/darvaza-proxy/core"
)

const (
	// DefaultPortAttempts indicates how many times we will try binding a port
	DefaultPortAttempts = 4

	// DefaultMaxRecvBufferSize indicates the receive buffer size of UDP listeners
	DefaultMaxRecvBufferSize = 2 * 1024 * 1024

	// MinimumMaxRecvBufferSize indicates the minimum receive buffer size of UDP listeners
	// If a lower value is given, it will be reset to DefaultMaxRecvBufferSize
	MinimumMaxRecvBufferSize = 32
)

// Config is the configuration for Bind()
type Config struct {
	// Interface is the list of interfaces to listen on
	Interfaces []string
	// Addresses is the list of addresses to listen on
	Addresses []string
	// Port is the port to listen on, for both TCP and UDP
	Port uint16
	// PortStrict tells us not to try other ports
	PortStrict bool
	// PortAttempts indicates how many times we will try finding a port
	PortAttempts int

	// ListenTCP is the helper to use to listen on TCP ports
	ListenTCP func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	// ListenUDP is the helper to use to listen on UDP ports
	ListenUDP func(network string, laddr *net.UDPAddr) (*net.UDPConn, error)

	// MaxRecvBufferSize is the buffer size we will attempt to set to
	// UDP listeners
	MaxRecvBufferSize int
}

// SetDefaults attempts to fill any configuration gap, specially
// the IP Addresses when interfaces are provided instead
func (cfg *Config) SetDefaults() error {
	// Port
	if cfg.PortAttempts < 1 {
		cfg.PortAttempts = DefaultPortAttempts
	}

	// UDP
	if cfg.MaxRecvBufferSize < MinimumMaxRecvBufferSize {
		cfg.MaxRecvBufferSize = DefaultMaxRecvBufferSize
	}

	// Callbacks
	if cfg.ListenTCP == nil {
		cfg.ListenTCP = net.ListenTCP
	}
	if cfg.ListenUDP == nil {
		cfg.ListenUDP = net.ListenUDP
	}

	// Addresses
	if len(cfg.Addresses) == 0 {
		addresses, err := cfg.getStringIPAddresses()
		if err != nil {
			return err
		}
		cfg.Addresses = addresses
	}
	return nil
}

func (cfg *Config) getStringIPAddresses() ([]string, error) {
	if len(cfg.Interfaces) > 0 {
		// Add addresses of the given interfaces
		return core.GetStringIPAddresses(cfg.Interfaces...)
	}
	return []string{"0.0.0.0"}, nil
}

// Addrs returns the Addresses list parsed into net.IP
func (cfg *Config) Addrs() ([]net.IP, error) {
	n := len(cfg.Addresses)
	out := make([]net.IP, 0, n)

	for _, s := range cfg.Addresses {
		ip, err := core.ParseNetIP(s)
		if err != nil {
			return out, err
		}

		out = append(out, ip)
	}

	return out, nil
}

func (cfg *Config) refresh(lsns []*net.TCPListener) {
	// Refresh cfg.Addresses
	addrs := make([]string, len(lsns))

	for i, lsn := range lsns {
		addr := lsn.Addr().(*net.TCPAddr)

		if i == 0 {
			// Refresh cfg.Port in case of 0 or non-strict
			cfg.Port = uint16(addr.Port)
		}

		addrs[i] = addr.IP.String()
	}

	cfg.Addresses = addrs
}

// Bind attempts to listen all specified addresses.
// TCP and UDP on the same port for all.
func (cfg *Config) Bind() ([]*net.TCPListener, []*net.UDPConn, error) {
	if err := cfg.SetDefaults(); err != nil {
		return nil, nil, err
	}

	addrs, err := cfg.Addrs()
	if err != nil {
		return nil, nil, err
	}

	tcp, udp, err := cfg.listen(addrs)
	if err == nil {
		// on success, refresh Port and Addresses
		cfg.refresh(tcp)
	}

	return tcp, udp, err
}

func (cfg *Config) listen(addrs []net.IP) ([]*net.TCPListener, []*net.UDPConn, error) {
	var tcp []*net.TCPListener
	var udp []*net.UDPConn
	var err error

	port := int(cfg.Port)

	if cfg.Port != 0 && cfg.PortStrict {
		// strict mode, try only once
		return cfg.tryListen(0, addrs, port)
	}

	for i := 0; i < cfg.PortAttempts; i++ {
		tcp, udp, err = cfg.tryListen(i, addrs, port)
		if err == nil {
			// success
			return tcp, udp, nil
		}
	}

	return nil, nil, err
}

func (cfg *Config) tryListen(pass int, addrs []net.IP, port int) (
	[]*net.TCPListener, []*net.UDPConn, error) {
	//
	if port != 0 {
		port = port + pass
	}
	return cfg.tryListenPort(addrs, port)
}

// revive:disable:cognitive-complexity

func (cfg *Config) tryListenPort(addrs []net.IP, port int) (
	[]*net.TCPListener, []*net.UDPConn, error) {
	// revive:enable:cognitive-complexity
	var ok bool

	n := len(addrs)
	tcpListeners := make([]*net.TCPListener, 0, n)
	udpListeners := make([]*net.UDPConn, 0, n)

	// close all on error
	defer func() {
		if !ok {
			for _, tcpLn := range tcpListeners {
				_ = tcpLn.Close()
			}
			for _, udpLn := range udpListeners {
				_ = udpLn.Close()
			}
		}
	}()

	for _, ip := range addrs {
		// TCP
		tcpAddr := &net.TCPAddr{IP: ip, Port: port}
		tcpLn, err := cfg.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return nil, nil, err
		}

		tcpListeners = append(tcpListeners, tcpLn)

		if port == 0 {
			// port was random, now we stick to it
			port = tcpLn.Addr().(*net.TCPAddr).Port
		}

		// UDP
		udpAddr := &net.UDPAddr{IP: ip, Port: port}
		udpLn, err := cfg.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, nil, err
		}

		udpListeners = append(udpListeners, udpLn)

		if _, err := cfg.setUDPRecvBuffer(udpLn); err != nil {
			return nil, nil, err
		}
	}

	// Success
	ok = true
	return tcpListeners, udpListeners, nil
}

func (cfg *Config) setUDPRecvBuffer(udpLn *net.UDPConn) (int, error) {
	var err error

	size := cfg.MaxRecvBufferSize
	for size > 0 {
		if err = udpLn.SetReadBuffer(size); err == nil {
			// success
			return size, nil
		}
	}

	return 0, err
}

// Bind attempts to listen all addresses specified by the given
// configuration. TCP and UDP on the same port for all.
func Bind(cfg *Config) ([]*net.TCPListener, []*net.UDPConn, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	return cfg.Bind()
}

// UseListener sets Bind's Config to use the provided
// ListenConfig
func (cfg *Config) UseListener(lc TCPUDPListener) {
	cfg.ListenTCP = lc.ListenTCP
	cfg.ListenUDP = lc.ListenUDP
}
