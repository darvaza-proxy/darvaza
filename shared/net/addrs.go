package net

import (
	"fmt"
	"net"
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
