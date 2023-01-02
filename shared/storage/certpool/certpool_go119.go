//go:build !go1.18

package certpool

import "crypto/x509"

func (s *CertPool) exportUnlocked() *x509.CertPool {
	p := s.cached

	if p == nil {
		p = x509.NewCertPool()
		for _, cert := range s.hashed {
			p.AddCert(cert)
		}
		s.cached = p
	}

	return p.Clone()
}
