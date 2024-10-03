package simple

import (
	"container/list"
	"context"
	"sync"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/x/tls"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"

	legacy "darvaza.org/darvaza/shared/storage/certpool"

	"golang.org/x/sync/singleflight"
)

// A Getter is a helper to get a certificate for a name
type Getter func(ctx context.Context,
	key x509utils.PrivateKey, name string) (*tls.Certificate, error)

// Store is a darvaza TLS Store that doesn't talk to anyone
// external service nor monitors for new files
type Store struct {
	mu sync.Mutex
	g  singleflight.Group

	logger slog.Logger

	roots   certpool.CertPool
	inter   certpool.CertPool
	bundler *tls.Bundler

	keys     []x509utils.PrivateKey
	certs    *list.List
	hashed   map[certpool.Hash]*certInfo
	names    map[string]*list.List
	patterns map[string]*list.List
}

type certInfo struct {
	c        *tls.Certificate
	hash     certpool.Hash
	names    []string
	patterns []string
}

// init unconditionally initializes the Store
func (s *Store) init() {
	s.logger = defaultLogger()

	_ = s.roots.Reset()
	_ = s.inter.Reset()
	s.bundler.Roots = &s.roots
	s.bundler.Inter = &s.inter

	s.keys = []x509utils.PrivateKey{}
	s.certs = list.New()
	s.hashed = make(map[certpool.Hash]*certInfo)
	s.names = make(map[string]*list.List)
	s.patterns = make(map[string]*list.List)
}

// lockInit will acquire a lock and initialize the Store if needed
func (s *Store) lockInit() {
	s.mu.Lock()
	if s.hashed == nil {
		s.init()
	}
}

// NewFromBuffer creates a Store from a given PoolBuffer
func NewFromBuffer(pb *legacy.PoolBuffer, base x509utils.CertPool) (*Store, error) {
	s := new(Store)
	s.init()
	if pb != nil {
		certs, err := pb.Certificates(base)
		if err != nil {
			return nil, err
		}

		pb.CopyPool(&s.roots)
		addCerts(s, certs...)
	}
	return s, nil
}

func addCerts(s *Store, certs ...*tls.Certificate) {
	for _, c := range certs {
		addOneCert(s, c)
	}
}

func addOneCert(s *Store, c *tls.Certificate) {
	key, ok := c.PrivateKey.(x509utils.PrivateKey)
	if !ok {
		// drop keyless certificates
		return
	}

	// contains key
	if !core.SliceContainsFn(s.keys, key, PrivateKeyEqual) {
		// new key
		s.keys = append(s.keys, key)
	}

	// contains cert
	hash, ok := certpool.HashCert(c.Leaf)
	if !ok {
		return
	}

	if _, found := s.hashed[hash]; !found {
		// new cert
		names, patterns := x509utils.Names(c.Leaf)

		ci := &certInfo{
			c:        c,
			hash:     hash,
			names:    names,
			patterns: patterns,
		}
		addCertInfo(s, ci)
	}
}

func addCertInfo(s *Store, ci *certInfo) {
	s.hashed[ci.hash] = ci
	s.certs.PushFront(ci)

	for _, name := range ci.names {
		core.MapListAppend(s.names, name, ci.c)
	}
	for _, pattern := range ci.patterns {
		core.MapListAppend(s.patterns, pattern, ci.c)
	}
}
