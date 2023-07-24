package simple

import "crypto/x509"

// AddCACert adds a CA Certificate to the Store
func (s *Store) AddCACert(cert *x509.Certificate) error {
	if cert == nil || !cert.IsCA {
		return ErrInvalidCert{
			Reason: "not a CA",
		}
	}

	s.lockInit()
	defer s.mu.Unlock()

	s.roots.AddCert(cert)
	return nil
}
