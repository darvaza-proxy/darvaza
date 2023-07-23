// Package autocert provides a TLS storage that fetches certificates
// and use its own CA as fallback
package autocert

import (
	"context"
	"crypto/x509"
	"sync"
	"time"

	"darvaza.org/darvaza/shared/storage/simple"
	"darvaza.org/darvaza/shared/x509utils"

	"darvaza.org/slog"
)

var (
	_ x509utils.ReadStore  = (*Store)(nil)
	_ x509utils.WriteStore = (*Store)(nil)
)

// Store is a TLS store that will try to acquire new certificates
// externally, or act as CA if it can't
type Store struct {
	mu     sync.Mutex
	logger slog.Logger

	pool    simple.Store
	getter  simple.Getter
	timeout time.Duration
}

// New creates a new Store based on the given Config
func New(cfg *Config) (*Store, error) {
	if cfg == nil {
		cfg = new(Config)
	}

	if err := cfg.SetDefaults(); err != nil {
		return nil, err
	}

	s := &Store{
		logger:  cfg.Logger,
		getter:  cfg.Getter,
		timeout: cfg.Timeout,
	}

	if err := cfg.ApplyFiles(s); err != nil {
		return nil, err
	}

	if err := s.Prepare(); err != nil {
		return nil, err
	}
	return s, nil
}

// Prepare fills any gap in the Store and validates it
func (*Store) Prepare() error {
	// TODO: Implement
	return nil
}

// ForEach iterates over all stored certificates
func (s *Store) ForEach(ctx context.Context, f x509utils.StoreIterFunc) error {
	return s.pool.ForEach(ctx, f)
}

// Get gets from the [Store] a certificate matching the given name
func (s *Store) Get(ctx context.Context, name string) (*x509.Certificate, error) {
	return s.pool.Get(ctx, name)
}

// Delete removes a certificate by name
func (s *Store) Delete(ctx context.Context, name string) error {
	return s.pool.Delete(ctx, name)
}

// DeleteCert removes a certificate from the store
func (s *Store) DeleteCert(ctx context.Context, cert *x509.Certificate) error {
	return s.pool.DeleteCert(ctx, cert)
}
