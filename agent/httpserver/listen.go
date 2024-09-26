package httpserver

import (
	"fmt"
	"io"
	"net"
	"net/netip"
	"syscall"

	"darvaza.org/core"
	"darvaza.org/x/net/bind"
)

// ServerListeners is the list of all listeners on a Server
type ServerListeners struct {
	Insecure []*net.TCPListener
	Secure   []*net.TCPListener
	Quic     []*net.UDPConn
}

// Close closes all listeners. Errors are ignored
func (sl *ServerListeners) Close() error {
	closeAll(sl.Insecure)
	closeAll(sl.Secure)
	closeAll(sl.Quic)
	return nil
}

func closeAll[T io.Closer](s []T) {
	for _, l := range s {
		_ = l.Close()
	}
}

func (*ServerListeners) checkCount(secure, quic, insecure int) error {
	// same number of secure and quic listeners, and at least one
	if secure == quic && secure > 0 {
		// either no insecure listeners, or the same number as secure ones
		if insecure == 0 || insecure == secure {
			return nil
		}
	}

	return fmt.Errorf("inconsistent listeners count (secure:%v, quic:%v, insecure:%v)",
		secure, quic, insecure)
}

func (*ServerListeners) checkSecureMatch(tcpAddr *net.TCPAddr, udpAddr *net.UDPAddr,
	port int) error {
	//
	if !tcpAddr.IP.Equal(udpAddr.IP) ||
		tcpAddr.Port != udpAddr.Port ||
		tcpAddr.Port != port {
		return fmt.Errorf("listener mismatch (secure:%q, quic:%q, port:%v)",
			tcpAddr.String(), udpAddr.String(), port)
	}

	return nil
}

func (*ServerListeners) checkInsecureMatch(tcpAddr *net.TCPAddr, ip net.IP, port int) error {
	if !tcpAddr.IP.Equal(ip) || tcpAddr.Port != port {
		expected := &net.TCPAddr{
			IP:   ip,
			Port: port,
		}

		return fmt.Errorf("listener mismatch (insecure:%q expected:%q)",
			tcpAddr.String(),
			expected.String())
	}
	return nil
}

func (sl *ServerListeners) secureAddress(i int) *net.TCPAddr {
	p, ok := sl.Secure[i].Addr().(*net.TCPAddr)
	if !ok {
		panic("unreachable")
	}
	return p
}

func (sl *ServerListeners) quicAddress(i int) *net.UDPAddr {
	p, ok := sl.Quic[i].LocalAddr().(*net.UDPAddr)
	if !ok {
		panic("unreachable")
	}
	return p
}

func (sl *ServerListeners) insecureAddress(i int) *net.TCPAddr {
	p, ok := sl.Insecure[i].Addr().(*net.TCPAddr)
	if !ok {
		panic("unreachable")
	}
	return p
}

// revive:disable:cognitive-complexity

// IPAddresses validates the ServerListeners and provides the list of
// IP Addresses as string
func (sl *ServerListeners) IPAddresses() ([]net.IP, error) {
	// revive:enable:cognitive-complexity
	nInsecure := len(sl.Insecure)
	nSecure := len(sl.Secure)
	nQuic := len(sl.Quic)

	if err := sl.checkCount(nSecure, nQuic, nInsecure); err != nil {
		return nil, err
	}

	count := nSecure
	port, insecure := 0, 0
	ips := make([]net.IP, count)

	for i := 0; i < count; i++ {
		tcpAddr := sl.secureAddress(i)
		udpAddr := sl.quicAddress(i)

		if port == 0 {
			port = tcpAddr.Port
		}

		if err := sl.checkSecureMatch(tcpAddr, udpAddr, port); err != nil {
			return nil, err
		}

		ips[i] = tcpAddr.IP
	}

	if nInsecure > 0 {
		for i := 0; i < count; i++ {
			tcpAddr := sl.insecureAddress(i)

			if insecure == 0 {
				insecure = tcpAddr.Port
			}

			if err := sl.checkInsecureMatch(tcpAddr, ips[i], insecure); err != nil {
				return nil, err
			}
		}
	}

	return ips, nil
}

