// Package simple provides a simple self-contained TLS Store
package simple

import (
	"container/list"
	"crypto/tls"
	"crypto/x509"
	"io/fs"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/storage"
	"github.com/darvaza-proxy/darvaza/shared/storage/certpool"
	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

var (
	_ storage.Store = (*Store)(nil)
)

// Store is a darvaza TLS Store that doesn't talk to anyone
// external service nor monitors for new files
type Store struct {
	pool     *certpool.CertPool
	certs    []*tls.Certificate
	names    map[string]*list.List
	patterns map[string]*list.List
}

// GetCAPool returns a reference to the Certificates Pool
func (s *Store) GetCAPool() *x509.CertPool {
	return s.pool.Export()
}

// GetCertificate returns the TLS Certificate that should be used
// for a given TLS request
func (s *Store) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cert := s.findMatchingCert(chi); cert != nil {
		// found
		return cert, nil
	}

	if cert := s.findAnyCert(chi); cert != nil {
		// better than nothing
		return cert, nil
	}

	return nil, fs.ErrNotExist
}

func (s *Store) findMatchingCert(chi *tls.ClientHelloInfo) *tls.Certificate {
	if name, ok := x509utils.SanitiseName(chi.ServerName); ok {
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
	}

	return nil
}

func (s *Store) findAnyCert(chi *tls.ClientHelloInfo) *tls.Certificate {
	for _, c := range s.certs {
		if err := chi.SupportsCertificate(c); err == nil {
			// works for me
			return c
		}
	}

	return nil
}

// New creates a Store using a list of PEM blocks, filenames, or directories
func New(blocks ...string) (*Store, error) {
	var pb certpool.PoolBuffer

	for _, s := range blocks {
		if s != "" {
			if err := pb.Add(s); err != nil {
				return nil, err
			}
		}
	}

	return NewFromBuffer(&pb, nil)
}

// revive:disable:cognitive-complexity

// NewFromBuffer creates a Store from a given PoolBuffer
func NewFromBuffer(pb *certpool.PoolBuffer, base *certpool.CertPool) (*Store, error) {
	// revive:enable:cognitive-complexity
	if base == nil {
		sys, err := certpool.SystemCertPool()
		if err != nil {
			return nil, core.Wrap(err, "certpool.SystemCertPool")
		}
		base = sys
	}

	certs, err := pb.Certificates(base)
	if err != nil {
		return nil, err
	}

	store := &Store{
		certs:    removeKeyless(certs),
		pool:     pb.Pool(),
		names:    make(map[string]*list.List),
		patterns: make(map[string]*list.List),
	}

	for _, c := range store.certs {
		names, patterns := x509utils.Names(c.Leaf)
		for _, s := range names {
			core.MapListAppend(store.names, s, c)
		}
		for _, s := range patterns {
			core.MapListAppend(store.patterns, s, c)
		}
	}

	return store, nil
}

func removeKeyless(certs []*tls.Certificate) []*tls.Certificate {
	var j int
	for i, c := range certs {
		if c.PrivateKey == nil {
			continue
		}

		if i != j {
			certs[j] = c
		}
		j++
	}

	return certs[:j]
}
