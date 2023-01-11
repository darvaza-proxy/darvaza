package gnocco

import (
	"github.com/darvaza-proxy/slog"

	"github.com/darvaza-proxy/gnocco/shared/cblog"
)

// Logger returns the internal slog.Logger
func (cf *Gnocco) Logger() slog.Logger {
	return cf.logger
}

func newLogger(cf *Gnocco) slog.Logger {
	logger := cblog.New()

	if cf.Log.Stdout {
		logger.SetLogger("console", nil)
	}

	if cf.Log.File != "" {
		cfg := map[string]interface{}{"file": cf.Log.File}
		logger.SetLogger("file", cfg)
		logger.Info().Print("Logger started")
	}

	return logger
}
