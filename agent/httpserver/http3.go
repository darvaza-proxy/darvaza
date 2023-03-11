package httpserver

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
)

const (
	// AltSvcHeader is the header label used to advertise
	// Quic support
	AltSvcHeader = "Alt-Svc"
)

func (*Server) newQuicConfig() *quic.Config {
	return &quic.Config{}
}

func (srv *Server) newTLSServerHTTP3Config() *tls.Config {
	tlsConf := srv.newTLSServerConfig()
	tlsConf = http3.ConfigureTLSConfig(tlsConf)
	return tlsConf
}

func (*Server) getQuicAltSvc() string {
	// TODO: implement
	return ""
}

// prepareQuicListeners wraps a set of UDP listeners to
// be used for the HTTP/3 server
func (srv *Server) prepareQuicListeners(listeners []*net.UDPConn) ([]quic.EarlyListener, error) {
	var out []quic.EarlyListener

	if len(listeners) > 0 {
		config := srv.newQuicConfig()
		tlsConf := srv.newTLSServerHTTP3Config()

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

// prepareAndSpawnQuic binds a quic.EarlyListener to an
// HTTP/3 server and spawn the corresponding worker
func (srv *Server) prepareAndSpawnQuic(lsn quic.EarlyListener, h http.Handler) error {
	h3s := &http3.Server{
		Handler: h,
	}

	addr := lsn.Addr()

	srv.wg.Go(func() error {
		srv.logListening("https+quic", addr)
		return h3s.ServeListener(lsn)
	})

	srv.wg.Go(func() error {
		<-srv.ctx.Done()
		srv.logClosing("https+quic", addr)
		return h3s.CloseGracefully(srv.cfg.GracefulTimeout)
	})

	return nil
}

func (srv *Server) spawnQuic(listeners []quic.EarlyListener) {
	h := srv.mux

	for _, lsn := range listeners {
		srv.wg.Go(func() error {
			return srv.prepareAndSpawnQuic(lsn, h)
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
