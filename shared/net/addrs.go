// Package net provides generic network helpers and proxies
// to some useful standard types and functions
package net

import (
	"fmt"
	"strconv"

	"golang.org/x/net/idna"

	"darvaza.org/core"
)

// SplitHostPort splits a network address into host and port,
// validating the host and port in the process
func SplitHostPort(hostport string) (string, uint16, error) {
	host, port, err := core.SplitHostPort(hostport)
	switch {
	case err != nil:
		return "", 0, err
	case port == "":
		return host, 0, nil
	default:
		u, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return "", 0, err
		}

		return host, uint16(u & 0xffff), nil
	}
}

// JoinHostPort combines a given host address and a port, validating
// the provided IP address or name in the process
func JoinHostPort(host string, port uint16) (string, error) {
	host, err := stringifyAddr(host)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%v", host, port), nil
}

func stringifyAddr(host string) (string, error) {
	addr, err := core.ParseAddr(host)
	if err == nil {
		// IP Address
		switch {
		case addr.IsUnspecified():
			host = ""
		case addr.Is6():
			host = fmt.Sprintf("[%s]", addr.String())
		case addr.Is4():
			host = addr.String()
		}

		return host, nil
	}

	s, err := idna.ToASCII(host)
	if err == nil {
		// good name
		return s, nil
	}

	// invalid host
	return "", err
}

// JoinAllHostPorts combines a list of addresses and a list of ports, validating
// the provided IP addresses or name in the process
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
