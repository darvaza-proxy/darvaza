package x509utils

import (
	"crypto/x509"
	"encoding/pem"
)

// CertPool extends the standard x509.CertPool
type CertPool struct {
	pool *x509.CertPool
}

func (pool *CertPool) getPool() *x509.CertPool {
	p := pool.pool
	if p == nil {
		p, _ = x509.SystemCertPool()
		if p == nil {
			p = x509.NewCertPool()
		}
		pool.pool = p
	}
	return p
}

// Pool returns a reference to our internal x509.CertPool
func (pool *CertPool) Pool() *x509.CertPool {
	return pool.getPool()
}

func (pool *CertPool) addCert(filename string, cert *x509.Certificate) {
	pool.pool.AddCert(cert)
}

func (pool *CertPool) addCertPEM(filename string, block *pem.Block) bool {

	if block.Type == "CERTIFICATE" {
		// block is cert
		certBytes := block.Bytes
		cert, err := x509.ParseCertificate(certBytes)
		if err == nil {
			pool.addCert(filename, cert)
		}
	}

	return false
}

// Add adds certificates to the CertPool
func (pool *CertPool) Add(s string) error {
	pool.getPool()
	return ReadPEM(s, pool.addCertPEM)
}

// AddCert adds parsed certificates to the CertPool
func (pool *CertPool) AddCert(cert *x509.Certificate) {
	pool.getPool()
	pool.addCert("", cert)
}
