package httpserver

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"darvaza.org/core"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

const (
	// AltSvcHeader is the header label used to advertise
	// Quic support
	AltSvcHeader = "Alt-Svc"

	// GrabQuicHeadersRetry indicates how long we wait to
	// grab the generated Alt-Svc header
	GrabQuicHeadersRetry = 10 * time.Millisecond
)

// NewQuicConfig returns the quic.Config to be used on the
// HTTP/3 server
func (*Server) NewQuicConfig() *quic.Config {
	return &quic.Config{}
}

// prepareQuicListeners wraps a set of UDP listeners to
// be used for the HTTP/3 server
func (srv *Server) prepareQuicListeners(listeners []*net.UDPConn) ([]*quic.EarlyListener, error) {
	var out []*quic.EarlyListener

	if len(listeners) > 0 {
		config := srv.NewQuicConfig()
		tlsConf := srv.NewTLSConfig()
		tlsConf = http3.ConfigureTLSConfig(tlsConf)

		for _, conn := range listeners {
			lsn, err := quic.ListenEarly(conn, tlsConf, config)
			if err != nil {
				return out, err
			}

			out = append(out, lsn)
		}
	}

	return out, nil
}

// prepareAndSpawnH3 binds a quic.EarlyListener to an
// HTTP/3 server and spawn the corresponding worker
func (srv *Server) prepareAndSpawnH3(lsn *quic.EarlyListener) error {
	h3s := &http3.Server{
		Addr:    lsn.Addr().String(),
		Handler: srv.mux,
	}

	addr := lsn.Addr()

	srv.wg.GoCatch(func() error {
		srv.logListening("quic", addr)
		return h3s.ServeListener(lsn)
	}, func(err error) error {
		switch err {
		case nil, quic.ErrServerClosed:
			srv.debug().Printf("%s: done", addr)
			err = nil
		default:
			srv.error(err).Printf("%s: failed", addr)
		}

		return err
	})

	srv.wg.Go(func() error {
		<-srv.ctx.Done()
		srv.logClosing("quic", addr)
		// TODO: h3s.CloseGracefully isn't implemented yet
		srv.warn(nil).Printf("%s: closing UDP listener abruptly", addr)
		return h3s.Close()
	})

	srv.wg.Go(func() error {
		return srv.grabQuicHeaders(srv.ctx, h3s)
	})

	return nil
}

func (srv *Server) spawnH3(listeners []*quic.EarlyListener) {
	for _, lsn := range listeners {
		srv.wg.Go(func() error {
			return srv.prepareAndSpawnH3(lsn)
		})
	}
}

// SetQuicHeaders appends Quic's Alt-Svc to the headers
func (srv *Server) SetQuicHeaders(hdr http.Header) error {
	if s := srv.getQuicAltSvc(); s != "" {
		hdr["Alt-Svc"] = append(hdr["Alt-Svc"], s)
	}
	return http3.ErrNoAltSvcPort
}

// QuicHeadersMiddleware creates is a middleware function
// that injects Alt-Svc on the http.Response headers
func (srv *Server) QuicHeadersMiddleware(next http.Handler) http.Handler {
	h := func(rw http.ResponseWriter, req *http.Request) {
		_ = srv.SetQuicHeaders(rw.Header())
		next.ServeHTTP(rw, req)
	}

	return http.HandlerFunc(h)
}

// grabQuicHeader tries periodically to grab the Alt-Svc headers corresponding
// to a server until it succeeds or the given context is cancelled
func (srv *Server) grabQuicHeaders(ctx context.Context, h3s *http3.Server) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(GrabQuicHeadersRetry):
			hdr := make(http.Header)

			if err := h3s.SetQuicHeaders(hdr); err == nil {
				// success
				srv.appendQuicHeaders(hdr[AltSvcHeader])
				return nil
			}
		}
	}
}

func (srv *Server) appendQuicHeaders(altSvcs []string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	s := strings.Split(",", srv.quicAltSvc)
	for i, hdr := range altSvcs {
		srv.debug().Printf("%s[%v]: %s", AltSvcHeader, i, hdr)

		for _, part := range strings.Split(",", hdr) {
			if !core.SliceContains(s, part) {
				s = append(s, part)
			}
		}
	}

	srv.quicAltSvc = strings.Join(s, ",")
}

func (srv *Server) getQuicAltSvc() string {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	return srv.quicAltSvc
}
