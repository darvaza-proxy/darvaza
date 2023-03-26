// Package simple provides a simple self-contained TLS Store
package simple

import (
	"container/list"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/fs"
	"sync"

	"golang.org/x/sync/singleflight"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/storage"
	"github.com/darvaza-proxy/darvaza/shared/storage/certpool"
	"github.com/darvaza-proxy/darvaza/shared/x509utils"
	"github.com/darvaza-proxy/slog"
)

var (
	_ storage.Store = (*Store)(nil)
)

// A Getter is a helper to get a certificate for a name
type Getter func(ctx context.Context,
	key x509utils.PrivateKey, name string) (*tls.Certificate, error)

// Config is a custom factory for the Store allowing the usage
// of a Logger and a roots base different that what the system provides
type Config struct {
	Base   x509utils.CertPooler
	Logger slog.Logger
}

// Store is a darvaza TLS Store that doesn't talk to anyone
// external service nor monitors for new files
type Store struct {
	mu sync.Mutex
	g  singleflight.Group

	pool     *certpool.CertPool
	keys     []x509utils.PrivateKey
	certs    []*tls.Certificate
	hashed   map[certpool.Hash]*tls.Certificate
	names    map[string]*list.List
	patterns map[string]*list.List
}

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

// GetCertificateWithCallback returns the TLS Certificate that should be used
// for a given TLS request. If one isn't available it call use
// a callback to acquire one
func (s *Store) GetCertificateWithCallback(chi *tls.ClientHelloInfo,
	getter Getter) (*tls.Certificate, error) {
	//
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
			// found or aqcuired
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
		pool:     pb.Pool(),
		keys:     []x509utils.PrivateKey{},
		certs:    []*tls.Certificate{},
		hashed:   make(map[certpool.Hash]*tls.Certificate),
		names:    make(map[string]*list.List),
		patterns: make(map[string]*list.List),
	}

	addCerts(store, certs...)
	return store, nil
}

func addCerts(s *Store, certs ...*tls.Certificate) {
	for _, c := range certs {
		key, ok := c.PrivateKey.(x509utils.PrivateKey)
		if !ok {
			// drop keyless certificates
			continue
		}

		// contains key
		if !core.SliceContainsFn(s.keys, key, pkEqual) {
			// new key
			s.keys = append(s.keys, key)
		}

		// contains cert
		hash := certpool.HashCert(c.Leaf)
		if _, found := s.hashed[hash]; !found {
			// new cert
			s.hashed[hash] = c
			names, patterns := x509utils.Names(c.Leaf)
			addCertWithNames(s, c, names, patterns)
		}
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

func pkEqual(a, b x509utils.PrivateKey) bool {
	return a.Equal(b)
}
