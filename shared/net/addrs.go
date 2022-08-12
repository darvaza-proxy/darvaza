package net

import (
	"fmt"
	"net"
	"strings"
)

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
