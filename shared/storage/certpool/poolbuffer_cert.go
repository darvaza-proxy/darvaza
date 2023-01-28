package certpool

import (
	"crypto/x509"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

type pbCerts struct {
	certs map[Hash]*pbCertData
	pool  *CertPool
}

type pbCertData struct {
	Filename string
	Cert     *x509.Certificate

	Hash Hash
	Pub  x509utils.PublicKey
}

func (m *pbCerts) Reset() {
	m.certs = nil
	m.pool = nil
}

func (m *pbCerts) AddCert(fn string, cert *x509.Certificate) bool {
	if cert == nil {
		// NOOP
		return false
	} else if m.certs == nil {
		// first
		m.certs = make(map[Hash]*pbCertData)
	}

	hash := HashCert(cert)
	if _, ok := m.certs[hash]; ok {
		// duplicate
		return false
	}

	// new
	m.certs[hash] = &pbCertData{
		Filename: fn,
		Cert:     cert,
		Hash:     hash,
		Pub:      cert.PublicKey.(x509utils.PublicKey),
	}
	m.pool = nil
	return true
}

func (m *pbCerts) Pool() *CertPool {
	pool := m.pool
	if pool == nil {
		pool = &CertPool{}
		if m.certs != nil {
			for _, cd := range m.certs {
				pool.addCertUnsafe(cd.Hash, "", cd.Cert)
			}
		}
		m.pool = pool
	}
	return pool
}

func (m *pbCerts) Count() int {
	return len(m.pool.hashed)
}

func (m *pbCerts) Export() *x509.CertPool {
	return m.Pool().Export()
}

func (pb *PoolBuffer) addCertUnlocked(fn string, cert *x509.Certificate) error {
	var pool *pbCerts

	if err := pb.printCert(fn, cert); err != nil {
		return err
	}

	if cert.IsCA {
		pool = &pb.roots
	} else {
		pool = &pb.certs
	}

	pool.AddCert(fn, cert)
	return nil
}
