// Package certpool provides a x509 Certificates store from ground up
package certpool

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
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

func (pool *CertPool) addCert(_ string, cert *x509.Certificate) {
	pool.pool.AddCert(cert)
}

func (pool *CertPool) addCertPEM(filename string, block *pem.Block) bool {
	if cert, _ := x509utils.BlockToCertificate(block); cert != nil {
		pool.addCert(filename, cert)
	}
	return false
}

// Add adds certificates to the CertPool
func (pool *CertPool) Add(s string) error {
	pool.getPool()
	return x509utils.ReadStringPEM(s, pool.addCertPEM)
}

//revive:disable:confusing-naming

// AddCert adds parsed certificates to the CertPool
func (pool *CertPool) AddCert(cert *x509.Certificate) {
	//revive:enable:confusing-naming
	pool.getPool()
	pool.addCert("", cert)
}
