package simple

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/fs"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/x509utils"
)

var (
	_ x509utils.ReadStore = (*Store)(nil)
)

// revive:disable:cognitive-complexity

// Get gets from the [Store] a certificate matching the given name
func (s *Store) Get(_ context.Context, name string) (*x509.Certificate, error) {
	// revive:enable:cognitive-complexity
	var out *x509.Certificate

	if name, ok := x509utils.SanitiseName(name); ok {
		s.lockInit()
		defer s.mu.Unlock()

		// IP
		if n, ok := x509utils.NameAsIP(name); ok {
			core.MapListForEach(s.names, n, func(c *tls.Certificate) bool {
				out = c.Leaf
				return out != nil
			})
		}

		if out != nil {
			// found
			return out, nil
		}

		// Name
		core.MapListForEach(s.names, name, func(c *tls.Certificate) bool {
			out = c.Leaf
			return out != nil
		})

		if out != nil {
			// found
			return out, nil
		}

		// Wildcard
		if suffix, ok := x509utils.NameAsSuffix(name); ok {
			core.MapListForEach(s.patterns, suffix, func(c *tls.Certificate) bool {
				out = c.Leaf
				return out != nil
			})
		}

		if out != nil {
			// found
			return out, nil
		}
	}

	return nil, fs.ErrNotExist
}

// revive:disable:cognitive-complexity

// ForEach iterates over all stored certificates
func (s *Store) ForEach(ctx context.Context, f x509utils.StoreIterFunc) error {
	// revive:enable:cognitive-complexity
	var err error

	if f != nil {
		s.lockInit()

		core.ListForEach(s.certs, func(ci *certInfo) bool {
			if ci.c.Leaf != nil {
				s.mu.Unlock()
				err = f(ci.c.Leaf)
				s.mu.Lock()
			}

			select {
			case <-ctx.Done():
				err = ctx.Err()
				return true
			default:
				return err != nil
			}
		})
		s.mu.Unlock()
	}

	return err
}
