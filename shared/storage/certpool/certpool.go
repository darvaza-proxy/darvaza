// Package certpool provides a x509 Certificates store from ground up
package certpool

import (
	"container/list"
	"crypto/x509"
	"sync"

	"github.com/darvaza-proxy/darvaza/shared/x"
	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

var (
	_ x509utils.CertPooler     = (*CertPool)(nil)
	_ x509utils.CertPoolWriter = (*CertPool)(nil)
)

// CertPool represents a collection of CA Certificates
type CertPool struct {
	mu sync.RWMutex

	cached *x509.CertPool
	hashed map[Hash]*certPoolEntry

	names    map[string]*list.List
	patterns map[string]*list.List
	subjects map[string]*list.List
}

type certPoolEntry struct {
	hash     Hash
	cert     *x509.Certificate
	names    []string
	patterns []string
}

// init reinitialises the CertPool
func (s *CertPool) init() {
	s.cached = nil
	s.hashed = make(map[Hash]*certPoolEntry)
	s.names = make(map[string]*list.List)
	s.patterns = make(map[string]*list.List)
	s.subjects = make(map[string]*list.List)
}

// Clone creates a copy of the CertPool
func (s *CertPool) Clone() x509utils.CertPooler {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := &CertPool{
		cached:   s.exportUnlocked(),
		hashed:   make(map[Hash]*certPoolEntry, len(s.hashed)),
		names:    x.CloneMapList(s.names),
		patterns: x.CloneMapList(s.patterns),
		subjects: x.CloneMapList(s.subjects),
	}

	for hash, d := range s.hashed {
		names := make([]string, len(d.names))
		patterns := make([]string, len(d.patterns))

		copy(names, d.names)
		copy(patterns, d.patterns)

		clone.hashed[hash] = &certPoolEntry{
			hash:     hash,
			cert:     d.cert,
			names:    names,
			patterns: patterns,
		}
	}

	return clone
}

// Count tells how many certificates are stored in the CertPool
func (s *CertPool) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.hashed)
}

// Export produces a standard *x509.CertPool containing the
// same CA certificates
func (s *CertPool) Export() *x509.CertPool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.exportUnlocked()
}

// Reset removes all certificates from the Pool
func (s *CertPool) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.init()
}
