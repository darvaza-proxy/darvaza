// Package certpool provides a x509 Certificates store from ground up
package certpool

import (
	"container/list"
	"crypto/x509"
	"sync"

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

// Export produces a standard *x509.CertPool containing the
// same CA certificates
func (s *CertPool) Export() *x509.CertPool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.exportUnlocked()
}
