package storage

import (
	"crypto/tls"
	"crypto/x509"
	"syscall"
)

// Store represents the most fundamental interface as needed to setup tls.Config{}
type Store interface {
	GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error)
	GetCAPool() *x509.CertPool
}

// SetupTLS configures a given tls.Config to use a given Store
func SetupTLS(conf *tls.Config, store Store) error {
	if conf != nil && store != nil {
		if pool := store.GetCAPool(); pool != nil {
			conf.GetCertificate = store.GetCertificate
			conf.RootCAs = pool
			conf.ClientCAs = pool
			return nil
		}
	}
	return syscall.EINVAL
}

// NewTLSConfig returns a tls.Config configured to use the given Store
func NewTLSConfig(store Store) (*tls.Config, error) {
	conf := &tls.Config{}
	if err := SetupTLS(conf, store); err != nil {
		return nil, err
	}
	return conf, nil
}
