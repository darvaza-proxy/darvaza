package x509utils

import (
	"crypto/x509"
	"encoding/pem"
)

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

func (pool *CertPool) Pool() *x509.CertPool {
	return pool.getPool()
}

func (pool *CertPool) addCertPEM(filename string, block *pem.Block) bool {

	if block.Type == "CERTIFICATE" {
		// block is cert
		certBytes := block.Bytes
		cert, err := x509.ParseCertificate(certBytes)
		if err == nil {
			pool.pool.AddCert(cert)
		}
	}

	return false
}

func (pool *CertPool) AddCert(s string) error {
	pool.getPool()
	return ReadPEM(s, pool.addCertPEM)
}
