package httpserver

import (
	"net"

	"darvaza.org/slog"
)

func (srv *Server) error(err error) slog.Logger {
	log := srv.cfg.Logger.Error()
	if err != nil {
		log = log.WithField(slog.ErrorFieldName, err)
	}
	return log
}

func (srv *Server) warn(err error) slog.Logger {
	log := srv.cfg.Logger.Warn()
	if err != nil {
		log = log.WithField(slog.ErrorFieldName, err)
	}
	return log
}

func (srv *Server) info() slog.Logger {
	return srv.cfg.Logger.Info()
}

func (srv *Server) withInfo() (slog.Logger, bool) {
	return srv.cfg.Logger.Info().WithEnabled()
}

func (srv *Server) debug() slog.Logger {
	return srv.cfg.Logger.Debug()
}

func (srv *Server) withDebug() (slog.Logger, bool) {
	return srv.cfg.Logger.Debug().WithEnabled()
}

func (srv *Server) logListening(scheme string, addr net.Addr) {
	if log, ok := srv.withInfo(); ok {
		for _, addr := range getStringAddrPort(addr) {
			log.Printf("%s %s://%s", "listening", scheme, addr)
		}
	}
}

func (srv *Server) logClosing(scheme string, addr net.Addr) {
	if log, ok := srv.withInfo(); ok {
		for _, addr := range getStringAddrPort(addr) {
			log.Printf("%s %s://%s", "closing", scheme, addr)
		}
	}
}
