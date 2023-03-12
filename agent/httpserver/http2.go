package httpserver

import (
	"context"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

// NewH2Server creates a new HTTP/2 capable http.Server
func (srv *Server) NewH2Server(h http.Handler) (*http.Server, error) {
	h1s := srv.NewHTTPServer()
	h1s.TLSConfig = srv.NewTLSConfig()
	h1s.Handler = h

	h2s := &http2.Server{}
	if err := http2.ConfigureServer(h1s, h2s); err != nil {
		return nil, err
	}

	return h1s, nil
}

// NewH2Handler returns the http.Handler to use on the H2 server
func (srv *Server) NewH2Handler() http.Handler {
	h := http.Handler(srv.mux)

	// Advertise Quic
	h = srv.QuicHeadersMiddleware(h)

	return h
}

func (srv *Server) prepareAndSpawnH2(lsn net.Listener, h http.Handler) error {
	w, err := srv.NewH2Server(h)
	if err != nil {
		return err
	}

	addr := lsn.Addr()

	srv.wg.Go(func() error {
		srv.logListening("https", addr)
		return w.Serve(lsn)
	})

	srv.wg.Go(func() error {
		<-srv.ctx.Done()
		srv.logClosing("https", addr)
		return w.Shutdown(context.Background())
	})

	return nil
}

func (srv *Server) spawnH2(listeners []net.Listener) {
	h := srv.NewH2Handler()

	for _, lsn := range listeners {
		srv.wg.Go(func() error {
			return srv.prepareAndSpawnH2(lsn, h)
		})
	}
}
