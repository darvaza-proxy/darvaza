package httpgroup

import "github.com/darvaza-proxy/slog"

func (heg *Group) withLogger(level slog.LogLevel) (slog.Logger, bool) {
	l, ok := heg.logger.Load().(slog.Logger)
	if !ok {
		return nil, false
	}

	return l.WithLevel(level).WithEnabled()
}

func (heg *Group) debug() (slog.Logger, bool) {
	return heg.withLogger(slog.Debug)
}

func (heg *Group) error(err error) (slog.Logger, bool) {
	if l, ok := heg.withLogger(slog.Error); ok {
		if err != nil {
			l = l.WithField(slog.ErrorFieldName, err)
		}
		return l, true
	}
	return nil, false
}
