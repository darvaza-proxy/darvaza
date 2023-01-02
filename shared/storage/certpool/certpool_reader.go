package certpool

import (
	"context"
	"crypto/x509"
	"io/fs"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

// ForEach iterates over all certificates
func (s *CertPool) ForEach(ctx context.Context, fn x509utils.StoreIterFunc) error {
	if fn == nil {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	for _, cert := range s.Certs() {
		term, err := iterStep(ctx, fn, cert)
		if term {
			return err
		}
	}
	return nil
}

func iterStep(ctx context.Context, fn x509utils.StoreIterFunc, cert *x509.Certificate) (
	term bool, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
		term = true
	default:
		if err = fn(cert); err != nil {
			term = true
		}
	}
	return term, err
}

// Certs returns an array of all certificates in the CertPool
func (s *CertPool) Certs() []*x509.Certificate {
	s.mu.Lock()
	q := make([]*x509.Certificate, 0, len(s.hashed))
	for _, p := range s.hashed {
		q = append(q, p.cert)
	}
	s.mu.Unlock()

	return q
}

// Get find a certificate by name
func (s *CertPool) Get(_ context.Context, name string) (*x509.Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e := s.getFirstByName(name); e != nil {
		return e.cert, nil
	}

	return nil, fs.ErrNotExist
}
