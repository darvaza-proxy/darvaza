package x509utils

import (
	"crypto/x509"
	"fmt"
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
