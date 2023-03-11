package httpserver

import (
	"io"
	"net"
	"net/netip"
	"syscall"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/net/bind"
)

func closeAll[T io.Closer](s []T) {
	for _, l := range s {
		_ = l.Close()
	}
}

// Listen listens to the addresses specified on the Config
func (srv *Server) Listen() error {
	if srv.slsn != nil {
		return syscall.EBUSY
	}

	lc := bind.NewListenConfig(srv.cfg.Context, srv.cfg.Bind.KeepAlive)
	return srv.ListenWithListener(lc)
}

// revive:disable:cognitive-complexity

// ListenWithListener uses a given TCPUDPListener to listen to the addresses
// specified on the Config
func (srv *Server) ListenWithListener(lc bind.TCPUDPListener) error {
	// revive:enable:cognitive-complexity
	var ok bool

	if srv.slsn != nil {
		return syscall.EBUSY
	}

	defer func() {
		if !ok {
			closeAll(srv.slsn)
			closeAll(srv.ilsn)
			closeAll(srv.ulsn)
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

	srv.slsn = secure
	srv.ulsn = quic

	// update config
	bc.RefreshFromTCPListeners(secure)
	cfg.Addresses = bc.Addresses
	cfg.Port = bc.Port

	if cfg.AllowInsecure {
		// insecure
		bc.Port = cfg.PortInsecure
		bc.DefaultPort = 80
		bc.OnlyTCP = true

		insecure, _, err := bc.Bind()
		if err != nil {
			return err
		}

		srv.ilsn = insecure

		// update config
		if len(insecure) == 0 {
			core.Panic("unreachable")
		}
		addr, ok := core.AddrPort(insecure[0].Addr())
		if !ok {
			core.Panic("unreachable")
		}
		cfg.PortInsecure = addr.Port()
	}

	ok = true
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
