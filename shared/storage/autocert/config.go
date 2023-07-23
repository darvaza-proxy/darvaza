package autocert

import (
	"time"

	"darvaza.org/darvaza/shared/storage/simple"
	"darvaza.org/slog"
)

const (
	// DefaultGetTimeout is how long we wait a Getter
	// if Timeout is set to 0ms
	DefaultGetTimeout = 1 * time.Second
)

// Config indicates how the Store will be configured
type Config struct {
	// Logger to be attached to the Store
	Logger slog.Logger

	// Getter is a helper called when we don't have the
	// required certificate
	Getter simple.Getter

	// Timeout indicates how long to wait for Getter before
	// issuing its own. Negative means no timeout, and zero
	// milliseconds falls back to a 1s default
	Timeout time.Duration

	Keys   string
	Certs  string
	CAKey  string
	CACert string
	Roots  string
}

// SetDefaults fills the gaps on a Config
func (cfg *Config) SetDefaults() error {
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}

	// round up to milliseconds
	t := (cfg.Timeout*1000 + 999) / 1000
	switch {
	case t == 0:
		cfg.Timeout = DefaultGetTimeout
	case t < 0:
		cfg.Timeout = -1
	default:
		cfg.Timeout = t
	}

	return nil
}

// ApplyFiles applies keys and certs on the Config to the
// given Store
func (cfg *Config) ApplyFiles(s *Store) error {
	if _, err := s.addKeyString(cfg.Keys, false); err != nil {
		return err
	}

	if _, err := s.addKeyString(cfg.CAKey, true); err != nil {
		return err
	}

	if _, err := s.addCertString(cfg.Roots, true); err != nil {
		return err
	}

	if _, err := s.addCertString(cfg.Certs, true); err != nil {
		return err
	}

	if _, err := s.addCertString(cfg.Certs, false); err != nil {
		return err
	}

	return nil
}
