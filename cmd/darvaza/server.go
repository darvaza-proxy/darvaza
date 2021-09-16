package main

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

	"github.com/darvaza-proxy/darvaza/tls/sni"
)

type Proxy struct {
	Protocol   string   `default:"http" hcl:"protocol,label"`
	ListenAddr []string `default:"[\":8080\"]" hcl:"listen"`
	pp         *proxy
}

type proxy struct {
	errGroup    *errgroup.Group
	errCtx      context.Context
	cancel      context.CancelFunc
	inShutdown  int32
	mu          sync.Mutex
	listeners   map[*net.Listener]struct{}
	activeConns map[*net.Conn]struct{}
}

func (p *proxy) shuttingDown() bool {
	return atomic.LoadInt32(&p.inShutdown) != 0
}

func (p *proxy) trackL(ln *net.Listener, add bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.listeners == nil {
		p.listeners = make(map[*net.Listener]struct{})
	}
	if add {
		if !p.shuttingDown() {
			p.listeners[ln] = struct{}{}
		}
	} else {
		delete(p.listeners, ln)
	}
}

func (p *proxy) trackConn(c *net.Conn, add bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.activeConns == nil {
		p.activeConns = make(map[*net.Conn]struct{})
	}
	if add {
		p.activeConns[c] = struct{}{}
	} else {
		delete(p.activeConns, c)
	}
}

func (p *Proxy) Run() {
	var pi = new(proxy)
	ctx, cancel := context.WithCancel(context.Background())
	pi.cancel = cancel
	pi.errGroup, pi.errCtx = errgroup.WithContext(ctx)

	for _, laddr := range p.ListenAddr {
		//TODO do we want UDP/IP and or others?
		l, err := net.Listen("tcp", laddr)
		if err != nil {
			log.Printf("cannot listen on %s.\n %q\n", laddr, err)
			continue
		}
		pi.trackL(&l, true)
	}

	p.pp = pi

	for l := range pi.listeners {
		//TODO: Go(func () error{}) means no l tag
		// https://golang.org/doc/faq#closures_and_goroutines
		l := l
		pi.errGroup.Go(func() error {
			for {
				conn, err := (*l).Accept()
				if err != nil {
					select {
					case <-pi.errCtx.Done():
						return fmt.Errorf("server shutting down")
					default:
						return err
					}
				}
				pi.trackConn(&conn, true)
				go handleConnection(conn)
			}
		})
	}

}

func (p *Proxy) Reload() error {
	err := p.Cancel()
	p.Run()
	return err
}

func (p *Proxy) Cancel() error {
	defer p.pp.cancel()
	atomic.StoreInt32(&p.pp.inShutdown, 1)

	for c := range p.pp.activeConns {
		(*c).Close()
		p.pp.trackConn(c, false)
	}

	err := p.pp.closeListeners()
	if err != nil {
		return err
	}

	return nil
}

func (p *proxy) closeListeners() error {
	var err error
	for ln := range p.listeners {
		if cerr := (*ln).Close(); cerr != nil && err == nil {
			err = cerr
		}
		p.trackL(ln, false)
	}
	return err
}

type prefixConn struct {
	net.Conn
	io.Reader
}

func (c prefixConn) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
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
	//TODO Deal with non TLS connections
	if sn != nil && sn.ServerName != "" {
		c := prefixConn{
			Conn:   conn,
			Reader: io.MultiReader(&buf, conn),
		}
		conn.SetReadDeadline(time.Time{})
		defer c.Close()
		var upstream net.Conn
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		//TODO after we will have backends we can drop the hardcoded 443
		upstream, err := net.Dial("tcp", fmt.Sprintf("%s:%d", sn.ServerName, 443))
		if err != nil {
			// TODO: Need to retry
			log.Println(err)
			return
		}
		defer upstream.Close()

		go io.Copy(upstream, io.MultiReader(bytes.NewReader(buf.Bytes()), c))
		io.Copy(c, upstream)
	}
}
