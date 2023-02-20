package x509utils

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

// Names returns a list of exact names and patterns the certificate
// supports
func Names(cert *x509.Certificate) ([]string, []string) {
	var names []string
	var patterns []string

	if len(cert.URIs) == 0 {
		// an "old" certificate, no SAN
		names = append(names, cert.Subject.CommonName)
	}

	for _, addr := range cert.IPAddresses {
		// See RFC 6125, Appendix B.2.
		names = append(names, fmt.Sprintf("[%s]", addr.String()))
	}

	for _, s := range cert.DNSNames {
		s = strings.ToLower(s)

		if strings.HasPrefix(s, "*.") {
			// pattern
			patterns = append(patterns, s[1:])
		} else {
			// literal
			names = append(names, s)
		}
	}

	return names, patterns
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
			if addr, err := netip.ParseAddr(host); err == nil {
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
