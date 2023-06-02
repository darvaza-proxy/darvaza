// Package simple provides a simple self-contained TLS Store
package simple

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/fs"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/storage"
	"darvaza.org/darvaza/shared/x509utils"
)

var (
	_ storage.Store = (*Store)(nil)
)

// GetCAPool returns a reference to the Certificates Pool
func (s *Store) GetCAPool() *x509.CertPool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.pool.Export()
}

// GetCertificate returns the TLS Certificate that should be used
// for a given TLS request
func (s *Store) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return s.GetCertificateWithCallback(chi, nil)
}

// revive:disable:cognitive-complexity

// GetCertificateWithCallback returns the TLS Certificate that should be used
// for a given TLS request. If one isn't available it call use
// a callback to acquire one
func (s *Store) GetCertificateWithCallback(chi *tls.ClientHelloInfo,
	getter Getter) (*tls.Certificate, error) {
	// revive:enable:cognitive-complexity
	name := chi.ServerName
	if name == "" {
		name = chi.Conn.LocalAddr().String()
	}

	name, ok := x509utils.SanitiseName(name)
	if ok {
		// find name match locally
		s.mu.Lock()
		cert := s.findMatchingCert(chi, name)
		s.mu.Unlock()

		if cert == nil && getter != nil {
			// try to acquire
			cert = s.getMatchingCert(chi.Context(), name, getter)
		}

		if cert != nil {
			// found or acquired
			return cert, nil
		}
	}

	// get me anything please
	s.mu.Lock()
	defer s.mu.Unlock()

	cert := s.findAnyCert(chi)
	if cert != nil {
		return cert, nil
	}

	return nil, fs.ErrNotExist
}

func (s *Store) getMatchingCert(ctx context.Context, name string, getter Getter) *tls.Certificate {
	var key x509utils.PrivateKey

	// attempt to reuse our existing key
	s.mu.Lock()
	if len(s.keys) > 0 {
		key = s.keys[0]
	}
	s.mu.Unlock()

	v, err, _ := s.g.Do(name, func() (any, error) {
		c, e := getter(ctx, key, name)
		return c, e
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	// singleflight.Do returned once, release them all
	s.g.Forget(name)

	if err == nil {
		if cert, ok := v.(*tls.Certificate); ok {
			// acquired. store
			addCerts(s, cert)
			return cert
		}
	}

	return nil
}

func (s *Store) findMatchingCert(chi *tls.ClientHelloInfo, name string) *tls.Certificate {
	// IP
	if n, ok := x509utils.NameAsIP(name); ok {
		return FindSupportedInMap(chi, n, s.names)
	}

	// exact
	if cert := FindSupportedInMap(chi, name, s.names); cert != nil {
		return cert
	}

	// wildcard
	if suffix, ok := x509utils.NameAsSuffix(name); ok {
		return FindSupportedInMap(chi, suffix, s.patterns)
	}

	return nil
}

func (s *Store) findAnyCert(chi *tls.ClientHelloInfo) *tls.Certificate {
	var cert *tls.Certificate

	core.ListForEach(s.certs, func(c *tls.Certificate) bool {
		if err := chi.SupportsCertificate(c); err == nil {
			// works for me
			cert = c
		}

		return cert != nil
	})

	return cert
}
