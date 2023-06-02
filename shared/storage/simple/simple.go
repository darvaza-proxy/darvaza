package simple

import (
	"container/list"
	"context"
	"crypto/tls"
	"sync"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/storage/certpool"
	"darvaza.org/darvaza/shared/x509utils"
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

	pool     *certpool.CertPool
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

func newStore(roots *certpool.CertPool) *Store {
	if roots == nil {
		roots = new(certpool.CertPool)
		roots.Reset()
	}

	return &Store{
		pool:     roots,
		keys:     []x509utils.PrivateKey{},
		certs:    list.New(),
		hashed:   make(map[certpool.Hash]*certInfo),
		names:    make(map[string]*list.List),
		patterns: make(map[string]*list.List),
	}
}

// NewFromBuffer creates a Store from a given PoolBuffer
func NewFromBuffer(pb *certpool.PoolBuffer, base x509utils.CertPooler) (*Store, error) {
	certs, err := pb.Certificates(base)
	if err != nil {
		return nil, err
	}

	s := newStore(pb.Pool())
	addCerts(s, certs...)
	return s, nil
}

func addCerts(s *Store, certs ...*tls.Certificate) {
	for _, c := range certs {
		key, ok := c.PrivateKey.(x509utils.PrivateKey)
		if !ok {
			// drop keyless certificates
			continue
		}

		// contains key
		if !core.SliceContainsFn(s.keys, key, PrivateKeyEqual) {
			// new key
			s.keys = append(s.keys, key)
		}

		// contains cert
		hash := certpool.HashCert(c.Leaf)
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
