package autocert

import (
	"context"
	"crypto/x509"
)

// AddCACert adds CA certificates to the store
func (s *Store) AddCACert(data string) error {
	_, err := s.addCertString(data, true)
	return err
}

// AddCert adds certificates to the store
func (s *Store) AddCert(data string) error {
	_, err := s.addCertString(data, false)
	return err
}

// Put adds a certificate to the store
func (s *Store) Put(_ context.Context, name string, cert *x509.Certificate) error {
	return s.pool.AddCert(name, cert)
}
