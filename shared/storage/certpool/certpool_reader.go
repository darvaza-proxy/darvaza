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

// Equal checks if another CertPooler is equal to this one
func (s *CertPool) Equal(x x509utils.CertPooler) bool {
	if x == nil {
		// s != nil
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if b, ok := x.(*CertPool); ok {
		// same type
		b.mu.RLock()
		defer b.mu.RUnlock()

		return equal(s, b)
	}

	// use StoreReader interface
	hashed, count := len(s.hashed), 0
	ctx := context.Background()

	iter := func(cert *x509.Certificate) error {
		// not larger
		count++
		if count > hashed {
			return fs.ErrNotExist
		}

		// and all certs there exist here
		h := HashCert(cert)
		if _, found := s.hashed[h]; !found {
			return fs.ErrNotExist
		}

		// continue
		return nil
	}

	if err := x.ForEach(ctx, iter); err != nil {
		// failed
		return false
	}
	// same size
	return count == hashed
}

func equal(a *CertPool, b *CertPool) bool {
	// same size
	if len(a.hashed) != len(b.hashed) {
		return false
	}
	// and all hashes here exist there
	for k := range a.hashed {
		if _, ok := b.hashed[k]; !ok {
			return false
		}
	}

	return true
}

// Minus produces a new CertPool without any certificate on the given Pool
func (s *CertPool) Minus(x x509utils.CertPooler) x509utils.CertPooler {
	out := s.Clone().(*CertPool)

	if x == nil {
		// nothing to remove
		return out
	}

	if b, ok := x.(*CertPool); ok {
		// same type
		b.mu.RLock()
		defer b.mu.RUnlock()

		return minus(out, b)
	}

	// use StoreReader interface
	ctx := context.Background()
	x.ForEach(ctx, func(cert *x509.Certificate) error {
		out.deleteHash(HashCert(cert))
		return nil
	})

	return out
}

func minus(out, b *CertPool) *CertPool {
	for hash := range b.hashed {
		_ = out.deleteHash(hash)
	}
	return out
}

// Plus produces a new CertPool with all certificate on the given Pool
func (s *CertPool) Plus(x x509utils.CertPooler) x509utils.CertPooler {
	out := s.Clone().(*CertPool)
	if x != nil {
		if b, ok := x.(*CertPool); ok {
			// same type
			b.mu.RLock()
			defer b.mu.RUnlock()

			return plus(out, b)
		}

		// use StoreReader interface
		ctx := context.Background()
		x.ForEach(ctx, func(cert *x509.Certificate) error {
			out.addCertUnsafe(HashCert(cert), "", cert)
			return nil
		})
	}

	return out
}

func plus(out, b *CertPool) *CertPool {
	for hash, cert := range b.hashed {
		out.addCertUnsafe(hash, "", cert.cert)
	}
	return out
}
