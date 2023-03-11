package httpserver

import (
	"net/http"
)

func (srv *Server) newBaseServer() *http.Server {
	return &http.Server{
		ReadTimeout:       srv.cfg.ReadTimeout,
		ReadHeaderTimeout: srv.cfg.ReadHeaderTimeout,
		WriteTimeout:      srv.cfg.WriteTimeout,
		IdleTimeout:       srv.cfg.IdleTimeout,
	}
}
