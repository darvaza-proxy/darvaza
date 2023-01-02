//go:build go1.18

package certpool

import "crypto/x509"

func (s *CertPool) exportUnlocked() *x509.CertPool {
	p := x509.NewCertPool()
	for _, cert := range s.hashed {
		p.AddCert(cert)
	}

	return p
}
