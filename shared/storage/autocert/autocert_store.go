package autocert

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"sync/atomic"
	"time"

	"darvaza.org/darvaza/shared/storage"
	"darvaza.org/darvaza/shared/x509utils"
)

var (
	_ storage.Store = (*Store)(nil)
)

// GetCAPool returns the set of trusted of CAs for tls.Config
func (s *Store) GetCAPool() *x509.CertPool {
	return s.pool.GetCAPool()
}

// GetCertificate finds or acquires a certificate for the given ClientHelloInfo
func (s *Store) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return s.pool.GetCertificateWithCallback(hello, s.tryGetCertificate)
}

func (s *Store) tryGetCertificate(ctx context.Context,
	pk x509utils.PrivateKey, name string) (*tls.Certificate, error) {
	//
	var cert *tls.Certificate
	var err error

	switch {
	case s.getter == nil:
		// issue Certificate
		cert, err = s.issueCertificate(ctx, pk, name)
	case s.timeout < 0:
		// get without timeout
		cert, err = s.fetchCertificate(ctx, pk, name)
	default:
		// try get with timeout
		cert, err = s.fetchCertificateWithTimeout(ctx, pk, name)
	}

	if err != nil {
		// failed, but we can't fail.
		s.warn(err).
			WithField("Host", name).
			Printf("%q: failed to acquire certificate", name)

		// issue certificate
		cert, err = s.issueCertificate(ctx, pk, name)
	}

	if cert != nil {
		// store certificate
		err = s.pool.AddCert(name, cert.Leaf)
	}

	return cert, err
}

func (s *Store) fetchCertificate(ctx context.Context,
	pk x509utils.PrivateKey, name string) (*tls.Certificate, error) {
	//
	cert, err := s.getter(ctx, pk, name)

	switch {
	case cert != nil && err == nil:
		// success
		s.debug().
			WithField("Host", name).
			Printf("%q: certificate %v acquired", name, cert)
		return cert, nil
	case cert == nil && err != nil:
		// regular error
		s.error(err).
			WithField("Host", name).
			Printf("%q: failed to acquire certificate", name)
		return nil, err
	default:
		// both or none
		s.error(err).
			WithField("Host", name).
			Printf("%q: invalid response from getter", name)

		if err == nil {
			err = mkcertErrUnknown(name)
		}
		return nil, err
	}
}

func (s *Store) fetchCertificateWithTimeout(ctx context.Context,
	pk x509utils.PrivateKey, name string) (*tls.Certificate, error) {
	// try acquire certificate
	var atomicErr atomic.Value
	var certCh = make(chan *tls.Certificate, 1)

	go func() {
		defer close(certCh)

		c, err := s.fetchCertificate(ctx, pk, name)

		if c != nil {
			// store even if too late
			err = s.pool.AddCert(name, c.Leaf)
		}

		switch {
		case err != nil:
			// remember error
			atomicErr.Store(err)
		default:
			// emit
			certCh <- c
		}
	}()

	select {
	case c := <-certCh:
		if c != nil {
			// acquired
			return c, nil
		}
	case <-time.After(s.timeout):
	case <-ctx.Done():
	}

	// failed
	return nil, mkcertErrTimeoutUnlessAtomic(name, &atomicErr)
}
