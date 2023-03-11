package httpserver

import (
	"context"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (srv *Server) newInsecureServer(h http.Handler) *http.Server {
	h1s := srv.newBaseServer()
	h2s := &http2.Server{}

	// H2C
	h = h2c.NewHandler(h, h2s)
	h1s.Handler = h

	return h1s
}

func (srv *Server) prepareAndSpawnInsecure(lsn net.Listener, h http.Handler) error {
	w := srv.newInsecureServer(h)

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

	return nil
}

func (srv *Server) spawnInsecure(listeners []*net.TCPListener) {
	var h http.Handler

	if srv.cfg.HandleInsecure {
		// same handler as secure then
		h = srv.mux
	} else {
		// only ACME-HTTP-01 and https redirect
		h = srv.newInsecureHandler()
	}

	// Advertise Quic
	h = srv.QuicHeadersMiddleware(h)

	for _, lsn := range listeners {
		srv.wg.Go(func() error {
			return srv.prepareAndSpawnInsecure(lsn, h)
		})
	}
}
