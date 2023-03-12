// Package tlsfwd implements a simple TLS forwarder
package tlsfwd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"sync/atomic"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/proxy"
	"github.com/darvaza-proxy/darvaza/shared/tls/sni"
	"github.com/darvaza-proxy/middleware"
)

// Server is a simple TLS Frowarder
type Server struct {
	ps        []string
	ds        []*sni.Dispatcher
	res       net.IP
	redir     int
	wg        core.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	cancelled atomic.Bool
	err       atomic.Value
}

// NewTLSFwd constructs and returns a new TLS forwarder from arguments
func NewTLSFwd(ports []uint16, options ...func(*Server) error) (*Server, error) {
	ctx1, cancel := context.WithCancel(context.Background())

	pts := make([]string, 0)
	for _, p := range ports {
		port := ":" + strconv.FormatUint(uint64(p), 10)
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

// Redir will allow http -> https redirection from port 80 to 443
func Redir(u int) func(s *Server) error {
	return func(s *Server) error {
		if 0 <= u && u <= 65535 {
			s.redir = u
			return nil
		}
		return fmt.Errorf("invalid unsecure port")
	}
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

	upstream, err := addrFromName(c.ServerName)
	if err != nil {
		return err
	}

	err = proxy.Forward(ctx, conn, upstream)
	return err
}

func addrFromName(s string) (netip.AddrPort, error) {
	upIPs, err := net.LookupIP(s)
	if err != nil {
		fmt.Println(err)
		return netip.AddrPort{}, err
	}

	addrToUse, ok := netip.AddrFromSlice(upIPs[0])
	if !ok {
		return netip.AddrPort{}, err
	}

	upstream := netip.AddrPortFrom(addrToUse, 443)

	return upstream, nil
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

	err := s.prepareRedir()
	if err != nil {
		return err
	}

	err = s.prepareDispatchers()
	if err != nil {
		return err
	}
	ok = true
	return s.wg.Wait()
}

func (s *Server) prepareRedir() error {
	if s.redir > 0 {
		redir := ":" + strconv.Itoa(s.redir)
		s.wg.Go(func() error {
			return http.ListenAndServe(redir,
				middleware.NewHTTPSRedirectHandler(0))
		})
	}
	return nil
}

func (s *Server) prepareDispatchers() error {
	for _, p := range s.ps {
		ll, err := net.Listen("tcp", p)
		if err != nil {
			return err
		}
		dispatcher := &sni.Dispatcher{
			GetHandler: getDefaultTLSHandler,
		}
		s.ds = append(s.ds, dispatcher)
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
