// Package net provides generic network helpers and proxies
// to some useful standard types and functions
package net

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
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
	ip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		// bad address
		return "", err
	} else if ip == nil || ip.IP.IsUnspecified() {
		// wildcard
		host = ""
	} else {
		host = ip.String()

		if strings.ContainsRune(host, ':') {
			// IPv6
			host = fmt.Sprintf("[%s]", host)
		}
	}

	return fmt.Sprintf("%s:%v", host, port), nil
}

// JoinAllHostPorts combines a list of addresses and a list of ports, validating
// the provided IP addresses in the process
func JoinAllHostPorts(addresses []string, ports []uint16) ([]string, error) {
	var out []string

	for _, s := range addresses {

		ip, err := net.ResolveIPAddr("ip", s)
		if err != nil {
			// bad address
			return out, err
		} else if ip == nil || ip.IP.IsUnspecified() {
			// wildcard
			s = ""
		} else {
			s = ip.String()

			if strings.ContainsRune(s, ':') {
				// IPv6
				s = fmt.Sprintf("[%s]", s)
			}
		}

		for _, p := range ports {
			out = append(out, fmt.Sprintf("%s:%v", s, p))
		}
	}

	return out, nil
}

// GetStringIPAddresses returns a list of text IP addresses bound
// to the given interfaces or all if none are given
func GetStringIPAddresses(ifaces ...string) ([]string, error) {
	addrs, err := GetIPAddresses(ifaces...)

	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {

		if addr.IsValid() {
			s := addr.String()
			out = append(out, s)
		}
	}

	return out, err
}

// GetNetIPAddresses returns a list of net.IP addresses bound to
// the given interfaces or all if none are given
func GetNetIPAddresses(ifaces ...string) ([]net.IP, error) {
	addrs, err := GetIPAddresses(ifaces...)

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

	return out, err
}

// GetIPAddresses returns a list of netip.Addr bound to the given
// interfaces or all if none are given
func GetIPAddresses(ifaces ...string) ([]netip.Addr, error) {
	var out []netip.Addr

	if len(ifaces) == 0 {
		var err error

		ifaces, err = GetInterfacesNames()
		if err != nil {
			return out, err
		}
	}

	for _, name := range ifaces {
		ifi, err := net.InterfaceByName(name)
		if err != nil {
			return out, err
		}

		addrs, err := ifi.Addrs()
		if err != nil {
			return out, err
		}

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
	}

	return out, nil
}

// GetInterfacesNames returns the list of interfaces,
// considering an optional exclusion list
func GetInterfacesNames(except ...string) ([]string, error) {
	var out []string

	s, err := net.Interfaces()
	if err != nil {
		return out, err
	}

	for _, ifi := range s {
		name := ifi.Name

		for _, nope := range except {
			if name == nope {
				name = ""
				break
			}
		}

		if name != "" {
			out = append(out, name)
		}
	}

	return out, nil
}
