package certpool

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

// AppendCertsFromPEM adds certificates to the Pool from a PEM encoded blob,
// and returns true if a new Certificate was effectivelt added
func (s *CertPool) AppendCertsFromPEM(b []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	var added bool

	x509utils.ReadPEM(b, func(_ string, block *pem.Block) bool {
		if cert, _ := x509utils.BlockToCertificate(block); cert != nil && cert.IsCA {
			if s.addCertUnsafe(HashCert(cert), cert) {
				added = true
			}
		}
		return false // continue
	})

	return added
}

// AddCert adds parsed certificates to the CertPool
func (s *CertPool) AddCert(cert *x509.Certificate) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cert != nil && cert.IsCA {
		return s.addCertUnsafe(HashCert(cert), cert)
	}
	return false
}

func (s *CertPool) addCertUnsafe(hash Hash, cert *x509.Certificate) bool {
	var added bool

	if s.hashed == nil {
		s.hashed = make(map[Hash]*x509.Certificate)
	}

	if _, ok := s.hashed[hash]; !ok {
		s.cached = nil // invalidate cache
		s.hashed[hash] = cert
		added = true
	}

	return added
}
