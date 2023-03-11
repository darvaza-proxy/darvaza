package httpserver

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/darvaza-proxy/darvaza/shared/tls/sni"
)

func (srv *Server) prepareSecureListeners(listeners []*net.TCPListener) []net.Listener {
	var out []net.Listener

	if l := len(listeners); l > 0 {
		tlsServerConfig := srv.newTLSServerConfig()
		rtio := srv.getReadHeaderTimeout()

		out = make([]net.Listener, 0, l)

		for _, tcpLsn := range listeners {
			var lsn net.Listener

			// sni.Dispatcher
			lsn = srv.applySNIDispatcher(tcpLsn, rtio)
			// tls.Listener
			lsn = tls.NewListener(lsn, tlsServerConfig)

			out = append(out, lsn)
		}
	}

	return out
}

func (srv *Server) newTLSServerConfig() *tls.Config {
	conf := srv.cfg.TLSConfig
	if conf == nil {
		conf = &tls.Config{}
	}

	if conf.GetCertificate == nil {
		conf.GetCertificate = srv.cfg.GetCertificate
	}
	if conf.GetConfigForClient == nil {
		conf.GetConfigForClient = srv.cfg.GetConfigForClient
	}

	if conf.ClientCAs == nil && srv.cfg.GetClientCAs != nil {
		// mTLS
		conf.ClientCAs = srv.cfg.GetClientCAs()
	}
	if conf.RootCAs == nil && srv.cfg.GetRootCAs != nil {
		conf.RootCAs = srv.cfg.GetRootCAs()
	}
	return conf
}

func (srv *Server) applySNIDispatcher(lsn net.Listener, rtio time.Duration) net.Listener {
	if cb := srv.cfg.GetHandlerForClient; cb != nil {
		d := &sni.Dispatcher{
			Logger:  srv.log,
			Context: srv.ctx,

			GetHandler: cb,
			OnError: func(err error) bool {
				srv.Fail(err)
				return true
			},
			OnAccept: func(conn net.Conn) (net.Conn, error) {
				_ = conn.SetReadDeadline(getTimeout(rtio))
				return conn, nil
			},
		}

		srv.spawnDispatcher(d, lsn)
		return d
	}
	return lsn
}

func (srv *Server) spawnDispatcher(d *sni.Dispatcher, lsn net.Listener) {
	addr := lsn.Addr()

	srv.wg.Go(func() error {
		srv.logListening("tls", addr)
		return d.Serve(lsn)
	})

	srv.wg.Go(func() error {
		<-srv.ctx.Done()
		srv.logClosing("tls", addr)
		return d.Shutdown(context.Background())
	})
}
