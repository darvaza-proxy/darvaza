// Package net provides generic network helpers and proxies
// to some useful standard types and functions
package net

import (
	"fmt"
	"strconv"

	"github.com/darvaza-proxy/core"
)

// SplitHostPort splits a network address into host and port,
// validating the port in the process
func SplitHostPort(hostport string) (string, uint16, error) {
	host, port, err := core.SplitHostPort(hostport)
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
	addr, err := core.ParseAddr(host)
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
