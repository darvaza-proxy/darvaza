package gnocco

import (
	"darvaza.org/slog"

	"darvaza.org/darvaza/shared/cblog"
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
		cfg := map[string]any{"file": cf.Log.File}
		logger.SetLogger("file", cfg)
		logger.Info().Print("Logger started")
	}

	return logger
}
