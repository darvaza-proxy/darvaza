// Package certpool provides a x509 Certificates store from ground up
package certpool

import (
	"container/list"
	"crypto/x509"
	"sync"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/x509utils"
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
	return s.Copy(nil)
}

// Copy replicate itself into a given CertPool
func (s *CertPool) Copy(out *CertPool) *CertPool {
	if out == nil {
		out = new(CertPool)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	*out = preClone(s)

	for hash, d := range s.hashed {
		names := make([]string, len(d.names))
		patterns := make([]string, len(d.patterns))

		copy(names, d.names)
		copy(patterns, d.patterns)

		out.hashed[hash] = &certPoolEntry{
			hash:     hash,
			cert:     d.cert,
			names:    names,
			patterns: patterns,
		}
	}

	return out
}

func preClone(s *CertPool) CertPool {
	return CertPool{
		cached:   s.exportUnlocked(),
		hashed:   make(map[Hash]*certPoolEntry, len(s.hashed)),
		names:    core.MapListCopy(s.names),
		patterns: core.MapListCopy(s.patterns),
		subjects: core.MapListCopy(s.subjects),
	}
}

// Count tells how many certificates are stored in the CertPool
func (s *CertPool) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.hashed)
}

// IsCA tells if all certificates in the store are CAs
func (s *CertPool) IsCA() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, e := range s.hashed {
		if !e.cert.IsCA {
			return false
		}
	}
	return true
}

// Export produces a standard *x509.CertPool containing the
// same CA certificates
func (s *CertPool) Export() *x509.CertPool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.exportUnlocked()
}

func (s *CertPool) exportUnlocked() *x509.CertPool {
	p := s.cached

	if p == nil {
		p = x509.NewCertPool()
		for _, cert := range s.hashed {
			p.AddCert(cert.cert)
		}
		s.cached = p
	}

	return p.Clone()
}

// Reset removes all certificates from the Pool
func (s *CertPool) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.init()
}
