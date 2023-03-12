// Package tlsfwd implements a simple TLS forwarder
package tlsfwd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/tls/sni"
	"github.com/darvaza-proxy/middleware"
)

// Server is a simple TLS Frowarder
type Server struct {
	ps        []string
	ds        []*sni.Dispatcher
	res       net.IP
	redir     bool
	wg        core.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled atomic.Bool
	err       atomic.Value
}

// NewTLSFwd constructs and returns a new TLS forwarder from arguments
func NewTLSFwd(ports []int, options ...func(*Server) error) (*Server, error) {
	ctx1, cancel := context.WithCancel(context.Background())

	pts := make([]string, 0)
	for _, p := range ports {
		port := ":" + strconv.Itoa(p)
		pts = append(pts, port)
	}

	srv := Server{
		ps:     pts,
		ctx:    ctx1,
		cancel: cancel,
	}

	for _, opt := range options {
		err := opt(&srv)
		if err != nil {
			return nil, err
		}
	}

	return &srv, nil
}

// SetRedir will allow http -> https redirection from port 80 to 443
func SetRedir(s *Server) error {
	s.redir = true
	return nil
}

func getDefaultTLSHandler(c *tls.ClientHelloInfo) sni.Handler {
	return func(ctx context.Context,
		conn net.Conn,
	) error {
		return tlsHandler(ctx, c, conn)
	}
}

func tlsHandler(ctx context.Context, c *tls.ClientHelloInfo,
	conn net.Conn,
) error {
	defer conn.Close()
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	upstream, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.ServerName, 443))
	if err != nil {
		// TODO: Need to retry?
		return err
	}
	defer upstream.Close()

	var wg core.WaitGroup

	wg.Go(func() error {
		// We just want to CloseWrite to signal that we finished writting
		// but naked net.Conn does not have it so we transform to TCPConn
		if cw, ok := conn.(interface{ CloseWrite() error }); ok {
			defer cw.CloseWrite()
		}
		_, err = io.Copy(conn, upstream)
		return err
	})
	if cwu, ok := upstream.(interface{ CloseWrite() error }); ok {
		defer cwu.CloseWrite()
	}
	_, err = io.Copy(upstream, conn)
	if err != nil {
		return err
	}

	wg.Wait()
	// We made it, no error to return
	return nil
}

// Serve will start serving connections caught by a Server
func (s *Server) Serve() error {
	var ok bool

	defer func() {
		if !ok {
			for _, k := range s.ds {
				_ = k.Close
			}
		}
	}()

	errHTTP := s.prepareRedir()
	if errHTTP != nil {
		return errHTTP
	}

	err := s.prepareDispatchers()
	if err != nil {
		return err
	}
	ok = true
	return s.wg.Wait()
}

func (s *Server) prepareRedir() error {
	if s.redir {
		s.wg.Go(func() error {
			return http.ListenAndServe(":80",
				middleware.NewHTTPSRedirectHandler(0))
		})
	}
	return nil
}

func (s *Server) prepareDispatchers() error {
	for _, p := range s.ps {
		dispatcher := &sni.Dispatcher{
			GetHandler: getDefaultTLSHandler,
		}
		s.ds = append(s.ds, dispatcher)
		ll, err := net.Listen("tcp", p)
		if err != nil {
			return err
		}
		s.wg.Go(func() error {
			return dispatcher.Serve(ll)
		})
	}
	return nil
}

// Err returns the error that caused the cancellation, if any
func (s *Server) Err() error {
	if err, ok := s.err.Load().(error); ok {
		return err
	}
	return nil
}

// Wait waits until all workers are done
func (s *Server) Wait() error {
	s.wg.Wait()
	return s.Err()
}

func (s *Server) tryCancel(err error) {
	// once
	if s.cancelled.CompareAndSwap(false, true) {
		const msg = "Initiating shutdown"

		if err != nil {
			fmt.Printf("%s: %s", msg, err.Error())
		} else {
			fmt.Println(msg)
		}

		s.doCancel(err)
	}
}

func (s *Server) doCancel(err error) {
	// store reason
	if err != nil {
		s.err.Store(err)
	}
	// cancel workers
	if s.cancel != nil {
		s.cancel()
	}
}
