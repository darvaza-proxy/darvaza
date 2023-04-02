// Package server provides an implementation of a TLSProxy.
package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"darvaza.org/darvaza/shared/tls/sni"
)

type emptyStruct struct{}

// ProxyConfig is a configuration for a TLSproxy.
type ProxyConfig struct {
	Protocol   string   `default:"http" hcl:"protocol,label"`
	ListenAddr []string `default:"[\":8080\"]" hcl:"listen"`
}

// Proxy implements a TLSproxy.
type Proxy struct {
	errGroup    *errgroup.Group
	errCtx      context.Context
	cancel      context.CancelFunc
	inShutdown  int32
	mu          sync.Mutex
	listeners   map[*net.Listener]emptyStruct
	activeConns map[*net.Conn]emptyStruct
	tlsHandler  func(net.Conn)
}

func (p *Proxy) shuttingDown() bool {
	return atomic.LoadInt32(&p.inShutdown) != 0
}

func (p *Proxy) trackL(ln *net.Listener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.listeners == nil {
		p.listeners = make(map[*net.Listener]emptyStruct)
	}
	if !p.shuttingDown() {
		if _, found := p.listeners[ln]; !found {
			p.listeners[ln] = emptyStruct{}
		} else {
			delete(p.listeners, ln)
		}
	}
}

func (p *Proxy) trackConn(c *net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.activeConns == nil {
		p.activeConns = make(map[*net.Conn]emptyStruct)
	}

	if !p.shuttingDown() {
		if _, found := p.activeConns[c]; !found {
			p.activeConns[c] = emptyStruct{}
		} else {
			delete(p.activeConns, c)
		}
	}
}

// New returns a pointer to a TLSproxy created from a TLSproxy configuration.
func (pc *ProxyConfig) New() *Proxy {
	var p = new(Proxy)

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.errGroup, p.errCtx = errgroup.WithContext(ctx)

	for _, laddr := range pc.ListenAddr {
		// TODO: do we want UDP/IP and or others?
		l, err := net.Listen("tcp", laddr)
		if err != nil {
			log.Printf("cannot listen on %s.\n %q\n", laddr, err)
			continue
		}
		p.trackL(&l)
	}
	p.tlsHandler = defaultTLSHandler
	return p
}

// TODO: fix revive
//revive:disable:cognitive-complexity

// Run is starting a TLSproxy that accepts connections.
func (p *Proxy) Run() error {
	//revive:enable:cognitive:complexity
	for l := range p.listeners {
		// TODO: Go(func () error{}) means no l tag
		// https://golang.org/doc/faq#closures_and_goroutines
		l := l
		p.errGroup.Go(func() error {
			for {
				if p.shuttingDown() {
					return fmt.Errorf("server shutting down")
				}
				conn, err := (*l).Accept()
				if err != nil {
					select {
					case <-p.errCtx.Done():
						return fmt.Errorf("server shutting down")
					default:
						return err
					}
				}
				p.trackConn(&conn)
				go p.tlsHandler(conn)
			}
		})
	}
	return p.errGroup.Wait()
}

// TODO: reimplemet and fix revive
//
//revive:disable:cognitive-complexity
func (p *Proxy) closeListeners() error {
	//revive:enable:cognitive-complexity
	var err error
	for ln := range p.listeners {
		cerr := (*ln).Close()
		if cerr != nil {
			if cerr.(*net.OpError).Unwrap().Error() != "use of closed network connection" {
				if err == nil {
					err = cerr
				}
			}
		}
		p.trackL(ln)
	}
	return err
}

// Reload wil re-read the configuration of a TLSproxy and apply it.
func (*Proxy) Reload() error {
	// TODO: For now this is a no-op that returns nil.
	return nil
}

// TLSHandler returns the handler function of a TLSproxy.
func (p *Proxy) TLSHandler(fn func(net.Conn)) {
	p.tlsHandler = fn
}

// Cancel is canceling/shutting down a TLSproxy.
func (p *Proxy) Cancel() error {
	defer p.cancel()

	atomic.StoreInt32(&p.inShutdown, 1)

	for c := range p.activeConns {
		err := (*c).Close()
		if err != nil && err.(*net.OpError).Unwrap().Error() != "use of closed network connection" {
			log.Println(err)
		}
		p.trackConn(c)
	}

	err := p.closeListeners()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type prefixConn struct {
	net.Conn
	io.Reader
}

// Read reads data into p.
// It returns the number of bytes read into p.
func (c prefixConn) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
}

func defaultTLSHandler(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, conn, 1+2+2); err != nil {
		log.Println(err)
		return
	}
	length := binary.BigEndian.Uint16(buf.Bytes()[3:5])
	if _, err := io.CopyN(&buf, conn, int64(length)); err != nil {
		log.Println(err)
		return
	}
	sn := sni.GetInfo(buf.Bytes())
	// TODO: Deal with non TLS connections
	if sn != nil && sn.ServerName != "" {
		c := prefixConn{
			Conn:   conn,
			Reader: io.MultiReader(&buf, conn),
		}
		_ = conn.SetReadDeadline(time.Time{})
		defer c.Close()
		var upstream net.Conn
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		// TODO: after we will have backends we can drop the hardcoded 443.
		upstream, err := net.Dial("tcp", fmt.Sprintf("%s:%d", sn.ServerName, 443))
		if err != nil {
			// TODO: Need to retry.
			log.Println(err)
			return
		}
		defer upstream.Close()

		go io.Copy(upstream, io.MultiReader(bytes.NewReader(buf.Bytes()), c))
		_, _ = io.Copy(c, upstream)
	}
}
