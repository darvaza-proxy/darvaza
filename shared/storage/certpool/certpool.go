// Package certpool provides a x509 Certificates store from ground up
package certpool

import (
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
	hashed map[Hash]*x509.Certificate
}

// Export produces a standard *x509.CertPool containing the
// same CA certificates
func (s *CertPool) Export() *x509.CertPool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.exportUnlocked()
}