// StringIPAddresses validates the ServerListeners and provides the list of
// IP Addresses as string
func (sl *ServerListeners) StringIPAddresses() ([]string, error) {
	ips, err := sl.IPAddresses()

	addrs := make([]string, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, ip.String())
	}

	return addrs, err
}

// Ports returns the port of the first secure listener and optionally
// the first insecure one
func (sl *ServerListeners) Ports() (secure uint16, insecure uint16, ok bool) {
	if len(sl.Secure) == 0 {
		return 0, 0, false
	}

	addr := sl.secureAddress(0)
	secure = uint16(addr.Port)

	if len(sl.Insecure) > 0 {
		addr := sl.insecureAddress(0)
		insecure = uint16(addr.Port)
	}

	return secure, insecure, true
}

// Listen listens to the addresses specified on the Config
func (srv *Server) Listen() error {
	if srv.sl != nil {
		return syscall.EBUSY
	}

	lc := bind.NewListenConfig(srv.cfg.Context, srv.cfg.Bind.KeepAlive)
	return srv.ListenWithListener(lc)
}

// ListenWithListener uses a given TCPUDPListener to listen to the addresses
// specified on the Config
func (srv *Server) ListenWithListener(lc bind.TCPUDPListener) error {
	var sl ServerListeners
	var ok bool

	if srv.sl != nil {
		return syscall.EBUSY
	}

	defer func() {
		if !ok {
			_ = sl.Close()
		}
	}()

	cfg := &srv.cfg.Bind

	// secure and quic
	bc := bind.Config{
		Interfaces:   cfg.Interfaces,
		Addresses:    cfg.Addresses,
		DefaultPort:  443,
		Port:         cfg.Port,
		PortStrict:   cfg.PortStrict,
		PortAttempts: cfg.PortAttempts,
	}
	bc.UseListener(lc)

	secure, quic, err := bc.Bind()
	if err != nil {
		return err
	}

	sl.Secure = secure
	sl.Quic = quic

	// update config
	bc.RefreshFromTCPListeners(secure)
	cfg.Addresses = bc.Addresses

	if cfg.AllowInsecure {
		// insecure
		bc.Port = cfg.PortInsecure
		bc.DefaultPort = 80
		bc.OnlyTCP = true

		insecure, _, err := bc.Bind()
		if err != nil {
			return err
		}

		sl.Insecure = insecure
	}

	// update config
	cfg.Port, cfg.PortInsecure, _ = sl.Ports()

	ok = true

	// Store
	srv.sl = &sl
	return nil
}

// WithListeners validates and attaches provided listeners
func (srv *Server) WithListeners(sl *ServerListeners) error {
	if srv.sl != nil {
		return syscall.EBUSY
	}

	// Validate
	addrs, err := sl.StringIPAddresses()
	if err != nil {
		return err
	}

	port, portInsecure, _ := sl.Ports()

	// Update config
	cfg := &srv.cfg.Bind

	if portInsecure != 0 && !cfg.AllowInsecure {
		srv.log.Warn().
			Printf("Insecure was disabled but listeners at %v were provided", portInsecure)
		cfg.AllowInsecure = true
	}

	cfg.Addresses = addrs
	cfg.Port = port
	cfg.PortInsecure = portInsecure

	// Store
	srv.sl = sl

	return nil
}

func getStringAddrPort(addr net.Addr) []string {
	if ap, ok := core.AddrPort(addr); ok {
		if addr := ap.Addr(); addr.IsUnspecified() {
			out, ok := getAllStringAddrPort(ap.Port())
			if ok {
				return out
			}
		}
		return []string{ap.String()}
	}
	return []string{addr.String()}
}

func getAllStringAddrPort(port uint16) ([]string, bool) {
	ifaces, _ := core.GetInterfacesNames("lo")
	addrs, _ := core.GetIPAddresses(ifaces...)
	if len(ifaces) > 0 && len(addrs) == 0 {
		addrs, _ = core.GetIPAddresses()
	}

	if l := len(addrs); l > 0 {
		out := make([]string, 0, l)
		for _, addr := range addrs {
			ap := netip.AddrPortFrom(addr, port)
			out = append(out, ap.String())
		}
		return out, true
	}

	return nil, false
}
