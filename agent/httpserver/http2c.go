package httpserver

import (
	"context"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// NewHTTPServer creates a new http.Server
func (srv *Server) NewHTTPServer() *http.Server {
	return &http.Server{
		ReadTimeout:       srv.cfg.ReadTimeout,
		ReadHeaderTimeout: srv.cfg.ReadHeaderTimeout,
		WriteTimeout:      srv.cfg.WriteTimeout,
		IdleTimeout:       srv.cfg.IdleTimeout,
	}
}

// NewH2CServer creates a new H2C capable http.Server
func (srv *Server) NewH2CServer(h http.Handler) *http.Server {
	h1s := srv.NewHTTPServer()
	h2s := &http2.Server{}

	h1s.Handler = h2c.NewHandler(h, h2s)
	return h1s
}

// NewH2CHandler returns the http.Handler to use on the H2C server
func (srv *Server) NewH2CHandler() http.Handler {
	var h http.Handler

	if srv.cfg.HandleInsecure {
		// same handler as secure then
		h = srv.mux
	} else {
		// only ACME-HTTP-01 and https redirect
		h = srv.NewHTTPSRedirectHandler()
	}

	// Advertise Quic
	h = srv.QuicHeadersMiddleware(h)

	return h
}

func (srv *Server) spawnH2C(listeners []*net.TCPListener) {
	h := srv.NewH2CHandler()

	for _, lsn := range listeners {
		w := srv.NewH2CServer(h)
		addr := lsn.Addr()

		srv.wg.Go(func() error {
			srv.logListening("http", addr)
			return w.Serve(lsn)
		})
		srv.wg.Go(func() error {
			<-srv.ctx.Done()
			srv.logClosing("http", addr)
			return w.Shutdown(context.Background())
		})
	}
}
