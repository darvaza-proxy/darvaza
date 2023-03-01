package x509utils

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"

	"github.com/darvaza-proxy/core"
)

// Names returns a list of exact names and patterns the certificate
// supports
func Names(cert *x509.Certificate) ([]string, []string) {
	names, patterns := splitDNSNames(cert.DNSNames)
	names = appendIPAddresses(names, cert.IPAddresses)

	if len(names) == 0 && len(patterns) == 0 {
		if len(cert.URIs) == 0 {
			// an "old" certificate, no SAN
			cn := cert.Subject.CommonName
			if cn != "" {
				names = append(names, strings.ToLower(cn))
			}
		}
	}
	return names, patterns
}

func splitDNSNames(dnsNames []string) (names []string, patterns []string) {
	for _, s := range dnsNames {
		s = strings.ToLower(s)

		if strings.HasPrefix(s, "*.") {
			// pattern
			patterns = append(patterns, s[1:])
		} else if s != "" {
			// literal
			names = append(names, s)
		}
	}

	return names, patterns
}

func appendIPAddresses(names []string, addrs []net.IP) []string {
	for _, ip := range addrs {
		if addr, ok := netip.AddrFromSlice(ip); ok {
			if addr.IsValid() {
				name := fmt.Sprintf("[%s]", addr.String())
				names = append(names, name)
			}
		}
	}
	return names
}

// Hostname returns a sanitised hostname for a parsed URL
func Hostname(u *url.URL) (string, bool) {
	return SanitiseName(u.Host)
}

// SanitiseName takes a Hostname and returns the name (or address)
// we will use for matching certificates
func SanitiseName(name string) (string, bool) {
	if name != "" {
		if host, _, err := net.SplitHostPort(name); err == nil {
			if addr, err := core.ParseAddr(host); err == nil {
				// IP
				addr.Unmap()
				addr.WithZone("")
				host = addr.String()
			} else {
				// Name
				host = removeZone(host)
			}
			return host, len(host) > 2
		}
	}
	return "", false
}

func removeZone(name string) string {
	idx := strings.LastIndexFunc(name, func(r rune) bool {
		return r == '%'
	})
	if idx < 0 {
		return name
	}
	return name[:idx]
}
