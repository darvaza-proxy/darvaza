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
	"github.com/darvaza-proxy/slog"
)

var (
	_ storage.Store = (*Store)(nil)
)

// Config is a custom factory for the Store allowing the usage
// of a Logger and a roots base different that what the system provides
type Config struct {
	Base   x509utils.CertPooler
	Logger slog.Logger
}

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
	var c Config
	return c.New(blocks...)
}

// New creates a Store using keys and certificates provided as
// files, directories, or direct PEM encoded content
func (c *Config) New(blocks ...string) (*Store, error) {
	var pb certpool.PoolBuffer

	if c.Logger != nil {
		pb.SetLogger(c.Logger)
	}

	for _, s := range blocks {
		if s != "" {
			if err := pb.Add(s); err != nil {
				return nil, err
			}
		}
	}

	return NewFromBuffer(&pb, c.Base)
}

// NewFromBuffer creates a Store from a given PoolBuffer
func NewFromBuffer(pb *certpool.PoolBuffer, base x509utils.CertPooler) (*Store, error) {
	certs, err := pb.Certificates(base)
	if err != nil {
		return nil, err
	}

	store := &Store{
		certs:    []*tls.Certificate{},
		pool:     pb.Pool(),
		names:    make(map[string]*list.List),
		patterns: make(map[string]*list.List),
	}

	addCerts(store, certs...)
	return store, nil
}

func addCerts(s *Store, certs ...*tls.Certificate) {
	for _, c := range certs {
		if c.PrivateKey == nil {
			// drop keyless certificates
			continue
		}

		names, patterns := x509utils.Names(c.Leaf)
		addCertWithNames(s, c, names, patterns)
	}
}

func addCertWithNames(s *Store, c *tls.Certificate,
	names, patterns []string) {
	//
	s.certs = append(s.certs, c)
	for _, name := range names {
		core.MapListAppend(s.names, name, c)
	}
	for _, pattern := range patterns {
		core.MapListAppend(s.patterns, pattern, c)
	}
}
