package simple

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/fs"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
)

var (
	_ x509utils.CertPool = (*Store)(nil)
)

// revive:disable:cognitive-complexity

// Get gets from the [Store] a certificate matching the given name
func (s *Store) Get(_ context.Context, name string) (*x509.Certificate, error) {
	// revive:enable:cognitive-complexity
	var out *x509.Certificate

	if name, ok := x509utils.SanitizeName(name); ok {
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

// ForEach iterates over all stored certificates
//
//revive:disable:cognitive-complexity
func (s *Store) ForEach(ctx context.Context, f func(context.Context, *x509.Certificate) bool) {
	//revive:enable:cognitive-complexity
	if f != nil {
		s.lockInit()

		core.ListForEach(s.certs, func(ci *certInfo) bool {
			ok := true

			if ci.c.Leaf != nil {
				s.mu.Unlock()
				ok = f(ctx, ci.c.Leaf)
				s.mu.Lock()
			}

			select {
			case <-ctx.Done():
				return true
			default:
				return !ok
			}
		})
		s.mu.Unlock()
	}
}

// Clone ...
func (s *Store) Clone() x509utils.CertPool { return s }

// Export ...
func (*Store) Export() *x509.CertPool { panic(core.ErrTODO) }
