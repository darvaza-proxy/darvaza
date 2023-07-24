package simple

import (
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

// SetLogger attaches a logger to the Store
func (s *Store) SetLogger(logger slog.Logger) {
	if logger == nil {
		logger = defaultLogger()
	}

	s.logger = logger
}

func defaultLogger() slog.Logger {
	return discard.New()
}
