// Package net provides generic network helpers and proxies
// to some useful standard types and functions
package net

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"

	"github.com/darvaza-proxy/darvaza/shared/data"
)

// SplitHostPort splits a network address into host and port,
// validating the port in the process
func SplitHostPort(hostport string) (string, uint16, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, err
	} else if port == "" {
		return host, 0, nil
	} else if u, err := strconv.ParseUint(port, 10, 16); err != nil {
		return "", 0, err
	} else {
		return host, uint16(u & 0xffff), nil
	}
}

// JoinHostPort combines a given host address and a port, validating
// the provided IP address in the process
func JoinHostPort(host string, port uint16) (string, error) {
	host, err := stringifyAddr(host)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%v", host, port), nil
}

func stringifyAddr(host string) (string, error) {
	if host == "0" {
		// special case
		return "", nil
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return "", err
	}

	if addr.IsUnspecified() {
		host = ""
	} else if addr.Is4() {
		host = addr.String()
	} else {
		host = fmt.Sprintf("[%s]", addr.String())
	}

	return host, nil
}

// JoinAllHostPorts combines a list of addresses and a list of ports, validating
// the provided IP addresses in the process
func JoinAllHostPorts(addresses []string, ports []uint16) ([]string, error) {
	var out []string

	for _, s := range addresses {
		addr, err := stringifyAddr(s)
		if err != nil {
			return out, err
		}

		for _, p := range ports {
			out = append(out, fmt.Sprintf("%s:%v", addr, p))
		}
	}

	return out, nil
}

// GetStringIPAddresses returns a list of text IP addresses bound
// to the given interfaces or all if none are given
func GetStringIPAddresses(ifaces ...string) ([]string, error) {
	addrs, err := GetIPAddresses(ifaces...)
	out := asStringIPAddresses(addrs...)

	return out, err
}

func asStringIPAddresses(addrs ...netip.Addr) []string {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if addr.IsValid() {
			s := addr.String()
			out = append(out, s)
		}
	}
	return out
}

// GetNetIPAddresses returns a list of net.IP addresses bound to
// the given interfaces or all if none are given
func GetNetIPAddresses(ifaces ...string) ([]net.IP, error) {
	addrs, err := GetIPAddresses(ifaces...)
	out := asNetIPAddresses(addrs...)
	return out, err
}

func asNetIPAddresses(addrs ...netip.Addr) []net.IP {
	out := make([]net.IP, len(addrs))
	for i, addr := range addrs {
		var ip net.IP

		if addr.Is4() {
			a4 := addr.As4()
			ip = a4[:]
		} else {
			a16 := addr.As16()
			ip = a16[:]
		}

		out[i] = ip
	}

	return out
}

// GetIPAddresses returns a list of netip.Addr bound to the given
// interfaces or all if none are given
func GetIPAddresses(ifaces ...string) ([]netip.Addr, error) {
	var out []netip.Addr

	if len(ifaces) == 0 {
		// all addresses
		addrs, err := net.InterfaceAddrs()
		out = appendNetIPAsIP(out, addrs...)

		return out, err
	}

	// only given
	for _, name := range ifaces {
		ifi, err := net.InterfaceByName(name)
		if err != nil {
			return out, err
		}

		addrs, err := ifi.Addrs()
		if err != nil {
			return out, err
		}

		out = appendNetIPAsIP(out, addrs...)
	}

	return out, nil
}

func appendNetIPAsIP(out []netip.Addr, addrs ...net.Addr) []netip.Addr {
	for _, addr := range addrs {
		var s []byte

		switch v := addr.(type) {
		case *net.IPAddr:
			s = v.IP
		case *net.IPNet:
			s = v.IP
		}

		if ip, ok := netip.AddrFromSlice(s); ok {
			out = append(out, ip.Unmap())
		}
	}

	return out
}

// GetInterfacesNames returns the list of interfaces,
// considering an optional exclusion list
func GetInterfacesNames(except ...string) ([]string, error) {
	s, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(s))

	for _, ifi := range s {
		if s := ifi.Name; s != "" {
			out = append(out, s)
		}
	}

	if len(except) > 0 {
		out = data.SliceMinus(out, except)
	}
	return out, nil
}

// revive:disable:cognitive-complexity

// AddrPort attempts to extract a netip.AddrPort from an object
func AddrPort(v any) (netip.AddrPort, bool) {
	if p, ok := v.(interface {
		AddrPort() netip.AddrPort
	}); ok {
		return p.AddrPort(), true
	}

	if p, ok := v.(interface {
		Addr() net.Addr
	}); ok {
		return AddrPort(p.Addr())
	}

	if p, ok := v.(interface {
		RemoteAddr() net.Addr
	}); ok {
		return AddrPort(p.RemoteAddr())
	}

	switch addr := v.(type) {
	case *net.TCPAddr:
		if ip, ok := netip.AddrFromSlice(addr.IP); ok {
			return netip.AddrPortFrom(ip, uint16(addr.Port)), true
		}
	case *net.UDPAddr:
		if ip, ok := netip.AddrFromSlice(addr.IP); ok {
			return netip.AddrPortFrom(ip, uint16(addr.Port)), true
		}
	}

	return netip.AddrPort{}, false
}

// revive:enable:cognitive-complexity
