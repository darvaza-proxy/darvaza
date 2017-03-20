package util

import (
	"net"
	"time"
)

var (
	// DialTimeout is timeout for dialing
	DialTimeout time.Duration

	// DefaultDialTimeout is default timeout, 300ms
	DefaultDialTimeout = 300 * time.Millisecond
)

func getTimeout() time.Duration {
	if DialTimeout != 0 {
		return DialTimeout
	}
	return DefaultDialTimeout
}

/*
Dial try to dial all the ip address one by one if addr is a domain name, util one is successed.



It tries to dail to ipv6 first, and then dial to ipv4 until one is successed

If  dial to all ip failed, it return a error.

*/
func Dial(network string, addr string) (net.Conn, error) {
	var ip net.IP
	var err error
	var ips []net.IP
	var conn net.Conn

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	/* test is ip address */
	ip = net.ParseIP(host)
	if ip == nil {
		/* domain name resolve */
		ips, err = ResolveDNS(host)
		if err != nil {
			return nil, err
		}
	} else {
		ips = append(ips, ip)
	}

	for _, ip = range ips {
		conn, err = net.DialTimeout(network, net.JoinHostPort(ip.String(), port), getTimeout())
		if err == nil {
			/* dial success, return */
			return conn, nil
		}
		/* continue try next ip */
	}

	/* return last error */
	return nil, err
}
