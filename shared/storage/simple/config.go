package simple

import (
	"darvaza.org/darvaza/shared/storage/certpool"
	"darvaza.org/darvaza/shared/x509utils"
	"darvaza.org/slog"
)

// Config is a custom factory for the Store allowing the usage
// of a Logger and a roots base different that what the system provides
type Config struct {
	Base   x509utils.CertPooler
	Logger slog.Logger
}

// New creates a Store using a list of PEM blocks, filenames, or directories
func New(blocks ...string) (*Store, error) {
	var c Config
	return c.New(blocks...)
}

// New creates a Store using keys and certificates provided as
// files, directories, or direct PEM encoded content
func (c *Config) New(blocks ...string) (*Store, error) {
	pb, err := c.newPoolBuffer(blocks...)
	if err != nil {
		return nil, err
	}

	s, err := NewFromBuffer(pb, c.Base)
	if err != nil {
		return nil, err
	}

	if c.Logger != nil {
		s.SetLogger(c.Logger)
	}

	return s, nil
}

func (c *Config) newPoolBuffer(blocks ...string) (*certpool.PoolBuffer, error) {
	var pb certpool.PoolBuffer

	if c.Logger != nil {
		pb.SetLogger(c.Logger)
	}

	for _, s := range blocks {
		if s != "" {
			if err := pb.Add(s); err != nil {
				return nil, err
			}
		}
	}

	return &pb, nil
}
