// Package httpserver provides a flexible H1/H2C/H2/H3 server
package httpserver

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// Server is an instance of our H1/H2C/H2/H3 server
type Server struct {
	mu        sync.Mutex
	wg        core.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled atomic.Bool
	err       atomic.Value
	onCancel  func(err error)

	log slog.Logger
	cfg Config

	mux *http.ServeMux
	sl  *ServerListeners

	quicAltSvc string
}

// New creates a new Server from a Config
func (cfg *Config) New() (*Server, error) {
	if err := cfg.SetDefaults(); err != nil {
		return nil, err
	}

	ctx1, cancel := context.WithCancel(cfg.Context)

	srv := &Server{
		ctx:    ctx1,
		cancel: cancel,
		log:    cfg.Logger,
		cfg:    *cfg,
	}

	return srv, nil
}

// OnCancel specifies a function to be called when shutdown
// is initiated, caused by errors or by calling Cancel()/Close()
func (srv *Server) OnCancel(fn func(error)) {
	srv.onCancel = fn
}

// Close tries to initiate a cancellation, and
// returns the reason if it was already cancelled
func (srv *Server) Close() error {
	srv.Cancel()
	return srv.Err()
}

// Cancel initiates a cancellation if it wasn't
// cancelled already
func (srv *Server) Cancel() {
	srv.tryCancel(nil)
}

// Fail initiates a cancellation with the given
// argument as reason
func (srv *Server) Fail(err error) {
	srv.tryCancel(err)
}

// Cancelled tells if the server has been cancelled
func (srv *Server) Cancelled() bool {
	return srv.cancelled.Load()
}

// Err returns the error that caused the cancellation, if any
func (srv *Server) Err() error {
	if err, ok := srv.err.Load().(error); ok {
		return err
	}
	return nil
}

func (srv *Server) tryCancel(err error) {
	// once
	if srv.cancelled.CompareAndSwap(false, true) {
		const msg = "Initiating shutdown"

		if err != nil {
			srv.log.Error().
				WithField(slog.ErrorFieldName, err).
				Printf("%s: %s", msg, err.Error())
		} else {
			srv.log.Info().Println(msg)
		}

		srv.doCancel(err)
	}
}

func (srv *Server) doCancel(err error) {
	// store reason
	if err != nil {
		srv.err.Store(err)
	}

	// cancel workers
	if srv.cancel != nil {
		srv.cancel()
	}

	// notify user
	if srv.onCancel != nil {
		srv.onCancel(err)
	}
}

// Wait waits until all workers are done
func (srv *Server) Wait() error {
	srv.wg.Wait()
	return srv.Err()
}

// ListenAndServe attempts to listen all the necessary port and then
// start the service as Serve does
func (srv *Server) ListenAndServe(h http.Handler) error {
	if err := srv.Listen(); err != nil {
		return err
	}

	return srv.Serve(h)
}

// Serve starts running the service. if a handler wasn't set on Server.Handler,
// you can provide one here. if you do it in both places the underlying
// http.ServeMux will panic
func (srv *Server) Serve(h http.Handler) error {
	var ok bool

	defer func() {
		if !ok {
			_ = srv.sl.Close()
		}
	}()

	if h != nil {
		// this will panic if the user has already set one.
		// pass `nil` in that case
		srv.Handle("/", h)
	}

	tlsListeners := srv.prepareSecureListeners(srv.sl.Secure)
	quicListeners, err := srv.prepareQuicListeners(srv.sl.Quic)
	if err != nil {
		return err
	}

	srv.wg.OnError(srv.onWorkerError)

	// from here onward we don't need to worry about the listeners
	ok = true
	srv.spawnH2(tlsListeners)
	srv.spawnH2C(srv.sl.Insecure)
	srv.spawnH3(quicListeners)
	return srv.wg.Wait()
}

func (srv *Server) onWorkerError(err error) error {
	switch err {
	case http.ErrServerClosed:
		srv.debug().Println("httpserver:", err.Error())
		return nil
	case nil:
		srv.debug().Println("httpserver:", "I hear crickets")
		return nil
	default:
		srv.error(err).Println("httpserver:", "worker failed", err.Error())
		srv.Fail(err)
		return err
	}
}
