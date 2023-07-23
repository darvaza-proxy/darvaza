package autocert

import (
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

func defaultLogger() slog.Logger {
	return discard.New()
}

// SetLogger attaches an [slog.Logger] to the store.
// if nil a [discard.Logger] will be used.
func (s *Store) SetLogger(l slog.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if l == nil {
		l = defaultLogger()
	}
	s.logger = l
}

func (s *Store) error(err error) slog.Logger {
	l := s.logger.Error()
	if err != nil {
		l = l.WithField(slog.ErrorFieldName, err)
	}
	return l
}

func (s *Store) warn(err error) slog.Logger {
	l := s.logger.Warn()
	if err != nil {
		l = l.WithField(slog.ErrorFieldName, err)
	}
	return l
}

func (s *Store) info() slog.Logger {
	return s.logger.Info()
}

func (s *Store) debug() slog.Logger {
	return s.logger.Debug()
}

func (s *Store) withInfo() (slog.Logger, bool) {
	return s.logger.Info().WithEnabled()
}

func (s *Store) withDebug() (slog.Logger, bool) {
	return s.logger.Debug().WithEnabled()
}
