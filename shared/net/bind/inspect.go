package bind

import (
	"net"
	"net/netip"

	"darvaza.org/core"
)

// AddrPortSliceTCPListener attempts to extract netip.AddrPort from a slice of
// *net.TCPListener, it also confirms if all entries were converted
func AddrPortSliceTCPListener(s []*net.TCPListener) ([]netip.AddrPort, bool) {
	var out []netip.AddrPort

	for _, lsn := range s {
		if addr, ok := core.AddrPort(lsn.Addr()); ok {
			if addr.IsValid() {
				out = append(out, addr)
			}
		}
	}
	return out, len(out) == len(s)
}

// AddrPortSliceUDPConn attempts to extract netip.AddrPort from a slice of
// *net.UDPConn, it also confirms if all entries were converted
func AddrPortSliceUDPConn(s []*net.UDPConn) ([]netip.AddrPort, bool) {
	var out []netip.AddrPort

	for _, lsn := range s {
		if addr, ok := core.AddrPort(lsn.LocalAddr()); ok {
			if addr.IsValid() {
				out = append(out, addr)
			}
		}
	}
	return out, len(out) == len(s)
}

// SamePort verifies if all addresses on a slice point to the same port,
// SamePort verifies if all addresses on a slice point to the same port,
// and tells us which.
// In case of error it will return the first port found.
func SamePort(s []netip.AddrPort) (uint16, bool) {
	var port uint16
	for _, ap := range s {
		if port == 0 {
			port = ap.Port()
		} else if ap.Port() != port {
			// fail
			return port, false
		}
	}
	return port, true
}

// IPAddresses extract all valid unique addresses on a slice, and if all
// were unique and valid
func IPAddresses(s []netip.AddrPort) ([]netip.Addr, bool) {
	out := make([]netip.Addr, 0, len(s))

	for _, ap := range s {
		if ap.IsValid() {
			addr := ap.Addr()
			if !core.SliceContains(out, addr) {
				out = append(out, addr)
			}
		}
	}

	return out, len(out) == len(s)
}

// StringIPAddresses extract all valid unique addresses on a slice,
// and if all were unique and valid
func StringIPAddresses(s []netip.AddrPort) ([]string, bool) {
	out := make([]string, 0, len(s))

	for _, ap := range s {
		if ap.IsValid() {
			addr := ap.Addr().String()
			if !core.SliceContains(out, addr) {
				out = append(out, addr)
			}
		}
	}

	return out, len(out) == len(s)
}
