package httpserver

import (
	"context"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

func (srv *Server) newSecureServer(h http.Handler) (*http.Server, error) {
	h1s := srv.newBaseServer()
	h1s.TLSConfig = srv.newTLSServerConfig()
	h1s.Handler = h

	h2s := &http2.Server{}
	if err := http2.ConfigureServer(h1s, h2s); err != nil {
		return nil, err
	}

	return h1s, nil
}

func (srv *Server) prepareAndSpawnSecure(lsn net.Listener, h http.Handler) error {
	w, err := srv.newSecureServer(h)
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

func (srv *Server) spawnSecure(listeners []net.Listener) {
	h := http.Handler(srv.mux)

	// Advertise Quic
	h = srv.QuicHeadersMiddleware(h)

	for _, lsn := range listeners {
		srv.wg.Go(func() error {
			return srv.prepareAndSpawnSecure(lsn, h)
		})
	}
}
